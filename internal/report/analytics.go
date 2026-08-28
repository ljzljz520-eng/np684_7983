package report

import (
	"sort"
	"ticketdesk/internal/model"
)

type Trend struct {
	Outcome string `json:"outcome"`
	Count   int    `json:"count"`
	Share   int    `json:"share"`
}

func BuildTrends(attempts []model.ValidationAttempt) []Trend {
	counts := map[string]int{}
	for _, attempt := range attempts {
		counts[attempt.Outcome]++
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]Trend, 0, len(keys))
	for _, key := range keys {
		share := 0
		if len(attempts) > 0 {
			share = counts[key] * 100 / len(attempts)
		}
		result = append(result, Trend{Outcome: key, Count: counts[key], Share: share})
	}
	return result
}

func SummarizeCodes(codes []model.TicketCode) map[string]int {
	counts := map[string]int{"pending": 0, "consumed": 0, "duplicate": 0, "invalid": 0}
	for _, code := range codes {
		if _, ok := counts[code.State]; ok {
			counts[code.State]++
		} else {
			counts[code.State] = 1
		}
	}
	return counts
}

func RecommendedAction(summary model.BatchSummary) string {
	if summary.Batch.Status == model.BatchPending {
		return "queue batch"
	}
	if summary.Pending > 0 {
		return "resume processing"
	}
	if len(RetryableFailures(summary.Failures)) > 0 {
		return "retry failures"
	}
	if summary.Batch.Failed > 0 {
		return "review failures"
	}
	return "archive batch"
}
