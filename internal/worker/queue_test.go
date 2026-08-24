package worker

import (
	"path/filepath"
	"testing"
	"ticketdesk/internal/batch"
	"ticketdesk/internal/store"
)

func TestQueueProcessesTask(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := batch.NewService(db)
	b, _, err := s.RegisterBatch("queue", "gate", "op", []string{"ABC123"})
	if err != nil {
		t.Fatal(err)
	}
	tasks, err := s.QueueBatch(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	q := NewQueue(s, 1)
	q.Start()
	defer q.Stop()
	if err := q.Submit(tasks); err != nil {
		t.Fatal(err)
	}
	for range tasks {
		select {
		case <-q.Results():
		case err := <-q.Errors():
			t.Fatal(err)
		}
	}
}
