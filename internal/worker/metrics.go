package worker

import (
	"sync"
	"ticketdesk/internal/model"
)

type Metrics struct {
	mu        sync.Mutex
	submitted int
	completed int
	failed    int
}

func (m *Metrics) RecordSubmission(count int) { m.mu.Lock(); m.submitted += count; m.mu.Unlock() }
func (m *Metrics) RecordResult(attempt model.ValidationAttempt) {
	m.mu.Lock()
	m.completed++
	if attempt.Outcome != model.AttemptSuccess {
		m.failed++
	}
	m.mu.Unlock()
}
func (m *Metrics) Snapshot() (int, int, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.submitted, m.completed, m.failed
}
func (m *Metrics) CompletionRate() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.submitted == 0 {
		return 0
	}
	return float64(m.completed) / float64(m.submitted)
}
