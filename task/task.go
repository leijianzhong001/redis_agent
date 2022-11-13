package task

import "time"

const (
	TODO     = iota // TODO 待办
	PROGRESS        // PROGRESS 执行中
	SUC             // SUC 任务执行成功
	FAIL            // FAIL 任务执行失败
)

type DataCleanTaskInfo struct {
	// 任务id
	TaskId int `json:"taskId"`
	// 当前游标
	Cursor uint64 `json:"cursor"`
	// 用户名称
	UserName string `json:"userName"`
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
