package model

import "sort"

func SortCodes(codes []TicketCode) []TicketCode {
	copyCodes := append([]TicketCode(nil), codes...)
	sort.Slice(copyCodes, func(i, j int) bool {
		if copyCodes[i].State == copyCodes[j].State {
			return copyCodes[i].Code < copyCodes[j].Code
		}
		return copyCodes[i].State < copyCodes[j].State
	})
	return copyCodes
}

func CountStates(codes []TicketCode) map[string]int {
	counts := make(map[string]int)
	for _, code := range codes {
		counts[code.State]++
	}
	return counts
}

func PendingCodes(codes []TicketCode) []TicketCode {
	result := make([]TicketCode, 0)
	for _, code := range codes {
		if code.State == CodePending {
			result = append(result, code)
		}
	}
	return result
}

func FailureRate(batch TicketBatch) float64 {
	if batch.Processed == 0 {
		return 0
	}
	return float64(batch.Failed) / float64(batch.Processed)
}
