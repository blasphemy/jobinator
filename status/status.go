package status

const (
	//Pending represents a job that has not yet run
	Pending = iota
	//Running represents a job that is in progress
	Running
	//Done is a job that has finished successfully
	Done
	//Retry is a job that has failed but is ready to be retried
	Retry
	//Failed is a job that has exceeded the retry limit and given up
	Failed
)
