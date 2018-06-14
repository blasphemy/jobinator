package memoryclient

import "github.com/blasphemy/jobinator"

/*
type InternalClient interface {
	InternalEnqueueJob(string, interface{}) error
	InternalSelectJob() (*job, error)
	InternalMarkJobFinished(*job) error
	InternalPendingJobs() (int, error)
	InternalRegisterWorker(string, WorkerFunc)
}
*/

type MemoryClient struct {
}

func (m *MemoryClient) InternalEnqueueJob(string, interface{}) error {
	return nil
}

func (m *MemoryClient) InternalSelectJob() (*jobinator.Job, error) {
	return nil, nil
}

func (m *MemoryClient) InternalMarkJobFinished(j *jobinator.Job) error {
	return nil
}

func (m *MemoryClient) InternalPendingJobs() (int, error) {
	return 0, nil
}

func (m *MemoryClient) InternalRegisterWorker(string, jobinator.WorkerFunc) {

}
