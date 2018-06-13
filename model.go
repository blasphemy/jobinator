package jobinator

import "time"

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
