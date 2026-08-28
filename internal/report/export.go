package report

import (
	"encoding/csv"
	"io"
	"strconv"

	"ticketdesk/internal/model"
)

func WriteFailureCSV(w io.Writer, failures []model.FailureDetail) error {
	c := csv.NewWriter(w)
	if err := c.Write([]string{"batch_id", "code", "category", "message", "retryable", "resolved"}); err != nil {
		return err
	}
	for _, failure := range failures {
		if err := c.Write([]string{failure.BatchID, failure.Code, failure.Category, failure.Message, strconv.FormatBool(failure.Retryable), strconv.FormatBool(failure.Resolved)}); err != nil {
			return err
		}
	}
	c.Flush()
	return c.Error()
}
