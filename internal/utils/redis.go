package utils

import (
	"github.com/go-redis/redis/v8"
	"strings"
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

// ParseInfoProp 解析Info原始信息中的指定值
func ParseInfoProp(info string, prop string) string {
	for _, ele := range strings.Split(info, "\r\n") {
		propAndValue := strings.Split(ele, ":")
		if len(propAndValue) < 2 {
			continue
		}
		name := propAndValue[0]
		value := propAndValue[1]
		if prop == name {
			return value
		}
	}
	return ""
}
