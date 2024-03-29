package utils

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strings"
	"sync"
)

var redisClient *redis.Client
var redisClusterClient *redis.ClusterClient
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
			Password: "123",
		})
	}
	return redisClient
}

func GetRedisClusterClient() *redis.ClusterClient {
	if redisClusterClient != nil {
		return redisClusterClient
	}

	// 为空，则初始化
	lock.Lock()
	defer lock.Unlock()
	if redisClusterClient == nil {
		redisClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        []string{"localhost:6379"},
			MaxRedirects: 5,
			Username:     "default",
			Password:     "123",
		})
	}
	return redisClusterClient
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
