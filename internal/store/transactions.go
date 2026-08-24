package store

import (
	"fmt"

	"go.etcd.io/bbolt"
	"ticketdesk/internal/model"
)

func (s *Store) SaveValidationResult(attempt model.ValidationAttempt, failure *model.FailureDetail) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		attemptData, err := model.Encode(attempt)
		if err != nil {
			return err
		}
		if err := tx.Bucket([]byte("attempts")).Put([]byte(attempt.ID), attemptData); err != nil {
			return err
		}
		if failure != nil {
			failureData, err := model.Encode(*failure)
			if err != nil {
				return err
			}
			if err := tx.Bucket([]byte("failures")).Put([]byte(failure.Key()), failureData); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) DeleteBatch(batchID string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		if err := tx.Bucket([]byte("batches")).Delete([]byte(batchID)); err != nil {
			return err
		}
		for _, bucketName := range [][]byte{[]byte("codes"), []byte("tasks"), []byte("attempts"), []byte("failures")} {
			bucket := tx.Bucket(bucketName)
			keys := make([][]byte, 0)
			if err := bucket.ForEach(func(k, v []byte) error {
				if len(v) > 0 && string(k) == batchID {
					return nil
				}
				var match string
				if bucketName[0] == 'c' {
					var code model.TicketCode
					if err := model.Decode(v, &code); err != nil {
						return err
					}
					match = code.BatchID
				}
				if bucketName[0] == 't' {
					var task model.WorkerTask
					if err := model.Decode(v, &task); err != nil {
						return err
					}
					match = task.BatchID
				}
				if bucketName[0] == 'a' {
					var attempt model.ValidationAttempt
					if err := model.Decode(v, &attempt); err != nil {
						return err
					}
					match = attempt.BatchID
				}
				if bucketName[0] == 'f' {
					var failure model.FailureDetail
					if err := model.Decode(v, &failure); err != nil {
						return err
					}
					match = failure.BatchID
				}
				if match == batchID {
					keys = append(keys, append([]byte(nil), k...))
				}
				return nil
			}); err != nil {
				return fmt.Errorf("scan %s: %w", bucketName, err)
			}
			for _, key := range keys {
				if err := bucket.Delete(key); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
