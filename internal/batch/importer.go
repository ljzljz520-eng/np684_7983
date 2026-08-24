package batch

import (
	"io"
	"ticketdesk/internal/model"
	"ticketdesk/internal/validate"
)

func (s *Service) RegisterFromReader(id, source, operator string, reader io.Reader) (model.TicketBatch, []validate.Issue, error) {
	raw, issues, err := validate.ParseCodeLines(reader)
	if err != nil {
		return model.TicketBatch{}, nil, err
	}
	if len(issues) > 0 {
		return model.TicketBatch{ID: id, Source: source, CreatedBy: operator}, issues, nil
	}
	return s.RegisterBatch(id, source, operator, raw)
}

func (s *Service) ValidateOnly(rawCodes []string) []validate.Issue {
	batch := model.TicketBatch{ID: "preview", Source: "preview", Total: len(rawCodes)}
	codes := make([]model.TicketCode, 0, len(rawCodes))
	for _, raw := range rawCodes {
		codes = append(codes, model.TicketCode{BatchID: batch.ID, Code: model.NormalizeCode(raw), State: model.CodePending})
	}
	return s.validator.ValidateBatch(batch, codes)
}
