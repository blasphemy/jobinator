package jobinator

import (
	"time"
)

//InternalClient is the interface a storage backend  has to implement. See memoryclient or gormclient for details.
type InternalClient interface {
	InternalEnqueueJob(*Job) error
	InternalSelectJob() (*Job, error)
	InternalMarkJobFinished(*Job) error
	InternalPendingJobs() (int, error)
	InternalRegisterWorker(string, WorkerFunc)
}

//Job is the internal representation of a job
type Job struct {
	ID        string
	Name      string
	Args      []byte
	CreatedAt time.Time
	Status    int
}

//JobRef is a reference to a job (and it's client). It is passed to a WorkerFunc to get the job args
type JobRef struct {
	argData []byte
	c       *Client
	j       *Job
}

//WorkerFunc is the type of function that must be implemented to be a worker
type WorkerFunc func(j *JobRef) error

//ClientConfig is settings that the client uses during runtime
type ClientConfig struct {
	WorkerSleepTime time.Duration
}
