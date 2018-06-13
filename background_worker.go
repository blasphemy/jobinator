package jobinator

import (
	"log"
	"time"
)

func (c *client) NewBackgroundWorker() *BackgroundWorker {
	bw := &BackgroundWorker{
		running: false,
		c:       c,
	}
	return bw
}

func (bw *BackgroundWorker) backgroundWorkerFunc() {
	bw.running = true
	for {
		select {
		case <-bw.quitChan:
			log.Println("stopping worker")
			bw.running = false
			return
		default:
			log.Printf("Sleeping for %s", bw.c.config.WorkerSleepTime)
			time.Sleep(bw.c.config.WorkerSleepTime)
			bw.c.ExecuteOneJob()
		}
	}
}

func (bw *BackgroundWorker) Start() {
	bw.quitChan = make(chan bool)
	go bw.backgroundWorkerFunc()
}

func (bw *BackgroundWorker) Stop() {
	bw.quitChan <- false
}
