package tsq

import (
	"time"
)

type Runner interface {
	Run(args interface{}) (interface{}, error)
}

type Job struct {
	UUID      string      `json:"uuid"`
	Name      string      `json:"name"`
	Status    string      `json:"status"`
	Arguments interface{} `json:"arguments"`
	Result    interface{} `json:"result"`
	Created   time.Time   `json:"created"`
	Updated   time.Time   `json:"updated"`
}

const (
	JOB_PENDING = "PENDING"
	JOB_RUNNING = "RUNNING"
	JOB_SUCCESS = "SUCCESS"
	JOB_FAILURE = "FAILURE"
)

type LifeCycle interface {
	Start() error
	Stop()
}

type JobStore interface {
	LifeCycle
	Store(job *Job) error
	GetJob(uuid string) (*Job, error)
	SetStatus(uuid string, status string, updated time.Time) error
	SetResult(uuid string, result interface{}) error
	GetJobs() ([]*Job, error)
}
