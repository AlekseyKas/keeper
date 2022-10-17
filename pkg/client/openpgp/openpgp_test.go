package openpgp

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/AlekseyKas/keeper/pkg/logging"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
	"github.com/denisbrodbeck/machineid"
	"github.com/stretchr/testify/require"
)

const (
	pubKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GopenPGP 2.4.10
Comment: https://gopenpgp.org

xjMEYyhjxhYJKwYBBAHaRw8BAQdABBGMYkg//P2HBC0JCUbjqhJy7I6X9RFm4XVd
dKDNqAzNMWFjNWM2YjEwZDgyYjRkYmVhZGQ1OTdiZWYwMTA1OTcxIDx0ZXN0QGVt
YWlsLmNvbT7CjAQTFggAPgUCYyhjxgmQjNtOjc05I/QWIQT+HnsCjM5BpqRmWPmM
206NzTkj9AIbAwIeAQIZAQMLCQcCFQgDFgACAiIBAABvtgEAk/WEMtwHzJhobY1J
zI8BlwEWcGimU4tTZlRHny1qaEoBAKU7nA+fK6ZBscwqkUGT5OBHGAnVatOF3ioT
CFU6cSANzjgEYyhjxhIKKwYBBAGXVQEFAQEHQFD9Bv7VAt2IOfO2DTRZDO5lZPad
XBUdBd/6FXk0F1dUAwEKCcJ4BBgWCAAqBQJjKGPGCZCM206NzTkj9BYhBP4eewKM
zkGmpGZY+YzbTo3NOSP0AhsMAADziAD/eKxXpURDFFGW2nH386BR0fz7Y3shDZmA
CoxeB8h3ohgA/RtSpyonZVmeAhLWmJT2zKlgYERG8g11iYjeo48itewK
=IyYM
-----END PGP PUBLIC KEY BLOCK-----`

	privkey = `-----BEGIN PGP PRIVATE KEY BLOCK-----
Version: GopenPGP 2.4.10
Comment: https://gopenpgp.org

xYYEYyro8BYJKwYBBAHaRw8BAQdAvpsXVpSKOlQPdw/oVZGvJrF7zisGZhJGb8jC
Ctlqji/+CQMIm+HROP8dhSlgaCPqmvWK3o1zFidzgCLW7jz4G3Fjmv2yQbyO6AC7
Ak3eW320T3VLYN7YDqJRJbRLCJCgzt1ff9jCjr4g8QyUT6wVcctkfc0xYWM1YzZi
MTBkODJiNGRiZWFkZDU5N2JlZjAxMDU5NzEgPHVzZXJAZW1haWwuY29tPsKMBBMW
CAA+BQJjKujwCZDgMsZsFNDkFBYhBBG1gnZ8+VJvsceu4uAyxmwU0OQUAhsDAh4B
AhkBAwsJBwIVCAMWAAICIgEAAEROAQDlWO7DgMODhcxzu2c/WYFfhr8qQ34ZTSa7
ZTO09EOMfQEAyGpfmBpMdkWfllTjr/KPoJw4lNQ63+Kb5VMB0pz9KwPHiwRjKujw
EgorBgEEAZdVAQUBAQdACBIDWwnJ7ZmFGd1G/uQ7ThKtXwOdG32qyIIQP8y3y0YD
AQoJ/gkDCCuHI1pwIahgYGLGD/nkrPxi7oIE9lZziu0w1d2q2vNK1a0FZ/px3Ixo
iysRKFMP3DoQGTyY8TXnfkXeejgmVMdqc3IjhwKxRtzex0dcAZfCeAQYFggAKgUC
Yyro8AmQ4DLGbBTQ5BQWIQQRtYJ2fPlSb7HHruLgMsZsFNDkFAIbDAAA8ekA/iRC
hPs8Mb5lncbgC9b3brnX2ZTBjXiLpBWVyCCf5EfSAP0Skg4EjvLRIxnwJ5YNLvus
wlLvJXQkM2ZQXKOUQYWBAw==
=V+1n
-----END PGP PRIVATE KEY BLOCK-----`

	encrypted = `-----BEGIN PGP MESSAGE-----
Version: GopenPGP 2.4.10
Comment: https://gopenpgp.org

wV4DElBhMCk5W8ASAQdA8zrgeJLiw5ebj2cFhedRwX+NQDe7mTPbkaqETU7SkHww
8T4Wv9pUvX02mStPGSU14lCvHKp1mQJl6QVYDyCCjMEwew1CBJ0RG5nvuJH0dHd6
0sCAAfqrxO4fjDM0jL4MFp5awG9L2B2WPm+qv1wbQBOncje0ws/Yxssnl+3sViy0
4JlOoB9gmWQKee4ICdxcdgm+2jxwcgM1h/bSLOCUMsTEWjS1qGgSYDCJvm2AYy0u
pK0ek6GZ7SEffLXPplhxjQ3SnLd79TJbRnAruTI3GNRtPnNJTzUGmR+jVTZXucF3
LcobhdAHTyGvsmptNzv8Cuw6dBkvPE1mag2Po/Ib5yV+pcsnsXLEyLUn/nbCu25A
9KR/9g1UfMCH49784tgjMGK8I6M6CRE7usi9+OKt7qXeFJk7sjYDXITP9TAagSWj
AEyVpCTUKzFmPFnEgp9NIzON06oBxbxsBrSsVOe0y5fE5XIEK4ITKNxWNJXRbcwY
/VbGO/YyM+kDMGVtBlUvr39NRrCnmaBLLvqA2cpTA5dysKg=
=oaoI
-----END PGP MESSAGE-----
`
)

func clean(t *testing.T) {
	content, err := os.ReadDir("./openpgp")
	require.NoError(t, err)
	for _, file := range content {
		path := fmt.Sprintf("./openpgp/%s", file.Name())
		err := os.Remove(path)
		require.NoError(t, err)
		log.Debugf("%s deleted", nil, path)
	}
	err = os.RemoveAll("./openpgp")
	require.NoError(t, err)
}

func TestGeneratePair(t *testing.T) {
	defer clean(t)
	type args struct {
		name       string
		email      string
		passphrase string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "genereate openpgp pair",
			args: args{
				name:       "tester",
				email:      "tester@test.com",
				passphrase: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zerolog.New().SetLevel(logging.TraceLevel)
			os.Setenv("KEEPER_PGP_PASSPHRASE", tt.args.passphrase)
			defer os.Unsetenv("KEEPER_PGP_PASSPHRASE")
			p, err := New()
			require.NoError(t, err)
			err = p.GeneratePair()
			require.NoError(t, err)
			log.Debugf("user %s pair created and tested", nil, tt.args.name)
		})
	}
}

func TestGetPubFromRing(t *testing.T) {
	defer clean(t)
	os.Setenv("KEEPER_PGP_PASSPHRASE", "test")
	defer os.Unsetenv("KEEPER_PGP_PASSPHRASE")
	p, err := New()
	require.NoError(t, err)
	err = p.GeneratePair()
	require.NoError(t, err)
	id, err := machineid.ID()
	require.NoError(t, err)
	f, err := os.Open(fmt.Sprintf("./openpgp/%s.pub", id))
	require.NoError(t, err)
	pub, err := io.ReadAll(f)
	require.NoError(t, err)
	got := p.GetPublicKey()
	require.Equal(t, string(pub), got)

}

func TestGenerateEncryptionAndDecryption(t *testing.T) {
	defer clean(t)
	os.Setenv("KEEPER_PGP_PASSPHRASE", "test")
	defer os.Unsetenv("KEEPER_PGP_PASSPHRASE")
	p, err := New()
	require.NoError(t, err)
	err = p.GeneratePair()
	require.NoError(t, err)
	testdata := []byte("very secret")
	encrypted, err := p.EncryptWithKeys(testdata)
	require.NoError(t, err)
	decrypted, err := p.DecryptWithKeys(encrypted)
	require.NoError(t, err)
	require.Equal(t, testdata, decrypted)
}

func TestReloadPublicKeys(t *testing.T) {
	defer clean(t)
	os.Setenv("KEEPER_PGP_PASSPHRASE", "test")
	defer os.Unsetenv("KEEPER_PGP_PASSPHRASE")
	p, err := New()
	require.NoError(t, err)
	err = p.GeneratePair()
	require.NoError(t, err)
	pp := []string{pubKey}
	err = p.ReloadPublicKeys(pp)
	require.NoError(t, err)

}

func TestAddPrivateKey(t *testing.T) {
	defer clean(t)
	os.Setenv("KEEPER_PGP_PASSPHRASE", "test")
	defer os.Unsetenv("KEEPER_PGP_PASSPHRASE")
	p, err := New()
	require.NoError(t, err)
	err = p.GeneratePair()
	require.NoError(t, err)
	err = p.AddPrivateKey([]byte(privkey))
	require.NoError(t, err)

}

func TestReadFolder(t *testing.T) {
	defer clean(t)
	os.Setenv("KEEPER_PGP_PASSPHRASE", "test")
	defer os.Unsetenv("KEEPER_PGP_PASSPHRASE")
	p, err := New()
	require.NoError(t, err)
	err = p.GeneratePair()
	require.NoError(t, err)
	err = p.ReadFolder("494477e0d52f4d3faf9ce15cef5addcf")
	require.NoError(t, err)

}

func TestEncryptEmpty(t *testing.T) {
	defer clean(t)
	os.Setenv("KEEPER_PGP_PASSPHRASE", "test")
	defer os.Unsetenv("KEEPER_PGP_PASSPHRASE")
	p, err := New()
	require.NoError(t, err)
	err = p.GeneratePair()
	require.NoError(t, err)
	_, err = p.EncryptWithKeys([]byte{})
	require.Error(t, err)

}

func TestDecryptEmpty(t *testing.T) {
	defer clean(t)
	os.Setenv("KEEPER_PGP_PASSPHRASE", "test")
	defer os.Unsetenv("KEEPER_PGP_PASSPHRASE")
	p, err := New()
	require.NoError(t, err)
	err = p.GeneratePair()
	require.NoError(t, err)
	_, err = p.DecryptWithKeys([]byte{})
	require.Error(t, err)

}

func TestAddPublic(t *testing.T) {
	defer clean(t)
	os.Setenv("KEEPER_PGP_PASSPHRASE", "test")
	defer os.Unsetenv("KEEPER_PGP_PASSPHRASE")
	p, err := New()
	require.NoError(t, err)
	err = p.GeneratePair()
	require.NoError(t, err)
	err = p.AddPublicKey([]byte(pubKey))
	require.NoError(t, err)

}

// func TestStream(t *testing.T) {
// 	p, err := New("passphrase")
// 	require.NoError(t, err)
// 	err = p.GeneratePair("test-name", "test@email.com")
// 	require.NoError(t, err)
// 	msg := []byte("testz<fdbafrbarfbadsfbvzfdbzdnbzd v<sdrfv<s")
// 	t.Run("compressed encryption", func(t *testing.T) {
// 		f, err := p.NewPGPStream("test.gpg")
// 		require.NoError(t, err)
// 		fmt.Fprint(f.ReadWriter(), string(msg))
// 		err = f.Close()
// 		require.NoError(t, err)
// 	})
// 	t.Run("compressed decryption", func(t *testing.T) {
// 		f, err := p.NewPGPStream("test.gpg")
// 		require.NoError(t, err)
// 		t.Log(f.buf.Bytes())
// 		data, err := io.ReadAll(f.ReadWriter())
// 		require.NoError(t, err)
// 		require.Equal(t, msg, data)
// 		err = f.Close()
// 		require.NoError(t, err)
// 	})

// }

func TestG(t *testing.T) {
	defer clean(t)
	p, err := emptyKeyring()
	require.NoError(t, err)
	require.NotNil(t, p)
	p.passphrase = []byte("test")
	err = p.ReadFolder("ttt")
	require.Error(t, err)
	err = p.AddPrivateKey([]byte(privkey))
	require.NoError(t, err)
	err = p.AddPublicKey([]byte(pubKey))
	require.NoError(t, err)

	err = p.AddPrivateKey([]byte{})
	require.Error(t, err)
	err = p.AddPublicKey([]byte{})
	require.Error(t, err)

	err = p.ReloadPublicKeys([]string{"test"})
	require.Error(t, err)

	_, err = p.DecryptWithKeys([]byte("data"))
	require.Error(t, err)

	_, err = p.DecryptWithKeys([]byte(encrypted))
	require.Error(t, err)
	err = os.MkdirAll("./openpgp", 0777)
	require.NoError(t, err)
	f, err := os.OpenFile("./openpgp/test.gpg", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	require.NoError(t, err)
	_, err = fmt.Fprint(f, "test")
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	err = p.ReadFolder("test")
	require.Error(t, err)

	p.passphrase = []byte("123")
	err = p.AddPrivateKey([]byte(privkey))
	require.Error(t, err)
}
