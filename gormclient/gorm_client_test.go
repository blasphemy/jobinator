package gormclient

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

func TestNewGormClient(t *testing.T) {
	c, err := NewGormClient("sqlite3", "file::memory:?mode=memory&cache=shared", jobinator.ClientConfig{
		WorkerSleepTime: time.Second / 10,
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
	})
}

func TestExecuteJob(t *testing.T) {
	err := g.ExecuteOneJob()
	assert.Nil(t, err)
	assert.Equal(t, 1, test.count)
}

func TestPendingJobs(t *testing.T) {
	g.EnqueueJob("willNotExecute", nil)
	num, err := g.PendingJobs()
	assert.Nil(t, err)
	assert.Equal(t, 1, num)
}