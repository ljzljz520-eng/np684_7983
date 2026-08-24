package worker

import (
	"ticketdesk/internal/batch"
	"ticketdesk/internal/model"
)

type Recovery struct{ service *batch.Service }

func NewRecovery(service *batch.Service) *Recovery { return &Recovery{service: service} }

func (r *Recovery) Pending(batchID string) ([]model.WorkerTask, error) {
	tasks := make([]model.WorkerTask, 0)
	all, err := r.service.Tasks(batchID)
	if err != nil {
		return nil, err
	}
	for _, task := range all {
		if task.State != model.TaskDone {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}
