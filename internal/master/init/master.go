package init

import (
	"fmt"
	"go-cron/internal/master"

	"runtime"
)

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func InitMaster(confFile string) {
	var (
		err error
	)

	// 初始化配置
	if err = master.InitConfig(confFile); err != nil {
		goto ERR
	}

	// 初始化线程
	initEnv()

	if err = master.InitJobLogMgr(); err != nil {
		goto ERR
	}

	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}

	// 初始化web api
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}

	for {
		select {}
	}

ERR:
	fmt.Println(err)
}
