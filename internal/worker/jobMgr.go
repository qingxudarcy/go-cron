package worker

import (
	"context"
	"go-cron/internal/common"
	"time"

	mvccpb "go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type JobMgr struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

var (
	G_jobMgr *JobMgr
)

func (jobMgr *JobMgr) watchJobs() (err error) {
	var (
		getRes             *clientv3.GetResponse
		keyVal             *mvccpb.KeyValue
		job                *common.Job
		watchStartRevision int64
		watchChan          clientv3.WatchChan
		watchRes           clientv3.WatchResponse
		event              *clientv3.Event
		jobEvent           *common.JobEvent
		jobName            string
	)

	if getRes, err = jobMgr.kv.Get(context.TODO(), common.JobKeyPrefix, clientv3.WithPrefix()); err != nil {
		return
	}

	for _, keyVal = range getRes.Kvs {
		if job, err = common.UnpackJob(keyVal.Value); err != nil {
			continue // TODO 打log
		} else {
			jobEvent = common.InitJobEvent(common.JobPutEvent, job)
			G_scheduler.PushJobevent(jobEvent)
			// 把这个Job交给shcheduler
		}
	}

	go func() { // 从revision开始监听后续时间变化
		watchStartRevision = getRes.Header.Revision
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JobKeyPrefix, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		for watchRes = range watchChan {
			for _, event = range watchRes.Events {
				switch event.Type {
				case clientv3.EventTypePut:
					if job, err = common.UnpackJob(event.Kv.Value); err != nil {
						continue // TODO 打log
					}
					jobEvent = common.InitJobEvent(common.JobPutEvent, job)

				case clientv3.EventTypeDelete:
					jobName = common.ExtractJobName(string(event.Kv.Key))

					jobEvent = common.InitJobEvent(common.JobDeleteEvent, &common.Job{Name: jobName})
				}

				G_scheduler.PushJobevent(jobEvent)
				// 推给scheduler
			}
		}

	}()

	return

}

func (jobMgr *JobMgr) watchKiller() {
	var (
		watchChan  clientv3.WatchChan
		watchResp  clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobName    string
		jobEvent   *common.JobEvent
	)

	go func() {
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JobKillerPrefix, clientv3.WithPrefix())
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT:
					jobName = common.ExtractKillerName(string(watchEvent.Kv.Key))

					jobEvent = common.InitJobEvent(common.JobKillEvent, &common.Job{Name: jobName})
				}

				G_scheduler.PushJobevent(jobEvent)

			}

		}

	}()
}

func InitJobMgr() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
	)

	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		return
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)

	G_jobMgr = &JobMgr{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}

	err = G_jobMgr.watchJobs()

	G_jobMgr.watchKiller()

	return
}

func (jobMgr *JobMgr) CreateJobLock(jobName string) (jobLock *JobLock) {
	jobLock = InitJobLock(jobName, jobMgr.kv, jobMgr.lease)

	return
}
