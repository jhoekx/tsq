package tsq

import (
	"errors"
	"sync"
	"time"
)

type TaskQueue struct {
	Tasks         map[string]Runner
	jobQueue      chan *Job
	jobMutex      sync.Mutex
	Jobs          []*Job
	CleanInterval time.Duration
	MaxAge        time.Duration
}

func (q *TaskQueue) Define(name string, r Runner) {
	q.Tasks[name] = r
}

func (q *TaskQueue) Submit(name string, arguments interface{}) (job *Job, err error) {
	uuid, err := newUUID()
	if err != nil {
		return
	}

	if _, ok := q.Tasks[name]; ok {
		now := time.Now()
		job = &Job{Name: name, UUID: uuid, Status: JOB_PENDING, Arguments: arguments, Created: now, Updated: now}
		q.add(job)
		q.jobQueue <- job
		return
	}
	err = errors.New("Unknown task: " + name)
	return
}

func (q *TaskQueue) add(job *Job) {
	q.jobMutex.Lock()
	q.Jobs = append(q.Jobs, job)
	q.jobMutex.Unlock()
	return
}

func (q *TaskQueue) GetJob(uuid string) (job *Job, err error) {
	for _, job := range q.Jobs {
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
		for {
			time.Sleep(q.CleanInterval * time.Second)
			q.clean()
		}
	}()
}

func (q *TaskQueue) run(job *Job) {
	job.SetStatus(JOB_RUNNING)
	result, err := q.Tasks[job.Name].Run(job.Arguments)
	job.Result = result
	if err != nil {
		if result == nil {
			job.Result = err.Error()
		}
		job.SetStatus(JOB_FAILURE)
		return
	}
	job.SetStatus(JOB_SUCCESS)
}

func (job *Job) SetStatus(status string) {
	job.Status = status
	job.Updated = time.Now()
}

func (q *TaskQueue) clean() {
	q.jobMutex.Lock()
	newList := make([]*Job, 0, len(q.Jobs))
	limit := time.Now().Add(-q.MaxAge * time.Second)
	for _, job := range q.Jobs {
		if job.Updated.After(limit) {
			newList = append(newList, job)
		}
	}
	q.Jobs = newList
	q.jobMutex.Unlock()
}
