package statistics

import (
	"context"
	"fmt"
	"github.com/leijianzhong001/redis_agent/internal/reader"
	"github.com/leijianzhong001/redis_agent/internal/utils"
	"github.com/leijianzhong001/redis_agent/task"
	"time"
)

var ctx = context.Background()

type Statisticians struct{}

func (statisticians *Statisticians) ExecuteStatistic(taskInfo *task.GenericTaskInfo) {
	err := taskInfo.CreateTask()
	if err != nil {
		return
	}

}

func (statisticians *Statisticians) count(taskInfo *task.GenericTaskInfo) {
	client := utils.GetRedisClient()

	// 1、获取角色信息， 非从节点不执行
	infoReplication, err := client.Info(ctx, "Replication").Result()
	if err != nil {
		taskInfo.AppendFailLog("info Replication command execute fail: " + err.Error())
		return
	}

	role := utils.ParseInfoProp(infoReplication, "role")
	if role == "master" {
		taskInfo.AppendFailLog("master can not execute statistics, process terminal.")
		return
	}

	// 2、执行bgsave
	_, err = client.BgSave(ctx).Result()
	if err != nil {
		taskInfo.AppendFailLog("bgsave command execute fail: " + err.Error())
		return
	}

	// 3、等待bgsave完成
	for {
		infoResult, err := client.Info(ctx, "Persistence").Result()
		if err != nil {
			taskInfo.AppendFailLog("info Persistence command execute fail: " + err.Error())
			return
		}

		rdbInfoProgress := utils.ParseInfoProp(infoResult, "rdb_bgsave_in_progress")
		if rdbInfoProgress == "0" {
			// 重新变为0，说明bgsave完成了
			break
		}
		// 1秒轮询一次
		time.Sleep(time.Second * 1)
	}

	infoResult, err := client.Info(ctx, "Persistence").Result()
	if err != nil {
		taskInfo.AppendFailLog("info Persistence command execute fail: " + err.Error())
		return
	}

	lastBgsaveStatus := utils.ParseInfoProp(infoResult, "rdb_last_bgsave_status")
	if lastBgsaveStatus != "ok" {
		taskInfo.AppendFailLog("the rdb_last_bgsave_status is not ok, process terminal.")
		return
	}

	// 4、从/data下读取dump.rdb文件
	rdbReader := reader.NewRDBReader("/data/dump.rdb")
	// 从这里接收key和value
	ch := rdbReader.StartRead()
	for entry := range ch {
		cmd := entry.Argv[0]
		switch cmd {
		case "set":
			key := entry.Argv[1]
			value := entry.Argv[2]
			fmt.Printf("key: %s, value: %s\n", key, value)
		case "hset":
		case "rpush":
		case "sadd":
		case "zadd":
		}
	}

}
