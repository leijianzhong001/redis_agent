package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/leijianzhong001/redis_agent/task"
	"testing"
	"time"
)

var ctx = context.Background()

func TestRedisClient(t *testing.T) {
	redisClient := GetRedisClient()
	result, err := redisClient.Info(ctx).Result()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(result)
}

func TestRedisClient2(t *testing.T) {
	taskInfo := task.DataCleanTaskInfo{
		TaskId:       1234,
		Cursor:       0,
		UserName:     "snrs",
		Status:       0,
		StartTime:    time.Now(),
		LastScanTime: time.Now(),
		KeyCount:     0,
	}

	marshal, err := json.Marshal(taskInfo)
	if err != nil {
		return
	}
	fmt.Println(string(marshal))
}