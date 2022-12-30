package worker

import (
	"context"
	"go-cron/internal/common"
	"net"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Register struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	localIP string
}

var (
	G_register *Register
)

func getLocalIP() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet
		isIpNet bool
	)

	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}

	for _, addr = range addrs {
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}

	err = common.ERR_NO_LOCAL_IP_FOUND

	return

}

func (register *Register) RegisterIP() {
	var (
		leaseGrantRes              *clientv3.LeaseGrantResponse
		leaseId                    clientv3.LeaseID
		cancelCtx                  context.Context
		cancelFunc                 context.CancelFunc
		leaseKeepAliveResponseChan <-chan *clientv3.LeaseKeepAliveResponse
		workerDir                  string
		leaseKeepRes               *clientv3.LeaseKeepAliveResponse
		err                        error
	)

	for {
		cancelFunc = nil
		if leaseGrantRes, err = register.lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}

		leaseId = leaseGrantRes.ID

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())

		if leaseKeepAliveResponseChan, err = register.lease.KeepAlive(cancelCtx, leaseId); err != nil {
			goto RETRY
		}

		workerDir = common.JobWorkerDir + register.localIP

		if _, err = register.kv.Put(cancelCtx, workerDir, "", clientv3.WithLease(leaseId)); err != nil {
			goto RETRY
		}

		for {
			select {
			case leaseKeepRes = <-leaseKeepAliveResponseChan:
				if leaseKeepRes == nil {
					goto RETRY
				}
			}
		}

	RETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}

func InitRegister() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		localIP string
	)

	if localIP, err = getLocalIP(); err != nil {
		return
	}

	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		return
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	G_register = &Register{
		client:  client,
		kv:      kv,
		lease:   lease,
		localIP: localIP,
	}

	go G_register.RegisterIP()

	return
}
