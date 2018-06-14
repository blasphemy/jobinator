package memoryclient

import (
	"sync"

	"github.com/blasphemy/jobinator"
	"github.com/blasphemy/jobinator/status"
)

type MemoryClient struct {
	jobs    []*jobinator.Job
	joblock sync.Mutex
	wfList  []string
}

func NewMemoryClient(config jobinator.ClientConfig) *jobinator.Client {
	newmc := &MemoryClient{
		jobs:    []*jobinator.Job{},
		joblock: sync.Mutex{},
		wfList:  []string{},
	}
	newc := jobinator.NewClient(newmc, config)
	return newc
}

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

func (m *MemoryClient) InternalSelectJob() (*jobinator.Job, error) {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	for _, x := range m.jobs {
		if m.listContains(x.Name) {
			if x.Status == status.Pending || x.Status == status.Retry {
				x.Status = status.Running
				return x, nil
			}
		}
	}
	return nil, nil
}

func (m *MemoryClient) SetStatus(j *jobinator.Job, status int) error {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	j.Status = status
	return nil
}

func (m *MemoryClient) IncRetryCount(j *jobinator.Job) error {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	j.RetryCount++
	return nil
}

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

func (m *MemoryClient) InternalRegisterWorker(name string, wf jobinator.WorkerFunc) {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	m.wfList = append(m.wfList, name)
}
func (m *MemoryClient) SetError(j *jobinator.Job, errtxt string, stack string) error {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	j.Error = errtxt
	j.ErrorStack = stack
	return nil
}
