package tsq

import "time"

type Config struct {
	QueueLength   int
	JobStore      JobStore
	CleanInterval time.Duration
	CleanStrategy CleanStrategy
}

func (config *Config) NewQueue() (q *TaskQueue) {
	q = &TaskQueue{
		tasks:         make(map[string]Runner),
		jobQueue:      make(chan *Job, config.QueueLength),
		jobStore:      config.getJobStore(),
		cleanInterval: config.CleanInterval,
		cleaner:       config.getCleanStrategy(),
	}
	return
}

func (config *Config) getJobStore() (store JobStore) {
	if config.JobStore != nil {
		store = config.JobStore
	} else {
		store = DefaultConfig.JobStore
	}
	return
}

func (config *Config) getCleanStrategy() (cleaner CleanStrategy) {
	if config.CleanStrategy != nil {
		cleaner = config.CleanStrategy
	} else {
		cleaner = DefaultConfig.CleanStrategy
	}
	return
}

var DefaultConfig Config = Config{
	QueueLength:   10,
	JobStore:      NewMemoryStore(),
	CleanInterval: time.Duration(24) * time.Hour,
	CleanStrategy: &TimeBasedCleanStrategy{time.Duration(7*24) * time.Hour},
}
