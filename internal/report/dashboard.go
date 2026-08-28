package report

import (
	"fmt"
	"ticketdesk/internal/model"
)

func DashboardLines(summary model.BatchSummary) []string {
	dashboard := BuildDashboard(summary)
	lines := []string{fmt.Sprintf("Batch: %s", dashboard.BatchID), fmt.Sprintf("Status: %s", dashboard.Status), fmt.Sprintf("Progress: %d%%", dashboard.Progress), fmt.Sprintf("Succeeded: %d", dashboard.Succeeded), fmt.Sprintf("Failed: %d", dashboard.Failed), fmt.Sprintf("Pending: %d", dashboard.Pending)}
	for category, count := range dashboard.FailureCategories {
		lines = append(lines, fmt.Sprintf("Failure %s: %d", category, count))
	}
	return lines
}

func GroupFailuresByCode(failures []model.FailureDetail) map[string][]model.FailureDetail {
	groups := make(map[string][]model.FailureDetail)
	for _, failure := range failures {
		groups[failure.Code] = append(groups[failure.Code], failure)
	}
	return groups
}

func HasActionableFailures(failures []model.FailureDetail) bool {
	for _, failure := range failures {
		if failure.Retryable && !failure.Resolved {
			return true
		}
	}
	return false
}
