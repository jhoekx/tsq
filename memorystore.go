package tsq

import (
	"errors"
	"sync"
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

func (s *MemoryStore) Store(job *Job) {
	s.jobMutex.Lock()
	s.jobs = append(s.jobs, job)
	s.jobMutex.Unlock()
}

func (s *MemoryStore) GetJobs() (jobs []*Job) {
	return s.jobs
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

func (s *MemoryStore) Clean(cleaner CleanStrategy) {
	s.jobMutex.Lock()
	defer s.jobMutex.Unlock()
	newList := make([]*Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		if !job.HasFinished() || cleaner.ShouldKeep(job) {
			newList = append(newList, job)
		}
	}
	s.jobs = newList
}
