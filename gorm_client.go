package jobinator

import (
	"fmt"
	"strings"
	"sync"

	"github.com/blasphemy/jobinator/status"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/vmihailenco/msgpack"
)

type GormClient struct {
	db          *gorm.DB
	dbLock      sync.Mutex
	workerFuncs map[string]WorkerFunc
}

func NewGormClient(dbtype string, dbconn string, config clientConfig) (*Client, error) {
	err := gormDriverIsValid(dbtype)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(dbtype, dbconn)
	db.AutoMigrate(&job{})
	if err != nil {
		return nil, err
	}
	newgc := &GormClient{
		db:          db,
		dbLock:      sync.Mutex{},
		workerFuncs: make(map[string]WorkerFunc),
	}
	newc := &Client{
		newgc,
		[]*BackgroundWorker{},
		config,
	}
	return newc, nil
}

func gormDriverIsValid(driverName string) error {
	vd := []string{"sqlite3"}
	for _, x := range vd {
		if driverName == x {
			return nil
		}
	}
	return fmt.Errorf("%s is not a valid driver, valid drivers are: %s", driverName, strings.Join(vd, " "))
}

func (c *GormClient) RegisterWorker(name string, wf WorkerFunc) {
	c.workerFuncs[name] = wf
}

func (c *GormClient) executeWorker(name string, ref *jobRef) error {
	_, ok := c.workerFuncs[name]
	if !ok {
		return fmt.Errorf("Worker %s is not available", name)
	}
	err := c.workerFuncs[name](ref)
	return err
}

func (c *GormClient) EnqueueJob(name string, args interface{}) error {
	contextMsg, err := msgpack.Marshal(args)
	if err != nil {
		return err
	}
	nj := &job{
		Name:   name,
		Args:   contextMsg,
		Status: status.STATUS_ENQUEUED,
	}
	c.dbLock.Lock()
	err = c.db.Save(nj).Error
	c.dbLock.Unlock()
	return err
}

func (c *GormClient) selectJob() (*job, error) {
	wf := []string{}
	for x := range c.workerFuncs {
		wf = append(wf, x)
	}
	c.dbLock.Lock()
	defer c.dbLock.Unlock()
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

func (c *GormClient) markJobFinished(j *job) {
	c.db.Model(j).Update("status", status.STATUS_DONE)
}

func (c *GormClient) pendingJobs() (int, error) {
	var n int
	c.dbLock.Lock()
	defer c.dbLock.Unlock()
	err := c.db.Model(&job{}).Where("status in (?)", []int{status.STATUS_ENQUEUED, status.STATUS_RETRY}).Count(&n).Error
	if err != nil {
		return 0, err
	}
	return n, nil
}
