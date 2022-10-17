package web

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"

	"github.com/AlekseyKas/keeper/pkg/client/config"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
	"github.com/AlekseyKas/keeper/pkg/models"
)

func TestNew(t *testing.T) {
	config.New()
	got := New()
	require.NotNil(t, got)
	_, ok := got.(*WebClient)
	require.True(t, ok)

	err := got.Close()
	require.NoError(t, err)
}

func TestSignIn(t *testing.T) {

	type test struct {
		name       string
		connection bool
		returnCode int
		returnBody string
		want       bool

		pgp     bool
		pgpCode int
	}
	tests := []test{
		{
			name:       "successfull auth",
			returnCode: http.StatusOK,
			connection: true,
			returnBody: "{\"status\": \"OK\"}",
			want:       false,
		},
		{
			name:       "failed connection",
			connection: false,
			want:       true,
		},
		{
			name:       "failed login",
			returnCode: http.StatusForbidden,
			connection: true,
			returnBody: "{\"status\": \"Forbidden\"}",
			want:       true,
		},
		{
			name:       "failed second login",
			returnCode: http.StatusAlreadyReported,
			connection: true,
			returnBody: "{\"status\": \"Forbidden\"}",
			want:       false,
		},
		{
			name:       "failed login with bad request",
			returnCode: http.StatusBadRequest,
			connection: true,
			returnBody: "{\"status\": \"Forbidden\"}",
			want:       true,
		},
		{
			name:       "wrong username or password",
			returnCode: http.StatusUnauthorized,
			connection: true,
			returnBody: "{\"status\": \"Unauthorized\"}",
			want:       true,
		},
		{
			name:       "internal server error (token not created)",
			returnCode: http.StatusInternalServerError,
			connection: true,
			want:       true,
		},
		{
			name:       "failed login",
			returnCode: http.StatusForbidden,
			connection: true,
			returnBody: "{\"status\": \"Forbidden\"}",
			want:       true,

			pgp:     true,
			pgpCode: http.StatusBadRequest,
		},
		{
			name:       "failed login",
			returnCode: http.StatusForbidden,
			connection: true,
			returnBody: "{\"status\": \"Forbidden\"}",
			want:       true,

			pgp:     true,
			pgpCode: http.StatusInternalServerError,
		},
		{
			name:       "failed login",
			returnCode: http.StatusForbidden,
			connection: true,
			returnBody: "{\"status\": \"Forbidden\"}",
			want:       false,

			pgp:     true,
			pgpCode: http.StatusCreated,
		},
	}
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	wc := &WebClient{
		client: resty.New(),
		jar:    jar,
		logger: zerolog.New().WithPrefix("http-client-test"),
		wsSig:  make(chan struct{}),
	}
	wc.client.SetBaseURL("https://127.0.0.1:8443")
	wc.client.SetCookieJar(wc.jar)
	wc.client.SetTLSClientConfig(&tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: true,
	})

	httpmock.ActivateNonDefault(wc.client.GetClient())
	defer httpmock.DeactivateAndReset()
	fakeUrl := "https://127.0.0.1:8443/api/v1/signin"
	fakePGP := "https://127.0.0.1:8443/api/v1/keeper/pgp/add"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.connection {
				fakeAnswer := tt.returnBody
				responder := httpmock.NewStringResponder(tt.returnCode, fakeAnswer)

				httpmock.RegisterResponder("POST", fakeUrl, responder)
				if tt.pgp {
					responderPGP := httpmock.NewStringResponder(tt.pgpCode, "")
					httpmock.RegisterResponder("POST", fakePGP, responderPGP)
				}
			}

			err = wc.Login("test", "test", "test")
			if err != nil {
				got = true
			}
			require.Equal(t, tt.want, got)
			httpmock.Reset()
		})
	}
}

func TestSignUP(t *testing.T) {

	type test struct {
		name       string
		connection bool
		returnCode int
		returnBody string
		want       bool
	}
	tests := []test{
		{
			name:       "successfull register",
			returnCode: http.StatusCreated,
			connection: true,
			returnBody: "{\"status\": \"OK\"}",
			want:       false,
		},
		{
			name:       "unsuccessfull register",
			returnCode: http.StatusBadRequest,
			connection: true,
			returnBody: "any json",
			want:       true,
		},
		{
			name:       "unsuccessfull register",
			returnCode: http.StatusInternalServerError,
			connection: true,
			returnBody: "any json",
			want:       true,
		},
		{
			name:       "successfull register",
			returnCode: http.StatusCreated,
			connection: false,
			want:       true,
		},
	}
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	wc := &WebClient{
		client: resty.New(),
		jar:    jar,
		logger: zerolog.New().WithPrefix("http-client-test"),
		wsSig:  make(chan struct{}),
	}
	wc.client.SetBaseURL("https://127.0.0.1:8443")
	wc.client.SetCookieJar(wc.jar)
	wc.client.SetTLSClientConfig(&tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: true,
	})

	httpmock.ActivateNonDefault(wc.client.GetClient())
	defer httpmock.DeactivateAndReset()
	fakeUrl := "https://127.0.0.1:8443/api/v1/signup"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.connection {
				fakeAnswer := tt.returnBody
				responder := httpmock.NewStringResponder(tt.returnCode, fakeAnswer)

				httpmock.RegisterResponder("POST", fakeUrl, responder)
			}

			err = wc.Register("test", "test", "test")
			if err != nil {
				got = true
			}
			require.Equal(t, tt.want, got)
			httpmock.Reset()
		})
	}
}

func TestDelete(t *testing.T) {

	type test struct {
		name       string
		connection bool
		returnCode int
		returnBody string
		want       bool
	}
	tests := []test{
		{
			name:       "success",
			returnCode: http.StatusAccepted,
			connection: true,
			returnBody: "{\"status\": \"OK\"}",
			want:       false,
		},
		{
			name:       "faile",
			returnCode: http.StatusBadRequest,
			connection: true,
			returnBody: "any json",
			want:       true,
		},
		{
			name:       "faile",
			returnCode: http.StatusInternalServerError,
			connection: true,
			returnBody: "any json",
			want:       true,
		},
		{
			name:       "fail",
			returnCode: http.StatusCreated,
			connection: false,
			want:       true,
		},
	}
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	wc := &WebClient{
		client: resty.New(),
		jar:    jar,
		logger: zerolog.New().WithPrefix("http-client-test"),
		wsSig:  make(chan struct{}),
	}
	wc.client.SetBaseURL("https://127.0.0.1:8443")
	wc.client.SetCookieJar(wc.jar)
	wc.client.SetTLSClientConfig(&tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: true,
	})

	httpmock.ActivateNonDefault(wc.client.GetClient())
	defer httpmock.DeactivateAndReset()
	fakeUrl := "https://127.0.0.1:8443/api/v1/keeper/remove"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.connection {
				fakeAnswer := tt.returnBody
				responder := httpmock.NewStringResponder(tt.returnCode, fakeAnswer)

				httpmock.RegisterResponder("DELETE", fakeUrl, responder)
			}

			err = wc.Delete()
			if err != nil {
				got = true
			}
			require.Equal(t, tt.want, got)
			httpmock.Reset()
		})
	}
}

func TestPush(t *testing.T) {

	type test struct {
		name       string
		connection bool
		returnCode int
		returnBody string
		want       bool
	}
	tests := []test{
		{
			name:       "success",
			returnCode: http.StatusOK,
			connection: true,
			returnBody: "{\"status\": \"OK\"}",
			want:       false,
		},
		{
			name:       "faile",
			returnCode: http.StatusBadRequest,
			connection: true,
			returnBody: "any json",
			want:       true,
		},
		{
			name:       "faile",
			returnCode: http.StatusInternalServerError,
			connection: true,
			returnBody: "any json",
			want:       true,
		},
		{
			name:       "fail",
			returnCode: http.StatusCreated,
			connection: false,
			want:       true,
		},
	}
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	wc := &WebClient{
		client: resty.New(),
		jar:    jar,
		logger: zerolog.New().WithPrefix("http-client-test"),
		wsSig:  make(chan struct{}),
	}
	wc.client.SetBaseURL("https://127.0.0.1:8443")
	wc.client.SetCookieJar(wc.jar)
	wc.client.SetTLSClientConfig(&tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: true,
	})

	httpmock.ActivateNonDefault(wc.client.GetClient())
	defer httpmock.DeactivateAndReset()
	fakeUrl := "https://127.0.0.1:8443/api/v1/keeper/push"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.connection {
				fakeAnswer := tt.returnBody
				responder := httpmock.NewStringResponder(tt.returnCode, fakeAnswer)

				httpmock.RegisterResponder("POST", fakeUrl, responder)
			}

			err = wc.Push("payload", "1234")
			if err != nil {
				got = true
			}
			require.Equal(t, tt.want, got)
			httpmock.Reset()
		})
	}
}

func TestConfirm(t *testing.T) {

	type test struct {
		name       string
		connection bool
		returnCode int
		returnBody string
		want       bool
	}
	tests := []test{
		{
			name:       "success",
			returnCode: http.StatusOK,
			connection: true,
			returnBody: "{\"status\": \"OK\"}",
			want:       false,
		},
		{
			name:       "fail",
			returnCode: http.StatusBadRequest,
			connection: true,
			returnBody: "any json",
			want:       true,
		},
		{
			name:       "fail",
			returnCode: http.StatusInternalServerError,
			connection: true,
			returnBody: "any json",
			want:       true,
		},
		{
			name:       "fail",
			returnCode: http.StatusCreated,
			connection: false,
			want:       true,
		},
	}
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	wc := &WebClient{
		client: resty.New(),
		jar:    jar,
		logger: zerolog.New().WithPrefix("http-client-test"),
		wsSig:  make(chan struct{}),
	}
	wc.client.SetBaseURL("https://127.0.0.1:8443")
	wc.client.SetCookieJar(wc.jar)
	wc.client.SetTLSClientConfig(&tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: true,
	})

	httpmock.ActivateNonDefault(wc.client.GetClient())
	defer httpmock.DeactivateAndReset()
	fakeUrl := "https://127.0.0.1:8443/api/v1/keeper/pgp/confirm"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.connection {
				fakeAnswer := tt.returnBody
				responder := httpmock.NewStringResponder(tt.returnCode, fakeAnswer)

				httpmock.RegisterResponder("POST", fakeUrl, responder)
			}

			err = wc.ConfirmPGP("1234")
			if err != nil {
				got = true
			}
			require.Equal(t, tt.want, got)
			httpmock.Reset()
		})
	}
}

func TestRevoke(t *testing.T) {

	type test struct {
		name       string
		connection bool
		returnCode int
		returnBody string
		want       bool
	}
	tests := []test{
		{
			name:       "success",
			returnCode: http.StatusGone,
			connection: true,
			returnBody: "any json",
			want:       false,
		},
		{
			name:       "fail",
			returnCode: http.StatusBadRequest,
			connection: true,
			returnBody: "any json",
			want:       true,
		},
		{
			name:       "fail",
			returnCode: http.StatusInternalServerError,
			connection: true,
			returnBody: "any json",
			want:       true,
		},
		{
			name:       "fail",
			returnCode: http.StatusCreated,
			connection: false,
			want:       true,
		},
	}
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	wc := &WebClient{
		client: resty.New(),
		jar:    jar,
		logger: zerolog.New().WithPrefix("http-client-test"),
		wsSig:  make(chan struct{}),
	}
	wc.client.SetBaseURL("https://127.0.0.1:8443")
	wc.client.SetCookieJar(wc.jar)
	wc.client.SetTLSClientConfig(&tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: true,
	})

	httpmock.ActivateNonDefault(wc.client.GetClient())
	defer httpmock.DeactivateAndReset()
	fakeUrl := "https://127.0.0.1:8443/api/v1/keeper/pgp/revoke"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.connection {
				fakeAnswer := tt.returnBody
				responder := httpmock.NewStringResponder(tt.returnCode, fakeAnswer)

				httpmock.RegisterResponder("POST", fakeUrl, responder)
			}

			err = wc.RevokePGP("1234")
			if err != nil {
				got = true
			}
			require.Equal(t, tt.want, got)
			httpmock.Reset()
		})
	}
}

func TestGetLogs(t *testing.T) {

	type test struct {
		name       string
		connection bool
		returnCode int
		returnBody []models.Action
		want       bool
	}
	tests := []test{
		{
			name:       "success",
			returnCode: http.StatusOK,
			connection: true,
			returnBody: []models.Action{{Action: "push", Checksum: "1234", IP: net.ParseIP("127.0.0.1"), Date: time.Now()}},
			want:       false,
		},
		{
			name:       "fail",
			returnCode: http.StatusBadRequest,
			connection: true,
			returnBody: nil,
			want:       true,
		},
		{
			name:       "success",
			returnCode: http.StatusNoContent,
			connection: true,
			returnBody: nil,
			want:       false,
		},
		{
			name:       "fail",
			returnCode: http.StatusCreated,
			connection: false,
			want:       true,
		},
	}
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	wc := &WebClient{
		client: resty.New(),
		jar:    jar,
		logger: zerolog.New().WithPrefix("http-client-test"),
		wsSig:  make(chan struct{}),
	}
	wc.client.SetBaseURL("https://127.0.0.1:8443")
	wc.client.SetCookieJar(wc.jar)
	wc.client.SetTLSClientConfig(&tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: true,
	})

	httpmock.ActivateNonDefault(wc.client.GetClient())
	defer httpmock.DeactivateAndReset()
	fakeUrl := "https://127.0.0.1:8443/api/v1/keeper/logs"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.connection {
				var fakeAnswer string
				if !tt.want {
					body, err := json.Marshal(tt.returnBody)
					require.NoError(t, err)
					fakeAnswer = string(body)
				} else {
					fakeAnswer = ""
				}
				responder := httpmock.NewStringResponder(tt.returnCode, fakeAnswer)

				httpmock.RegisterResponder("GET", fakeUrl, responder)
			}

			_, err := wc.GetLogs()
			if err != nil {
				got = true
			}
			require.Equal(t, tt.want, got)
			httpmock.Reset()
		})
	}
}

func TestPull(t *testing.T) {

	type test struct {
		name       string
		connection bool
		returnCode int
		returnBody []byte
		want       bool
	}
	tests := []test{
		{
			name:       "success",
			returnCode: http.StatusOK,
			connection: true,
			returnBody: []byte("hello"),
			want:       false,
		},
		{
			name:       "fail",
			returnCode: http.StatusBadRequest,
			connection: true,
			returnBody: nil,
			want:       true,
		},
		{
			name:       "success",
			returnCode: http.StatusNoContent,
			connection: true,
			returnBody: nil,
			want:       false,
		},
		{
			name:       "fail",
			returnCode: http.StatusCreated,
			connection: false,
			want:       true,
		},
	}
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	wc := &WebClient{
		client: resty.New(),
		jar:    jar,
		logger: zerolog.New().WithPrefix("http-client-test"),
		wsSig:  make(chan struct{}),
	}
	wc.client.SetBaseURL("https://127.0.0.1:8443")
	wc.client.SetCookieJar(wc.jar)
	wc.client.SetTLSClientConfig(&tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: true,
	})

	httpmock.ActivateNonDefault(wc.client.GetClient())
	defer httpmock.DeactivateAndReset()
	fakeUrl := "https://127.0.0.1:8443/api/v1/keeper/pull?checksum=1234"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.connection {
				responder := httpmock.NewStringResponder(tt.returnCode, string(tt.returnBody))
				httpmock.RegisterResponder("GET", fakeUrl, responder)
			}

			_, err := wc.Pull("1234")
			if err != nil {
				got = true
			}
			require.Equal(t, tt.want, got)
			httpmock.Reset()
		})
	}
}

func TestVersions(t *testing.T) {

	type test struct {
		name       string
		connection bool
		returnCode int
		returnBody []models.Version
		want       bool
	}
	tests := []test{
		{
			name:       "success",
			returnCode: http.StatusOK,
			connection: true,
			returnBody: []models.Version{{Date: time.Now(), Hash: "1234"}},
			want:       false,
		},
		{
			name:       "success",
			returnCode: http.StatusNoContent,
			connection: true,
			returnBody: []models.Version{},
			want:       false,
		},
		{
			name:       "fail",
			returnCode: http.StatusBadRequest,
			connection: true,
			returnBody: nil,
			want:       true,
		},
		{
			name:       "fail",
			returnCode: http.StatusCreated,
			connection: false,
			want:       true,
		},
	}
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	wc := &WebClient{
		client: resty.New(),
		jar:    jar,
		logger: zerolog.New().WithPrefix("http-client-test"),
		wsSig:  make(chan struct{}),
	}
	wc.client.SetBaseURL("https://127.0.0.1:8443")
	wc.client.SetCookieJar(wc.jar)
	wc.client.SetTLSClientConfig(&tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: true,
	})

	httpmock.ActivateNonDefault(wc.client.GetClient())
	defer httpmock.DeactivateAndReset()
	fakeUrl := "https://127.0.0.1:8443/api/v1/keeper/pull/versions"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.connection {
				var fakeAnswer string
				if !tt.want {
					body, err := json.Marshal(tt.returnBody)
					require.NoError(t, err)
					fakeAnswer = string(body)
				} else {
					fakeAnswer = ""
				}
				responder := httpmock.NewStringResponder(tt.returnCode, fakeAnswer)

				httpmock.RegisterResponder("GET", fakeUrl, responder)
			}

			_, err := wc.Versions()
			if err != nil {
				got = true
			}
			require.Equal(t, tt.want, got)
			httpmock.Reset()
		})
	}
}

func TestListPGP(t *testing.T) {

	type test struct {
		name       string
		connection bool
		returnCode int
		returnBody []models.PGP
		want       bool
	}
	tests := []test{
		{
			name:       "success",
			returnCode: http.StatusOK,
			connection: true,
			returnBody: []models.PGP{{Date: time.Now(), Publickey: "1234", Confirmed: true}},
			want:       false,
		},
		{
			name:       "success",
			returnCode: http.StatusNoContent,
			connection: true,
			returnBody: []models.PGP{},
			want:       false,
		},
		{
			name:       "fail",
			returnCode: http.StatusBadRequest,
			connection: true,
			returnBody: nil,
			want:       true,
		},
		{
			name:       "fail",
			returnCode: http.StatusCreated,
			connection: false,
			want:       true,
		},
	}
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	wc := &WebClient{
		client: resty.New(),
		jar:    jar,
		logger: zerolog.New().WithPrefix("http-client-test"),
		wsSig:  make(chan struct{}),
	}
	wc.client.SetBaseURL("https://127.0.0.1:8443")
	wc.client.SetCookieJar(wc.jar)
	wc.client.SetTLSClientConfig(&tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: true,
	})

	httpmock.ActivateNonDefault(wc.client.GetClient())
	defer httpmock.DeactivateAndReset()
	fakeUrl := "https://127.0.0.1:8443/api/v1/keeper/pgp/list"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.connection {
				var fakeAnswer string
				if !tt.want {
					body, err := json.Marshal(tt.returnBody)
					require.NoError(t, err)
					fakeAnswer = string(body)
				} else {
					fakeAnswer = ""
				}
				responder := httpmock.NewStringResponder(tt.returnCode, fakeAnswer)

				httpmock.RegisterResponder("GET", fakeUrl, responder)
			}

			_, err := wc.ListPGP()
			if err != nil {
				got = true
			}
			require.Equal(t, tt.want, got)
			httpmock.Reset()
		})
	}
}
