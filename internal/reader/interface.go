package reader

import "github.com/leijianzhong001/redis_agent/internal/entry"

type Reader interface {
	StartRead() chan *entry.Entry
}
