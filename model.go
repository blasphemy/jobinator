package jobinator

import (
	"sync"
	"time"
)

type InternalClient interface {
	InternalEnqueueJob(string, interface{}) error
	InternalSelectJob() (*job, error)
	InternalMarkJobFinished(*job) error
	InternalPendingJobs() (int, error)
	InternalRegisterWorker(string, WorkerFunc)
}

type job struct {
	ID        int64
	Name      string
	Args      []byte
	CreatedAt time.Time
	Status    int
}

type jobRef struct {
	argData []byte
	c       *Client
	j       *job
}

type WorkerFunc func(j *jobRef) error

type clientConfig struct {
	WorkerSleepTime time.Duration
}

type BackgroundWorker struct {
	c        *Client
	quitChan chan bool
	running  bool
	runMutex sync.Mutex
}
