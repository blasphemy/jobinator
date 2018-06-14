package memoryclient

import (
	"sync"

	"github.com/blasphemy/jobinator"
	"github.com/blasphemy/jobinator/status"
)

/*
type InternalClient interface {
	InternalEnqueueJob(string, interface{}) error
	InternalSelectJob() (*job, error)
	InternalMarkJobFinished(*job) error
	InternalPendingJobs() (int, error)
	InternalRegisterWorker(string, WorkerFunc)
}
*/

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
	m.jobs = append(m.jobs)
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
			if x.Status == status.STATUS_ENQUEUED || x.Status == status.STATUS_RETRY {
				return x, nil
			}
		}
	}
	return nil, nil
}

func (m *MemoryClient) InternalMarkJobFinished(j *jobinator.Job) error {
	return nil
}

func (m *MemoryClient) InternalPendingJobs() (int, error) {
	return 0, nil
}

func (m *MemoryClient) InternalRegisterWorker(string, jobinator.WorkerFunc) {

}
