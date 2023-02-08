package types

import (
	"github.com/leijianzhong001/redis_agent/internal/rdb/structure"
	"github.com/leijianzhong001/redis_agent/internal/utils"
	"io"
)

type StringObject struct {
	value string
	key   string
}

// LoadFromBuffer 从指定的reader中读取一个字符串，并填充到当前对象
func (o *StringObject) LoadFromBuffer(rd io.Reader, key string, _ byte) {
	o.key = key
	o.value = structure.ReadString(rd)
}

func (o *StringObject) Rewrite() []RedisCmd {
	cmd := RedisCmd{}
	cmd = append(cmd, "set", o.key, o.value)
	return []RedisCmd{cmd}
}

// MemOverhead 计算当前key加载到redis中以后的内存开销
// 一个简单的`key-value键值对`最终会产生4个消耗内存的结构，中间free掉的不考虑：
// 		1个`dictEntry`结构，24字节，负责保存具体的键值对；
// 		1个SDS结构，用作key字符串，视字符串长短占`4~18`个字节；
// 		1个`redisObject`结构，16字节，用作val对象（这个`redisObject`对象就是`dictEntry`中的共用体v）；
// 		1个SDS结构，用作val字符串，占`4~18`个字节;
func (o *StringObject) MemOverhead() uint64 {
	return utils.DictEntryOverhead() + utils.SdsOverhead(o.key) + utils.StringValueOverhead(o.value)
}
