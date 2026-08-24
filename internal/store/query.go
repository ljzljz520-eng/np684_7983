package store

import (
	"go.etcd.io/bbolt"
	"strings"
	"ticketdesk/internal/model"
)

func (s *Store) FindCodes(batchID, state string) ([]model.TicketCode, error) {
	result := make([]model.TicketCode, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("codes")).ForEach(func(_, value []byte) error {
			var code model.TicketCode
			if err := model.Decode(value, &code); err != nil {
				return err
			}
			if (batchID == "" || code.BatchID == batchID) && (state == "" || code.State == state) {
				result = append(result, code)
			}
			return nil
		})
	})
	return result, err
}

func (s *Store) FindFailures(batchID, category string, retryableOnly bool) ([]model.FailureDetail, error) {
	result := make([]model.FailureDetail, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("failures")).ForEach(func(_, value []byte) error {
			var failure model.FailureDetail
			if err := model.Decode(value, &failure); err != nil {
				return err
			}
			if batchID != "" && failure.BatchID != batchID {
				return nil
			}
			if category != "" && !strings.EqualFold(category, failure.Category) {
				return nil
			}
			if retryableOnly && (!failure.Retryable || failure.Resolved) {
				return nil
			}
			result = append(result, failure)
			return nil
		})
	})
	return result, err
}

func (s *Store) FindAttempts(batchID, outcome string) ([]model.ValidationAttempt, error) {
	result := make([]model.ValidationAttempt, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("attempts")).ForEach(func(_, value []byte) error {
			var attempt model.ValidationAttempt
			if err := model.Decode(value, &attempt); err != nil {
				return err
			}
			if (batchID == "" || attempt.BatchID == batchID) && (outcome == "" || attempt.Outcome == outcome) {
				result = append(result, attempt)
			}
			return nil
		})
	})
	return result, err
}

func (s *Store) MarkFailureResolved(id string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("failures"))
		var targetKey []byte
		if err := bucket.ForEach(func(k, value []byte) error {
			var failure model.FailureDetail
			if err := model.Decode(value, &failure); err != nil {
				return err
			}
			if failure.ID == id {
				targetKey = append([]byte(nil), k...)
			}
			return nil
		}); err != nil {
			return err
		}
		if targetKey == nil {
			return ErrNotFound
		}
		var failure model.FailureDetail
		if err := model.Decode(bucket.Get(targetKey), &failure); err != nil {
			return err
		}
		failure.Resolved = true
		data, err := model.Encode(failure)
		if err != nil {
			return err
		}
		return bucket.Put(targetKey, data)
	})
}
