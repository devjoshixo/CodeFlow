package metrics

import "sync"

type Metrics struct {
	mu                  sync.Mutex
	ExecutionsStarted   int64
	ExecutionsCompleted int64
	Errors              int64
}

func New() *Metrics {
	return &Metrics{}
}

func (m *Metrics) IncrementStarted() {
	m.mu.Lock()
	m.ExecutionsStarted++
	m.mu.Unlock()
}

func (m *Metrics) IncrementCompleted() {
	m.mu.Lock()
	m.ExecutionsCompleted++
	m.mu.Unlock()
}

func (m *Metrics) GetMetrics() (started, completed, errors int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ExecutionsStarted, m.ExecutionsCompleted, m.Errors
}
