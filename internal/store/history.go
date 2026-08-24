package store

import (
	"fmt"
	"go.etcd.io/bbolt"
	"ticketdesk/internal/model"
)

func (s *Store) SaveLifecycle(event model.LifecycleEvent) error {
	if event.ID == "" || event.BatchID == "" {
		return fmt.Errorf("lifecycle identity is required")
	}
	data, err := model.Encode(event)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("history"))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(event.ID), data)
	})
}

func (s *Store) ListLifecycle(batchID string) ([]model.LifecycleEvent, error) {
	result := make([]model.LifecycleEvent, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("history"))
		if bucket == nil {
			return nil
		}
		return bucket.ForEach(func(_, value []byte) error {
			var event model.LifecycleEvent
			if err := model.Decode(value, &event); err != nil {
				return err
			}
			if batchID == "" || event.BatchID == batchID {
				result = append(result, event)
			}
			return nil
		})
	})
	return result, err
}

func (s *Store) LatestLifecycle(batchID, code string) (model.LifecycleEvent, error) {
	events, err := s.ListLifecycle(batchID)
	if err != nil {
		return model.LifecycleEvent{}, err
	}
	var latest model.LifecycleEvent
	found := false
	for _, event := range events {
		if event.Code == code && (!found || event.Sequence > latest.Sequence) {
			latest = event
			found = true
		}
	}
	if !found {
		return latest, ErrNotFound
	}
	return latest, nil
}
