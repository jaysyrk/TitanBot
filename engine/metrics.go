package engine

import "time"

type Metrics struct {
	TotalRequests int
	SuccessCount  int
	FailureCount  int
	Latencies     []time.Duration
	StartTime     time.Time
	EndTime       time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{
		Latencies: make([]time.Duration, 0),
		StartTime: time.Now(),
	}
}

func (m *Metrics) AddResult(success bool, latency time.Duration) {
	m.TotalRequests++
	if success {
		m.SuccessCount++
	} else {
		m.FailureCount++
	}
	m.Latencies = append(m.Latencies, latency)
}

func (m *Metrics) Merge(other *Metrics) {
	m.TotalRequests += other.TotalRequests
	m.SuccessCount += other.SuccessCount
	m.FailureCount += other.FailureCount
	m.Latencies = append(m.Latencies, other.Latencies...)

	if other.StartTime.Before(m.StartTime) {
		m.StartTime = other.StartTime
	}
	if other.EndTime.After(m.EndTime) {
		m.EndTime = other.EndTime
	}
}
