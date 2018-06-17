package jobinator

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

func TestBackgroundWokers(t *testing.T) {
	td["NUM"] = 0
	amount := 5
	ta := testArgs{
		Amount: 1,
	}

	for i := 0; i < amount; i++ {
		c.EnqueueJob("inc", ta, JobConfig{
			MaxRetry: 0,
		})
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

func TestRetry(t *testing.T) {
	td["NUM"] = 0
	wf := func(j *JobRef) error {
		tdLock.Lock()
		defer tdLock.Unlock()
		td["NUM"]++
		return errors.New("should error")
	}
	c.RegisterWorker("retry_test", wf)
	c.EnqueueJob("retry_test", nil, JobConfig{
		MaxRetry: 1,
	})
	c.NewBackgroundWorker()
	c.NewBackgroundWorker()
	c.StartAllWorkers()
	time.Sleep(time.Second)
	c.DestroyAllWorkers()
	//num should be 2, one for the initial try, one for the retry
	assert.Equal(t, 2, td["NUM"])
}

func TestStopBlockingLongRunning(t *testing.T) {
	wf := func(j *JobRef) error {
		time.Sleep(10 * time.Second)
		return nil
	}
	c.RegisterWorker("sleep", wf)
	c.EnqueueJob("sleep", nil, JobConfig{
		MaxRetry: 0,
	})
	c.NewBackgroundWorker()
	c.NewBackgroundWorker()
	c.StartAllWorkers()
	time.Sleep(1 * time.Second)
	c.StopAllWorkersBlocking()
	for _, x := range c.workers {
		assert.False(t, x.IsRunning())
	}
	c.DestroyAllWorkers()
}

func TestStartWorkerTwice(t *testing.T) {
	c.NewBackgroundWorker()
	c.StartAllWorkers()
	c.StartAllWorkers()
	time.Sleep(1)
	c.DestroyAllWorkers()
}
