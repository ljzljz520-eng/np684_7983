package batch

import (
	"ticketdesk/internal/model"
	"ticketdesk/internal/store"
)

func (s *Service) GetTask(id string) (model.WorkerTask, error) {
	tasks, err := s.store.ListTasks("")
	if err != nil {
		return model.WorkerTask{}, err
	}
	for _, task := range tasks {
		if task.ID == id {
			return task, nil
		}
	}
	return model.WorkerTask{}, store.ErrNotFound
}

func (s *Service) SetTaskWorker(id, worker string) (model.WorkerTask, error) {
	task, err := s.GetTask(id)
	if err != nil {
		return task, err
	}
	return s.store.UpdateTask(id, func(t *model.WorkerTask) error { t.Worker = worker; return nil })
}

func (s *Service) Tasks(batchID string) ([]model.WorkerTask, error) {
	return s.store.ListTasks(batchID)
}

func (s *Service) Batches() ([]model.TicketBatch, error) { return s.store.ListBatches() }
