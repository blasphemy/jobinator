package jobinator

import (
	"fmt"

	"github.com/vmihailenco/msgpack"
)

type Client struct {
	InternalClient
	workers     []*BackgroundWorker
	config      clientConfig
	workerFuncs map[string]WorkerFunc
}

func (c *Client) RegisterWorker(name string, wf WorkerFunc) {
	c.workerFuncs[name] = wf
	c.InternalRegisterWorker(name, wf)
}

func (c *Client) ExecuteOneJob() error {
	j, err := c.SelectJob()
	if err != nil {
		return err
	}
	ja := &jobRef{
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

func (j *jobRef) ScanArgs(v interface{}) error {
	err := msgpack.Unmarshal(j.argData, v)
	return err
}

func (c *Client) stopAllWorkers() {
	for _, x := range c.workers {
		x.Stop()
	}
}

func (c *Client) stopAllWorkersBlocking() {
	for _, x := range c.workers {
		x.StopBlocking()
	}
}

func (c *Client) startAllWorkers() {
	for _, x := range c.workers {
		x.Start()
	}
}

func (c *Client) destroyAllWorkers() {
	for _, x := range c.workers {
		if x.IsRunning() {
			x.Stop()
		}
	}
	c.workers = []*BackgroundWorker{}
}

func (c *Client) executeWorker(name string, ref *jobRef) error {
	_, ok := c.workerFuncs[name]
	if !ok {
		return fmt.Errorf("Worker %s is not available", name)
	}
	err := c.workerFuncs[name](ref)
	return err
}

func (c *Client) SelectJob() (*job, error) {
	return c.InternalSelectJob()
}

func (c *Client) MarkJobFinished(j *job) error {
	return c.InternalMarkJobFinished(j)
}

func (c *Client) EnqueueJob(name string, args interface{}) error {
	return c.InternalEnqueueJob(name, args)
}

func (c *Client) PendingJobs() (int, error) {
	return c.InternalPendingJobs()
}
