package utils

import (
	"github.com/go-redis/redis/v8"
	"sync"
)

var RedisClient *redis.Client
var lock sync.Mutex

func GetRedisClient() *redis.Client {
	if RedisClient != nil {
		return RedisClient
	}

	// 为空，则初始化
	lock.Lock()
	defer lock.Unlock()
	if RedisClient == nil {
		RedisClient = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379", //"localhost:6379"
			Username: "default",
			Password: "c4b883c1cba107078b6e0eb6f5677b6a4fcf4046639f2d89a5ec43620efe6e12",
		})
	}
	return RedisClient
}
