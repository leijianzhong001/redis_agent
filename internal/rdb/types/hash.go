package types

import (
	"github.com/leijianzhong001/redis_agent/internal/log"
	"github.com/leijianzhong001/redis_agent/internal/rdb/structure"
	"github.com/leijianzhong001/redis_agent/internal/utils"
	"io"
)

type HashObject struct {
	key   string
	value map[string]string
}

func (o *HashObject) LoadFromBuffer(rd io.Reader, key string, typeByte byte) {
	o.key = key
	o.value = make(map[string]string)
	switch typeByte {
	case rdbTypeHash:
		o.readHash(rd)
	case rdbTypeHashZipmap:
		o.readHashZipmap(rd)
	case rdbTypeHashZiplist:
		o.readHashZiplist(rd)
	case rdbTypeHashListpack:
		o.readHashListpack(rd)
	default:
		log.Panicf("unknown hash type. typeByte=[%d]", typeByte)
	}
}

func (o *HashObject) readHash(rd io.Reader) {
	size := int(structure.ReadLength(rd))
	for i := 0; i < size; i++ {
		key := structure.ReadString(rd)
		value := structure.ReadString(rd)
		o.value[key] = value
	}
}

func (o *HashObject) readHashZipmap(rd io.Reader) {
	log.Panicf("not implemented rdbTypeZipmap")
}

func (o *HashObject) readHashZiplist(rd io.Reader) {
	list := structure.ReadZipList(rd)
	size := len(list)
	for i := 0; i < size; i += 2 {
		key := list[i]
		value := list[i+1]
		o.value[key] = value
	}
}

func (o *HashObject) readHashListpack(rd io.Reader) {
	list := structure.ReadListpack(rd)
	size := len(list)
	for i := 0; i < size; i += 2 {
		key := list[i]
		value := list[i+1]
		o.value[key] = value
	}
}

func (o *HashObject) Rewrite() []RedisCmd {
	var cmds []RedisCmd
	for k, v := range o.value {
		cmd := RedisCmd{"hset", o.key, k, v}
		cmds = append(cmds, cmd)
	}
	return cmds
}

// MemOverhead 计算当前key加载到redis中以后的内存开销
// 一个`Hash存储结构`最终会产生以下几个消耗内存的结构(相关代码可查阅`t_hash.c`中的`hashTypeLookupWriteOrCreate`函数)：
//		- 1个`dictEntry`结构，24字节，负责保存当前的哈希对象；
//		- 1个`SDS`结构，用作`key`字符串，占`4~18`个字节；
//		- 1个`redisObject`结构，`16`字节，指向当前`key`下属的`dict`结构；
//		- 1个`dict`结构，96字节，负责保存哈希对象的键值对；
//		- n个`dictEntry`结构，`24*n`字节，负责保存具体的`field`和`value`，n等于`field`个数；
//		- n个`SDS`结构，（`field`长度＋`4~18`）*n字节，用作field字符串；
//		- n个`SDS`结构，（`value`长度＋`4~18`）*n字节，用作value字符串；
// 因为hash类型内部有两个`dict`结构，所以最终会有产生两种`rehash`，一种`rehash`基准是`field`个数，另一种`rehash`基准是`key`个数，结合`jemalloc`内存分配规则，`hash`类型的容量评估模型为：
// 		总内存消耗 = [dictEntry大小 + key_SDS大小 + redisObject大小 + dict大小 + (dictEntry大小 + field_SDS大小 + val_SDS大小) * field个数 + field_bucket个数 * 指针大小] * key个数 + key_bucket个数 * 指针大小
func (o *HashObject) MemOverhead() uint64 {
	// dictEntry大小 + key_SDS大小 + redisObject大小 + dict大小
	topLevelObjOverhead := utils.DictEntryOverhead() + utils.SdsOverhead(o.key) + utils.RedisObjOverhead() + utils.DictOverhead()

	// (dictEntry大小 + field_SDS大小 + val_SDS大小) * field个数
	var dataOverhead uint64
	for field, value := range o.value {
		dataOverhead += utils.DictEntryOverhead() + utils.SdsOverhead(field) + utils.SdsOverhead(value)
	}

	// field_bucket个数 * 指针大小
	fieldBucketOverhead := utils.FieldBucketOverhead(uint64(len(o.value)))
	return topLevelObjOverhead + dataOverhead + fieldBucketOverhead
}
