package batch

import (
	"fmt"
	"ticketdesk/internal/model"
	"ticketdesk/internal/report"
)

func (s *Service) Dashboard(batchID string) (report.Dashboard, error) {
	summary, err := s.Summary(batchID)
	if err != nil {
		return report.Dashboard{}, err
	}
	return report.BuildDashboard(summary), nil
}

func (s *Service) FailureExport(batchID string) ([]model.FailureDetail, error) {
	summary, err := s.Summary(batchID)
	if err != nil {
		return nil, err
	}
	return report.SortFailures(summary.Failures), nil
}

func (s *Service) ProgressMessage(batchID string) (string, error) {
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return "", err
	}
	return model.CompletionMessage(batch), nil
}

func (s *Service) RequeuePending(batchID string) ([]model.WorkerTask, error) {
	codes, err := s.store.FindCodes(batchID, model.CodePending)
	if err != nil {
		return nil, err
	}
	tasks := make([]model.WorkerTask, 0, len(codes))
	for index, code := range codes {
		tasks = append(tasks, model.WorkerTask{ID: fmt.Sprintf("%s:resume:%04d", batchID, index+1), BatchID: batchID, Code: code.Code, State: model.TaskQueued})
	}
	for _, task := range tasks {
		if err := s.store.PutTask(task); err != nil {
			return nil, err
		}
	}
	return tasks, nil
}
