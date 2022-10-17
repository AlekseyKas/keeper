package storage

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/AlekseyKas/keeper/pkg/logging"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
	"github.com/AlekseyKas/keeper/pkg/models"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
)

// THIS IS INTEGRATION TESTING
// FOR CORRECT TEST PLEASE DELPOY POSTGRES DATABASE SERVER

const (
	testDSN = `postgresql://keeper:keeper@127.0.0.1:5432/keeper`
)

func startTest(t *testing.T) *psqlStorage {
	s := &psqlStorage{
		logger: zerolog.New().WithPrefix("storage-testing"),
	}
	db, err := pgxpool.Connect(context.Background(), testDSN)
	require.NoError(t, err)
	s.db = db
	return s
}

func endTest(t *testing.T, s *psqlStorage) {
	s.DeleteUser("test", net.ParseIP("127.0.0.1"))
	err := s.Close()
	require.NoError(t, err)
}

func TestNew(t *testing.T) {
	zerolog.New().SetLevel(logging.TraceLevel)

	type test struct {
		name string
		dsn  string
		want bool
	}
	tests := []test{
		{
			name: "valid dsn",
			dsn:  testDSN,
			want: false,
		},
		// {
		// 	name: "invalid dsn",
		// 	dsn:  "postgresql://keeper:keeper@192.168.0.1:5432/keeper",
		// 	want: true,
		// },
	}
	for _, tt := range tests {
		var got bool
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("DSN_ADDRESS", tt.dsn)
			_, err := New()
			if err != nil {
				got = true
			}
			require.Equal(t, tt.want, got)
			os.Unsetenv("DSN_ADDRESS")

		})
	}
}

func TestPing(t *testing.T) {
	s := startTest(t)
	err := s.Ping()
	require.NoError(t, err)
	err = s.Close()
	require.NoError(t, err)
	err = s.Ping()
	require.Error(t, err)
}

func TestSignUp(t *testing.T) {
	s := startTest(t)
	defer endTest(t, s)
	err := s.SignUp("test", "password", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.SignUp("test", "password", net.ParseIP("127.0.0.1"))
	require.Error(t, err)

}

func TestSignIn(t *testing.T) {
	s := startTest(t)
	defer endTest(t, s)
	err := s.SignUp("test", "password", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	u := models.User{
		Username:  "test",
		Password:  "password",
		PublicKey: "PublicKey",
	}
	err = s.SignIn(u, net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	u1 := models.User{
		Username:  "test1",
		Password:  "password1",
		PublicKey: "PublicKey1",
	}
	err = s.SignIn(u1, net.ParseIP("127.0.0.1"))
	require.Error(t, err)
	u2 := models.User{
		Username:  "",
		Password:  "password1",
		PublicKey: "PublicKey1",
	}
	err = s.SignIn(u2, net.ParseIP("127.0.0.1"))
	require.Error(t, err)
	u3 := models.User{
		Username:  "test",
		Password:  "",
		PublicKey: "PublicKey",
	}
	err = s.SignIn(u3, net.ParseIP("127.0.0.1"))
	require.Error(t, err)
	u4 := models.User{
		Username:  "",
		Password:  "",
		PublicKey: "PublicKey",
	}
	err = s.SignIn(u4, net.ParseIP("127.0.0.1"))
	require.Error(t, err)
	u5 := models.User{
		Username:  "11111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111",
		Password:  "2222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222aaaaaaaaaaaaaaaaaaaaaaaaaaaAAAAAAAAAAAAAAAAAAAAAAaaaaaAAAAAAAAAAAAAAAAAAAAaaaaAAAAAaaaAAAAAaaaaAAAAaaaaAAAAaaaaAAAAa",
		PublicKey: "PublicKey",
	}
	err = s.SignIn(u5, net.ParseIP("127.0.0.1"))
	require.Error(t, err)
	u6 := models.User{
		Username:  "test",
		Password:  "2222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222aaaaaaaaaaaaaaaaaaaaaaaaaaaAAAAAAAAAAAAAAAAAAAAAAaaaaaAAAAAAAAAAAAAAAAAAAAaaaaAAAAAaaaAAAAAaaaaAAAAaaaaAAAAaaaaAAAAa",
		PublicKey: "PublicKey",
	}
	err = s.SignIn(u6, net.ParseIP("127.0.0.1"))
	require.Error(t, err)
}

func TestDeleteUSer(t *testing.T) {
	s := startTest(t)
	defer endTest(t, s)
	err := s.SignUp("test", "password", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.DeleteUser("test", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.DeleteUser("test", net.ParseIP("127.0.0.1"))
	require.Error(t, err)
}

func TestPush(t *testing.T) {
	s := startTest(t)
	defer endTest(t, s)
	err := s.SignUp("test", "password", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.Push("test", "1234", "secret data", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.Push("test1", "1234", "secret data", net.ParseIP("127.0.0.1"))
	require.Error(t, err)

}

func TestPull(t *testing.T) {
	s := startTest(t)
	defer endTest(t, s)
	err := s.SignUp("test", "password", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.Push("test", "1234", "secret data", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	b, err := s.Pull("test", "1234", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	require.Equal(t, []byte("secret data"), b)
	_, err = s.Pull("test", "4321", net.ParseIP("127.0.0.1"))
	require.Error(t, err)
	_, err = s.Pull("test1", "1234", net.ParseIP("127.0.0.1"))
	require.Error(t, err)
}

func TestVersions(t *testing.T) {
	s := startTest(t)
	defer endTest(t, s)
	err := s.SignUp("test", "password", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.Push("test", "1234", "secret data1", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.Push("test", "4321", "secret data2", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	v, err := s.Versions("test", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	require.Equal(t, 2, len(v))
	v, err = s.Versions("test1", net.ParseIP("127.0.0.1"))
	require.Error(t, err)
}

func TestSaveLog(t *testing.T) {
	s := startTest(t)
	defer endTest(t, s)
	err := s.SaveLog("test-save", "test func", "no check sum", net.ParseIP("127.0.0.1"), time.Now())
	require.NoError(t, err)
}

func TestGetLog(t *testing.T) {
	s := startTest(t)
	defer endTest(t, s)
	logs, err := s.GetLog("test-save", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	require.NotEmpty(t, logs)
}

func TestListPGP(t *testing.T) {
	s := startTest(t)
	defer endTest(t, s)
	err := s.SignUp("test", "password", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.AddPGP("test", "PublicKey", true, net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.AddPGP("test", "PublicKey1", false, net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.AddPGP("test", "PublicKey2", false, net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.RevokePGP("test", "PublicKey2", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.AddPGP("test1", "PublicKey", true, net.ParseIP("127.0.0.1"))
	require.Error(t, err)
	list, err := s.ListPGP("test", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	require.NotEmpty(t, list)
	require.Equal(t, 2, len(list))
	_, err = s.ListPGP("bla", net.ParseIP("127.0.0.1"))
	require.Error(t, err)
}

func TestRevokePGP(t *testing.T) {
	s := startTest(t)
	defer endTest(t, s)
	err := s.SignUp("test", "password", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.AddPGP("test", "PublicKey", true, net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.RevokePGP("test", "PublicKey", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.RevokePGP("test1", "PublicKey", net.ParseIP("127.0.0.1"))
	require.Error(t, err)
	err = s.RevokePGP("test", "PublicKey1", net.ParseIP("127.0.0.1"))
	require.Error(t, err)
	err = s.RevokePGP("test1", "PublicKey1", net.ParseIP("127.0.0.1"))
	require.Error(t, err)
}

func TestConfirmPGP(t *testing.T) {
	s := startTest(t)
	defer endTest(t, s)
	err := s.SignUp("test", "password", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.AddPGP("test", "PublicKey", false, net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.ConfirmPGP("test", "PublicKey", net.ParseIP("127.0.0.1"))
	require.NoError(t, err)
	err = s.ConfirmPGP("test1", "PublicKey", net.ParseIP("127.0.0.1"))
	require.Error(t, err)
	err = s.ConfirmPGP("test", "PublicKey1", net.ParseIP("127.0.0.1"))
	require.Error(t, err)
	err = s.ConfirmPGP("test1", "PublickKey1", net.ParseIP("127.0.0.1"))
	require.Error(t, err)

}

func TestClose(t *testing.T) {
	s := startTest(t)
	err := s.Close()
	require.NoError(t, err)
}
