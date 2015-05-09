package tsq

import (
	"errors"
	"sync"
	"time"
)

type TaskQueue struct {
	tasks         map[string]Runner
	jobQueue      chan *Job
	jobMutex      sync.Mutex
	jobs          []*Job
	cleanInterval time.Duration
	maxAge        time.Duration
}

type Config struct {
	QueueLength   int
	CleanInterval time.Duration
	MaxAge        time.Duration
}

func (config *Config) NewQueue() (q *TaskQueue) {
	q = &TaskQueue{}
	q.tasks = make(map[string]Runner)
	q.jobQueue = make(chan *Job, config.QueueLength)
	q.jobs = make([]*Job, 0, 10)
	q.cleanInterval = config.CleanInterval
	q.maxAge = config.MaxAge
	return
}

var DefaultConfig Config = Config{
	QueueLength:   10,
	CleanInterval: time.Duration(24) * time.Hour,
	MaxAge:        time.Duration(7*24) * time.Hour,
}

func New() *TaskQueue {
	return DefaultConfig.NewQueue()
}

func (q *TaskQueue) Define(name string, r Runner) {
	q.tasks[name] = r
}

func (q *TaskQueue) Submit(name string, arguments interface{}) (job *Job, err error) {
	uuid, err := newUUID()
	if err != nil {
		return
	}

	if _, ok := q.tasks[name]; ok {
		now := time.Now()
		job = &Job{
			Name:      name,
			UUID:      uuid,
			Status:    JOB_PENDING,
			Arguments: arguments,
			Created:   now,
			Updated:   now,
		}
		q.add(job)
		q.jobQueue <- job
		return
	}
	err = errors.New("Unknown task: " + name)
	return
}

func (q *TaskQueue) add(job *Job) {
	q.jobMutex.Lock()
	q.jobs = append(q.jobs, job)
	q.jobMutex.Unlock()
	return
}

func (q *TaskQueue) GetJob(uuid string) (job *Job, err error) {
	for _, job := range q.jobs {
		if job.UUID == uuid {
			return job, err
		}
	}
	err = errors.New("Job " + uuid + " not found")
	return
}

func (q *TaskQueue) Start() {
	go func() {
		for {
			job := <-q.jobQueue
			q.run(job)
		}
	}()

	go func() {
		if q.cleanInterval <= 0 {
			panic("Incorrect cleaning interval")
		}
		for {
			time.Sleep(q.cleanInterval)
			q.clean()
		}
	}()
}

func (q *TaskQueue) run(job *Job) {
	job.setStatus(JOB_RUNNING)
	result, err := q.tasks[job.Name].Run(job.Arguments)
	job.Result = result
	if err != nil {
		if result == nil {
			job.Result = err.Error()
		}
		job.setStatus(JOB_FAILURE)
		return
	}
	job.setStatus(JOB_SUCCESS)
}

func (job *Job) setStatus(status string) {
	job.Status = status
	job.Updated = time.Now()
}

func (job *Job) HasFinished() bool {
	return job.Status == JOB_SUCCESS || job.Status == JOB_FAILURE
}

func (q *TaskQueue) clean() {
	q.jobMutex.Lock()
	newList := make([]*Job, 0, len(q.jobs))
	limit := time.Now().Add(-q.maxAge)
	for _, job := range q.jobs {
		if !job.HasFinished() || job.Updated.After(limit) {
			newList = append(newList, job)
		}
	}
	q.jobs = newList
	q.jobMutex.Unlock()
}
