package init

import (
	"fmt"
	"go-cron/internal/worker"

	"runtime"
)

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func InitWorker(confFile string) {
	var (
		err error
	)

	// 初始化配置
	if err = worker.InitConfig(confFile); err != nil {
		goto ERR
	}

	// 初始化线程
	initEnv()

	if err = worker.InitRegister(); err != nil {
		goto ERR
	}

	// 启动日志协程
	if err = worker.InitLogSink(); err != nil {
		goto ERR
	}

	worker.InitExcuter()

	worker.InitScheduler()

	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}

	for {
		select {}
	}

ERR:
	fmt.Println(err)
}
