package utils

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strings"
	"sync"
)

var redisClient *redis.Client
var lock sync.Mutex
var Ctx = context.Background()

func GetRedisClient() *redis.Client {
	if redisClient != nil {
		return redisClient
	}

	// 为空，则初始化
	lock.Lock()
	defer lock.Unlock()
	if redisClient == nil {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379", //"localhost:6379"
			Username: "default",
			Password: "c4b883c1cba107078b6e0eb6f5677b6a4fcf4046639f2d89a5ec43620efe6e12",
		})
	}
	return redisClient
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
