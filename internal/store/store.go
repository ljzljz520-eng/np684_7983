package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.etcd.io/bbolt"
	"ticketdesk/internal/model"
)

var (
	ErrNotFound        = errors.New("record not found")
	ErrAlreadyConsumed = errors.New("ticket already consumed")
	ErrInvalidState    = errors.New("invalid state transition")
)

var bucketNames = [][]byte{[]byte("batches"), []byte("codes"), []byte("attempts"), []byte("failures"), []byte("tasks")}

type Store struct {
	db      *bbolt.DB
	mu      sync.RWMutex
	barrier *consumeBarrier
}

type consumeBarrier struct {
	arrived chan struct{}
	release chan struct{}
	parties int
	once    sync.Once
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 2 * time.Second, NoSync: false})
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.initialize(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) initialize() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range bucketNames {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) SetConsumeBarrier(parties int) {
	if parties < 1 {
		parties = 1
	}
	s.mu.Lock()
	s.barrier = &consumeBarrier{arrived: make(chan struct{}, parties), release: make(chan struct{}), parties: parties}
	s.mu.Unlock()
}

func (s *Store) clearBarrier() {
	s.mu.Lock()
	s.barrier = nil
	s.mu.Unlock()
}

func (s *Store) waitAtBarrier() {
	s.mu.RLock()
	b := s.barrier
	s.mu.RUnlock()
	if b == nil {
		return
	}
	b.arrived <- struct{}{}
	if len(b.arrived) == b.parties {
		b.once.Do(func() { close(b.release) })
	}
	<-b.release
}

func (s *Store) PutBatch(batch model.TicketBatch) error {
	if batch.ID == "" {
		return fmt.Errorf("batch id is required")
	}
	if !model.ValidBatchStatus(batch.Status) {
		return fmt.Errorf("unknown batch status %q", batch.Status)
	}
	data, err := model.Encode(batch)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket([]byte("batches")).Put([]byte(batch.ID), data) })
}

func (s *Store) GetBatch(id string) (model.TicketBatch, error) {
	var batch model.TicketBatch
	err := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte("batches")).Get([]byte(id))
		if v == nil {
			return ErrNotFound
		}
		return model.Decode(v, &batch)
	})
	return batch, err
}

func (s *Store) ListBatches() ([]model.TicketBatch, error) {
	result := make([]model.TicketBatch, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("batches")).ForEach(func(_, v []byte) error {
			var b model.TicketBatch
			if err := model.Decode(v, &b); err != nil {
				return err
			}
			result = append(result, b)
			return nil
		})
	})
	return result, err
}

func (s *Store) PutCode(code model.TicketCode) error {
	if code.BatchID == "" || code.Code == "" {
		return fmt.Errorf("batch and code are required")
	}
	if !model.ValidCodeState(code.State) {
		return fmt.Errorf("unknown code state %q", code.State)
	}
	data, err := model.Encode(code)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket([]byte("codes")).Put([]byte(code.Key()), data) })
}

func (s *Store) GetCode(batchID, code string) (model.TicketCode, error) {
	var result model.TicketCode
	err := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte("codes")).Get([]byte(batchID + ":" + code))
		if v == nil {
			return ErrNotFound
		}
		return model.Decode(v, &result)
	})
	return result, err
}

func (s *Store) ListCodes(batchID string) ([]model.TicketCode, error) {
	result := make([]model.TicketCode, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("codes")).ForEach(func(k, v []byte) error {
			if len(batchID) > 0 && len(k) <= len(batchID) || (len(batchID) > 0 && string(k[:len(batchID)]) != batchID) {
				return nil
			}
			var c model.TicketCode
			if err := model.Decode(v, &c); err != nil {
				return err
			}
			result = append(result, c)
			return nil
		})
	})
	return result, err
}

func (s *Store) ConsumeTicketCode(batchID, code, worker string) (model.TicketCode, error) {
	// The barrier lets tests synchronize concurrent consumers so they race on the
	// same ticket; it is a no-op in production (no barrier is configured). It must
	// run before the consume transaction so it never splits the atomic read-check-write.
	s.waitAtBarrier()

	var current model.TicketCode
	key := []byte(batchID + ":" + code)
	err := s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("codes"))
		v := bucket.Get(key)
		if v == nil {
			return ErrNotFound
		}
		if err := model.Decode(v, &current); err != nil {
			return err
		}
		// Re-check state inside the write transaction. bbolt serializes writers, so
		// this read-check-write is atomic: a concurrent consumer that already consumed
		// the ticket will have committed before this transaction begins, and this one
		// observes the updated state instead of clobbering it.
		if current.State != model.CodePending {
			return ErrAlreadyConsumed
		}
		current.State = model.CodeConsumed
		current.Holder = worker
		current.Attempts++
		current.Validated = true
		data, err := model.Encode(current)
		if err != nil {
			return err
		}
		return bucket.Put(key, data)
	})
	return current, err
}

func (s *Store) PutAttempt(attempt model.ValidationAttempt) error {
	data, err := model.Encode(attempt)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket([]byte("attempts")).Put([]byte(attempt.ID), data) })
}

func (s *Store) ListAttempts(batchID string) ([]model.ValidationAttempt, error) {
	result := make([]model.ValidationAttempt, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("attempts")).ForEach(func(_, v []byte) error {
			var a model.ValidationAttempt
			if err := model.Decode(v, &a); err != nil {
				return err
			}
			if batchID == "" || a.BatchID == batchID {
				result = append(result, a)
			}
			return nil
		})
	})
	return result, err
}

func (s *Store) PutFailure(failure model.FailureDetail) error {
	data, err := model.Encode(failure)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket([]byte("failures")).Put([]byte(failure.Key()), data) })
}

func (s *Store) ListFailures(batchID string) ([]model.FailureDetail, error) {
	result := make([]model.FailureDetail, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("failures")).ForEach(func(_, v []byte) error {
			var f model.FailureDetail
			if err := model.Decode(v, &f); err != nil {
				return err
			}
			if batchID == "" || f.BatchID == batchID {
				result = append(result, f)
			}
			return nil
		})
	})
	return result, err
}

func (s *Store) PutTask(task model.WorkerTask) error {
	data, err := model.Encode(task)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket([]byte("tasks")).Put([]byte(task.ID), data) })
}

func (s *Store) UpdateTask(taskID string, update func(*model.WorkerTask) error) (model.WorkerTask, error) {
	var result model.WorkerTask
	err := s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("tasks"))
		value := bucket.Get([]byte(taskID))
		if value == nil {
			return ErrNotFound
		}
		if err := model.Decode(value, &result); err != nil {
			return err
		}
		if err := update(&result); err != nil {
			return err
		}
		data, err := model.Encode(result)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(taskID), data)
	})
	return result, err
}

func (s *Store) ListTasks(batchID string) ([]model.WorkerTask, error) {
	result := make([]model.WorkerTask, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("tasks")).ForEach(func(_, v []byte) error {
			var t model.WorkerTask
			if err := model.Decode(v, &t); err != nil {
				return err
			}
			if batchID == "" || t.BatchID == batchID {
				result = append(result, t)
			}
			return nil
		})
	})
	return result, err
}

func (s *Store) UpdateBatch(batchID string, update func(*model.TicketBatch) error) (model.TicketBatch, error) {
	var result model.TicketBatch
	err := s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("batches"))
		value := bucket.Get([]byte(batchID))
		if value == nil {
			return ErrNotFound
		}
		if err := model.Decode(value, &result); err != nil {
			return err
		}
		if err := update(&result); err != nil {
			return err
		}
		data, err := model.Encode(result)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(batchID), data)
	})
	return result, err
}
