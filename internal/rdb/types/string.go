package types

import (
	"github.com/leijianzhong001/redis_agent/internal/rdb/structure"
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
func (o *StringObject) MemOverhead() uint64 {

}
