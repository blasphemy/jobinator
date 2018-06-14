package jobinator

import (
	"sync"
	"time"
)

//BackgroundWorker represents a worker thread that runs in the background
type BackgroundWorker struct {
	c        *Client
	quitChan chan bool
	running  bool
	runMutex sync.Mutex
}

//NewBackgroundWorker returns a background worker handle, as well as registers it in the client. You can either keep it to start it yourself or use the client to start all background worker threads.
func (c *Client) NewBackgroundWorker() *BackgroundWorker {
	bw := &BackgroundWorker{
		running:  false,
		c:        c,
		runMutex: sync.Mutex{},
	}
	c.workers = append(c.workers, bw)
	return bw
}

func (bw *BackgroundWorker) backgroundWorkerFunc() {
	bw.runMutex.Lock()
	bw.running = true
	bw.runMutex.Unlock()
	for {
		select {
		case <-bw.quitChan:
			bw.runMutex.Lock()
			bw.running = false
			bw.runMutex.Unlock()
			return
		default:
			time.Sleep(bw.c.config.WorkerSleepTime)
			bw.c.backgroundExecute()
		}
	}
}

//IsRunning returns whether or not the background worker is running.
func (bw *BackgroundWorker) IsRunning() bool {
	bw.runMutex.Lock()
	defer bw.runMutex.Unlock()
	return bw.running
}

//Start starts the background worker. If it is already running, nothing happens.
func (bw *BackgroundWorker) Start() {
	if bw.IsRunning() {
		return
	}
	bw.quitChan = make(chan bool)
	go bw.backgroundWorkerFunc()
}

//Stop stops the background worker. This is non blocking, so it may take some time before the background worker is fully stopped. Use IsRunning() to see if it's still running.
func (bw *BackgroundWorker) Stop() {
	if bw.IsRunning() {
		bw.quitChan <- true
	}
}

//StopBlocking stops the background worker. It will block until the background worker has stopped running.
func (bw *BackgroundWorker) StopBlocking() {
	bw.Stop()
	for bw.IsRunning() == true {
		time.Sleep(bw.c.config.WorkerSleepTime / 2)
	}
}
