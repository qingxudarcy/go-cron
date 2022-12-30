package common

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"
)

type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

func UnpackJob(value []byte) (*Job, error) {
	var (
		job *Job
		err error
	)

	job = &Job{}
	err = json.Unmarshal(value, job)
	return job, err
}

func ExtractJobName(key string) string {
	return strings.TrimPrefix(key, JobKeyPrefix)
}

func ExtractKillerName(killerKey string) string {
	return strings.TrimPrefix(killerKey, JobKillerPrefix)
}

func ExtractNodeIP(key string) string {
	return strings.TrimPrefix(key, JobWorkerDir)
}

type JobEvent struct {
	JobType int // PUT DELETE
	Job     *Job
}

func InitJobEvent(jobType int, job *Job) *JobEvent {
	return &JobEvent{
		JobType: jobType,
		Job:     job,
	}
}

// 任务调度计划
type JobSchedulerPlan struct {
	Job      *Job
	Expr     *cronexpr.Expression
	NextTime time.Time
}

func BuildJobSchedulerPlan(job *Job) (jobSchedulerPlan *JobSchedulerPlan, err error) {
	var (
		expr *cronexpr.Expression
	)

	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	jobSchedulerPlan = &JobSchedulerPlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}

	return
}

type JobExcuteInfo struct {
	Job        *Job
	PlanTime   time.Time // 计划执行时间
	RealTime   time.Time // 实际执行时间
	CancelCtx  context.Context
	CancelFunc context.CancelFunc
}

func BuildJobExcuteInfo(jobPlan *JobSchedulerPlan) (jobExcuteInfo *JobExcuteInfo) {
	var (
		cancelCtx  context.Context
		cancelFunc context.CancelFunc
	)

	cancelCtx, cancelFunc = context.WithCancel(context.Background())
	jobExcuteInfo = &JobExcuteInfo{
		Job:        jobPlan.Job,
		PlanTime:   jobPlan.NextTime,
		RealTime:   time.Now(),
		CancelCtx:  cancelCtx,
		CancelFunc: cancelFunc,
	}

	return
}

type JobExcuteResult struct {
	JobExcuteInfo *JobExcuteInfo
	Output        string
	Err           error
	StartTime     time.Time
	EndTime       time.Time
}

type JobLog struct {
	JobName      string `bson:"jobName"`
	Command      string `bson:"command"`
	Err          string `bson:"err"`
	Output       string `bson:"output"`
	PlanTime     int64  `bson:"planTime"`
	ScheduleTime int64  `bson:"scheduleTime"`
	StartTime    int64  `bson:"startTime"`
	EndTime      int64  `bson:"endTime"`
}

type LogBatch struct {
	Logs []interface{}
}

type Response struct {
	ErrNo int         `json:"errNo"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

func SuccessRes(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	resp := &Response{
		ErrNo: 0,
		Msg:   "success",
		Data:  data,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func ErrRes(w http.ResponseWriter, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	resp := &Response{
		ErrNo: -1,
		Msg:   errMsg,
		Data:  nil,
	}

	jsonRes, _ := json.Marshal(resp)
	_, _ = w.Write(jsonRes)
}
