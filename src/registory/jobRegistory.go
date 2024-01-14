package registory

import (
	"errors"
	"fmt"
	"sync"
)

type JobRegistry struct {
	JobMap             map[string]*Job
	mutexForRegistryOp *sync.Mutex
	status             JobRegistryStatus
	runningJobList     map[string]bool
}

func (jobRegistry *JobRegistry) Initialize() {
	jobRegistry.mutexForRegistryOp.Lock()
	defer jobRegistry.mutexForRegistryOp.Unlock()

	jobRegistry.status = JOB_REGISTRY_INIT

}

func (jobRegistry *JobRegistry) Finalize() {
	jobRegistry.mutexForRegistryOp.Lock()
	defer jobRegistry.mutexForRegistryOp.Unlock()

	if jobRegistry.status != JOB_REGISTRY_FINAL {
		jobRegistry.status = JOB_REGISTRY_FINAL
	}
}

// ShutDownJob: to stop the job if it's running
func (jobRegistry *JobRegistry) ShutDownJob(jobName string) {
	fmt.Println("job-to-stop:" + jobName)

	job := jobRegistry.JobMap[jobName]
	job.ShutDownHook(job.JobContext.Cancel)
	jobRegistry.runningJobList[jobName] = false
}

// StartJob: to start the job if it's not running.
func (jobRegistry *JobRegistry) StartJob(jobName string) {
	fmt.Println("job-to-start:" + jobName)

	job := jobRegistry.JobMap[jobName]
	job.JobContext.Refresh()
	go job.JobFunc(job.JobContext.Ctx)
	jobRegistry.runningJobList[jobName] = true
}

func (jobRegistry *JobRegistry) GetJobList() []string {
	allJobs := make([]string, len(jobRegistry.JobMap))
	i := 0
	for name := range jobRegistry.JobMap {
		allJobs[i] = name
		i += 1
	}

	return allJobs
}

func (jobRegistry *JobRegistry) SetRunningJobList(runningJobList map[string]bool) {
	jobRegistry.runningJobList = runningJobList
}

func (jobRegistry *JobRegistry) RunningJobList() map[string]bool {
	return jobRegistry.runningJobList
}

func (jobRegistry *JobRegistry) AddJob(jobKey string, job *Job) error {
	jobRegistry.mutexForRegistryOp.Lock()
	defer jobRegistry.mutexForRegistryOp.Unlock()

	if jobRegistry.status == JOB_REGISTRY_FINAL {
		return errors.New("JobRegistry is finalized. Job cant be added")
	}

	if _, ok := jobRegistry.JobMap[jobKey]; ok {
		return errors.New("job already exists")
	}

	jobRegistry.JobMap[jobKey] = job

	return nil
}

func (jobRegistry *JobRegistry) AllocAllJobs(nodes []string, strategy AllocStrategy, myNodeId string) error {
	if jobRegistry.status != JOB_REGISTRY_FINAL {
		return errors.New("jobRegistry is not finalised")
	}

	allJobs := jobRegistry.GetJobList()
	fmt.Println("all jobs:", allJobs)
	return jobRegistry.AllocJobs(nodes, allJobs, strategy, myNodeId)

}

func (jobRegistry *JobRegistry) RefreshAllJobs(nodes []string, strategy AllocStrategy, myNodeId string) error {
	if jobRegistry.status != JOB_REGISTRY_FINAL {
		return errors.New("jobRegistry is not finalised")
	}

	allJobs := jobRegistry.GetJobList()
	return jobRegistry.RefreshJobs(nodes, allJobs, strategy, myNodeId)

}

func (jobRegistry *JobRegistry) AllocJobs(nodes []string, jobNames []string, strategy AllocStrategy, myNodeId string) (err error) {
	jobRegistry.mutexForRegistryOp.Lock()
	defer jobRegistry.mutexForRegistryOp.Unlock()

	if jobRegistry.status != JOB_REGISTRY_FINAL {
		return errors.New("jobRegistry is not finalised")
	}

	for i := range jobNames {
		jobName := jobNames[i]
		res, _ := strategy(jobName, myNodeId, nodes)
		/*
			- res == false
				- job == running -> job.shutDown()
			- res == true
				- job != running -> job.StartJob()
		*/
		if !res && jobRegistry.runningJobList[jobName] {
			jobRegistry.ShutDownJob(jobName)
		}
		if res && !jobRegistry.runningJobList[jobName] {
			jobRegistry.StartJob(jobName)
		}
	}

	return nil
}

func (jobRegistry *JobRegistry) RefreshJobs(nodes []string, jobNames []string, strategy AllocStrategy, myNodeId string) (err error) {
	jobRegistry.mutexForRegistryOp.Lock()
	defer jobRegistry.mutexForRegistryOp.Unlock()

	if jobRegistry.status != JOB_REGISTRY_FINAL {
		return errors.New("jobRegistry is not finalised")
	}
	for i := range jobNames {
		jobName := jobNames[i]
		res, _ := strategy(jobName, myNodeId, nodes)
		/*
			- res == false
				- job == running -> job.shutDown()
			- res == true
				- job == running -> job.shutDown()
				job.StartJob()
		*/
		if jobRegistry.runningJobList[jobName] {
			jobRegistry.ShutDownJob(jobName)
		}
		if res {
			jobRegistry.StartJob(jobName)
		}
	}

	return nil

}
