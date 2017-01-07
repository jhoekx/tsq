package tsq

import (
	"errors"
	"sync"
	"time"
)

type MemoryStore struct {
	jobMutex sync.Mutex
	jobs     []*Job
}

func NewMemoryStore() JobStore {
	store := &MemoryStore{}
	store.jobs = make([]*Job, 0, 10)
	return store
}

func (s *MemoryStore) Start() (err error) {
	return
}

func (s *MemoryStore) Stop() {
}

func (s *MemoryStore) Store(job *Job) error {
	s.jobMutex.Lock()
	s.jobs = append(s.jobs, job)
	s.jobMutex.Unlock()
	return nil
}

func (s *MemoryStore) GetJobs() ([]*Job, error) {
	return s.jobs, nil
}

func (s *MemoryStore) GetJob(uuid string) (job *Job, err error) {
	for _, job := range s.jobs {
		if job.UUID == uuid {
			return job, err
		}
	}
	err = errors.New("Job " + uuid + " not found")
	return
}

func (s *MemoryStore) SetStatus(uuid string, status string, updated time.Time) (err error) {
	job, err := s.GetJob(uuid)
	if err != nil {
		return
	}
	job.Status = status
	job.Updated = updated
	return
}

func (s *MemoryStore) SetResult(uuid string, result interface{}) (err error) {
	job, err := s.GetJob(uuid)
	if err != nil {
		return
	}
	job.Result = result
	return
}
