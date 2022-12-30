package worker

import (
	"fmt"
	"go-cron/internal/common"
	"time"
)

type Scheduler struct {
	jobEventChan   chan *common.JobEvent
	jobPlanTable   map[string]*common.JobSchedulerPlan
	jobExcuteTable map[string]*common.JobExcuteInfo
	jobResultChan  chan *common.JobExcuteResult
}

var (
	G_scheduler *Scheduler
)

func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulerPlan *common.JobSchedulerPlan
		isExisted        bool
		jobExcuteInfo    *common.JobExcuteInfo
		err              error
	)

	switch jobEvent.JobType {
	case common.JobPutEvent:
		if jobSchedulerPlan, err = common.BuildJobSchedulerPlan(jobEvent.Job); err != nil {
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulerPlan

	case common.JobDeleteEvent:
		if _, isExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; isExisted {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	case common.JobKillEvent:
		if jobExcuteInfo, isExisted = scheduler.jobExcuteTable[jobEvent.Job.Name]; isExisted {
			jobExcuteInfo.CancelFunc()
			delete(scheduler.jobExcuteTable, jobEvent.Job.Name)
		}
	}
}

func (scheduler *Scheduler) handleJobResult(jobResult *common.JobExcuteResult) {
	var (
		jobLog *common.JobLog
	)

	delete(scheduler.jobExcuteTable, jobResult.JobExcuteInfo.Job.Name)

	if jobResult.Err != common.ERR_LOCK_ALREADY_REQUIRED {
		jobLog = &common.JobLog{
			JobName:      jobResult.JobExcuteInfo.Job.Name,
			Command:      jobResult.JobExcuteInfo.Job.Command,
			Output:       jobResult.Output,
			PlanTime:     jobResult.JobExcuteInfo.PlanTime.UnixMilli(),
			ScheduleTime: jobResult.JobExcuteInfo.RealTime.UnixMilli(),
			StartTime:    jobResult.StartTime.UnixMilli(),
			EndTime:      jobResult.EndTime.UnixMilli(),
		}
		if jobResult.Err != nil {
			jobLog.Err = jobResult.Err.Error()
		} else {
			jobLog.Err = ""
		}
		G_logSink.Append(jobLog)
	}

	fmt.Println("任务执行完成", jobResult.JobExcuteInfo.Job.Name, jobResult.Output, jobResult.Err)
}

func (scheduler *Scheduler) tryStartJob(jobPlan *common.JobSchedulerPlan) {
	var (
		existed       bool
		jobExcuteInfo *common.JobExcuteInfo
	)

	if _, existed = scheduler.jobExcuteTable[jobPlan.Job.Name]; existed {
		fmt.Printf("%s 任务在执行，跳过\n", jobPlan.Job.Name)
		return
	}

	jobExcuteInfo = common.BuildJobExcuteInfo(jobPlan)
	scheduler.jobExcuteTable[jobPlan.Job.Name] = jobExcuteInfo

	fmt.Println("执行任务", jobExcuteInfo.Job.Name, jobExcuteInfo.PlanTime, jobExcuteInfo.RealTime)
	G_excuter.ExcuteJob(jobExcuteInfo)

}

func (scheduler *Scheduler) tryScheduler() (schedulerAfter time.Duration) {
	var (
		jobPlan  *common.JobSchedulerPlan
		now      time.Time
		nearTime *time.Time
	)

	if len(scheduler.jobPlanTable) == 0 { // 初始化无任务时，隔500ms去轮询
		schedulerAfter = 500 * time.Millisecond
		return
	}

	now = time.Now()

	for _, jobPlan = range scheduler.jobPlanTable {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			scheduler.tryStartJob(jobPlan)
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}

		// 统计最近一个要过期的任务时间
		if nearTime == nil || jobPlan.NextTime.After(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}

	// 下次调度时间 = 最近要执行的任务时间 - 现在的时间
	schedulerAfter = nearTime.Sub(now)

	return
}

// 调度协程
func (scheduler *Scheduler) schedulerLoop() {
	var (
		jobEvent       *common.JobEvent
		schedulerAfter time.Duration
		schedulerTimer *time.Timer
		jobResult      *common.JobExcuteResult
	)

	schedulerAfter = scheduler.tryScheduler()
	schedulerTimer = time.NewTimer(schedulerAfter)

	for {
		select {
		case jobEvent = <-scheduler.jobEventChan:
			scheduler.handleJobEvent(jobEvent)
		case <-schedulerTimer.C: // 最近的任务要执行了
		case jobResult = <-scheduler.jobResultChan:
			scheduler.handleJobResult(jobResult)
		}

		schedulerAfter = scheduler.tryScheduler()
		schedulerTimer.Reset(schedulerAfter)
	}
}

func (scheduler *Scheduler) PushJobevent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

func InitScheduler() {
	G_scheduler = &Scheduler{
		jobEventChan:   make(chan *common.JobEvent, 1000),
		jobPlanTable:   make(map[string]*common.JobSchedulerPlan),
		jobExcuteTable: make(map[string]*common.JobExcuteInfo),
		jobResultChan:  make(chan *common.JobExcuteResult, 1000),
	}

	go G_scheduler.schedulerLoop()
}

func (scheduler *Scheduler) PushJobResult(jobResult *common.JobExcuteResult) {
	scheduler.jobResultChan <- jobResult
}
