package cleaner

import (
	"context"
	"github.com/leijianzhong001/redis_agent/internal/utils"
	"github.com/leijianzhong001/redis_agent/task"
	log "github.com/sirupsen/logrus"
	"time"
)

var ctx = context.Background()

// 批次数量
var batchCount = 2000

// 批次浮动值 每次操作的值大于batchCount-batchFloat就可以执行
var batchFloat = 500

type SystemDataCleaner struct{}

func (cleaner *SystemDataCleaner) ExecuteClean(taskInfo *task.GenericTaskInfo) {
	err := cleaner.Clean(taskInfo)
	if err != nil {
		// 更新任务状态为失败
		taskInfo.Status = task.FAIL
		// 追加日志
		taskInfo.TaskLog = append(taskInfo.TaskLog, task.FormatLog(err.Error()))
		return
	}
	// 更新任务状态为成功
	taskInfo.Status = task.SUC
	taskInfo.TaskLog = append(taskInfo.TaskLog, task.FormatLog("data cleaning task successfully completed"))
}

// Clean system data.
func (cleaner *SystemDataCleaner) Clean(taskInfo *task.GenericTaskInfo) error {
	// 访问该接口的密钥，开始清理时的游标，  keyspace中key的数量
	// 开始游标, 默认为0
	cleanTaskParam, err := taskInfo.CleanTaskParam()
	if err != nil {
		log.Error("get clean task param occurred error", err)
		return err
	}

	cursor := cleanTaskParam.Cursor
	// 要清理数据的用户空间
	userName := cleanTaskParam.UserName
	// 获得redis客户端
	client := utils.GetRedisClient()

	// 16380/20 = 820
	keyGroupBySlot := make(map[int][]string, 820)
	for {
		var keys []string
		var err error
		taskInfo.LastScanTime = time.Now()
		// 这里的2000只是个建议值，并且添加了match参数之后，返回的key数量时不确定的，但可以肯定小于2000
		keys, cursor, err = client.Scan(ctx, cursor, userName+":*", 2000).Result()
		if err != nil {
			log.Error("scan redis occurred error", err)
			return err
		}

		// 按照slot进行分组
		for _, key := range keys {
			// 获得该key的slot
			slot := utils.Slot(key)

			_, ok := keyGroupBySlot[slot]
			if !ok {
				// 如果对应的位置没有切片，则初始化一个，不直接使用append的原因为为了防止扩容
				keyGroupBySlot[slot] = make([]string, 0, batchCount+batchFloat)
			}

			// 添加到对应的keySlot中
			keyGroupBySlot[slot] = append(keyGroupBySlot[slot], key)
			if len(keyGroupBySlot[slot]) >= batchCount-batchFloat {
				// 如果当前slot中的key满足一定的数量，则执行一次unlink
				_, err := client.Unlink(ctx, keyGroupBySlot[slot]...).Result()
				if err != nil {
					log.Error("unlink keys occurred error", err)
					return err
				}

				// 执行unlink之后，清空当前切片内容，但底层数组不变
				keyGroupBySlot[slot] = keyGroupBySlot[slot][0:0]
			}
		}

		// 记录最新的游标
		cleanTaskParam.Cursor = cursor

		// 一旦游标再次为0，则退出遍历
		if cursor == 0 {
			break
		}
	}

	for _, keys := range keyGroupBySlot {
		if len(keys) != 0 {
			// 如果有slot对应的key还没有删除
			_, err := client.Unlink(ctx, keys...).Result()
			if err != nil {
				log.Error("final unlink keys occurred error", err)
				return err
			}
		}
	}

	log.Infof("task %d scan and unlink done", taskInfo.TaskId)
	return nil
}

func NewCleaner() (*SystemDataCleaner, error) {
	cleaner := &SystemDataCleaner{}
	return cleaner, nil
}
