package report

import (
	"bytes"
	"testing"
	"ticketdesk/internal/model"
)

func TestDashboardAndCSV(t *testing.T) {
	summary := model.BatchSummary{Batch: model.TicketBatch{ID: "r", Status: model.BatchFailed, Processed: 2, Total: 2, Failed: 1}, Percent: 100, Failures: []model.FailureDetail{{BatchID: "r", Code: "BAD123", Category: "format", Message: "bad", Retryable: true}}}
	dashboard := BuildDashboard(summary)
	if dashboard.FailureCategories["format"] != 1 {
		t.Fatal(dashboard)
	}
	var output bytes.Buffer
	if err := WriteFailureCSV(&output, summary.Failures); err != nil {
		t.Fatal(err)
	}
	if output.Len() == 0 {
		t.Fatal("empty csv")
	}
}
