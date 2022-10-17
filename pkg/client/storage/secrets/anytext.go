package secrets

const scopeAnyText = "anytext"

// AnyText - type for binary secret
type AnyText struct {
	Text string
}

// NewText - create text secret
func NewText(text string) *AnyText {
	txt := AnyText{
		Text: text,
	}
	return &txt
}

// Scope - secret scope
func (s *AnyText) Scope() string {
	return scopeAnyText
}

// Value - secret value
func (s *AnyText) Value() interface{} {
	return s
}
