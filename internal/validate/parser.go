package validate

import (
	"bufio"
	"io"
	"strings"

	"ticketdesk/internal/model"
)

func ParseCodeLines(r io.Reader) ([]string, []Issue, error) {
	scanner := bufio.NewScanner(r)
	codes := make([]string, 0)
	issues := make([]Issue, 0)
	line := 0
	for scanner.Scan() {
		line++
		value := strings.TrimSpace(scanner.Text())
		if value == "" {
			continue
		}
		if len(value) > 64 {
			issues = append(issues, Issue{Code: "line_length", Message: "line exceeds import limit"})
			continue
		}
		codes = append(codes, model.NormalizeCode(value))
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	if line == 0 {
		issues = append(issues, Issue{Code: "empty_file", Message: "input has no lines"})
	}
	return codes, issues, nil
}

func DistinctCodes(codes []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(codes))
	for _, code := range codes {
		normal := model.NormalizeCode(code)
		if !seen[normal] {
			seen[normal] = true
			result = append(result, normal)
		}
	}
	return result
}
