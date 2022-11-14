package main

import (
	"context"
	"github.com/leijianzhong001/redis_agent/internal/cleaner"
	"github.com/leijianzhong001/redis_agent/server"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	format := &log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05", // 据说这个日期是GO语言的诞生时间,格式化时就必须要传这个时间,传其他的时间都会有问题.
	}
	log.SetFormatter(format)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	cleanerX, err := cleaner.NewCleaner()
	if err != nil {
		panic(err)
	}

	srv := server.NewRedisAgentServer(":6389", cleanerX)

	errChan, err := srv.ListenAndServe()
	if err != nil {
		log.Println("redis-agent server start failed:", err)
		return
	}
	log.Println("redis-agent server start ok...")
	log.Println("Submit a Get  request to http://ip:6389/serverStatus to test server is ok")
	log.Println("Submit a Post request to http://ip:6389/cleanTask to start cleaning up the data, param like this {\"taskId\":1234,\"cursor\":0,\"userName\":\"snrs\",\"status\":0,\"startTime\":null,\"lastScanTime\":null,\"keyCount\":0}")
	log.Println("Submit a Get  request to http://ip:6389/cleanTask/{taskId} to get task information")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-errChan:
		log.Println("web server run failed:", err)
		return
	case <-c:
		log.Println("redis-agent program is exiting...")
		ctx, cf := context.WithTimeout(context.Background(), time.Second)
		defer cf()
		err = srv.Shutdown(ctx)
	}

	if err != nil {
		log.Println("redis-agent program exit error:", err)
		return
	}
	log.Println("redis-agent program exit ok")
}
