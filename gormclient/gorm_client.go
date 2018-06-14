package gormclient

import (
	"fmt"
	"strings"
	"sync"

	"github.com/blasphemy/jobinator"
	"github.com/blasphemy/jobinator/status"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" //needed for sqlite support
)

//GormClient represents a client using a Gorm backend (SQL)
type GormClient struct {
	db     *gorm.DB
	dbLock sync.Mutex
	wfList []string
}

//NewGormClient returns a new *Client backed by gorm. It requires a driver type and connection string, as well as a ClientConfig.
func NewGormClient(dbtype string, dbconn string, config jobinator.ClientConfig) (*jobinator.Client, error) {
	err := gormDriverIsValid(dbtype)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(dbtype, dbconn)
	db.AutoMigrate(&jobinator.Job{})
	if err != nil {
		return nil, err
	}
	newgc := &GormClient{
		db:     db,
		dbLock: sync.Mutex{},
		wfList: []string{},
	}
	newc := jobinator.NewClient(newgc, config)
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

//InternalRegisterWorker adds a worker to the list of registered workers for the internal client. This allows it to determine which jobs this node can execute.
func (c *GormClient) InternalRegisterWorker(name string, wf jobinator.WorkerFunc) {
	c.wfList = append(c.wfList, name)
}

//InternalEnqueueJob queues up a job.
func (c *GormClient) InternalEnqueueJob(j *jobinator.Job) error {
	c.dbLock.Lock()
	err := c.db.Save(j).Error
	c.dbLock.Unlock()
	return err
}

//InternalSelectJob selects a job from the database and marks it as in progress.
func (c *GormClient) InternalSelectJob() (*jobinator.Job, error) {
	wf := []string{}
	for _, x := range c.wfList {
		wf = append(wf, x)
	}
	c.dbLock.Lock()
	defer c.dbLock.Unlock()
	tx := c.db.Begin()
	j := &jobinator.Job{}
	statuses := []int{
		status.STATUS_PENDING,
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

//InternalMarkJobFinished marks a job as done.
func (c *GormClient) InternalMarkJobFinished(j *jobinator.Job) error {
	err := c.db.Model(j).Update("status", status.STATUS_DONE).Error
	return err
}

//InternalPendingJobs returns the amount of jobs that are pending.
func (c *GormClient) InternalPendingJobs() (int, error) {
	var n int
	c.dbLock.Lock()
	defer c.dbLock.Unlock()
	err := c.db.Model(&jobinator.Job{}).Where("status in (?)", []int{status.STATUS_PENDING, status.STATUS_RETRY}).Count(&n).Error
	if err != nil {
		return 0, err
	}
	return n, nil
}
