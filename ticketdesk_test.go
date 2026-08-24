package ticketdesk

import (
	"path/filepath"
	"sync"
	"testing"
	"ticketdesk/internal/batch"
	"ticketdesk/internal/store"
)

func TestTicketCodeConsumedOnlyOnce(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := batch.NewService(db)
	b, issues, err := s.RegisterBatch("concurrent", "gate", "op", []string{"ABC123"})
	if err != nil || len(issues) != 0 {
		t.Fatalf("register err=%v issues=%v", err, issues)
	}
	tasks, err := s.QueueBatch(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	db.SetConsumeBarrier(2)
	var wg sync.WaitGroup
	results := make(chan string, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			attempt, processErr := s.ProcessTask(tasks[0])
			if processErr != nil {
				results <- processErr.Error()
				return
			}
			results <- attempt.Outcome
		}()
	}
	wg.Wait()
	close(results)
	successes, duplicates := 0, 0
	for result := range results {
		if result == "success" {
			successes++
		}
		if result == "duplicate" {
			duplicates++
		}
	}
	if successes != 1 || duplicates != 1 {
		t.Fatalf("successes=%d duplicates=%d", successes, duplicates)
	}
}
