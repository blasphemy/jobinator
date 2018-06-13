package jobinator

import (
	"log"
	"strings"

	"github.com/blasphemy/jobinator/status"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack"
)

var (
	validDrivers = []string{"sqlite3"}
)

func NewClient(dbtype string, dbconn string, config clientConfig) (*client, error) {
	err := driverIsValid(dbtype)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(dbtype, dbconn)
	db.LogMode(true)
	db.AutoMigrate(&job{})
	if err != nil {
		return nil, err
	}
	newc := &client{
		db:          db,
		workerFuncs: make(map[string]WorkerFunc),
		config:      config,
	}
	return newc, nil
}

func driverIsValid(driverName string) error {
	for _, x := range validDrivers {
		if driverName == x {
			return nil
		}
	}
	return errors.Errorf("%s is not a valid driver, valid drivers are: %s", strings.Join(validDrivers, " "))
}

func (c *client) RegisterWorker(name string, wf WorkerFunc) {
	log.Println("Registering worker for " + name)
	c.workerFuncs[name] = wf
}

func (c *client) executeWorker(name string, ref *jobRef) error {
	_, ok := c.workerFuncs[name]
	if !ok {
		return errors.Errorf("Worker %s is not available", name)
	}
	err := c.workerFuncs[name](ref)
	return err
}

func (c *client) EnqueueJob(name string, args interface{}) error {
	contextMsg, err := msgpack.Marshal(args)
	if err != nil {
		return err
	}
	nj := &job{
		Name:   name,
		Args:   contextMsg,
		Status: status.STATUS_ENQUEUED,
	}
	err = c.db.Save(nj).Error
	return err
}

func (c *client) selectJob() (*job, error) {
	wf := []string{}
	for x := range c.workerFuncs {
		wf = append(wf, x)
	}
	tx := c.db.Begin()
	j := &job{}
	statuses := []int{
		status.STATUS_ENQUEUED,
		status.STATUS_RETRY,
	}
	err := tx.First(j, "status in (?) AND name in (?)", statuses, wf).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Model(j).Update("status", status.STATUS_RUNNING).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (c *client) ExecuteOneJob() error {
	log.Println("grabbing a job")
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

func (c *client) markJobFinished(j *job) {
	c.db.Model(j).Update("status", status.STATUS_DONE)
}

func (c *client) pendingJobs() (int, error) {
	var n int
	err := c.db.Model(&job{}).Where("status in (?)", []int{status.STATUS_ENQUEUED, status.STATUS_RETRY}).Count(&n).Error
	if err != nil {
		return 0, err
	}
	return n, nil
}
