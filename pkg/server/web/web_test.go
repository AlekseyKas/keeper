package web

import (
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
	"github.com/AlekseyKas/keeper/pkg/models"
	mockStorage "github.com/AlekseyKas/keeper/pkg/server/storage/mock_storage"
	ssl "github.com/AlekseyKas/keeper/pkg/server/tls"
	"github.com/AlekseyKas/keeper/pkg/server/web/static"
)

func clean(t *testing.T) {
	list, err := os.ReadDir("./ssl")
	require.NoError(t, err)
	for _, item := range list {
		err := os.Remove(fmt.Sprintf("./ssl/%s", item.Name()))
		require.NoError(t, err)
	}
	err = os.RemoveAll("./ssl")
	require.NoError(t, err)
	os.Unsetenv("WEB_ADDRESS")
}

func client(t *testing.T) *resty.Client {
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := resty.New()
	client.SetCookieJar(jar)
	client.SetTLSClientConfig(&tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: true,
	})
	return client
}

func login(c *resty.Client) (*resty.Response, error) {
	response, err := c.R().
		SetHeader("Content-Type", "application/json").
		SetBody(models.User{
			Username:  "test",
			Password:  "test",
			PublicKey: "test",
		}).
		Post("api/v1/signin")
	return response, err
}

func TestServer(t *testing.T) {
	rbi, err := rand.Int(rand.Reader, big.NewInt(60000))
	require.NoError(t, err)
	port := int(rbi.Int64())
	if port < 10000 {
		port += 10000
	}
	webbind := fmt.Sprintf("127.0.0.1:%d", port)
	os.Setenv("WEB_ADDRESS", webbind)
	ssl.Prepare()
	defer clean(t)
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	db := mockStorage.NewMockStorage(ctl)
	ip := net.ParseIP("127.0.0.1")
	errDummy := errors.New("dummy error")

	gomock.InOrder(
		//signin
		db.EXPECT().SignIn(models.User{Username: "test", Password: "test", PublicKey: "test"}, ip).Return(nil),
		db.EXPECT().ListPGP("test", ip).Return([]models.PGP{{Date: time.Now(), Publickey: "test", Confirmed: true}}, nil),

		db.EXPECT().SignIn(models.User{Username: "test", Password: "test", PublicKey: "test"}, ip).Return(errDummy),

		db.EXPECT().SignIn(models.User{Username: "test", Password: "test", PublicKey: "test"}, ip).Return(nil),
		db.EXPECT().ListPGP("test", ip).Return([]models.PGP{}, nil),

		db.EXPECT().SignIn(models.User{Username: "test", Password: "test", PublicKey: "test"}, ip).Return(nil),
		db.EXPECT().ListPGP("test", ip).Return(nil, errDummy),

		db.EXPECT().SignIn(models.User{Username: "test", Password: "test", PublicKey: "test"}, ip).Return(nil),
		db.EXPECT().ListPGP("test", ip).Return([]models.PGP{{Date: time.Now(), Publickey: "test", Confirmed: false}}, nil),

		db.EXPECT().SignIn(models.User{Username: "test", Password: "test", PublicKey: "test"}, ip).Return(nil),
		db.EXPECT().ListPGP("test", ip).Return([]models.PGP{{Date: time.Now(), Publickey: "test1", Confirmed: false}}, nil),

		//delete
		db.EXPECT().SignIn(models.User{Username: "test", Password: "test", PublicKey: "test"}, ip).Return(nil),
		db.EXPECT().ListPGP("test", ip).Return([]models.PGP{{Date: time.Now(), Publickey: "test", Confirmed: true}}, nil),
		db.EXPECT().DeleteUser("test", ip).Return(nil),

		//delete
		db.EXPECT().SignIn(models.User{Username: "test", Password: "test", PublicKey: "test"}, ip).Return(nil),
		db.EXPECT().ListPGP("test", ip).Return([]models.PGP{{Date: time.Now(), Publickey: "test", Confirmed: true}}, nil),
		db.EXPECT().DeleteUser("test", ip).Return(errDummy),

		//register
		db.EXPECT().SignUp("test", "test", ip).Return(nil),
		db.EXPECT().AddPGP("test", "test", true, ip).Return(nil),
	)

	wg := new(sync.WaitGroup)
	s := &Server{
		echo: echo.New(),
		log:  zerolog.New().WithPrefix("web-server-test"),
		db:   db,
		sig:  make(chan struct{}),
		msg:  make(map[string]chan models.Message),
	}
	s.applyMiddlewares()
	static.ApplyStatic(s.echo)
	s.createRouter()
	go func() {
		err := s.Start(wg)
		require.NoError(t, err)
	}()
	time.Sleep(2 * time.Second)
	t.Run("test login", func(t *testing.T) {
		c := client(t)
		c.SetBaseURL(fmt.Sprintf("https://%s", webbind))

		type test struct {
			name string
			body models.User

			want bool
			code int
		}

		tests := []test{
			{
				name: "successfull login",
				body: models.User{
					Username:  "test",
					Password:  "test",
					PublicKey: "test",
				},

				want: false,
				code: http.StatusOK,
			},
			{
				name: "fail",
				body: models.User{
					Username:  "test",
					Password:  "test",
					PublicKey: "test",
				},

				want: false,
				code: http.StatusUnauthorized,
			},
			{
				name: "fail",
				body: models.User{
					Username:  "test",
					Password:  "test",
					PublicKey: "test",
				},

				want: false,
				code: http.StatusInternalServerError,
			},
			{
				name: "fail",
				body: models.User{
					Username:  "test",
					Password:  "test",
					PublicKey: "test",
				},

				want: false,
				code: http.StatusInternalServerError,
			},
			{
				name: "fail",
				body: models.User{
					Username:  "test",
					Password:  "test",
					PublicKey: "test",
				},

				want: false,
				code: http.StatusAlreadyReported,
			},
			{
				name: "fail",
				body: models.User{
					Username:  "test",
					Password:  "test",
					PublicKey: "test",
				},

				want: false,
				code: http.StatusForbidden,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				response, _ := c.R().
					SetHeader("Content-Type", "application/json").
					SetBody(tt.body).
					Post("api/v1/signin")
				require.Equal(t, tt.code, response.StatusCode())
				if response.StatusCode() == http.StatusOK {
					require.NotEmpty(t, response.Cookies())
					var token bool
					for _, cookie := range response.Cookies() {
						if cookie.Name == "token" {
							token = true
							require.NotEmpty(t, cookie.Value)
						}
					}
					require.True(t, token)
				}
			})
		}

	})

	t.Run("test remove", func(t *testing.T) {
		c := client(t)
		c.SetBaseURL(fmt.Sprintf("https://%s", webbind))

		type test struct {
			name string

			want bool
			code int
		}

		tests := []test{
			{
				name: "successfull delete",
				want: false,
				code: http.StatusAccepted,
			},
			{
				name: "fail delete",
				want: false,
				code: http.StatusInternalServerError,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r, err := login(c)
				require.Equal(t, http.StatusOK, r.StatusCode())
				require.NoError(t, err)
				response, _ := c.R().
					SetHeader("Content-Type", "application/json").
					Delete("api/v1/keeper/remove")
				require.Equal(t, tt.code, response.StatusCode())
				if response.StatusCode() == http.StatusOK {
					require.NotEmpty(t, response.Cookies())
					var token bool
					for _, cookie := range response.Cookies() {
						if cookie.Name == "token" {
							token = true
							require.NotEmpty(t, cookie.Value)
						}
					}
					require.True(t, token)
				}
			})
		}

	})

	t.Run("test register", func(t *testing.T) {
		c := client(t)
		c.SetBaseURL(fmt.Sprintf("https://%s", webbind))

		type test struct {
			name string
			body models.User

			want bool
			code int
		}

		tests := []test{
			{
				name: "successfull register",
				body: models.User{
					Username:  "test",
					Password:  "test",
					PublicKey: "test",
				},
				want: false,
				code: http.StatusCreated,
			},
			// {
			// 	name: "fail delete",
			// 	want: false,
			// 	code: http.StatusInternalServerError,
			// },
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				response, _ := c.R().
					SetHeader("Content-Type", "application/json").
					SetBody(tt.body).
					Post("api/v1/signup")
				require.Equal(t, tt.code, response.StatusCode())
				if response.StatusCode() == http.StatusOK {
					require.NotEmpty(t, response.Cookies())
					var token bool
					for _, cookie := range response.Cookies() {
						if cookie.Name == "token" {
							token = true
							require.NotEmpty(t, cookie.Value)
						}
					}
					require.True(t, token)
				}
			})
		}

	})

	err = s.Stop()
	require.NoError(t, err)
	s.wg.Wait()
}
