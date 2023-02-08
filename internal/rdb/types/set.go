package types

import (
	"github.com/leijianzhong001/redis_agent/internal/log"
	"github.com/leijianzhong001/redis_agent/internal/rdb/structure"
	"github.com/leijianzhong001/redis_agent/internal/utils"
	"io"
)

type SetObject struct {
	key      string
	elements []string
}

func (o *SetObject) LoadFromBuffer(rd io.Reader, key string, typeByte byte) {
	o.key = key
	switch typeByte {
	case rdbTypeSet:
		o.readSet(rd)
	case rdbTypeSetIntset:
		o.elements = structure.ReadIntset(rd)
	default:
		log.Panicf("unknown set type. typeByte=[%d]", typeByte)
	}
}

func (o *SetObject) readSet(rd io.Reader) {
	size := int(structure.ReadLength(rd))
	o.elements = make([]string, size)
	for i := 0; i < size; i++ {
		val := structure.ReadString(rd)
		o.elements[i] = val
	}
}

func (o *SetObject) Rewrite() []RedisCmd {
	cmds := make([]RedisCmd, len(o.elements))
	for inx, ele := range o.elements {
		cmd := RedisCmd{"sadd", o.key, ele}
		cmds[inx] = cmd
	}
	return cmds
}

// MemOverhead 计算当前key加载到redis中以后的内存开销
// 一个`set存储结构`最终会产生以下几个消耗内存的结构(相关代码可查阅`t_set.c`中的`saddCommand`函数)：
//		- 1个`dictEntry`结构，24字节，负责保存当前的集合对象；
//		- 1个`SDS`结构，用作`key`字符串，占`4~18`个字节；
//		- 1个`redisObject`结构，`16`字节，指向当前`key`下属的`dict`结构；
//		- 1个`dict`结构，`96`字节，负责保存集合对象的元素；
//		- n个`dictEntry`结构，`24*n`字节，负责保存具体value。 其中value保存在dictEntry对象的key成员中，v实际上始终为空。n等于`field`个数；
//		- n个`SDS`结构，（`value`长度＋`4~18`）*n字节，用作`value`字符串；
// set类型中的`value`也直接就都是sds了。具体可以查阅`t_set.c`中的`setTypeAdd`函数
//
// 单key内存开销 = dictEntry大小 + key_SDS大小 + redisObject大小 + dict大小 + (dictEntry大小 + val_SDS大小) * value个数 + value_bucket个数 * 指针大小
func (o *SetObject) MemOverhead() uint64 {
	// `dictEntry`结构大小 + key_SDS大小 + redisObject大小 + dict大小
	topLevelObjOverhead := utils.DictEntryOverhead() + utils.SdsOverhead(o.key) + utils.RedisObjOverhead() + utils.DictOverhead()
	var dataOverhead uint64
	for _, element := range o.elements {
		dataOverhead += utils.DictEntryOverhead() + utils.SdsOverhead(element)
	}
	valueBucketOverhead := utils.FieldBucketOverhead(uint64(len(o.elements)))
	return topLevelObjOverhead + dataOverhead + valueBucketOverhead
}
