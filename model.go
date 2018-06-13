package jobinator

import (
	"time"

	"github.com/jinzhu/gorm"
)

type job struct {
	ID        int64
	Name      string
	Args      []byte
	CreatedAt time.Time
	Status    int
}

type jobRef struct {
	argData []byte
	c       *client
	j       *job
}

type WorkerFunc func(j *jobRef) error

type client struct {
	db          *gorm.DB
	workerFuncs map[string]WorkerFunc
	config      clientConfig
}

type clientConfig struct {
	WorkerSleepTime time.Duration
}

type BackgroundWorker struct {
	c        *client
	quitChan chan bool
	running  bool
}
