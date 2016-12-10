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

func (s *CleanedMemoryStore) Store(job *Job) error {
	s.jobMutex.Lock()
	s.jobs = append(s.jobs, job)
	s.jobMutex.Unlock()
	return nil
}

func (s *CleanedMemoryStore) GetJobs() ([]*Job, error) {
	return s.jobs, nil
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

func (s *CleanedMemoryStore) SetStatus(uuid string, status string, updated time.Time) (err error) {
	job, err := s.GetJob(uuid)
	if err != nil {
		return
	}
	job.Status = status
	job.Updated = updated
	return
}

func (s *CleanedMemoryStore) SetResult(uuid string, result interface{}) (err error) {
	job, err := s.GetJob(uuid)
	if err != nil {
		return
	}
	job.Result = result
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
