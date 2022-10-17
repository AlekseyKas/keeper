package secrets

const scopeAnyBinary = "anybinary"

// AnyBinary - type for binary secret
type AnyBinary struct {
	Bytes []byte
}

// NewBinary - create binary secret
func NewBinary(b []byte) *AnyBinary {
	bb := AnyBinary{
		Bytes: b,
	}
	return &bb
}

// Scope - secret scope
func (s *AnyBinary) Scope() string {
	return scopeAnyBinary
}

// Value - secret value
func (s *AnyBinary) Value() interface{} {
	return s
}
