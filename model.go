package jobinator

import (
	"time"
)

//InternalClient is the interface a storage backend  has to implement. See memoryclient or gormclient for details.
type InternalClient interface {
	InternalEnqueueJob(*Job) error
	InternalSelectJob() (*Job, error)
	InternalPendingJobs() ([]*Job, error)
	InternalRegisterWorker(string, WorkerFunc)
	IncRetryCount(*Job) error
	SetStatus(*Job, int) error
	SetFinishedAt(*Job, int64) error
	SetNextRun(*Job, int64) error
	SetError(*Job, string, string) error
	InternalCleanup(CleanUpConfig) error
}

//Job is the internal representation of a job
type Job struct {
	ID             string
	Name           string `gorm:"index"`
	Args           []byte
	CreatedAt      int64
	Status         int `gorm:"index"`
	RetryCount     int
	MaxRetry       int
	Error          string
	ErrorStack     string
	FinishedAt     int64
	Repeat         bool
	RepeatInterval time.Duration
	NextRun        int64
	NamedJob       string `gorm:"index"` //named jobs should be unique but we don't want to require a name, so I'm not using unique_index
}

//JobConfig includes options for when a job is queued
type JobConfig struct {
	MaxRetry       int
	Repeat         bool
	RepeatInterval time.Duration
	Identifier     string
}

//JobRef is a reference to a job (and it's client). It is passed to a WorkerFunc to get the job args
type JobRef struct {
	c *Client
	j *Job
}

//WorkerFunc is the type of function that must be implemented to be a worker
type WorkerFunc func(j *JobRef) error

//ClientConfig is settings that the client uses during runtime
type ClientConfig struct {
	WorkerSleepTime time.Duration
}

//CleanUpConfig includes options for CleanUp methods
type CleanUpConfig struct {
	MaxAge        time.Duration
	IncludeFailed bool
}
