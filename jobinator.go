package jobinator

import (
	"fmt"
	"runtime/debug"

	"github.com/blasphemy/jobinator/status"
	"github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"
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
	}
	c.SetStatus(j, status.Done)
	return
}

//ExecuteOneJob pulls one job from the backend and executes it. This is mostly for testing or if you do not want to use background workers. This is a blocking action. It also has no retry logic, and just returns the error.
func (c *Client) ExecuteOneJob() error {
	j, err := c.selectJob()
	if err != nil {
		return err
	}
	if j == nil {
		return fmt.Errorf("No Jobs available")
	}
	ja := &JobRef{
		j: j,
		c: c,
	}
	err = c.executeWorker(j.Name, ja)
	if err != nil {
		return err
	}
	c.SetStatus(j, status.Done)
	return nil
}

//ScanArgs scans the job's arguments into your struct of choice.
func (j *JobRef) ScanArgs(v interface{}) error {
	err := msgpack.Unmarshal(j.j.Args, v)
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
	ctx, err := msgpack.Marshal(args)
	if err != nil {
		return err
	}
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	j := &Job{
		ID:       id.String(),
		Name:     name,
		Args:     ctx,
		Status:   status.Pending,
		MaxRetry: config.MaxRetry,
	}
	return c.InternalEnqueueJob(j)
}

//PendingJobs returns the number of pending jobs.
func (c *Client) PendingJobs() (int, error) {
	return c.InternalPendingJobs()
}
