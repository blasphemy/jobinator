package jobinator

import (
	"log"
	"sync"
	"time"
)

func (c *client) NewBackgroundWorker() *BackgroundWorker {
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
			log.Println("stopping worker")
			bw.runMutex.Lock()
			bw.running = false
			bw.runMutex.Unlock()
			return
		default:
			log.Printf("Sleeping for %s", bw.c.config.WorkerSleepTime)
			time.Sleep(bw.c.config.WorkerSleepTime)
			bw.c.ExecuteOneJob()
		}
	}
}

func (bw *BackgroundWorker) IsRunning() bool {
	bw.runMutex.Lock()
	defer bw.runMutex.Unlock()
	return bw.running
}

func (bw *BackgroundWorker) Start() {
	if bw.IsRunning() {
		return
	}
	bw.quitChan = make(chan bool)
	go bw.backgroundWorkerFunc()
}

func (bw *BackgroundWorker) Stop() {
	bw.quitChan <- false
}

func (bw *BackgroundWorker) StopBlocking() {
	bw.Stop()
	for bw.IsRunning() == true {
		time.Sleep(1 * time.Second)
	}
}
