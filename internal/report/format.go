package report

import (
	"fmt"
	"strings"
	"ticketdesk/internal/model"
)

func FormatCodeTable(codes []model.TicketCode) string {
	lines := []string{"CODE | STATE | HOLDER", "---- | ----- | ------"}
	for _, code := range model.SortCodes(codes) {
		lines = append(lines, fmt.Sprintf("%s | %s | %s", code.Code, code.State, code.Holder))
	}
	return strings.Join(lines, "\n")
}

func FormatFailureTable(failures []model.FailureDetail) string {
	lines := []string{"CODE | CATEGORY | MESSAGE", "---- | -------- | -------"}
	for _, failure := range SortFailures(failures) {
		lines = append(lines, fmt.Sprintf("%s | %s | %s", failure.Code, failure.Category, failure.Message))
	}
	return strings.Join(lines, "\n")
}

func FormatTrendTable(trends []Trend) string {
	lines := []string{"OUTCOME | COUNT | SHARE", "------- | ----- | -----"}
	for _, trend := range trends {
		lines = append(lines, fmt.Sprintf("%s | %d | %d%%", trend.Outcome, trend.Count, trend.Share))
	}
	return strings.Join(lines, "\n")
}
