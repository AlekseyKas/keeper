package secrets

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserPass(t *testing.T) {
	exp := &UserPass{
		Username: "user",
		Password: "Password",
	}
	tt := NewUserPass("user", "Password")
	require.Equal(t, scopeUserPass, tt.Scope())
	require.Equal(t, exp, tt.Value().(*UserPass))
}
