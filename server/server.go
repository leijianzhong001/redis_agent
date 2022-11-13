package server

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/leijianzhong001/redis_agent/internal/cleaner"
	"github.com/leijianzhong001/redis_agent/server/middleware"
	"github.com/leijianzhong001/redis_agent/task"
	"github.com/leijianzhong001/redis_agent/utils"
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
	// 数据清理
	router.HandleFunc("/cleanTask", agentServer.clean).Methods("POST")
	// 获取清理任务状态
	router.HandleFunc("/cleanTask/{taskId}", agentServer.reportProgress).Methods("GET")

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
	if utils.RedisClient != nil {
		if result, err := utils.RedisClient.Shutdown(ctx).Result(); err != nil {
			log.Error("shutdown redis client error", result, err)
		}
	}

	if err := agentServer.httpServer.Shutdown(ctx); err != nil {
		log.Error("shutdown http server error", err)
		return err
	}
	return nil
}

// clean 清理指定用户的数据
func (agentServer *RedisAgentServer) clean(w http.ResponseWriter, req *http.Request) {
	dec := json.NewDecoder(req.Body)
	var taskInfo task.DataCleanTaskInfo
	if err := dec.Decode(&taskInfo); err != nil {
		log.Error("decode Request.body error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("recieve clean task param %+v", taskInfo)

	// 另外启动一个goroutine异步执行
	go func() {
		agentServer.cleaner.ExecuteClean(&taskInfo)
	}()

	response(w, SucWithMsg("提交数据清理任务成功"))
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
		http.Error(w, "parseInt taskId error: " + taskId, http.StatusBadRequest)
		return
	}

	taskInfo := agentServer.cleaner.Report(taskIdInt)
	response(w, SucWithData(taskInfo))
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