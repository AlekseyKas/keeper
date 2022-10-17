package secrets

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnyBinary(t *testing.T) {
	b := []byte("test")
	bb := NewBinary(b)
	require.Equal(t, scopeAnyBinary, bb.Scope())
	require.Equal(t, b, bb.Value().(*AnyBinary).Bytes)
}
