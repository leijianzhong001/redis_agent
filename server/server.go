package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/leijianzhong001/redis_agent/internal/cleaner"
	"github.com/leijianzhong001/redis_agent/internal/memanalysis"
	"github.com/leijianzhong001/redis_agent/internal/utils"
	"github.com/leijianzhong001/redis_agent/server/middleware"
	"github.com/leijianzhong001/redis_agent/task"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

type RedisAgentServer struct {
	cleaner    *cleaner.SystemDataCleaner
	httpServer *http.Server
}

func NewRedisAgentServer(addr string, cleaner *cleaner.SystemDataCleaner) *RedisAgentServer {
	agentServer := &RedisAgentServer{
		cleaner: cleaner,
		httpServer: &http.Server{
			Addr: addr,
		},
	}

	router := mux.NewRouter()
	// 创建任务
	router.HandleFunc("/task", agentServer.createTask).Methods("POST")
	// 获取清理任务状态
	router.HandleFunc("/task/{taskId}", agentServer.reportProgress).Methods("GET")

	router.HandleFunc("/serverStatus", agentServer.serverStatus).Methods("GET")

	// 获取全量任务列表
	router.HandleFunc("/tasks", agentServer.getAllTask).Methods("GET")

	// 获取数据分析结果
	router.HandleFunc("/analysisInfo", agentServer.analysisInfo).Methods("GET")

	agentServer.httpServer.Handler = middleware.Logging(middleware.Validating(router))
	return agentServer
}

func (agentServer *RedisAgentServer) ListenAndServe() (<-chan error, error) {
	var err error
	errChan := make(chan error)
	go func() {
		err = agentServer.httpServer.ListenAndServe()
		errChan <- err
	}()

	select {
	case err = <-errChan:
		return nil, err
	case <-time.After(time.Second):
		return errChan, nil
	}
}

func (agentServer *RedisAgentServer) Shutdown(ctx context.Context) error {
	if redisClient := utils.GetRedisClient(); redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Error("shutdown redis client error", err)
		}
	}
	if clusterClient := utils.GetRedisClusterClient(); clusterClient != nil {
		if err := clusterClient.Close(); err != nil {
			log.Error("shutdown redis cluster client error", err)
		}
	}

	if err := agentServer.httpServer.Shutdown(ctx); err != nil {
		log.Error("shutdown http server error", err)
		return err
	}
	return nil
}

// createTask 清理指定用户的数据
func (agentServer *RedisAgentServer) createTask(w http.ResponseWriter, req *http.Request) {
	dec := json.NewDecoder(req.Body)
	var taskInfo task.GenericTaskInfo
	if err := dec.Decode(&taskInfo); err != nil {
		log.Error("decode Request.body error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("recieve create task param %+v", taskInfo)

	// 检查任务类型
	if err := taskInfo.CheckTaskType(); err != nil {
		response(w, FailWithMsg(err.Error()))
		return
	}

	// 参数校验
	if err := agentServer.checkParam(taskInfo); err != nil {
		response(w, FailWithMsg(err.Error()))
		return
	}

	// 创建任务
	if err := taskInfo.CreateTask(); err != nil {
		response(w, FailWithMsg(err.Error()))
		return
	}

	// // 另外启动一个goroutine异步执行, 不要阻塞http请求
	go func() {
		agentServer.startTask(&taskInfo)
	}()

	response(w, SucWithMsg("succeeded in creating a task"))
}

func (agentServer *RedisAgentServer) startTask(taskInfo *task.GenericTaskInfo) {
	// 匿名函数会以闭包的方式访问外围函数的变量 err, 所以后面的逻辑如果导致了err有值，那么defer函数中访问err也会有值
	var err error
	// 注册一个异常处理函数，防止出现异常导致主函数停止
	defer func() {
		if ex := recover(); ex != nil {
			exMsg := fmt.Sprintf("task is panic %v", ex)
			log.Errorf(exMsg)
			err = errors.New(exMsg)
		}

		if err != nil {
			taskInfo.AppendFailLog(err.Error())
			return
		}

		logMsg := fmt.Sprintf("successful completion of task %s", taskInfo.UniqueIdentifier())
		taskInfo.AppendSucLog(logMsg)
	}()

	// 执行任务
	switch taskInfo.TaskType {
	case task.CLEAN:
		err = agentServer.cleaner.ExecuteClean(taskInfo)
	case task.STATISTIC:
		err = memanalysis.ExecuteStatistic()
	case task.GENERATE:
		var generateUserDataParam *task.GenerateUserDataParam
		generateUserDataParam, err = taskInfo.GenerateUserDataParam()
		if err == nil {
			// 生成数据
			err = generateUserDataParam.GenerateData()
		}
	}
}

func (agentServer *RedisAgentServer) checkParam(taskInfo task.GenericTaskInfo) error {
	if taskInfo.TaskType == task.GENERATE {
		generateParam, err := taskInfo.GenerateUserDataParam()
		if err != nil {
			return err
		}

		if err := generateParam.CheckRedisType(); err != nil {
			return err
		}

		if generateParam.Count <= 0 || generateParam.Count > 99999 {
			return err
		}
	}

	if taskInfo.TaskType == task.STATISTIC {
		if task.HasProcessStatisticTask() {
			// 有正在进行中的数据分析任务, 直接返回
			return errors.New("there are already ongoing data analysis tasks in progress, refusing to submit new tasks")
		}
	}
	return nil
}

// reportProgress 上报清理进度
func (agentServer *RedisAgentServer) reportProgress(w http.ResponseWriter, req *http.Request) {
	//err := agentServer.cleaner.Report()
	taskId, ok := mux.Vars(req)["taskId"]
	if !ok {
		http.Error(w, "no taskId found in request", http.StatusBadRequest)
		return
	}

	log.Infof("report task progress %s", taskId)
	taskIdInt, err := strconv.Atoi(taskId)
	if err != nil {
		http.Error(w, "parseInt taskId error: "+taskId, http.StatusBadRequest)
		return
	}

	taskInfo := task.Report(taskIdInt)
	response(w, SucWithData(taskInfo))
}

func (agentServer *RedisAgentServer) serverStatus(w http.ResponseWriter, _ *http.Request) {
	response(w, SucWithMsg("OK"))
}

// getAllTask 获取任务列表
func (agentServer *RedisAgentServer) getAllTask(w http.ResponseWriter, _ *http.Request) {
	taskList := task.GetTaskList()
	response(w, SucWithData(taskList))
}

func (agentServer *RedisAgentServer) analysisInfo(w http.ResponseWriter, _ *http.Request) {
	userAndOverhead := memanalysis.GetUserAndOverhead()
	response(w, SucWithData(userAndOverhead))
}

func response(w http.ResponseWriter, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
