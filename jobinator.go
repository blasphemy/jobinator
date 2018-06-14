package jobinator

import (
	"fmt"

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

//ExecuteOneJob pulls one job from the backend and executes it. This is mostly for testing or if you do not want to use background workers. This is a blocking action
func (c *Client) ExecuteOneJob() error {
	j, err := c.SelectJob()
	if err != nil {
		return err
	}
	ja := &JobRef{
		argData: j.Args,
		j:       j,
		c:       c,
	}
	err = c.executeWorker(j.Name, ja)
	if err != nil {
		return err
	}
	c.MarkJobFinished(j)
	return nil
}

//ScanArgs scans the job's arguments into your struct of choice.
func (j *JobRef) ScanArgs(v interface{}) error {
	err := msgpack.Unmarshal(j.argData, v)
	return err
}

func (c *Client) StopAllWorkers() {
	for _, x := range c.workers {
		x.Stop()
	}
}

func (c *Client) StopAllWorkersBlocking() {
	for _, x := range c.workers {
		x.StopBlocking()
	}
}

func (c *Client) StartAllWorkers() {
	for _, x := range c.workers {
		x.Start()
	}
}

func (c *Client) DestroyAllWorkers() {
	for _, x := range c.workers {
		if x.IsRunning() {
			x.Stop()
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

func (c *Client) SelectJob() (*Job, error) {
	return c.InternalSelectJob()
}

func (c *Client) MarkJobFinished(j *Job) error {
	return c.InternalMarkJobFinished(j)
}

func (c *Client) EnqueueJob(name string, args interface{}) error {
	ctx, err := msgpack.Marshal(args)
	if err != nil {
		return err
	}
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	j := &Job{
		ID:     id.String(),
		Name:   name,
		Args:   ctx,
		Status: status.STATUS_ENQUEUED,
	}
	return c.InternalEnqueueJob(j)
}

func (c *Client) PendingJobs() (int, error) {
	return c.InternalPendingJobs()
}
