package jobinator

import (
	"time"
)

type InternalClient interface {
	InternalEnqueueJob(string, interface{}) error
	InternalSelectJob() (*Job, error)
	InternalMarkJobFinished(*Job) error
	InternalPendingJobs() (int, error)
	InternalRegisterWorker(string, WorkerFunc)
}

type Job struct {
	ID        int64
	Name      string
	Args      []byte
	CreatedAt time.Time
	Status    int
}

type jobRef struct {
	argData []byte
	c       *Client
	j       *Job
}

type WorkerFunc func(j *jobRef) error

type clientConfig struct {
	WorkerSleepTime time.Duration
}
