package etcdClient

import (
	"context"
	"fmt"
	"job-allocator/src/allocator"
	"job-allocator/src/registory"
	"job-allocator/src/strategy"
	"time"

	"github.com/google/uuid"
	etcd "go.etcd.io/etcd/client/v3"
)

func Setup() {
	alloc := allocator.Configure(
		allocator.WithConfig(etcd.Config{
			Endpoints:   []string{"0.0.0.0:2379"},
			DialTimeout: 2 * time.Second,
		}),
		allocator.WithMyNodeId("my-node-id-"+uuid.NewString()),
		allocator.WithBasePath("/etcd-test/test/registry"),
		allocator.WithAllocStrategy(strategy.HashAllocator),
		allocator.WithKeyListenerMap(GetKeyActionMap()),
	)

	jobConfigs := GetEtcdJobs()
	SubmitJobs(alloc, jobConfigs)
	alloc.FinalizeJobs()
}

func SubmitJobs(alloc *allocator.EtcdAlloc, jobConfigs []Job) map[string]bool {
	resp := make(map[string]bool)
	for i := range jobConfigs {
		job := jobConfigs[i]

		success, err := alloc.AppendOrOverwriteJob(job.Name, &registory.Job{
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

func RemoveJobs(etcdAlloc *allocator.EtcdAlloc, jobs []Job) (resp map[string]bool) {
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
