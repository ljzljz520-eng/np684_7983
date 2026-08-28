package report

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"ticketdesk/internal/model"
)

type Dashboard struct {
	BatchID           string         `json:"batch_id"`
	Status            string         `json:"status"`
	Progress          int            `json:"progress"`
	Succeeded         int            `json:"succeeded"`
	Failed            int            `json:"failed"`
	Pending           int            `json:"pending"`
	FailureCategories map[string]int `json:"failure_categories"`
}

func BuildDashboard(summary model.BatchSummary) Dashboard {
	categories := make(map[string]int)
	for _, failure := range summary.Failures {
		categories[failure.Category]++
	}
	return Dashboard{BatchID: summary.Batch.ID, Status: summary.Batch.Status, Progress: summary.Percent, Succeeded: summary.Batch.Succeeded, Failed: summary.Batch.Failed, Pending: summary.Pending, FailureCategories: categories}
}

func RenderText(summary model.BatchSummary) string {
	lines := []string{fmt.Sprintf("Batch %s", summary.Batch.ID), fmt.Sprintf("Status: %s", summary.Batch.Status), fmt.Sprintf("Progress: %d%%", summary.Percent), fmt.Sprintf("Succeeded: %d", summary.Batch.Succeeded), fmt.Sprintf("Failed: %d", summary.Batch.Failed), fmt.Sprintf("Pending: %d", summary.Pending)}
	if len(summary.Failures) > 0 {
		lines = append(lines, "Failures:")
	}
	failures := append([]model.FailureDetail(nil), summary.Failures...)
	sort.Slice(failures, func(i, j int) bool { return failures[i].Code < failures[j].Code })
	for _, failure := range failures {
		lines = append(lines, fmt.Sprintf("- %s: %s", failure.Code, failure.Message))
	}
	return strings.Join(lines, "\n")
}

func MarshalDashboard(summary model.BatchSummary) ([]byte, error) {
	return json.Marshal(BuildDashboard(summary))
}

func FilterFailures(summary model.BatchSummary, category string) []model.FailureDetail {
	result := make([]model.FailureDetail, 0)
	for _, failure := range summary.Failures {
		if category == "" || failure.Category == category {
			result = append(result, failure)
		}
	}
	return result
}
