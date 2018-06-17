package jobinator

import (
	"sync"
	"time"

	"github.com/blasphemy/jobinator/status"
)

type MockClient struct {
	joblock sync.Mutex
	jobs    []*Job
	wfList  []string
}

func newMockClient(c ClientConfig) *Client {
	mc := &MockClient{
		joblock: sync.Mutex{},
		jobs:    []*Job{},
		wfList:  []string{},
	}
	return NewClient(mc, c)
}

func (m *MockClient) InternalEnqueueJob(j *Job) error {
	m.joblock.Lock()
	m.jobs = append(m.jobs, j)
	m.joblock.Unlock()
	return nil
}

func (m *MockClient) InternalMarkJobFinished(j *Job) error {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	j.Status = status.Done
	return nil
}

func (m *MockClient) InternalPendingJobs() (int, error) {
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

func (m *MockClient) InternalRegisterWorker(name string, wf WorkerFunc) {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	m.wfList = append(m.wfList, name)
}

func (m *MockClient) InternalSelectJob() (*Job, error) {
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
				if time.Now().Unix() >= x.NextRun {
					x.Status = status.Running
					return x, nil
				}
			}
		}
	}
	return nil, nil
}

func (m *MockClient) listContains(name string) bool {
	for _, x := range m.wfList {
		if name == x {
			return true
		}
	}
	return false
}

func (m *MockClient) IncRetryCount(j *Job) error {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	j.RetryCount++
	return nil
}

func (m *MockClient) SetStatus(j *Job, status int) error {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	j.Status = status
	return nil
}

func (m *MockClient) SetError(j *Job, errtxt string, stack string) error {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	j.Error = errtxt
	j.ErrorStack = stack
	return nil
}

func (m *MockClient) SetFinishedAt(j *Job, t int64) error {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	j.FinishedAt = t
	return nil
}

func (m *MockClient) SetNextRun(j *Job, t int64) error {
	m.joblock.Lock()
	defer m.joblock.Unlock()
	j.NextRun = t
	return nil
}
