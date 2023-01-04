package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/leijianzhong001/redis_agent/internal/cleaner"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	TODO     = iota // TODO 待办
	PROGRESS        // PROGRESS 执行中
	SUC             // SUC 任务执行成功
	FAIL            // FAIL 任务执行失败
)

const (
	CLEAN     = iota // CLEAN 数据清理
	STATISTIC        // STATISTIC 内存占用统计
)

var locker sync.RWMutex
var tasks = make(map[int]*GenericTaskInfo, 10)

type GenericTaskInfo struct {
	// 任务id
	TaskId int `json:"taskId"`
	// 任务类型
	TaskType int `json:"taskType"`
	// 任务参数
	TaskParam map[string]string `json:"taskParam"`
	// 任务状态
	Status int `json:"status"`
	// 任务开始时间 默认 0001-01-01 00:00:00 +0000 UTC
	StartTime time.Time `json:"startTime"`
	// 最近一次scan时间
	LastScanTime time.Time `json:"lastScanTime"`
	// 任务开始时key的数量
	KeyCount int `json:"keyCount"`
	// 任务日志
	TaskLog []string `json:"taskLog"`
}

func (taskInfo *GenericTaskInfo) CleanTaskParam() (*cleaner.CleanTaskParam, error) {
	if taskInfo.TaskType != CLEAN {
		return nil, errors.New(fmt.Sprintf("Task type error: %d, you can't call this method", taskInfo.TaskType))
	}

	paramJson, err := json.Marshal(taskInfo.TaskParam)
	if err != nil {
		return nil, err
	}

	var taskParam cleaner.CleanTaskParam
	err = json.Unmarshal(paramJson, &taskParam)
	if err != nil {
		return nil, err
	}
	return &taskParam, nil
}

func (taskInfo *GenericTaskInfo) CheckTaskType() bool {
	if taskInfo.TaskType != CLEAN && taskInfo.TaskType != STATISTIC {
		return false
	}
	return true
}

func (taskInfo *GenericTaskInfo) CreateTask() error {
	locker.Lock()
	defer locker.Unlock()
	_, ok := tasks[taskInfo.TaskId]
	if !ok {
		// 说明该任务第一次执行
		taskInfo.Status = TODO
		taskInfo.StartTime = time.Now()
		taskInfo.LastScanTime = time.Now()
		taskInfo.TaskLog = make([]string, 0, 10)
		tasks[taskInfo.TaskId] = taskInfo
	}

	if taskInfo.Status == SUC {
		// 说明已经成功执行过一次了
		logMsg := fmt.Sprintf("%d clean task is already exec successfully, can't execute again!", taskInfo.TaskId)
		log.Info(logMsg)
		taskInfo.TaskLog = append(taskInfo.TaskLog, FormatLog(logMsg))
		return errors.New(logMsg)
	}

	if taskInfo.Status == PROGRESS {
		// 说明是重复提交
		logMsg := fmt.Sprintf("%d The clean task is in progress and cannot be executed again!", taskInfo.TaskId)
		log.Info(logMsg)
		taskInfo.TaskLog = append(taskInfo.TaskLog, FormatLog(logMsg))
		return errors.New(logMsg)
	}

	// 保存该任务到内存中
	taskInfo.StartTime = time.Now()
	taskInfo.Status = PROGRESS
	taskInfo.TaskLog = append(taskInfo.TaskLog, FormatLog("data cleaning task starts"))
	return nil
}

func FormatLog(log string) string {
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s] %s", timeStr, log)
}

// Report task status to snrs
func Report(taskId int) *GenericTaskInfo {
	return tasks[taskId]
}
