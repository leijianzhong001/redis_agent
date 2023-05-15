package task

import (
	"errors"
	"fmt"
	"github.com/leijianzhong001/redis_agent/internal/utils"
	"github.com/lucasepe/codename"
	log "github.com/sirupsen/logrus"
	"math/rand"
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
	Count     uint64 `json:"count,string"`
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
func (param GenerateUserDataParam) GenerateData() error {
	var err error
	switch param.RedisType {
	case RedisTypeString:
		err = param.generateString()
	case RedisTypeList:
		err = param.generateList()
	case RedisTypeHash:
		err = param.generateHash()
	case RedisTypeSet:
		err = param.generateSet()
	case RedisTypeZSet:
		err = param.generateZSet()
	}
	return err
}

func (param GenerateUserDataParam) generateString() error {
	clusterClient := utils.GetRedisClusterClient()
	rng, _ := codename.DefaultRNG()
	for i := uint64(0); i < param.Count; i++ {
		value := codename.Generate(rng, rand.Intn(70))
		key := param.UserName + ":" + fmt.Sprintf("%d", i)
		_, err := clusterClient.Set(utils.Ctx, key, value, 0).Result()
		if err != nil {
			log.Errorf("generate string data error: %v", err)
			return err
		}

		if i != 0 && i%1000 == 0 {
			log.Infof("user %s 1000 string hash inserted", param.UserName)
		}
	}
	log.Infof("user %s generate string %d data done", param.UserName, param.Count)
	return nil
}

func (param GenerateUserDataParam) generateList() error {
	return nil
}

func (param GenerateUserDataParam) generateHash() error {
	return nil
}

func (param GenerateUserDataParam) generateSet() error {
	return nil
}

func (param GenerateUserDataParam) generateZSet() error {
	return nil
}
