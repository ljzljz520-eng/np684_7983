package store

import (
	"fmt"
	"go.etcd.io/bbolt"
	"ticketdesk/internal/model"
)

func (s *Store) PutBatchWithCodes(batch model.TicketBatch, codes []model.TicketCode) error {
	if err := batch.Validate(); err != nil {
		return err
	}
	if len(codes) != batch.Total {
		return fmt.Errorf("code count does not match batch")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		batchData, err := model.Encode(batch)
		if err != nil {
			return err
		}
		if err := tx.Bucket([]byte("batches")).Put([]byte(batch.ID), batchData); err != nil {
			return err
		}
		bucket := tx.Bucket([]byte("codes"))
		for _, code := range codes {
			if err := code.Validate(); err != nil {
				return err
			}
			data, err := model.Encode(code)
			if err != nil {
				return err
			}
			if err := bucket.Put([]byte(code.Key()), data); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) ReplaceCodes(batchID string, codes []model.TicketCode) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("codes"))
		keys := make([][]byte, 0)
		if err := bucket.ForEach(func(key, value []byte) error {
			var code model.TicketCode
			if err := model.Decode(value, &code); err != nil {
				return err
			}
			if code.BatchID == batchID {
				keys = append(keys, append([]byte(nil), key...))
			}
			return nil
		}); err != nil {
			return err
		}
		for _, key := range keys {
			if err := bucket.Delete(key); err != nil {
				return err
			}
		}
		for _, code := range codes {
			data, err := model.Encode(code)
			if err != nil {
				return err
			}
			if err := bucket.Put([]byte(code.Key()), data); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) CountCodes(batchID, state string) (int, error) {
	codes, err := s.FindCodes(batchID, state)
	if err != nil {
		return 0, err
	}
	return len(codes), nil
}
