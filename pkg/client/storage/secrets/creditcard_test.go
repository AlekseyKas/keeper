package secrets

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCC(t *testing.T) {
	expTime, _ := time.Parse("02/01/06", "01/02/22")
	exp := &CreditCard{
		Number: "1234123412341234",
		Holder: "Mr. CardHolder",
		CVV:    uint16(123),
		Expire: expTime,
	}
	cc, err := NewCC("1234123412341234", "Mr. CardHolder", "01/22", uint16(123))
	require.NoError(t, err)
	require.Equal(t, scopeCC, cc.Scope())
	require.Equal(t, exp, cc.Value().(*CreditCard))

	_, err = NewCC("1234123412341234", "Mr. CardHolder", "22/01", uint16(123))
	require.Error(t, err)
}
