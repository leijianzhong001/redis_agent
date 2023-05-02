package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/leijianzhong001/redis_agent/internal/log"
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
	GENERATE
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

	// 任务参数对象，从TaskParam中反序列化得到
	TaskParamObj interface{}
}

// CleanTaskParam 数据清理任务独有参数
type CleanTaskParam struct {
	// 当前游标
	Cursor uint64 `json:"cursor,string"`
	// 用户名称
	UserName string `json:"userName"`
}

// CleanTaskParam 从map中得到CleanTaskParam参数
func (taskInfo *GenericTaskInfo) CleanTaskParam() (*CleanTaskParam, error) {
	if taskInfo.TaskType != CLEAN {
		return nil, errors.New(fmt.Sprintf("Task type error: %d, you can't call this method CleanTaskParam", taskInfo.TaskType))
	}

	paramJson, err := json.Marshal(taskInfo.TaskParam)
	if err != nil {
		return nil, err
	}

	var taskParam CleanTaskParam
	err = json.Unmarshal(paramJson, &taskParam)
	if err != nil {
		return nil, err
	}
	return &taskParam, nil
}

// GenerateUserDataParam 从map中得到CleanTaskParam参数
func (taskInfo *GenericTaskInfo) GenerateUserDataParam() (*GenerateUserDataParam, error) {
	if taskInfo.TaskType != GENERATE {
		return nil, errors.New(fmt.Sprintf("Task type error: %d, you can't call this method generateUserDataParam", taskInfo.TaskType))
	}

	if taskInfo.TaskParamObj != nil {
		obj := taskInfo.TaskParamObj
		if v, ok := obj.(*GenerateUserDataParam); ok {
			return v, nil
		}
	}

	paramJson, err := json.Marshal(taskInfo.TaskParam)
	if err != nil {
		return nil, err
	}

	var taskParam GenerateUserDataParam
	err = json.Unmarshal(paramJson, &taskParam)
	if err != nil {
		return nil, err
	}

	taskInfo.TaskParamObj = &taskParam
	return &taskParam, nil
}

func (taskInfo *GenericTaskInfo) CheckTaskType() error {
	if taskInfo.TaskType != CLEAN && taskInfo.TaskType != STATISTIC && taskInfo.TaskType != GENERATE {
		return errors.New("task Type must be 0/1/2")
	}
	return nil
}

func (taskInfo *GenericTaskInfo) CreateTask() error {
	locker.Lock()
	defer locker.Unlock()
	_, ok := tasks[taskInfo.TaskId]
	if ok {
		logMsg := fmt.Sprintf("%d task is already exists, can't create again!", taskInfo.TaskId)
		log.Infof(logMsg)
		return errors.New(logMsg)
	}

	// 说明该任务第一次执行 保存该任务到内存中
	taskInfo.Status = PROGRESS
	taskInfo.StartTime = time.Now()
	taskInfo.LastScanTime = time.Now()
	taskInfo.TaskLog = make([]string, 0, 10)
	taskInfo.TaskLog = append(taskInfo.TaskLog, FormatLog("task starts"))
	tasks[taskInfo.TaskId] = taskInfo
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

// AppendFailLog 追加失败日志
func (taskInfo *GenericTaskInfo) AppendFailLog(log string) {
	// 更新任务状态为失败
	taskInfo.Status = FAIL
	// 追加日志
	taskInfo.TaskLog = append(taskInfo.TaskLog, FormatLog(log))
}

// AppendSucLog 追加成功日志
func (taskInfo *GenericTaskInfo) AppendSucLog(log string) {
	// 更新任务状态为失败
	taskInfo.Status = SUC
	// 追加日志
	taskInfo.TaskLog = append(taskInfo.TaskLog, FormatLog(log))
}

func (taskInfo *GenericTaskInfo) UniqueIdentifier() string {
	return fmt.Sprintf("%d-%d", taskInfo.TaskId, taskInfo.TaskType)
}

func GetTaskList() map[int]*GenericTaskInfo {
	var newTaskList map[int]*GenericTaskInfo
	for k, v := range tasks {
		newTaskList[k] = v
	}
	return newTaskList
}

func GetStatisticTaskList() map[int]*GenericTaskInfo {
	statisticTasks := make(map[int]*GenericTaskInfo, 10)
	for key, value := range tasks {
		if value.TaskType != STATISTIC {
			continue
		}
		statisticTasks[key] = value
	}
	return statisticTasks
}

func HasProcessStatisticTask() bool {
	locker.Lock()
	defer locker.Unlock()
	for _, taskInfo := range tasks {
		if taskInfo.TaskType == STATISTIC && taskInfo.Status == PROGRESS {
			return true
		}
	}
	return false
}
