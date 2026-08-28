package batch

import (
	"ticketdesk/internal/model"
	"ticketdesk/internal/store"
)

func (s *Service) Transition(batchID, code, target, actor string) (model.TicketCode, error) {
	current, err := s.store.GetCode(batchID, code)
	if err != nil {
		return current, err
	}
	updated, event, err := model.TransitionCode(current, target, actor)
	if err != nil {
		return current, err
	}
	if err := s.store.PutCode(updated); err != nil {
		return current, err
	}
	if err := s.store.SaveLifecycle(event); err != nil {
		return current, err
	}
	return updated, nil
}

func (s *Service) Audit(batchID string) ([]model.LifecycleEvent, error) {
	return s.store.ListLifecycle(batchID)
}

func (s *Service) Reconcile(batchID string) error {
	codes, err := s.store.ListCodes(batchID)
	if err != nil {
		return err
	}
	for _, code := range codes {
		if !model.TerminalCode(code.State) {
			continue
		}
		if _, err := s.store.LatestLifecycle(batchID, code.Code); err == store.ErrNotFound {
			event := model.LifecycleEvent{ID: code.Key() + ":reconcile", BatchID: batchID, Code: code.Code, From: model.CodePending, To: code.State, Actor: "reconciler", Sequence: code.Attempts}
			if err := s.store.SaveLifecycle(event); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) ResolveFailure(id string) error { return s.store.MarkFailureResolved(id) }
