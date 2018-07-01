package gormclient

import (
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

//NewExistingGormClient is like NewGormClient(), except instead of adding connection params, it uses an existing gorm handle.
func NewExistingGormClient(db *gorm.DB, config jobinator.ClientConfig) (*jobinator.Client, error) {
	db.AutoMigrate(&jobinator.Job{})
	newgc := &GormClient{
		db:     db,
		wfList: []string{},
	}
	newc := jobinator.NewClient(newgc, config)
	return newc, nil
}

//NewGormClient returns a new *Client backed by gorm. It requires a driver type and connection string, as well as a ClientConfig.
func NewGormClient(dbtype string, dbconn string, config jobinator.ClientConfig) (*jobinator.Client, error) {
	db, err := gorm.Open(dbtype, dbconn)
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

//InternalRegisterWorker adds a worker to the list of registered workers for the internal client. This allows it to determine which jobs this node can execute.
func (c *GormClient) InternalRegisterWorker(name string, wf jobinator.WorkerFunc) {
	c.wfList = append(c.wfList, name)
}

//InternalEnqueueJob queues up a job.
func (c *GormClient) InternalEnqueueJob(j *jobinator.Job) error {
	if j.NamedJob != "" {
		//this is a named job. let's see if we can find it
		ej := &jobinator.Job{}
		nf := c.db.First(ej, "named_job = ?", j.NamedJob).RecordNotFound()
		if !nf { //found
			updates := &jobinator.Job{
				Args:           j.Args,
				MaxRetry:       j.MaxRetry,
				Repeat:         j.Repeat,
				RepeatInterval: j.RepeatInterval,
				Name:           j.Name,
			}
			if j.Repeat {
				updates.NextRun = ej.FinishedAt + int64(j.RepeatInterval.Seconds())
			}
			err := c.db.Model(ej).Update(updates).Error
			return err
		}
	}
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
	err := tx.Order("finished_at asc").First(j, "status = ? OR (status = ? AND (repeat = false) OR (repeat = true AND ? >= next_run)) AND name in (?)", status.Retry, status.Pending, time.Now().Unix(), wf).Error
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

//InternalPendingJobs returns all pending jobs
func (c *GormClient) InternalPendingJobs() ([]*jobinator.Job, error) {
	var j []*jobinator.Job
	err := c.db.Find(&j, "status = ? OR (status = ? AND (repeat = false) OR (repeat = true AND ? >= next_run))", status.Retry, status.Pending, time.Now().Unix()).Error
	if err != nil {
		return []*jobinator.Job{}, err
	}
	return j, nil
}

//IncRetryCount increments the retry count for the job
func (c *GormClient) IncRetryCount(j *jobinator.Job) error {
	err := c.db.Model(j).Update("retry_count", gorm.Expr("retry_count + ?", 1)).Error
	j.RetryCount++
	return err
}

//SetError sets the error and stacktrace on the job, usually occurring before a retry.
func (c *GormClient) SetError(j *jobinator.Job, errtxt string, stack string) error {
	err := c.db.Model(j).Updates(&jobinator.Job{Error: errtxt, ErrorStack: stack}).Error
	return err
}

//SetFinishedAt marks the time that the job finished at
func (c *GormClient) SetFinishedAt(j *jobinator.Job, t int64) error {
	err := c.db.Model(j).Update("finished_at", t).Error
	return err
}

//SetNextRun updates the next run field of the job
func (c *GormClient) SetNextRun(j *jobinator.Job, t int64) error {
	err := c.db.Model(j).Update("next_run", t).Error
	return err
}

//InternalCleanup deletes all jobs that have been finished (or optionally failed jobs) older than specified in the CleanUpConfig
func (c *GormClient) InternalCleanup(config jobinator.CleanUpConfig) error {
	statuses := []int{
		status.Done,
	}
	if config.IncludeFailed {
		statuses = append(statuses, status.Failed)
	}
	return c.db.Delete(&jobinator.Job{}, "status in (?) AND ? > (finished_at + ?)", statuses, time.Now().Unix(), int64(config.MaxAge.Seconds())).Error
}

func (c *GormClient) GetNamedJob(name string) (*jobinator.Job, error) {
	j := &jobinator.Job{}
	err := c.db.First(j, "named_job = ?", name).Error
	if err != nil {
		return nil, err
	}
	return j, nil
}
