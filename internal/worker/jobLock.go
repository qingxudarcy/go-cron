package worker

import (
	"context"
	"go-cron/internal/common"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type JobLock struct {
	kv         clientv3.KV
	lease      clientv3.Lease
	jobName    string
	cancelFunc context.CancelFunc // 用于终止自动续租
	leaseId    clientv3.LeaseID
	isLocked   bool
}

func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		kv:      kv,
		lease:   lease,
		jobName: jobName,
	}

	return
}

func (jobLock *JobLock) TryLock() (err error) {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
		leaseId        clientv3.LeaseID
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		txn            clientv3.Txn
		lockKey        string
		txnResp        *clientv3.TxnResponse
	)
	if leaseGrantResp, err = jobLock.lease.Grant(context.TODO(), 5); err != nil {
		return
	}

	cancelCtx, cancelFunc = context.WithCancel(context.TODO())
	leaseId = leaseGrantResp.ID

	keepRespChan, err = jobLock.lease.KeepAlive(cancelCtx, leaseId)
	if err != nil {
		goto FAIL
	}

	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)
		for {
			select {
			case keepResp = <-keepRespChan:
				if keepResp == nil {
					goto END
				}
			}
		}
	END:
	}()

	txn = jobLock.kv.Txn(context.TODO())

	lockKey = common.JobLockDir + jobLock.jobName

	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))

	if txnResp, err = txn.Commit(); err != nil {
		goto FAIL
	}

	if !txnResp.Succeeded {
		err = common.ERR_LOCK_ALREADY_REQUIRED
		goto FAIL
	}

	jobLock.leaseId = leaseId
	jobLock.cancelFunc = cancelFunc
	jobLock.isLocked = true

	return

FAIL:
	cancelFunc()
	_, _ = jobLock.lease.Revoke(context.TODO(), leaseId)
	return

}

func (jobLock *JobLock) UnLock() {
	if jobLock.isLocked {
		jobLock.cancelFunc()
		_, _ = jobLock.lease.Revoke(context.TODO(), jobLock.leaseId)
	}

}
