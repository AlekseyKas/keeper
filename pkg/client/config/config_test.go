package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/AlekseyKas/keeper/pkg/logging"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
	"github.com/stretchr/testify/require"
)

func TestNewWithDefault(t *testing.T) {
	want := Config{
		file:    "config.json",
		Mode:    "standalone",
		Storage: "secrets.db",
	}
	t.Run("test default values", func(t *testing.T) {
		c := &Config{}
		c.logger = zerolog.New().WithPrefix("configuration")
		c.SetDefaults()
		require.Equal(t, want.file, c.file)
		require.Equal(t, want.Mode, c.Mode)
		require.Equal(t, want.Storage, c.Storage)
		require.NotNil(t, c.logger)
		c.log().Debugf("\n%v\n", nil, c)
	})
}

func TestNew(t *testing.T) {
	want := Config{
		file:     "config.json",
		Mode:     "standalone",
		Storage:  "secrets.db",
		LogLevel: uint8(logging.InfoLevel),
	}
	t.Run("test New()", func(t *testing.T) {
		c := New()
		require.Equal(t, want.file, c.file)
		require.Equal(t, want.Mode, c.Mode)
		require.Equal(t, want.Storage, c.Storage)
		require.Equal(t, want.LogLevel, c.LogLevel)
		require.NotNil(t, c.logger)
	})
}

func TestConfig_SetByEnv(t *testing.T) {
	test := struct {
		name string
		want Config
	}{
		name: "reading os environments",
		want: Config{
			Username:   "username",
			Password:   "password",
			RemoteHTTP: "https://localhost.ltd:8443",
		},
	}
	os.Setenv("KEEPER_PGP_PASSPHRASE", "passphrase")
	os.Setenv("KEEPER_REMOTE_USERNAME", "username")
	os.Setenv("KEEPER_REMOTE_PASSWORD", "password")
	os.Setenv("KEEPER_REMOTE_URL", "https://localhost.ltd:8443")
	defer os.Unsetenv("KEEPER_PGP_PASSPHRASE")
	defer os.Unsetenv("KEEPER_REMOTE_USERNAME")
	defer os.Unsetenv("KEEPER_REMOTE_PASSWORD")
	defer os.Unsetenv("KEEPER_REMOTE_URL")
	t.Run(test.name, func(t *testing.T) {
		c := Config{}
		c.SetDefaults()
		c.SetByEnv()
		require.Equal(t, test.want.Username, c.Username)
		require.Equal(t, test.want.Password, c.Password)
		require.Equal(t, test.want.RemoteHTTP, c.RemoteHTTP)
	})
}

func TestConfig_SetByFlags(t *testing.T) {
	_ = zerolog.New().WithPrefix("configuration")
	tests := []struct {
		name string
		args []string
		want Config
	}{
		{
			name: "set up flags with warn log level",
			args: []string{"-remote-http", "https://localhost.ltd:8443", "-config", "remove_me.json", "-storage", "secrets.db", "-username", "username", "-password", "password", "-loglevel", "warn"},
			want: Config{
				Mode:       "client-server",
				Username:   "username",
				Password:   "password",
				RemoteHTTP: "https://localhost.ltd:8443",

				LogLevel: uint8(logging.WarnLevel),
				Storage:  "secrets.db",
				file:     "remove_me.json",
			},
		},
		{
			name: "set up flags with fatal log level",
			args: []string{"-remote-http", "https://localhost.ltd:8443", "-config", "remove_me.json", "-storage", "secrets.db", "-username", "username", "-password", "password", "-loglevel", "fatal"},
			want: Config{
				Mode:       "client-server",
				Username:   "username",
				Password:   "password",
				RemoteHTTP: "https://localhost.ltd:8443",

				LogLevel: uint8(logging.FatalLevel),
				Storage:  "secrets.db",
				file:     "remove_me.json",
			},
		},
		{
			name: "set up flags with panic log level",
			args: []string{"-remote-http", "https://localhost.ltd:8443", "-config", "remove_me.json", "-storage", "secrets.db", "-username", "username", "-password", "password", "-loglevel", "panic"},
			want: Config{
				Mode:       "client-server",
				Username:   "username",
				Password:   "password",
				RemoteHTTP: "https://localhost.ltd:8443",

				LogLevel: uint8(logging.PanicLevel),
				Storage:  "secrets.db",
				file:     "remove_me.json",
			},
		},
		{
			name: "set up flags with debug log level",
			args: []string{"-remote-http", "https://localhost.ltd:8443", "-config", "remove_me.json", "-storage", "secrets.db", "-username", "username", "-password", "password", "-loglevel", "debug"},
			want: Config{
				Mode:       "client-server",
				Username:   "username",
				Password:   "password",
				RemoteHTTP: "https://localhost.ltd:8443",

				LogLevel: uint8(logging.DebugLevel),
				Storage:  "secrets.db",
				file:     "remove_me.json",
			},
		},
		{
			name: "set up flags with info log level",
			args: []string{"-remote-http", "https://localhost.ltd:8443", "-config", "remove_me.json", "-storage", "secrets.db", "-username", "username", "-password", "password", "-loglevel", "info"},
			want: Config{
				Mode:       "client-server",
				Username:   "username",
				Password:   "password",
				RemoteHTTP: "https://localhost.ltd:8443",

				LogLevel: uint8(logging.InfoLevel),
				Storage:  "secrets.db",
				file:     "remove_me.json",
			},
		},
		{
			name: "set up flags with error log level",
			args: []string{"-remote-http", "https://localhost.ltd:8443", "-config", "remove_me.json", "-storage", "secrets.db", "-username", "username", "-password", "password", "-loglevel", "error"},
			want: Config{
				Mode:       "client-server",
				Username:   "username",
				Password:   "password",
				RemoteHTTP: "https://localhost.ltd:8443",

				LogLevel: uint8(logging.ErrorLevel),
				Storage:  "secrets.db",
				file:     "remove_me.json",
			},
		},
		{
			name: "set up flags with trace log level",
			args: []string{"-remote-http", "https://localhost.ltd:8443", "-config", "remove_me.json", "-storage", "secrets.db", "-username", "username", "-password", "password", "-loglevel", "trace"},
			want: Config{
				Mode:       "client-server",
				Username:   "username",
				Password:   "password",
				RemoteHTTP: "https://localhost.ltd:8443",

				LogLevel: uint8(logging.TraceLevel),
				Storage:  "secrets.db",
				file:     "remove_me.json",
			},
		},
		{
			name: "set up flags with wrong log level",
			args: []string{"-remote-http", "https://localhost.ltd:8443", "-config", "remove_me.json", "-storage", "secrets.db", "-username", "username", "-password", "password", "-loglevel", "bla"},
			want: Config{
				Mode:       "client-server",
				Username:   "username",
				Password:   "password",
				RemoteHTTP: "https://localhost.ltd:8443",
				LogLevel:   uint8(logging.InfoLevel),
				Storage:    "secrets.db",
				file:       "remove_me.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			c := Config{}
			c.SetDefaults()
			os.Args = append(os.Args, tt.args...)
			c.SetByFlags()
			require.Equal(t, tt.want.Username, c.Username)
			require.Equal(t, tt.want.Password, c.Password)
			require.Equal(t, tt.want.RemoteHTTP, c.RemoteHTTP)
			require.Equal(t, tt.want.RemoteGRPC, c.RemoteGRPC)
			require.Equal(t, tt.want.LogLevel, c.LogLevel)
			require.Equal(t, tt.want.file, c.file)
		})
	}
}

func TestConfig_SetByFile(t *testing.T) {
	tests := []struct {
		name string
		want *Config
	}{
		{
			name: "existed test file",
			want: &Config{
				Username:   "username",
				Password:   "password",
				RemoteHTTP: "https://localhost.ltd:8443",

				LogLevel: 4,
				file:     "remove_me.json",
			},
		},
		{
			name: "not existed test file",
			want: &Config{
				Username:   "username",
				Password:   "password",
				RemoteHTTP: "https://localhost.ltd:8443",

				LogLevel: 4,
				file:     "not_create_me.json",
			},
		},
	}
	for _, tt := range tests {
		tt.want.logger = zerolog.New().WithPrefix("configuration")
		if tt.want.file != "not_create_me.json" {
			f, err := os.OpenFile(tt.want.file, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
			require.NoError(t, err)
			jsonConfig, err := json.MarshalIndent(tt.want, "", "  ")
			require.NoError(t, err)
			f.Write(jsonConfig)
			f.Close()
		}
		t.Run(tt.name, func(t *testing.T) {
			got := tt.want.SetByFile()
			require.Equal(t, tt.want.Username, got.Username)
			require.Equal(t, tt.want.Password, got.Password)
			require.Equal(t, tt.want.RemoteHTTP, got.RemoteHTTP)
			require.Equal(t, tt.want.RemoteGRPC, got.RemoteGRPC)
			require.Equal(t, tt.want.LogLevel, got.LogLevel)
			require.Equal(t, tt.want.file, got.file)
		})
		if tt.want.file != "not_create_me.json" {
			err := os.Remove(tt.want.file)
			require.NoError(t, err)
		}
	}
}
