package tsq

import (
	"errors"
	"sync"
	"time"
)

type CleanedMemoryStore struct {
	jobMutex sync.Mutex
	jobs     []*Job
	cleaner  *Cleaner
}

func NewCleanedMemoryStore(cleanInterval time.Duration, maxAge time.Duration) JobStore {
	store := &CleanedMemoryStore{}
	store.jobs = make([]*Job, 0, 10)
	cleanStrategy := &TimeBasedCleanStrategy{MaxAge: maxAge}
	store.cleaner = NewCleaner(store, cleanInterval, cleanStrategy)
	return store
}

func (s *CleanedMemoryStore) Start() (err error) {
	err = s.cleaner.Start()
	return
}

func (s *CleanedMemoryStore) Stop() {
	s.cleaner.Stop()
}

func (s *CleanedMemoryStore) Store(job *Job) {
	s.jobMutex.Lock()
	s.jobs = append(s.jobs, job)
	s.jobMutex.Unlock()
}

func (s *CleanedMemoryStore) GetJobs() (jobs []*Job) {
	return s.jobs
}

func (s *CleanedMemoryStore) GetJob(uuid string) (job *Job, err error) {
	for _, job := range s.jobs {
		if job.UUID == uuid {
			return job, err
		}
	}
	err = errors.New("Job " + uuid + " not found")
	return
}

func (s *CleanedMemoryStore) Clean(cleaner CleanStrategy) {
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
