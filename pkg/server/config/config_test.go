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
		WebBind: ":8443",
		// GRPCBind: ":3200",
	}
	t.Run("test default values", func(t *testing.T) {
		c := &Config{}
		c.logger = zerolog.New().WithPrefix("configuration")
		c.SetDefaults()
		// require.Equal(t, want.GRPCBind, c.GRPCBind)
		require.Equal(t, want.WebBind, c.WebBind)
		require.NotNil(t, c.logger)
		c.log().WithPrefix("configuration").Debugf("\n%v\n", nil, c)
	})
}

func TestNew(t *testing.T) {
	want := Config{
		WebBind: ":8443",
		// GRPCBind: ":3200",
		LogLevel: uint8(logging.DebugLevel),
	}
	t.Run("test New()", func(t *testing.T) {
		c := New()
		// require.Equal(t, want.GRPCBind, c.GRPCBind)
		require.Equal(t, want.WebBind, c.WebBind)
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
			WebBind: "127.0.0.1:443",
			// GRPCBind: "127.0.0.1:2300",
			DSN: "postgresql://user:password@netloc:port/dbname?param1=value1",
		},
	}
	os.Setenv("WEB_ADDRESS", "127.0.0.1:443")
	os.Setenv("GRPC_ADDRESS", "127.0.0.1:2300")
	os.Setenv("DSN_ADDRESS", "postgresql://user:password@netloc:port/dbname?param1=value1")
	defer os.Unsetenv("WEB_ADDRESS")
	defer os.Unsetenv("GRPC_ADDRESS")
	defer os.Unsetenv("DSN_ADDRESS")
	t.Run(test.name, func(t *testing.T) {
		c := Config{}
		c.SetDefaults()
		c.SetByEnv()
		require.Equal(t, test.want.WebBind, c.WebBind)
		// require.Equal(t, test.want.GRPCBind, c.GRPCBind)
		require.Equal(t, test.want.DSN, c.DSN)
	})
}

func TestConfig_SetByFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want Config
	}{
		{
			name: "set up flags with warn log level",
			args: []string{"-http", "127.0.0.0:8080", "-loglevel", "warn", "-config", "remove_me.json", "-dsn", "postgresql://user:password@netloc:port/dbname?param1=value1"},
			want: Config{
				WebBind: "127.0.0.0:8080",
				// GRPCBind: "192.168.0.1:5555",
				LogLevel: uint8(logging.WarnLevel),
				DSN:      "postgresql://user:password@netloc:port/dbname?param1=value1",
				file:     "remove_me.json",
			},
		},
		{
			name: "set up flags with fatal log level",
			args: []string{"-http", "127.0.0.0:8080", "-loglevel", "fatal", "-config", "remove_me.json", "-dsn", "postgresql://user:password@netloc:port/dbname?param1=value1"},
			want: Config{
				WebBind: "127.0.0.0:8080",
				// GRPCBind: "192.168.0.1:5555",
				LogLevel: uint8(logging.FatalLevel),
				DSN:      "postgresql://user:password@netloc:port/dbname?param1=value1",
				file:     "remove_me.json",
			},
		},
		{
			name: "set up flags with panic log level",
			args: []string{"-http", "127.0.0.0:8080", "-loglevel", "panic", "-config", "remove_me.json", "-dsn", "postgresql://user:password@netloc:port/dbname?param1=value1"},
			want: Config{
				WebBind: "127.0.0.0:8080",
				// GRPCBind: "192.168.0.1:5555",
				LogLevel: uint8(logging.PanicLevel),
				DSN:      "postgresql://user:password@netloc:port/dbname?param1=value1",
				file:     "remove_me.json",
			},
		},
		{
			name: "set up flags with debug log level",
			args: []string{"-http", "127.0.0.0:8080", "-loglevel", "debug", "-config", "remove_me.json", "-dsn", "postgresql://user:password@netloc:port/dbname?param1=value1"},
			want: Config{
				WebBind: "127.0.0.0:8080",
				// GRPCBind: "192.168.0.1:5555",
				LogLevel: uint8(logging.DebugLevel),
				DSN:      "postgresql://user:password@netloc:port/dbname?param1=value1",
				file:     "remove_me.json",
			},
		},
		{
			name: "set up flags with info log level",
			args: []string{"-http", "127.0.0.0:8080", "-loglevel", "info", "-config", "remove_me.json", "-dsn", "postgresql://user:password@netloc:port/dbname?param1=value1"},
			want: Config{
				WebBind: "127.0.0.0:8080",
				// GRPCBind: "192.168.0.1:5555",
				LogLevel: uint8(logging.InfoLevel),
				DSN:      "postgresql://user:password@netloc:port/dbname?param1=value1",
				file:     "remove_me.json",
			},
		},
		{
			name: "set up flags with error log level",
			args: []string{"-http", "127.0.0.0:8080", "-loglevel", "error", "-config", "remove_me.json", "-dsn", "postgresql://user:password@netloc:port/dbname?param1=value1"},
			want: Config{
				WebBind: "127.0.0.0:8080",
				// GRPCBind: "192.168.0.1:5555",
				LogLevel: uint8(logging.ErrorLevel),
				DSN:      "postgresql://user:password@netloc:port/dbname?param1=value1",
				file:     "remove_me.json",
			},
		},
		{
			name: "set up flags with trace log level",
			args: []string{"-http", "127.0.0.0:8080", "-loglevel", "trace", "-config", "remove_me.json", "-dsn", "postgresql://user:password@netloc:port/dbname?param1=value1"},
			want: Config{
				WebBind: "127.0.0.0:8080",
				// GRPCBind: "192.168.0.1:5555",
				LogLevel: uint8(logging.TraceLevel),
				DSN:      "postgresql://user:password@netloc:port/dbname?param1=value1",
				file:     "remove_me.json",
			},
		},
		{
			name: "set up flags with wrong log level",
			args: []string{"-http", "127.0.0.0:8080", "-loglevel", "bla", "-config", "remove_me.json", "-dsn", "postgresql://user:password@netloc:port/dbname?param1=value1"},
			want: Config{
				WebBind: "127.0.0.0:8080",
				// GRPCBind: "192.168.0.1:5555",
				LogLevel: uint8(logging.DebugLevel),
				DSN:      "postgresql://user:password@netloc:port/dbname?param1=value1",
				file:     "remove_me.json",
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
			require.Equal(t, tt.want.WebBind, c.WebBind)
			// require.Equal(t, tt.want.GRPCBind, c.GRPCBind)
			require.Equal(t, tt.want.LogLevel, c.LogLevel)
			require.Equal(t, tt.want.DSN, c.DSN)
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
				WebBind: "127.0.0.1:443",
				// GRPCBind: "127.0.0.1:2300",
				DSN:      "postgresql://user:password@netloc:port/dbname?param1=value1",
				LogLevel: 4,
				file:     "remove_me.json",
			},
		},
		{
			name: "not existed test file",
			want: &Config{
				WebBind: "127.0.0.1:443",
				// GRPCBind: "127.0.0.1:2300",
				DSN:      "postgresql://user:password@netloc:port/dbname?param1=value1",
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
			require.Equal(t, tt.want.WebBind, got.WebBind)
			// require.Equal(t, tt.want.GRPCBind, got.GRPCBind)
			require.Equal(t, tt.want.DSN, got.DSN)
			require.Equal(t, tt.want.LogLevel, got.LogLevel)
			require.Equal(t, tt.want.file, got.file)
		})
		if tt.want.file != "not_create_me.json" {
			err := os.Remove(tt.want.file)
			require.NoError(t, err)
		}
	}
}
