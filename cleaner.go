package tsq

import "time"

type Cleanable interface {
	Clean(cleanStrategy CleanStrategy)
}

type CleanStrategy interface {
	ShouldKeep(job *Job) bool
}

type Cleaner struct {
	ticker        *time.Ticker
	cleanInterval time.Duration
	cleanStrategy CleanStrategy
	store         Cleanable
	stopQueue     chan bool
}

func NewCleaner(store Cleanable, cleanInterval time.Duration, cleanStrategy CleanStrategy) (cleaner *Cleaner) {
	return &Cleaner{
		stopQueue:     make(chan bool, 1),
		store:         store,
		cleanInterval: cleanInterval,
		cleanStrategy: cleanStrategy,
	}
}

func (c *Cleaner) Start() (err error) {
	c.ticker = time.NewTicker(c.cleanInterval)
	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.Clean()
			case <-c.stopQueue:
				return
			}
		}
	}()
	return
}

func (c *Cleaner) Stop() {
	c.ticker.Stop()
	c.stopQueue <- true
}

func (c *Cleaner) Clean() {
	c.store.Clean(c.cleanStrategy)
}

type TimeBasedCleanStrategy struct {
	MaxAge time.Duration
}

func (s *TimeBasedCleanStrategy) ShouldKeep(job *Job) bool {
	limit := time.Now().Add(-s.MaxAge)
	return job.Updated.After(limit)
}
