package jobinator

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/blasphemy/jobinator/status"
)

//Client is the main handle for a jobinator instance. It is where you will perform most actions.
type Client struct {
	InternalClient
	workers     []*BackgroundWorker
	config      ClientConfig
	workerFuncs map[string]WorkerFunc
}

//NewClient will wrap a client implementation and return the resulting client. Meant to be used for implementing storage backends.
func NewClient(ic InternalClient, config ClientConfig) *Client {
	newc := &Client{
		ic,
		[]*BackgroundWorker{},
		config,
		make(map[string]WorkerFunc),
	}
	return newc
}

//RegisterWorker registers a new workerfunc against it's identifier(name)
func (c *Client) RegisterWorker(name string, wf WorkerFunc) {
	c.workerFuncs[name] = wf
	c.InternalRegisterWorker(name, wf)
}

func (c *Client) backgroundExecute() {
	j, err := c.selectJob()
	if err != nil || j == nil {
		return
	}
	ja := &JobRef{
		j: j,
		c: c,
	}
	err = c.executeWorker(j.Name, ja)
	c.SetFinishedAt(j, time.Now().Unix())
	if err != nil {
		errtxt := err.Error()
		errstack := string(debug.Stack())
		c.SetError(j, errtxt, errstack)
		c.IncRetryCount(j)
		if j.RetryCount > j.MaxRetry {
			c.SetStatus(j, status.Failed)
			return
		}
		c.SetStatus(j, status.Retry)
		return
	} else {
		c.SetError(j, "", "")
	}
	if j.Repeat {
		c.SetNextRun(j, time.Now().Add(j.RepeatInterval).Unix())
		c.SetStatus(j, status.Pending)
	} else {
		c.SetStatus(j, status.Done)
	}
	return
}

//ScanArgs scans the job's arguments into your struct of choice.
func (j *JobRef) ScanArgs(v interface{}) error {
	err := json.Unmarshal(j.j.Args, v)
	return err
}

//StopAllWorkers stops all workers registered with the client. This is non blocking, so you may need to wait before they are all finished.
func (c *Client) StopAllWorkers() {
	for _, x := range c.workers {
		x.Stop()
	}
}

//StopAllWorkersBlocking stops all workers registered with the client. This is blocking.
func (c *Client) StopAllWorkersBlocking() {
	for _, x := range c.workers {
		x.StopBlocking()
	}
}

//StartAllWorkers starts all workers registered with the client.
func (c *Client) StartAllWorkers() {
	for _, x := range c.workers {
		x.Start()
	}
}

//DestroyAllWorkers stops all workers registered with the client and destroys them.
func (c *Client) DestroyAllWorkers() {
	for _, x := range c.workers {
		if x.IsRunning() {
			x.StopBlocking()
		}
	}
	c.workers = []*BackgroundWorker{}
}

func (c *Client) executeWorker(name string, ref *JobRef) error {
	_, ok := c.workerFuncs[name]
	if !ok {
		return fmt.Errorf("Worker %s is not available", name)
	}
	err := c.workerFuncs[name](ref)
	return err
}

func (c *Client) selectJob() (*Job, error) {
	return c.InternalSelectJob()
}

//EnqueueJob queues up a job to be run by a worker.
func (c *Client) EnqueueJob(name string, args interface{}, config JobConfig) error {
	ctx, err := json.Marshal(args)
	if err != nil {
		return err
	}
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	j := &Job{
		ID:             id.String(),
		Name:           name,
		Args:           ctx,
		Status:         status.Pending,
		MaxRetry:       config.MaxRetry,
		Repeat:         config.Repeat,
		RepeatInterval: config.RepeatInterval,
		NamedJob:       config.Identifier,
	}
	if config.Repeat {
		j.NextRun = time.Now().Add(j.RepeatInterval).Unix()
	}
	return c.InternalEnqueueJob(j)
}

//PendingJobs returns the number of pending jobs.
func (c *Client) PendingJobs() ([]*Job, error) {
	return c.InternalPendingJobs()
}

//CleanUp deletes all jobs that have been finished (or optionally failed jobs) older than specified in the CleanUpConfig
func (c *Client) CleanUp(config CleanUpConfig) error {
	return c.InternalCleanup(config)
}

func (c *Client) NamedJobInfo(name string) (JobInfo, error) {
	j, err := c.GetNamedJob(name)
	if err != nil {
		return JobInfo{}, err
	}
	fa := time.Unix(j.FinishedAt, 0)
	nr := time.Unix(j.NextRun, 0)
	jinfo := JobInfo{
		Identifier:     j.NamedJob,
		ID:             j.ID,
		LastRun:        fa,
		NextRun:        nr,
		RepeatInterval: j.RepeatInterval,
		Repeat:         j.Repeat,
		Args:           j.Args,
	}
	return jinfo, nil
}
