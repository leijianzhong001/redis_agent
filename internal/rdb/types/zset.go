package types

import (
	"fmt"
	"github.com/leijianzhong001/redis_agent/internal/log"
	"github.com/leijianzhong001/redis_agent/internal/rdb/structure"
	"github.com/leijianzhong001/redis_agent/internal/utils"
	"io"
)

type ZSetEntry struct {
	Member string
	Score  string
}

type ZsetObject struct {
	key      string
	elements []ZSetEntry
}

func (o *ZsetObject) LoadFromBuffer(rd io.Reader, key string, typeByte byte) {
	o.key = key
	switch typeByte {
	case rdbTypeZSet:
		o.readZset(rd)
	case rdbTypeZSet2:
		o.readZset2(rd)
	case rdbTypeZSetZiplist:
		o.readZsetZiplist(rd)
	case rdbTypeZSetListpack:
		o.readZsetListpack(rd)
	default:
		log.Panicf("unknown zset type. typeByte=[%d]", typeByte)
	}
}

func (o *ZsetObject) readZset(rd io.Reader) {
	size := int(structure.ReadLength(rd))
	o.elements = make([]ZSetEntry, size)
	for i := 0; i < size; i++ {
		o.elements[i].Member = structure.ReadString(rd)
		score := structure.ReadFloat(rd)
		o.elements[i].Score = fmt.Sprintf("%f", score)
	}
}

func (o *ZsetObject) readZset2(rd io.Reader) {
	size := int(structure.ReadLength(rd))
	o.elements = make([]ZSetEntry, size)
	for i := 0; i < size; i++ {
		o.elements[i].Member = structure.ReadString(rd)
		score := structure.ReadDouble(rd)
		o.elements[i].Score = fmt.Sprintf("%f", score)
	}
}

func (o *ZsetObject) readZsetZiplist(rd io.Reader) {
	list := structure.ReadZipList(rd)
	size := len(list)
	if size%2 != 0 {
		log.Panicf("zset listpack size is not even. size=[%d]", size)
	}
	o.elements = make([]ZSetEntry, size/2)
	for i := 0; i < size; i += 2 {
		o.elements[i/2].Member = list[i]
		o.elements[i/2].Score = list[i+1]
	}
}

func (o *ZsetObject) readZsetListpack(rd io.Reader) {
	list := structure.ReadListpack(rd)
	size := len(list)
	if size%2 != 0 {
		log.Panicf("zset listpack size is not even. size=[%d]", size)
	}
	o.elements = make([]ZSetEntry, size/2)
	for i := 0; i < size; i += 2 {
		o.elements[i/2].Member = list[i]
		o.elements[i/2].Score = list[i+1]
	}
}

func (o *ZsetObject) Rewrite() []RedisCmd {
	cmds := make([]RedisCmd, len(o.elements))
	for inx, ele := range o.elements {
		cmd := RedisCmd{"zadd", o.key, ele.Score, ele.Member}
		cmds[inx] = cmd
	}
	return cmds
}

// MemOverhead 计算当前key加载到redis中以后的内存开销
// 一个`sortedSet存储结构`最终会产生以下几个消耗内存的结构(相关代码可查阅`zaddGenericCommand.c`中的`saddCommand`函数)：
//		- 1个`dictEntry`结构，24字节，负责保存当前的集合对象；
//		- 1个`SDS`结构，用作`key`字符串，占`4~18`个字节；
//		- 1个`redisObject`结构，`16`字节，指向当前`key`下属的`skiplist`结构；
//		- 1个zset结构，`16`字节，作为sortedset的整体数据结构；
//		- 1个dict结构，`96`字节，该结构用与存储元素到其分值的映射
//		- n个dictEntry结构，`24`字节,n为sortedset中的元素个数. dictEntry中存储的key为元素的值，value为分数，这二者和下面的zskiplistNode中的ele和score是同一份数据，所以这里不计算entry中key和valuye的长度，而是在下面的zskiplistNode中计算
//		- n个`elementbucket`结构，8字节，n为sortedset中的元素个数, elementbucketCount 要满足 $elementbucketCount= 2^b, elementbucketCount>= n$
//		- 1个`zskiplist`结构，`32`字节，负责保存集合对象的元素和分值；
//		- n个`zskiplistNode`结构， `4~18 + 16 + 16 * levelSize`，levelSize为`level`数组的大小
// 这里着重说明一下level数组的大小。头节点的level数组长度固定为32，其他节点的level数组大小是随机生成的。level数组的大小其实就是跳表的索引层数，从跳表本身的第一层开始（所有元素都有第一层），所有节点有25%的概率生成第二层索引，所有节点有25% * 25%的概率生成第三层索引，以此类推。。。
// 如果跳表中的元素足够多的话，  那么最终会比较均匀的形成一个类似四叉树的结构，k层索引节点的数量会是k+1层的4倍
// 4^n = C， `n`为索引节点层数编号,`C`为该层元素数量元素数量。 假如元素总量为3000w, 那么这里的n的最大值为`12`，即最大可能有12层索引节点($4^{12}=16777216$)。
// 单key内存开销 = dictEntry + key_sds + value_redisObject + zset + dict + dictEntry * n + 8 * elementbucketCount($elementbucketCount= 2^b, elementbucketCount >= n$) + skiplist + zskiplistNode * n
func (o *ZsetObject) MemOverhead() uint64 {
	// `dictEntry`结构大小 + key_SDS大小 + redisObject大小 + dict大小
	topLevelObjOverhead := utils.DictEntryOverhead() + utils.SdsOverhead(o.key) + utils.RedisObjOverhead() + utils.ZsetOverhead() + utils.DictOverhead() + utils.FieldBucketOverhead(uint64(len(o.elements))) + utils.ZskiplistOverhead()
	var dataOverhead uint64
	for _, element := range o.elements {
		dataOverhead += utils.DictEntryOverhead() + utils.ZskiplistNodeOverhead(element.Member)
	}

	//  todo
	valueBucketOverhead := utils.FieldBucketOverhead(uint64(len(o.elements)))
	return topLevelObjOverhead + dataOverhead + valueBucketOverhead
}
