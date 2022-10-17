package secrets

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOTP(t *testing.T) {
	exp := &OTP{
		Method:        "TOTP",
		Issuer:        "https://localhost.ltd",
		Secret:        "owivneri39nv3",
		AccountName:   "username",
		RecoveryCodes: []string{"1234", "abcd"},
	}
	tt, err := NewOTP("TOTP", "https://localhost.ltd", "owivneri39nv3", "username", "1234", "abcd")
	require.NoError(t, err)
	require.Equal(t, scopeOTP, tt.Scope())
	require.Equal(t, exp, tt.Value().(*OTP))
}

func TestOTPWrong(t *testing.T) {
	tt, err := NewOTP("XOTP", "https://localhost.ltd", "owivneri39nv3", "username", "1234", "abcd")
	require.Error(t, err)
	require.Equal(t, scopeOTP, tt.Scope())
	require.Nil(t, tt.Value().(*OTP))
}
