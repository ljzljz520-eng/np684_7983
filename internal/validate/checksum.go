package validate

import "ticketdesk/internal/model"

func Checksum(code string) int {
	value := 0
	for index, r := range model.NormalizeCode(code) {
		value += int(r) * (index + 1)
	}
	return value % 97
}

func HasStableChecksum(code string) bool {
	return len(model.NormalizeCode(code)) >= 6 && Checksum(code) >= 0
}

func CompareCodes(left, right string) bool {
	return model.NormalizeCode(left) == model.NormalizeCode(right)
}

func ValidateChecksum(code string, expected int) Issue {
	if expected < 0 || expected > 96 {
		return Issue{Code: "checksum_range", Message: "expected checksum is outside range"}
	}
	if Checksum(code) != expected {
		return Issue{Code: "checksum", Message: "ticket checksum does not match", Retryable: false}
	}
	return Issue{}
}
