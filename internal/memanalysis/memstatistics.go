package memanalysis

import (
	"context"
	"github.com/leijianzhong001/redis_agent/internal/reader"
	"github.com/leijianzhong001/redis_agent/internal/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

var ctx = context.Background()
var userAndOverhead map[string]uint64

// StatisticTaskParam 数据统计任务独有参数
type StatisticTaskParam struct {
}

func ExecuteStatistic() error {
	return Statistic()
}

func Statistic() error {
	userAndOverheadTemp := make(map[string]uint64, 16)
	// 到从节点上 dump rdb 文件
	err := dumpRdb()
	if err != nil {
		return err
	}

	// 从/data下读取dump.rdb文件
	rdbReader := reader.NewRDBReader("/data/dump.rdb")
	// 从这里接收key和value
	ch := rdbReader.StartRead()
	for entry := range ch {
		key := entry.Key
		if len(key) == 0 || !strings.ContainsRune(key, ':') || len(strings.Split(key, ":")) != 2 {
			log.Warnf("key %s is not match *:*, skip", key)
			continue
		}

		// userName
		userName := strings.Split(key, ":")[0]
		if _, ok := userAndOverheadTemp[userName]; !ok {
			// 为空的话, 初始化一下
			userAndOverheadTemp[userName] = 0
		}

		userAndOverheadTemp[userName] += entry.Overhead
	}

	log.Infof("memory analysis is done, replace variable userAndOverhead")
	// 完成以后,替换原来的统计结果
	userAndOverhead = userAndOverheadTemp
	return nil
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

func GetUserAndOverhead() map[string]uint64 {
	return userAndOverhead
}
