package memoryclient

import (
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

func TestNewMemoryClient(t *testing.T) {
	c := NewMemoryClient(jobinator.ClientConfig{
		WorkerSleepTime: time.Second / 10,
	})
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
	g.StartAllWorkers()
	time.Sleep(1 * time.Second)
	g.DestroyAllWorkers()
	assert.Equal(t, 1, test.count)
}

func TestPendingJobs(t *testing.T) {
	g.EnqueueJob("willNotExecute", nil, jobinator.JobConfig{
		MaxRetry: 0,
	})
	jobs, err := g.PendingJobs()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(jobs))
}
