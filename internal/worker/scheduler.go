package worker

import (
	"fmt"
	"sync"
	"ticketdesk/internal/batch"
	"ticketdesk/internal/model"
)

type Scheduler struct {
	service *batch.Service
	metrics *Metrics
	mu      sync.Mutex
	running bool
}

func NewScheduler(service *batch.Service) *Scheduler {
	return &Scheduler{service: service, metrics: &Metrics{}}
}

func (s *Scheduler) Begin(batchID string, workers int) ([]model.WorkerTask, error) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil, fmt.Errorf("scheduler already running")
	}
	s.running = true
	s.mu.Unlock()
	plan, err := s.service.BuildPlan(batchID, workers)
	if err != nil {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return nil, err
	}
	if err := s.service.SavePlan(plan); err != nil {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return nil, err
	}
	s.metrics.RecordSubmission(len(plan.Tasks))
	return plan.Tasks, nil
}

func (s *Scheduler) Finish()           { s.mu.Lock(); s.running = false; s.mu.Unlock() }
func (s *Scheduler) Running() bool     { s.mu.Lock(); defer s.mu.Unlock(); return s.running }
func (s *Scheduler) Metrics() *Metrics { return s.metrics }

func (s *Scheduler) ProcessSequential(tasks []model.WorkerTask, workerName string) error {
	for _, task := range tasks {
		task.Worker = workerName
		attempt, err := s.service.ProcessTask(task)
		if err != nil {
			return err
		}
		s.metrics.RecordResult(attempt)
	}
	s.Finish()
	return nil
}
