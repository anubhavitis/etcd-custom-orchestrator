package etcdClient

import (
	"context"
	cs_alloc "etcd_test/cs-alloc"
	hash_alloc "etcd_test/cs-alloc/alloc_strategy"
	etcd_alloc "etcd_test/cs-alloc/etcd"
	"fmt"
	"time"

	"github.com/google/uuid"
	etcd "go.etcd.io/etcd/client/v3"
)

func Setup() {
	alloc := etcd_alloc.Configure(
		etcd_alloc.WithConfig(etcd.Config{
			Endpoints:   []string{"0.0.0.0:2379"},
			DialTimeout: 2 * time.Second,
		}),
		etcd_alloc.WithMyNodeId("my-node-id-"+uuid.NewString()),
		etcd_alloc.WithBasePath("/etcd-test/test/registry"),
		etcd_alloc.WithAllocStrategy(hash_alloc.HashAllocator),
		etcd_alloc.WithKeyListenerMap(GetKeyActionMap()),
	)

	jobConfigs := GetEtcdJobs()
	SubmitJobs(alloc, jobConfigs)
	alloc.FinalizeJobs()
}

func SubmitJobs(alloc *etcd_alloc.EtcdAlloc, jobConfigs []Job) map[string]bool {
	resp := make(map[string]bool)
	for i := range jobConfigs {
		job := jobConfigs[i]

		success, err := alloc.AppendOrOverwriteJob(job.Name, &cs_alloc.Job{
			JobFunc: func(ctx context.Context) {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						fmt.Println("Processing job: ", job.Name)
						time.Sleep(1 * time.Second)
					}
				}
			},
			ShutDownHook: func(cancel context.CancelFunc) {
				fmt.Println("Cancelling job: ", job.Name)
				cancel()
			},
		})
		if err != nil {
			fmt.Println("Errow while submitting job: ", err)
		}
		resp[job.Name] = success
	}

	return resp
}

func RemoveJobs(etcdAlloc *etcd_alloc.EtcdAlloc, jobs []Job) (resp map[string]bool) {
	resp = make(map[string]bool)
	jobRegistry := etcdAlloc.JobRegistry()
	runningJobs := jobRegistry.RunningJobList()
	for i := range jobs {
		job := jobs[i]
		jobName := job.Name
		resp[jobName] = false
		if runningJobs[jobName] {
			fmt.Println("job-to-stop:" + jobName)
			// Concel the Job
			job := jobRegistry.JobMap[jobName]
			job.ShutDownHook(job.JobContext.Cancel)

			//Set JobStatus as False
			runningJobs[jobName] = false

			// Update RunningJobs list in JobRegistry
			jobRegistry.SetRunningJobList(runningJobs)

			resp[jobName] = true
		}
	}

	return
}
