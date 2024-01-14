package allocator

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"job-allocator/src/registory"
	"job-allocator/src/strategy"

	"go.etcd.io/etcd/api/v3/mvccpb"
	etcd "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
)

type EtcdAlloc struct {
	client             *etcd.Client
	basePath           string
	myNodeId           string
	allocStrategy      strategy.AllocStrategy
	jobRegistry        *registory.JobRegistry
	runtimeJobRegistry []string
	keyListenerMap     *map[string]func(watchChan *etcd.WatchChan)
}

func (etcdAlloc *EtcdAlloc) MyNodeId() string {
	return etcdAlloc.myNodeId
}

func (etcdAlloc *EtcdAlloc) AllocStrategy() strategy.AllocStrategy {
	return etcdAlloc.allocStrategy
}

func (etcdAlloc *EtcdAlloc) JobRegistry() *registory.JobRegistry {
	return etcdAlloc.jobRegistry
}

var etcdAllocInstance *EtcdAlloc
var once sync.Once

func GetInstance() *EtcdAlloc {
	return etcdAllocInstance
}

func Configure(opts ...EtcdAllocOption) *EtcdAlloc {
	op := &EtcdAllocOp{}
	op.applyOpts(opts)

	client, _ := etcd.New(op.config)

	once.Do(func() {
		etcdAllocInstance = &EtcdAlloc{
			client:         client,
			basePath:       op.basePath,
			myNodeId:       op.nodeId,
			allocStrategy:  op.allocStrategy,
			jobRegistry:    registory.NewJobRegistry(),
			keyListenerMap: op.keyListenerMap,
		}
	})

	etcdAllocInstance.listenForPeers()
	etcdAllocInstance.setThisNodeAlive()
	return etcdAllocInstance
}

func (etcdAlloc *EtcdAlloc) Client() *etcd.Client {
	return etcdAlloc.client
}

func (etcdAlloc *EtcdAlloc) setThisNodeAlive() {

	lease := etcd.NewLease(etcdAlloc.client)
	resp, _ := lease.Grant(context.TODO(), 5)

	etcdAlloc.client.Put(context.TODO(), etcdAlloc.basePath+"/"+etcdAlloc.myNodeId, etcdAlloc.myNodeId, etcd.WithLease(resp.ID))

	go etcdAlloc.KeepAliveLease(resp.ID, 2*time.Second)
}

func (etcdAlloc *EtcdAlloc) KeepAliveLease(id etcd.LeaseID, duration time.Duration) {

	keepAliveChan, err := etcdAlloc.client.KeepAlive(context.TODO(), id)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case resp := <-keepAliveChan:
			if resp == nil {
				panic("keep-alive channel closed")
			}
			// fmt.Println("keep-alive response received:", resp)
		}
	}
}

func (etcdAlloc *EtcdAlloc) listenForPeers() {
	for key, actionFunc := range *etcdAlloc.keyListenerMap {
		go etcdAlloc.watchKeyAndPerformAction(key, actionFunc)
	}
}

func (etcdAlloc *EtcdAlloc) listener(channel *etcd.WatchChan) {
	for resp := range *channel {
		switch resp.Events[0].Type {

		case mvccpb.DELETE:
			fmt.Println("DELETE : something changed")
			array := etcdAlloc.FetchNodes()
			etcdAlloc.jobRegistry.RefreshAllJobs(array, etcdAlloc.allocStrategy, etcdAlloc.myNodeId)

		case mvccpb.PUT:
			fmt.Println("PUT : something changed")
			array := etcdAlloc.FetchNodes()
			etcdAlloc.jobRegistry.RefreshAllJobs(array, etcdAlloc.allocStrategy, etcdAlloc.myNodeId)
		}
	}

}

func (etcdAlloc *EtcdAlloc) watchKeyAndPerformAction(key string, action func(channel *etcd.WatchChan)) {
	watchChan := etcdAlloc.Client().Watch(context.TODO(), key, etcd.WithPrefix())
	action(&watchChan)
}

func (etcdAlloc *EtcdAlloc) AppendOrOverwriteJob(jobKey string, job *registory.Job) (bool, error) {
	err := etcdAlloc.jobRegistry.AddJob(jobKey, job)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (etcdAlloc *EtcdAlloc) ShutdownJobs() error {
	jobRegistry := etcdAlloc.jobRegistry
	runningJobs := jobRegistry.RunningJobList()
	for name, job := range jobRegistry.JobMap {
		if runningJobs[name] {
			fmt.Println("job-to-stop:" + name)
			job.ShutDownHook(job.JobContext.Cancel)
			runningJobs[name] = false
			jobRegistry.SetRunningJobList(runningJobs)
		}
	}
	return nil
}

func (etcdAlloc *EtcdAlloc) FinalizeJobs() error {

	etcdAlloc.jobRegistry.Finalize()

	array := etcdAlloc.FetchNodes()

	etcdAlloc.jobRegistry.AllocAllJobs(array, etcdAlloc.allocStrategy, etcdAlloc.myNodeId)

	return nil
}

func (etcdAlloc *EtcdAlloc) FetchNodes() []string {
	getResponse, _ := etcdAlloc.client.Get(context.TODO(), etcdAlloc.basePath, etcd.WithPrefix())

	var array = make([]string, len(getResponse.Kvs))
	var i = 0
	for _, v := range getResponse.Kvs {
		fqNodeName := string(v.Key)
		array[i] = strings.Replace(fqNodeName, etcdAlloc.basePath+"/", "", 1)
		i++
	}

	return array
}
