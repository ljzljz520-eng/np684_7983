package model

import "strings"

func ValidIdentifier(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	for _, r := range trimmed {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

func CanonicalWorker(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func CopyBatch(batch TicketBatch) TicketBatch { return batch }
