package etcdClient

import (
	"encoding/json"
	"fmt"

	etcdalloc "etcd_test/cs-alloc/etcd"

	"go.etcd.io/etcd/api/v3/mvccpb"
	etcd "go.etcd.io/etcd/client/v3"
)

func GetKeyActionMap() *map[string]func(watchChan *etcd.WatchChan) {

	keyToAction := map[string]func(watchChan *etcd.WatchChan){
		"/etcd-test/test/registry":     registryListener,
		"/etcd-test/test/jobs/updated": updateJobsListener,
	}
	return &keyToAction
}

// Listeners
func updateJobsListener(channel *etcd.WatchChan) {
	for resp := range *channel {
		key := string(resp.Events[0].Kv.Key)
		switch resp.Events[0].Type {

		case mvccpb.DELETE:
			fmt.Println("DELETE : something changed for key: ", key)

		case mvccpb.PUT:
			jobName := string(resp.Events[0].Kv.Value)
			fmt.Println("PUT : something changed for key : ", key, "and value: ", jobName)
			UpdateJobHandler(jobName)
		}
	}

}

func registryListener(channel *etcd.WatchChan) {
	etcdAlloc := etcdalloc.GetInstance()
	for resp := range *channel {
		key := string(resp.Events[0].Kv.Key)

		switch resp.Events[0].Type {

		case mvccpb.DELETE:
			fmt.Println("DELETE : something changed for key: ", key)
			array := etcdAlloc.FetchNodes()
			etcdAlloc.JobRegistry().AllocAllJobs(array, etcdAlloc.AllocStrategy(), etcdAlloc.MyNodeId())

		case mvccpb.PUT:
			fmt.Println("PUT : something changed for key: ", key)
			array := etcdAlloc.FetchNodes()
			etcdAlloc.JobRegistry().AllocAllJobs(array, etcdAlloc.AllocStrategy(), etcdAlloc.MyNodeId())
		}
	}
}

func UpdateJobHandler(jobName string) {
	etcdAlloc := etcdalloc.GetInstance()
	jobRegistry := etcdAlloc.JobRegistry()

	var updatedJobs []string
	json.Unmarshal([]byte(jobName), &updatedJobs)

	var newJobs []Job
	etcdJobMap := GetEtcdJobMap()

	for i := range updatedJobs {
		newJobs = append(newJobs, etcdJobMap[updatedJobs[i]])
	}

	// Add new job to job registry
	jobRegistry.Initialize()
	SubmitJobs(etcdAlloc, newJobs)
	jobRegistry.Finalize()

	// Refresh job in job registry
	jobRegistry.RefreshJobs(
		etcdAlloc.FetchNodes(),
		updatedJobs,
		etcdAlloc.AllocStrategy(),
		etcdAlloc.MyNodeId(),
	)
}
