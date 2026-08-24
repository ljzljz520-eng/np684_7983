package batch

import (
	"fmt"
	"sort"

	"ticketdesk/internal/model"
	"ticketdesk/internal/store"
	"ticketdesk/internal/validate"
)

type Service struct {
	store     *store.Store
	validator *validate.CodeValidator
}

func NewService(db *store.Store) *Service {
	return &Service{store: db, validator: validate.NewCodeValidator()}
}

func (s *Service) RegisterBatch(id, source, operator string, rawCodes []string) (model.TicketBatch, []validate.Issue, error) {
	codes := make([]model.TicketCode, 0, len(rawCodes))
	for _, raw := range rawCodes {
		codes = append(codes, model.TicketCode{BatchID: id, Code: model.NormalizeCode(raw), State: model.CodePending})
	}
	batch := model.TicketBatch{ID: id, Source: source, CreatedBy: operator, Status: model.BatchPending, Total: len(codes)}
	issues := s.validator.ValidateBatch(batch, codes)
	if len(issues) > 0 {
		return batch, issues, nil
	}
	if err := s.store.PutBatch(batch); err != nil {
		return batch, nil, err
	}
	for _, code := range codes {
		if err := s.store.PutCode(code); err != nil {
			return batch, nil, err
		}
	}
	return batch, nil, nil
}

func (s *Service) QueueBatch(batchID string) ([]model.WorkerTask, error) {
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return nil, err
	}
	codes, err := s.store.ListCodes(batchID)
	if err != nil {
		return nil, err
	}
	tasks := make([]model.WorkerTask, 0, len(codes))
	for i, code := range codes {
		task := model.WorkerTask{ID: fmt.Sprintf("%s-%04d", batchID, i+1), BatchID: batchID, Code: code.Code, State: model.TaskQueued}
		if err := s.store.PutTask(task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	_, err = s.store.UpdateBatch(batch.ID, func(b *model.TicketBatch) error { b.Status = model.BatchRunning; return nil })
	return tasks, err
}

func (s *Service) ProcessTask(task model.WorkerTask) (model.ValidationAttempt, error) {
	if task.State == model.TaskDone {
		return model.ValidationAttempt{}, fmt.Errorf("task already completed")
	}
	if _, err := s.store.UpdateTask(task.ID, func(t *model.WorkerTask) error { t.State = model.TaskWorking; t.Tries++; return nil }); err != nil {
		return model.ValidationAttempt{}, err
	}
	code, err := s.store.GetCode(task.BatchID, task.Code)
	if err != nil {
		return model.ValidationAttempt{}, err
	}
	issue := s.validator.ValidateFormat(code.Code)
	_, consumeErr := s.store.ConsumeTicketCode(task.BatchID, task.Code, task.Worker)
	outcome := validate.DecideCodeOutcome(code, issue, consumeErr)
	attempt := model.ValidationAttempt{ID: fmt.Sprintf("%s-attempt-%d", task.ID, task.Tries), BatchID: task.BatchID, Code: task.Code, Worker: task.Worker, Outcome: model.AttemptSuccess, Message: outcome.Message, Sequence: task.Tries}
	if outcome.State != model.CodeConsumed {
		attempt.Outcome = model.AttemptFailure
		if outcome.State == model.CodeDuplicate {
			attempt.Outcome = model.AttemptDuplicate
		}
	}
	if err := s.store.PutAttempt(attempt); err != nil {
		return attempt, err
	}
	if outcome.State != model.CodeConsumed {
		failure := model.FailureDetail{ID: attempt.ID, BatchID: task.BatchID, Code: task.Code, Category: s.validator.Classify(outcome.Issue), Message: outcome.Message, Retryable: s.validator.Retryable(outcome.Issue)}
		if outcome.State == model.CodeDuplicate {
			failure.Category = "duplicate_consumption"
			failure.Retryable = false
		}
		if err := s.store.PutFailure(failure); err != nil {
			return attempt, err
		}
	}
	if _, err := s.store.UpdateTask(task.ID, func(t *model.WorkerTask) error { t.State = model.TaskDone; return nil }); err != nil {
		return attempt, err
	}
	if err := s.recalculateBatch(task.BatchID); err != nil {
		return attempt, err
	}
	return attempt, nil
}

func (s *Service) recalculateBatch(batchID string) error {
	codes, err := s.store.ListCodes(batchID)
	if err != nil {
		return err
	}
	processed, succeeded, failed := 0, 0, 0
	for _, code := range codes {
		if model.TerminalCode(code.State) {
			processed++
		}
		if code.State == model.CodeConsumed {
			succeeded++
		}
		if code.State == model.CodeInvalid || code.State == model.CodeDuplicate {
			failed++
		}
	}
	_, err = s.store.UpdateBatch(batchID, func(b *model.TicketBatch) error {
		b.Processed = processed
		b.Succeeded = succeeded
		b.Failed = failed
		b.Status = validate.BatchState(processed, succeeded, failed, b.Total)
		return nil
	})
	return err
}

func (s *Service) RetryFailures(batchID string) ([]model.WorkerTask, error) {
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return nil, err
	}
	if batch.RetryCount >= 3 {
		return nil, fmt.Errorf("retry limit reached")
	}
	failures, err := s.store.ListFailures(batchID)
	if err != nil {
		return nil, err
	}
	tasks := make([]model.WorkerTask, 0)
	for i, failure := range failures {
		if !validate.IsRetryAllowed(failure, batch.RetryCount) {
			continue
		}
		task := model.WorkerTask{ID: fmt.Sprintf("%s-retry-%d-%d", batchID, batch.RetryCount+1, i+1), BatchID: batchID, Code: failure.Code, State: model.TaskQueued}
		if err := s.store.PutTask(task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	_, err = s.store.UpdateBatch(batchID, func(b *model.TicketBatch) error { b.RetryCount++; b.Status = model.BatchRunning; return nil })
	return tasks, err
}

func (s *Service) Summary(batchID string) (model.BatchSummary, error) {
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return model.BatchSummary{}, err
	}
	codes, err := s.store.ListCodes(batchID)
	if err != nil {
		return model.BatchSummary{}, err
	}
	failures, err := s.store.ListFailures(batchID)
	if err != nil {
		return model.BatchSummary{}, err
	}
	attempts, err := s.store.ListAttempts(batchID)
	if err != nil {
		return model.BatchSummary{}, err
	}
	pending := 0
	for _, code := range codes {
		if code.State == model.CodePending {
			pending++
		}
	}
	sort.Slice(codes, func(i, j int) bool { return codes[i].Code < codes[j].Code })
	sort.Slice(failures, func(i, j int) bool { return failures[i].Code < failures[j].Code })
	sort.Slice(attempts, func(i, j int) bool { return attempts[i].Sequence < attempts[j].Sequence })
	return model.BatchSummary{Batch: batch, Codes: codes, Failures: failures, Attempts: attempts, Pending: pending, Percent: batch.Progress()}, nil
}
