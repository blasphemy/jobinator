package memoryclient

import (
	"sync"
	"time"

	"github.com/blasphemy/jobinator"
	"github.com/blasphemy/jobinator/status"
)

//MemoryClient is the internal client of a memory backed jobinator instance
type MemoryClient struct {
	jobs    []*jobinator.Job
	joblock sync.Mutex
	wfList  []string
}

//NewMemoryClient returns a new jobinator client that stores all jobs in memory
func NewMemoryClient(config jobinator.ClientConfig) *jobinator.Client {
	newmc := &MemoryClient{
		jobs:    []*jobinator.Job{},
		joblock: sync.Mutex{},
		wfList:  []string{},
	}
	newc := jobinator.NewClient(newmc, config)
	return newc
}

//InternalEnqueueJob queues up a job on the in memory store
func (m *MemoryClient) InternalEnqueueJob(j *jobinator.Job) error {
	m.joblock.Lock()
	m.jobs = append(m.jobs, j)
	m.joblock.Unlock()
	return nil
}

func (m *MemoryClient) listContains(name string) bool {
	for _, x := range m.wfList {
		if name == x {
			return true
		}
	}
	return false
}

//InternalSelectJob selects a job and marks it as running.
func (m *MemoryClient) InternalSelectJob() (*jobinator.Job, error) {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	for _, x := range m.jobs {
		if m.listContains(x.Name) {
			if x.Status == status.Retry {
				x.Status = status.Running
				return x, nil
			}
			if x.Status == status.Pending {
				if !x.Repeat {
					x.Status = status.Running
					return x, nil
				}
				nextRun := x.FinishedAt + x.RepeatInterval
				if time.Now().Unix() > nextRun {
					x.Status = status.Running
					return x, nil
				}
			}
		}
	}
	return nil, nil
}

//SetStatus updates the job status.
func (m *MemoryClient) SetStatus(j *jobinator.Job, status int) error {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	j.Status = status
	return nil
}

//IncRetryCount increases the retry count of a job
func (m *MemoryClient) IncRetryCount(j *jobinator.Job) error {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	j.RetryCount++
	return nil
}

//InternalPendingJobs returns the number of pending jobs
func (m *MemoryClient) InternalPendingJobs() (int, error) {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	count := 0
	for _, x := range m.jobs {
		if x.Status == status.Pending || x.Status == status.Retry {
			count++
		}
	}
	return count, nil
}

//InternalRegisterWorker adds the worker to the list of workers internally
func (m *MemoryClient) InternalRegisterWorker(name string, wf jobinator.WorkerFunc) {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	m.wfList = append(m.wfList, name)
}

//SetError sets the job's error status.
func (m *MemoryClient) SetError(j *jobinator.Job, errtxt string, stack string) error {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	j.Error = errtxt
	j.ErrorStack = stack
	return nil
}

//SetFinishedAt marks the time that the job finished at
func (m *MemoryClient) SetFinishedAt(j *jobinator.Job, t int64) error {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	j.FinishedAt = t
	return nil
}
