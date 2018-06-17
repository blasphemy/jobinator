package gormclient

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/blasphemy/jobinator"
)

var g *jobinator.Client

type testData struct {
	count int
}

var test *testData

func TestNewGormClient(t *testing.T) {
	c, err := NewGormClient("sqlite3", "file::memory:?mode=memory&cache=shared", jobinator.ClientConfig{
		WorkerSleepTime: time.Second / 2,
	})
	assert.Nil(t, err)
	assert.NotNil(t, c)
	g = c
}

func TestRegisterWorker(t *testing.T) {
	type testArgs struct {
		Amount int
	}
	td := &testData{
		count: 0,
	}
	test = td
	wf := func(j *jobinator.JobRef) error {
		args := &testArgs{}
		err := j.ScanArgs(args)
		if err != nil {
			return err
		}
		test.count = +args.Amount
		return nil
	}
	g.RegisterWorker("inc", wf)
}

func TestEnqueJob(t *testing.T) {
	type testArgs struct {
		Amount int
	}
	g.EnqueueJob("inc", &testArgs{
		Amount: 1,
	}, jobinator.JobConfig{
		MaxRetry: 0,
	})
}

func TestExecuteJob(t *testing.T) {
	g.NewBackgroundWorker()
	g.NewBackgroundWorker()
	g.StartAllWorkers()
	time.Sleep(1 * time.Second)
	g.DestroyAllWorkers()
	assert.Equal(t, 1, test.count)
}

func TestPendingJobs(t *testing.T) {
	g.EnqueueJob("willNotExecute", nil, jobinator.JobConfig{
		MaxRetry: 0,
	})
	num, err := g.PendingJobs()
	assert.Nil(t, err)
	assert.Equal(t, 1, num)
}

var td = make(map[string]int)
var tdLock = sync.Mutex{}

func TestRepeatingJob(t *testing.T) {
	td["repeater"] = 0
	wf := func(j *jobinator.JobRef) error {
		tdLock.Lock()
		td["repeater"]++
		tdLock.Unlock()
		return nil
	}
	g.RegisterWorker("repeater", wf)
	g.EnqueueJob("repeater", nil, jobinator.JobConfig{
		Repeat:         true,
		RepeatInterval: time.Second * 1,
	})
	g.NewBackgroundWorker()
	g.StartAllWorkers()
	time.Sleep(time.Second * 3)
	g.DestroyAllWorkers()
	assert.Equal(t, 3, td["repeater"])
}
