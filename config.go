package tsq

import "time"

type Config struct {
	QueueLength int
	JobStore    JobStore
}

func (config *Config) NewQueue() (q *TaskQueue) {
	q = &TaskQueue{
		stopQueue: make(chan bool, 1),
		tasks:     make(map[string]Runner),
		jobQueue:  make(chan *Job, config.getQueueLength()),
		jobStore:  config.getJobStore(),
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

func (config *Config) getQueueLength() (queueLength int) {
	if config.QueueLength > 0 {
		queueLength = config.QueueLength
	} else {
		queueLength = DefaultConfig.QueueLength
	}
	return
}

var DefaultConfig Config = Config{
	QueueLength: 10,
	JobStore:    NewCleanedMemoryStore(time.Duration(24)*time.Hour, time.Duration(7*24)*time.Hour),
}
