package model

type BatchMetrics struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Consumed  int `json:"consumed"`
	Duplicate int `json:"duplicate"`
	Invalid   int `json:"invalid"`
}

func MetricsForCodes(codes []TicketCode) BatchMetrics {
	metrics := BatchMetrics{Total: len(codes)}
	for _, code := range codes {
		switch code.State {
		case CodePending:
			metrics.Pending++
		case CodeConsumed:
			metrics.Consumed++
		case CodeDuplicate:
			metrics.Duplicate++
		case CodeInvalid:
			metrics.Invalid++
		}
	}
	return metrics
}

func (m BatchMetrics) Processed() int { return m.Consumed + m.Duplicate + m.Invalid }
func (m BatchMetrics) Complete() bool { return m.Total > 0 && m.Processed() == m.Total }
func (m BatchMetrics) Percent() int {
	if m.Total == 0 {
		return 0
	}
	return m.Processed() * 100 / m.Total
}
func (m BatchMetrics) Failed() int { return m.Duplicate + m.Invalid }
func (m BatchMetrics) SuccessRate() int {
	if m.Total == 0 {
		return 0
	}
	return m.Consumed * 100 / m.Total
}

func (m BatchMetrics) Status() string {
	if m.Total == 0 {
		return BatchPending
	}
	if !m.Complete() {
		return BatchRunning
	}
	if m.Failed() > 0 {
		return BatchFailed
	}
	return BatchComplete
}
