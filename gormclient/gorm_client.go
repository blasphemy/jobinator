package gormclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/blasphemy/jobinator"
	"github.com/blasphemy/jobinator/status"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" //needed for sqlite support
)

//GormClient represents a client using a Gorm backend (SQL)
type GormClient struct {
	db     *gorm.DB
	wfList []string
}

//NewGormClient returns a new *Client backed by gorm. It requires a driver type and connection string, as well as a ClientConfig.
func NewGormClient(dbtype string, dbconn string, config jobinator.ClientConfig) (*jobinator.Client, error) {
	err := gormDriverIsValid(dbtype)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(dbtype, dbconn)
	db.LogMode(false)
	db.AutoMigrate(&jobinator.Job{})
	if err != nil {
		return nil, err
	}
	newgc := &GormClient{
		db:     db,
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
	err := c.db.Create(j).Error
	return err
}

//InternalSelectJob selects a job from the database and marks it as in progress.
func (c *GormClient) InternalSelectJob() (*jobinator.Job, error) {
	wf := []string{}
	for _, x := range c.wfList {
		wf = append(wf, x)
	}
	tx := c.db.Begin()
	j := &jobinator.Job{}
	err := tx.First(j, "status = ? OR (status = ? AND (repeat = false) OR (repeat = true AND ? >= next_run)) AND name in (?)", status.Retry, status.Pending, time.Now().Unix(), wf).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Model(j).Update("status", status.Running).Error
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

//SetStatus sets the status for the job. See jobinator/status package for more info
func (c *GormClient) SetStatus(j *jobinator.Job, status int) error {
	err := c.db.Model(j).Update("status", status).Error
	return err
}

//InternalPendingJobs returns the amount of jobs that are pending.
func (c *GormClient) InternalPendingJobs() (int, error) {
	var n int
	err := c.db.Model(&jobinator.Job{}).Where("status in (?)", []int{status.Pending, status.Retry}).Count(&n).Error
	if err != nil {
		return 0, err
	}
	return n, nil
}

//IncRetryCount increments the retry count for the job
func (c *GormClient) IncRetryCount(j *jobinator.Job) error {
	err := c.db.Model(j).Update("retry_count", gorm.Expr("retry_count + ?", 1)).Error
	return err
}

//SetError sets the error and stacktrace on the job, usually occuring before a retry.
func (c *GormClient) SetError(j *jobinator.Job, errtxt string, stack string) error {
	err := c.db.Model(j).Updates(&jobinator.Job{Error: errtxt, ErrorStack: stack}).Error
	return err
}

//SetFinishedAt marks the time that the job finished at
func (c *GormClient) SetFinishedAt(j *jobinator.Job, t int64) error {
	err := c.db.Model(j).Update("finished_at", t).Error
	return err
}

func (c *GormClient) SetNextRun(j *jobinator.Job, t int64) error {
	err := c.db.Model(j).Update("next_run", t).Error
	return err
}
