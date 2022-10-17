package secrets

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnyText(t *testing.T) {
	txt := "test"
	tt := NewText(txt)
	require.Equal(t, scopeAnyText, tt.Scope())
	require.Equal(t, txt, tt.Value().(*AnyText).Text)
}
