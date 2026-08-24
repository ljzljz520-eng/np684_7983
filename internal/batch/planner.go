package batch

import (
	"fmt"
	"sort"
	"ticketdesk/internal/model"
)

type Plan struct {
	BatchID          string
	Tasks            []model.WorkerTask
	Pending          int
	Invalid          int
	EstimatedWorkers int
}

func (s *Service) BuildPlan(batchID string, workerCount int) (Plan, error) {
	if workerCount < 1 {
		workerCount = 1
	}
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return Plan{}, err
	}
	codes, err := s.store.FindCodes(batchID, model.CodePending)
	if err != nil {
		return Plan{}, err
	}
	sort.Slice(codes, func(i, j int) bool { return codes[i].Code < codes[j].Code })
	plan := Plan{BatchID: batch.ID, Pending: len(codes), EstimatedWorkers: workerCount, Tasks: make([]model.WorkerTask, 0, len(codes))}
	for index, code := range codes {
		issue := s.validator.ValidateFormat(code.Code)
		if issue.Code != "" {
			plan.Invalid++
			continue
		}
		plan.Tasks = append(plan.Tasks, model.WorkerTask{ID: fmt.Sprintf("%s-plan-%04d", batchID, index+1), BatchID: batchID, Code: code.Code, State: model.TaskQueued})
	}
	return plan, nil
}

func (s *Service) SavePlan(plan Plan) error {
	for _, task := range plan.Tasks {
		if err := s.store.PutTask(task); err != nil {
			return err
		}
	}
	_, err := s.store.UpdateBatch(plan.BatchID, func(batch *model.TicketBatch) error {
		if plan.Pending > 0 {
			batch.Status = model.BatchRunning
		}
		return nil
	})
	return err
}

func SplitTasks(tasks []model.WorkerTask, groups int) [][]model.WorkerTask {
	if groups < 1 {
		groups = 1
	}
	result := make([][]model.WorkerTask, groups)
	for index, task := range tasks {
		target := index % groups
		result[target] = append(result[target], task)
	}
	return result
}

func TaskDistribution(tasks []model.WorkerTask) map[string]int {
	counts := make(map[string]int)
	for _, task := range tasks {
		counts[task.State]++
	}
	return counts
}
