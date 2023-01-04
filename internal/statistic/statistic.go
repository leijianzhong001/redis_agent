package statistic

import "github.com/leijianzhong001/redis_agent/task"

type Statisticians struct{}

func (statisticians *Statisticians) ExecuteStatistic(taskInfo *task.GenericTaskInfo) {
	err := taskInfo.CreateTask()
	if err != nil {
		return
	}

}
