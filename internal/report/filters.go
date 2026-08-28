package report

import (
	"sort"
	"ticketdesk/internal/model"
)

func SortFailures(failures []model.FailureDetail) []model.FailureDetail {
	result := append([]model.FailureDetail(nil), failures...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Category == result[j].Category {
			return result[i].Code < result[j].Code
		}
		return result[i].Category < result[j].Category
	})
	return result
}

func RetryableFailures(failures []model.FailureDetail) []model.FailureDetail {
	result := make([]model.FailureDetail, 0)
	for _, failure := range failures {
		if failure.Retryable && !failure.Resolved {
			result = append(result, failure)
		}
	}
	return result
}
