package sender

// SMSSender interface for sending SMS verification codes
type SMSSender interface {
	Send(to, templateID, code string) error
}

// noopSMSSender provides a no-op implementation of SMSSender so Wire can
// resolve it. In production, a real SMS provider should be wired.
type noopSMSSender struct{}

func (noopSMSSender) Send(_, _, _ string) error { return nil }

func NewNoopSMSSender() SMSSender { return noopSMSSender{} }
