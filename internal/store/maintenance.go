package store

import (
	"fmt"

	"go.etcd.io/bbolt"
	"ticketdesk/internal/model"
)

func (s *Store) Health() error {
	if s.db == nil {
		return fmt.Errorf("store is closed")
	}
	return s.db.View(func(tx *bbolt.Tx) error {
		if tx.Bucket([]byte("batches")) == nil {
			return fmt.Errorf("batches bucket missing")
		}
		return nil
	})
}

func (s *Store) SnapshotCounts() (map[string]int, error) {
	counts := make(map[string]int)
	err := s.db.View(func(tx *bbolt.Tx) error {
		for _, name := range bucketNames {
			bucket := tx.Bucket(name)
			if bucket == nil {
				return fmt.Errorf("bucket %s missing", name)
			}
			count := 0
			err := bucket.ForEach(func(_, _ []byte) error { count++; return nil })
			if err != nil {
				return err
			}
			counts[string(name)] = count
		}
		return nil
	})
	return counts, err
}

func (s *Store) PurgeResolved(before int64) (int, error) {
	removed := 0
	err := s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("failures"))
		keys := make([][]byte, 0)
		err := bucket.ForEach(func(k, v []byte) error {
			var failure model.FailureDetail
			if err := model.Decode(v, &failure); err != nil {
				return err
			}
			if failure.Resolved && before > 0 {
				keys = append(keys, append([]byte(nil), k...))
			}
			return nil
		})
		if err != nil {
			return err
		}
		for _, key := range keys {
			if err := bucket.Delete(key); err != nil {
				return err
			}
			removed++
		}
		return nil
	})
	return removed, err
}
