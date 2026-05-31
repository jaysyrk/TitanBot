package engine

import (
	"sync"
	"time"
)

type Result struct {
	Success bool
	Latency time.Duration
}

type Runner struct {
	Concurrency int
	Duration    time.Duration
	Target      string
	Results     chan Result
	Metrics     *Metrics
}

func NewRunner(concurrency int, duration time.Duration, target string) *Runner {
	return &Runner{
		Concurrency: concurrency,
		Duration:    duration,
		Target:      target,
		Results:     make(chan Result, concurrency*2), // Buffered channel
		Metrics:     NewMetrics(),
	}
}

func (r *Runner) StartCollector(wg *sync.WaitGroup) {
	defer wg.Done()
	for res := range r.Results {
		r.Metrics.AddResult(res.Success, res.Latency)
	}
	r.Metrics.EndTime = time.Now()
}
