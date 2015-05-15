package tsq

import (
	"errors"
	"time"
)

type TaskQueue struct {
	stopQueue chan bool
	tasks     map[string]Runner
	jobQueue  chan *Job
	jobStore  JobStore
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
		q.jobStore.Store(job)
		q.jobQueue <- job
		return
	}
	err = errors.New("Unknown task: " + name)
	return
}

func (q *TaskQueue) GetJobs() (jobs []*Job) {
	return q.jobStore.GetJobs()
}

func (q *TaskQueue) GetJob(uuid string) (job *Job, err error) {
	return q.jobStore.GetJob(uuid)
}

func (q *TaskQueue) Start() (err error) {
	q.jobStore.Start()
	go func() {
		for {
			select {
			case job := <-q.jobQueue:
				q.run(job)
			case <-q.stopQueue:
				return
			}

		}
	}()
	return
}

func (q *TaskQueue) Stop() {
	q.jobStore.Stop()
	q.stopQueue <- true
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
