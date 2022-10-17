package storage

import (
	"sync"
	"testing"

	"github.com/AlekseyKas/keeper/pkg/client/storage/secrets"
	"github.com/AlekseyKas/keeper/pkg/helpers"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
	"github.com/stretchr/testify/require"
)

func TestClientStorage(t *testing.T) {
	k := New()
	up := secrets.NewUserPass("user", "password")
	otp, err := secrets.NewOTP("TOTP", "https://localhost.ltd", "SeCrEt", "1234", "abcd")
	require.NoError(t, err)
	cc, err := secrets.NewCC("1234123412341234", "Mr.CardHolder", "01/23", 123)
	require.NoError(t, err)
	txt := secrets.NewText("this is secret text")
	bin := secrets.NewBinary([]byte("this is secret binary"))
	type test struct {
		name        string
		description string
		secret      Secret
	}
	tests := []test{
		{
			name:        "user-pass",
			description: "test of user password",
			secret:      up,
		},
		{
			name:        "otp",
			description: "test of otp",
			secret:      otp,
		},
		{
			name:        "creditcard",
			description: "test of creditcard",
			secret:      cc,
		},
		{
			name:        "anytext",
			description: "test of any text",
			secret:      txt,
		},
		{
			name:        "anybinary",
			description: "test of any binary",
			secret:      bin,
		},
	}
	round := 0
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			round++
			k.InsertSecret(tt.name, tt.description, tt.secret)
			list := k.ListSecrets()
			require.Equal(t, round, len(list))
			val := k.GetSecret(tt.name)
			switch val.Scope() {
			case "user-password":
				require.Equal(t, tt.secret, val.Value().(*secrets.UserPass))
			case "otp":
				require.Equal(t, tt.secret, val.Value().(*secrets.OTP))
			case "creditcard":
				require.Equal(t, tt.secret, val.Value().(*secrets.CreditCard))
			case "anytext":
				require.Equal(t, tt.secret, val.Value().(*secrets.AnyText))
			case "anybinary":
				require.Equal(t, tt.secret, val.Value().(*secrets.AnyBinary))
			}
		})
	}
	t.Run("test duplicate secret", func(t *testing.T) {
		k.InsertSecret("anytext", "test of any text", txt)
		list := k.ListSecrets()
		require.Equal(t, round, len(list))
	})
	t.Run("get exist secret", func(t *testing.T) {
		secret := k.GetSecret("anytext")
		require.Equal(t, txt, secret.Value().(*secrets.AnyText))
	})
	t.Run("get not exist secret", func(t *testing.T) {
		secret := k.GetSecret("text")
		require.Nil(t, secret)
	})
	t.Run("delete exist secret", func(t *testing.T) {
		k.DeleteSecret("anytext")
		list := k.ListSecrets()
		round--
		require.Equal(t, round, len(list))
	})
	t.Run("delete not exist secret", func(t *testing.T) {
		k.DeleteSecret("text")
		list := k.ListSecrets()
		require.Equal(t, round, len(list))
	})
	saved := []byte{}
	t.Run("test saving of secrets", func(t *testing.T) {
		saved, err = k.Save()
		require.NoError(t, err)
		require.NotEmpty(t, saved)
		require.Equal(t, k.HashSum(), helpers.GenHash(saved))

	})
	t.Run("test reenrypt", func(t *testing.T) {
		_, err := k.ReEncrypt()
		require.NoError(t, err)
	})

	t.Run("test reload", func(t *testing.T) {
		_, err := k.Save()
		require.Error(t, err)
		s := secrets.NewText("new text secret")
		k.InsertSecret("bla-bla", "bla", s)
		b, err := k.Save()
		require.NoError(t, err)
		kk := &Keeper{
			secrets: make([]Value, 0),
			logger:  zerolog.New().WithPrefix("storage"),
			rwMutex: new(sync.RWMutex),
			hashsum: "",
		}
		err = kk.Load(b)
		require.NoError(t, err)
		err = kk.Load(b)
		require.Error(t, err)
		err = kk.Load(nil)
		require.NoError(t, err)
	})
}
