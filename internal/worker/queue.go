package worker

import (
	"fmt"
	"sync"

	"ticketdesk/internal/batch"
	"ticketdesk/internal/model"
)

type Queue struct {
	service *batch.Service
	jobs    chan model.WorkerTask
	results chan model.ValidationAttempt
	errors  chan error
	workers int
	wg      sync.WaitGroup
	closed  chan struct{}
}

func NewQueue(service *batch.Service, workers int) *Queue {
	if workers < 1 {
		workers = 1
	}
	return &Queue{service: service, jobs: make(chan model.WorkerTask), results: make(chan model.ValidationAttempt, workers), errors: make(chan error, workers), workers: workers, closed: make(chan struct{})}
}

func (q *Queue) Start() {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.run(i + 1)
	}
}

func (q *Queue) run(index int) {
	defer q.wg.Done()
	workerName := fmt.Sprintf("worker-%d", index)
	for {
		select {
		case task, ok := <-q.jobs:
			if !ok {
				return
			}
			task.Worker = workerName
			attempt, err := q.service.ProcessTask(task)
			if err != nil {
				q.errors <- err
			} else {
				q.results <- attempt
			}
		case <-q.closed:
			return
		}
	}
}

func (q *Queue) Submit(tasks []model.WorkerTask) error {
	for _, task := range tasks {
		select {
		case q.jobs <- task:
		case <-q.closed:
			return fmt.Errorf("queue is closed")
		}
	}
	return nil
}

func (q *Queue) Results() <-chan model.ValidationAttempt { return q.results }
func (q *Queue) Errors() <-chan error                    { return q.errors }

func (q *Queue) Stop() {
	select {
	case <-q.closed:
		return
	default:
		close(q.closed)
	}
	close(q.jobs)
	q.wg.Wait()
}
