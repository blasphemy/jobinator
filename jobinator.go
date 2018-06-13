package jobinator

import (
	"github.com/vmihailenco/msgpack"
)

func (c *Client) ExecuteOneJob() error {
	j, err := c.selectJob()
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
	c.markJobFinished(j)
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
