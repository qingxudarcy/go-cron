package worker

import (
	"context"
	"go-cron/internal/common"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LogSink struct {
	client         *mongo.Client
	logCollection  *mongo.Collection
	logChan        chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

var (
	G_logSink *LogSink
)

func (logSink *LogSink) writeMany(logBatch *common.LogBatch) {
	_, _ = logSink.logCollection.InsertMany(context.TODO(), logBatch.Logs)
}

func (logSink *LogSink) writeLoop() {
	var (
		jobLog       *common.JobLog
		logBatch     *common.LogBatch
		commitTimer  *time.Timer
		timeoutBatch *common.LogBatch
	)

	for {
		select {
		case jobLog = <-logSink.logChan:
			if logBatch == nil {
				logBatch = &common.LogBatch{}
				commitTimer = time.AfterFunc(time.Duration(G_config.LogCommitTimeout)*time.Millisecond, func(logBatch *common.LogBatch) func() {
					return func() {
						logSink.autoCommitChan <- logBatch
					}
				}(logBatch))
			}
			logBatch.Logs = append(logBatch.Logs, jobLog)
			if len(logBatch.Logs) >= G_config.LogBatchSize {
				logSink.writeMany(logBatch)
				logBatch = nil
				commitTimer.Stop()
			}
		case timeoutBatch = <-logSink.autoCommitChan:
			if timeoutBatch != logBatch {
				continue
			}
			logSink.writeMany(timeoutBatch)
			logBatch = nil
		}
	}
}

func InitLogSink() (err error) {
	var (
		client *mongo.Client
	)

	if client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(G_config.MongoUri).SetConnectTimeout(5*time.Millisecond)); err != nil {
		return
	}

	G_logSink = &LogSink{
		client:         client,
		logCollection:  client.Database("cron").Collection("log"),
		logChan:        make(chan *common.JobLog, 10000),
		autoCommitChan: make(chan *common.LogBatch, 1000),
	}

	go G_logSink.writeLoop()

	return

}

func (logSink *LogSink) Append(jobLog *common.JobLog) {
	select {
	case logSink.logChan <- jobLog:
	default:
	}
}
