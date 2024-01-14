package registory

import (
	"sync"

	"golang.org/x/net/context"
)

type JobContext struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

type Job struct {
	JobFunc      func(ctx context.Context)
	ShutDownHook func(cancel context.CancelFunc)
	JobContext   JobContext
}

func (jc *JobContext) Refresh() {
	if jc.Ctx == nil || jc.Ctx.Err() != nil {
		jc.Ctx, jc.Cancel = context.WithCancel(context.Background())
	}
}

type Alloc interface {
	AppendOrOverwriteJob(jobKey string, job Job) (bool, error)
	FinalizeJobs() error
}

type JobRegistryStatus string

const (
	JOB_REGISTRY_INIT  JobRegistryStatus = "INIT"
	JOB_REGISTRY_FINAL JobRegistryStatus = "FINAL"
)

func NewJobRegistry() *JobRegistry {
	jr := &JobRegistry{
		JobMap:             make(map[string]*Job),
		mutexForRegistryOp: &sync.Mutex{},
		status:             JOB_REGISTRY_INIT,
		runningJobList:     make(map[string]bool),
	}

	return jr
}
