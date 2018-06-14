package jobinator

import (
	"sync"
	"testing"
	"time"

	"github.com/blasphemy/jobinator/status"
	"github.com/stretchr/testify/assert"
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
			if x.Status == status.Pending || x.Status == status.Retry {
				x.Status = status.Running
				return x, nil
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

var c *Client

type testArgs struct {
	Amount int
}

var td = make(map[string]int)
var tdLock = sync.Mutex{}

func TestEmpty(t *testing.T) {
	c = newMockClient(ClientConfig{
		WorkerSleepTime: time.Second / 10,
	})
	assert.NotNil(t, c)
}

func TestRegisterWorker(t *testing.T) {
	wf := func(j *JobRef) error {
		args := &testArgs{}
		err := j.ScanArgs(args)
		if err != nil {
			return err
		}
		tdLock.Lock()
		td["NUM"] += args.Amount
		tdLock.Unlock()
		return nil
	}
	c.RegisterWorker("inc", wf)
}

func TestEnqueueJob(t *testing.T) {
	args := &testArgs{
		Amount: 1,
	}
	err := c.EnqueueJob("inc", args)
	assert.Nil(t, err)
}

func TestExecuteOneJob(t *testing.T) {
	err := c.ExecuteOneJob()
	assert.Nil(t, err)
	assert.Equal(t, 1, td["NUM"])
}

func TestBackgroundWokers(t *testing.T) {
	td["NUM"] = 0
	amount := 5
	ta := testArgs{
		Amount: 1,
	}

	for i := 0; i < amount; i++ {
		c.EnqueueJob("inc", ta)
	}
	for i := 0; i < 4; i++ {
		c.NewBackgroundWorker()
	}
	c.StartAllWorkers()
	time.Sleep(time.Second)
	c.StopAllWorkersBlocking()
	for _, x := range c.workers {
		assert.False(t, x.IsRunning())
	}
	assert.Equal(t, amount, td["NUM"])
}

func TestStopAllWorkers(t *testing.T) {
	c.StartAllWorkers()
	time.Sleep(1 * time.Second)
	c.StopAllWorkers()
	time.Sleep(time.Second)
	for _, x := range c.workers {
		assert.False(t, x.IsRunning())
	}
}

func TestDestroyAllWorkers(t *testing.T) {
	assert.NotZero(t, len(c.workers))
	c.DestroyAllWorkers()
	assert.Zero(t, len(c.workers))
}
