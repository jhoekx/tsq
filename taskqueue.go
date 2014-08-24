package tsq

import (
	"errors"
	"sync"
)

type TaskQueue struct {
	Tasks    map[string]Runner
	jobQueue chan *Job
	jobMutex sync.Mutex
	Jobs     []*Job
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
		job = &Job{Name: name, UUID: uuid, Status: JOB_PENDING, Arguments: arguments}
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
}

func (q *TaskQueue) run(job *Job) {
	job.Status = JOB_RUNNING
	result, err := q.Tasks[job.Name].Run(job.Arguments)
	job.Result = result
	if err != nil {
        if result == nil {
            job.Result = err.Error()
        }
		job.Status = JOB_FAILURE
		return
	}
	job.Status = JOB_SUCCESS
}
