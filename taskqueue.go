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

	if _, ok := q.tasks[name]; !ok {
		err = errors.New("Unknown task: " + name)
		return
	}

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

func (q *TaskQueue) GetJobs() []*Job {
	return q.jobStore.GetJobs()
}

func (q *TaskQueue) GetJob(uuid string) (*Job, error) {
	return q.jobStore.GetJob(uuid)
}

func (q *TaskQueue) Start() (err error) {
	err = q.jobStore.Start()
	if err != nil {
		return
	}
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
	q.jobStore.SetStatus(job.UUID, JOB_RUNNING)
	result, err := q.tasks[job.Name].Run(job.Arguments)
	q.jobStore.SetResult(job.UUID, result)
	if err != nil {
		if result == nil {
			q.jobStore.SetResult(job.UUID, err.Error())
		}
		q.jobStore.SetStatus(job.UUID, JOB_FAILURE)
		return
	}
	q.jobStore.SetStatus(job.UUID, JOB_SUCCESS)
}

func (job *Job) HasFinished() bool {
	return job.Status == JOB_SUCCESS || job.Status == JOB_FAILURE
}
