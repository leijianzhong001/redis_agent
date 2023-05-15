package memanalysis

import (
	"context"
	"encoding/json"
	"github.com/leijianzhong001/redis_agent/internal/reader"
	"github.com/leijianzhong001/redis_agent/internal/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

var ctx = context.Background()
var userAndOverhead map[string]*UserOverhead

// UserOverhead 用户和开销数据
type UserOverhead struct {
	// 用户
	UserName string `json:"userName"`
	// key数量
	KeyCount uint64 `json:"keyCount"`
	// 过期key数量
	ExpireKeyCount uint64 `json:"expireKeyCount"`
	// 当前用户的内存开销
	Overhead uint64 `json:"overhead"`
	// 内存分析时间
	AnalysisDate time.Time `json:"analysisDate"`
}

func ExecuteStatistic() error {
	return Statistic()
}

func Statistic() error {
	userAndOverheadTemp := make(map[string]*UserOverhead, 16)
	// 到从节点上 dump rdb 文件
	err := dumpRdb()
	if err != nil {
		log.Errorf("dump rdb error: %v", err)
		return err
	}
	log.Infof("dump rdb success")

	// 从/data下读取dump.rdb文件
	rdbReader := reader.NewRDBReader("/data/dump.rdb")
	// 从这里接收key和value
	ch := rdbReader.StartRead()
	for entry := range ch {
		key := entry.Key
		if len(key) == 0 || !strings.ContainsRune(key, ':') || len(strings.Split(key, ":")) < 2 {
			log.Warnf("key %s is not match *:*, skip", key)
			continue
		}

		// userName
		userName := strings.Split(key, ":")[0]
		if _, ok := userAndOverheadTemp[userName]; !ok {
			// 为空的话, 初始化一下
			userAndOverheadTemp[userName] = &UserOverhead{
				UserName:     userName,
				AnalysisDate: time.Now(),
			}
		}

		userOverhead := userAndOverheadTemp[userName]
		userOverhead.Overhead += entry.Overhead
		userOverhead.KeyCount++
		if entry.IsExpireKey {
			userOverhead.ExpireKeyCount++
		}
	}

	// 计算每个系统的key的rehash的开销
	keyRehashOverhead(userAndOverheadTemp)

	// 计算过期字典rehash开销
	expireKeyRehashOverhead(userAndOverheadTemp)

	log.Infof("memory analysis is done, replace variable userAndOverhead")
	// 完成以后,替换原来的统计结果
	userAndOverhead = userAndOverheadTemp

	for sys, overhead := range userAndOverhead {
		data, _ := json.Marshal(overhead)
		log.Infof("sys: %s, overhead: %s", sys, data)
	}
	return nil
}

// keyRehashOverhead 计算key的rehash开销
func keyRehashOverhead(userAndOverheadTemp map[string]*UserOverhead) {
	// key的总数量
	var allKeyCount uint64
	for _, overhead := range userAndOverheadTemp {
		allKeyCount += overhead.KeyCount
	}

	// 以当前的key为规模,计算rehash的额外开销
	allRehashOverhead := utils.FieldBucketOverhead(allKeyCount)
	for _, overhead := range userAndOverheadTemp {
		// 这里计算当前用户的key数量占总数量的比例, 不需要精确计算, 得到大概的比例就行
		proportion := float64(overhead.KeyCount) / float64(allKeyCount)
		// 将这里的比例应用到rehash的占用上
		currentUserRehashOverhead := float64(allRehashOverhead) * proportion
		// 将当前用户的rehash开销累加到总值上
		overhead.Overhead += uint64(currentUserRehashOverhead)
	}
}

// expireKeyRehashOverhead 计算过期key的rehash开销
func expireKeyRehashOverhead(userAndOverheadTemp map[string]*UserOverhead) {
	// key的总数量
	var allExpireKeyCount uint64
	for _, overhead := range userAndOverheadTemp {
		allExpireKeyCount += overhead.ExpireKeyCount
	}

	// 以当前的key为规模,计算rehash的额外开销
	allRehashOverhead := utils.FieldBucketOverhead(allExpireKeyCount)
	for _, overhead := range userAndOverheadTemp {
		// 这里计算当前用户的key数量占总数量的比例, 不需要精确计算, 得到大概的比例就行
		proportion := float64(overhead.ExpireKeyCount) / float64(allExpireKeyCount)
		// 将这里的比例应用到rehash的占用上
		currentUserRehashOverhead := float64(allRehashOverhead) * proportion
		// 将当前用户的rehash开销累加到总值上
		overhead.Overhead += uint64(currentUserRehashOverhead)
	}
}

// dumpRdb 到从节点上dump rdb文件
func dumpRdb() error {
	client := utils.GetRedisClient()

	// 1、获取角色信息， 非从节点不执行
	infoReplication, err := client.Info(ctx, "Replication").Result()
	if err != nil {
		return errors.New("info Replication command execute fail: " + err.Error())
	}

	role := utils.ParseInfoProp(infoReplication, "role")
	if role == "master" {
		return errors.New("master can not execute statistics, process terminal.")
	}

	// 2、执行bgsave
	_, err = client.BgSave(ctx).Result()
	if err != nil {
		return errors.New("bgsave command execute fail: " + err.Error())
	}

	// 3、等待bgsave完成
	var infoResult string
	for {
		infoResult, err = client.Info(ctx, "Persistence").Result()
		if err != nil {
			return errors.New("info Persistence command execute fail: " + err.Error())
		}

		rdbInfoProgress := utils.ParseInfoProp(infoResult, "rdb_bgsave_in_progress")
		if rdbInfoProgress == "0" {
			// 重新变为0，说明bgsave完成了
			break
		}
		// 1秒轮询一次
		time.Sleep(time.Second * 1)
	}

	lastBgsaveStatus := utils.ParseInfoProp(infoResult, "rdb_last_bgsave_status")
	if lastBgsaveStatus != "ok" {
		return errors.New("the rdb_last_bgsave_status is not ok, process terminal.")
	}
	return nil
}

func GetUserAndOverhead() map[string]*UserOverhead {
	return userAndOverhead
}
