package worker

import (
	"go-cron/internal/common"
	"math/rand"
	"os/exec"
	"time"
)

var (
	G_excuter *Excuter
)

type Excuter struct {
}

func (excuter *Excuter) ExcuteJob(info *common.JobExcuteInfo) {
	go func() {
		var (
			cmd     *exec.Cmd
			output  []byte
			err     error
			result  *common.JobExcuteResult
			jobLock *JobLock
		)

		result = &common.JobExcuteResult{
			JobExcuteInfo: info,
		}

		jobLock = G_jobMgr.CreateJobLock(info.Job.Name)

		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond) // 随机睡眠不到1s，解决多个worker出现有worker节点饿死的问题
		err = jobLock.TryLock()
		defer jobLock.UnLock()

		result.StartTime = time.Now()

		if err != nil {
			result.Err = err
			result.EndTime = time.Now()
		} else {
			cmd = exec.CommandContext(info.CancelCtx, "/bin/bash", "-c", info.Job.Command)
			output, err = cmd.CombinedOutput()
			result.EndTime = time.Now()
			result.Output = string(output)
			result.Err = err
		}

		G_scheduler.PushJobResult(result)
	}()

}

func InitExcuter() {
	G_excuter = &Excuter{}
}
