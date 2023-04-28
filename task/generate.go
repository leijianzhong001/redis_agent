package task

import (
	"errors"
	"fmt"
	"github.com/leijianzhong001/redis_agent/internal/utils"
	"github.com/lucasepe/codename"
)

const (
	RedisTypeString = "string"
	RedisTypeList   = "list"
	RedisTypeHash   = "hash"
	RedisTypeSet    = "set"
	RedisTypeZSet   = "zset"
)

type GenerateUserDataParam struct {
	UserName  string `json:"userName"`
	RedisType string `json:"redisType"`
	Count     uint64 `json:"count"`
}

// CheckRedisType 检查redis数据类型是否合法
func (param GenerateUserDataParam) CheckRedisType() error {
	if param.RedisType == RedisTypeString ||
		param.RedisType == RedisTypeList ||
		param.RedisType == RedisTypeHash ||
		param.RedisType == RedisTypeSet ||
		param.RedisType == RedisTypeZSet {
		return nil
	}
	return errors.New("redisType illegality")
}

// GenerateData 生成redis数据
func (param GenerateUserDataParam) GenerateData() {
	switch param.RedisType {
	case RedisTypeString:
		param.generateString()
	case RedisTypeList:
		param.generateList()
	case RedisTypeHash:
		param.generateHash()
	case RedisTypeSet:
		param.generateSet()
	case RedisTypeZSet:
		param.generateZSet()
	}
}

func (param GenerateUserDataParam) generateString() {
	clusterClient := utils.GetRedisClient()
	rng, _ := codename.DefaultRNG()
	for i := uint64(0); i < param.Count; i++ {
		value := codename.Generate(rng, 50)
		key := param.UserName + ":" + fmt.Sprintf("%d", i)
		clusterClient.Set(utils.Ctx, key, value, 0)
	}
}

func (param GenerateUserDataParam) generateList() {

}

func (param GenerateUserDataParam) generateHash() {

}

func (param GenerateUserDataParam) generateSet() {

}

func (param GenerateUserDataParam) generateZSet() {

}
