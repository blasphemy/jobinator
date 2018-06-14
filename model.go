package jobinator

import (
	"time"
)

type InternalClient interface {
	InternalEnqueueJob(*Job) error
	InternalSelectJob() (*Job, error)
	InternalMarkJobFinished(*Job) error
	InternalPendingJobs() (int, error)
	InternalRegisterWorker(string, WorkerFunc)
}

type Job struct {
	ID        string
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

type ClientConfig struct {
	WorkerSleepTime time.Duration
}
