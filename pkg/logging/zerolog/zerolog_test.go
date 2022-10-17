package zerolog

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/AlekseyKas/keeper/pkg/logging"
)

func TestNew(t *testing.T) {
	logger := New()
	test := struct {
		name string
	}{
		name: "logger creation",
	}
	t.Run(test.name, func(t *testing.T) {
		require.NotNil(t, logger)
	})
}

func Test_zeroLogger_Print(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "simple word",
			args: []string{"word"},
		},
		{
			name: "many words",
			args: []string{"word1", "word2"},
		},
		{
			name: "many ints",
			args: []int{123, 12},
		},
	}
	for _, tt := range tests {
		out := &bytes.Buffer{}
		re := regexp.MustCompile(`\?\?\?.\[.+\]`)
		t.Run(tt.name, func(t *testing.T) {
			log := new(zeroLogger)
			console := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
			log.logger = zerolog.New(console)
			log.Print(nil, tt.args)
			// t.Log(out.String())
			require.True(t, re.Match(out.Bytes()))
			out.Reset()
		})
	}
}

func Test_zeroLogger_Printf(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "simple word",
			args: []string{"word"},
		},
		{
			name: "many words",
			args: []string{"word1", "word2"},
		},
		{
			name: "many ints",
			args: []int{123, 12},
		},
	}
	for _, tt := range tests {
		out := &bytes.Buffer{}
		re := regexp.MustCompile(`format.\[.+\]`)
		t.Run(tt.name, func(t *testing.T) {
			log := new(zeroLogger)
			console := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
			log.logger = zerolog.New(console)
			log.Printf("format=%v", nil, tt.args)
			// t.Log(out.String())
			require.True(t, re.Match(out.Bytes()))
			out.Reset()
		})
	}
}

func Test_zeroLogger_Tracef(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "simple word",
			args: []string{"word"},
		},
		{
			name: "many words",
			args: []string{"word1", "word2"},
		},
		{
			name: "many ints",
			args: []int{123, 12},
		},
	}
	re := regexp.MustCompile(`format.\[.+\]`)
	re1 := regexp.MustCompile(`TRC`)
	for _, tt := range tests {
		out := &bytes.Buffer{}
		t.Run(tt.name, func(t *testing.T) {
			log := new(zeroLogger)
			console := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
			log.logger = zerolog.New(console)
			log.Tracef("format=%v", nil, tt.args)
			// t.Log(out.String())
			require.True(t, re.Match(out.Bytes()))
			require.True(t, re1.Match(out.Bytes()))
			out.Reset()
		})
	}
}

func Test_zeroLogger_Trace(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "simple word",
			args: []string{"word"},
		},
		{
			name: "many words",
			args: []string{"word1", "word2"},
		},
		{
			name: "many ints",
			args: []int{123, 12},
		},
	}
	re := regexp.MustCompile(`\[.+\]`)
	re1 := regexp.MustCompile(`TRC`)
	for _, tt := range tests {
		out := &bytes.Buffer{}
		t.Run(tt.name, func(t *testing.T) {
			log := new(zeroLogger)
			console := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
			log.logger = zerolog.New(console)
			log.Trace(nil, tt.args)
			// t.Log(out.String())
			require.True(t, re.Match(out.Bytes()))
			require.True(t, re1.Match(out.Bytes()))
			out.Reset()
		})
	}
}

func Test_zeroLogger_Debugf(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "simple word",
			args: []string{"word"},
		},
		{
			name: "many words",
			args: []string{"word1", "word2"},
		},
		{
			name: "many ints",
			args: []int{123, 12},
		},
	}
	re := regexp.MustCompile(`format.\[.+\]`)
	re1 := regexp.MustCompile(`DBG`)
	for _, tt := range tests {
		out := &bytes.Buffer{}
		t.Run(tt.name, func(t *testing.T) {
			log := new(zeroLogger)
			console := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
			log.logger = zerolog.New(console)
			log.Debugf("format=%v", nil, tt.args)
			// t.Log(out.String())
			require.True(t, re.Match(out.Bytes()))
			require.True(t, re1.Match(out.Bytes()))
			out.Reset()
		})
	}
}

func Test_zeroLogger_Debug(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "simple word",
			args: []string{"word"},
		},
		{
			name: "many words",
			args: []string{"word1", "word2"},
		},
		{
			name: "many ints",
			args: []int{123, 12},
		},
	}
	re := regexp.MustCompile(`\[.+\]`)
	re1 := regexp.MustCompile(`DBG`)
	for _, tt := range tests {
		out := &bytes.Buffer{}
		t.Run(tt.name, func(t *testing.T) {
			log := new(zeroLogger)
			console := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
			log.logger = zerolog.New(console)
			log.Debug(nil, tt.args)
			// t.Log(out.String())
			require.True(t, re.Match(out.Bytes()))
			require.True(t, re1.Match(out.Bytes()))
			out.Reset()
		})
	}
}

func Test_zeroLogger_Infof(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "simple word",
			args: []string{"word"},
		},
		{
			name: "many words",
			args: []string{"word1", "word2"},
		},
		{
			name: "many ints",
			args: []int{123, 12},
		},
	}
	re := regexp.MustCompile(`format.\[.+\]`)
	re1 := regexp.MustCompile(`INF`)
	for _, tt := range tests {
		out := &bytes.Buffer{}
		t.Run(tt.name, func(t *testing.T) {
			log := new(zeroLogger)
			console := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
			log.logger = zerolog.New(console)
			log.Infof("format=%v", nil, tt.args)
			// t.Log(out.String())
			require.True(t, re.Match(out.Bytes()))
			require.True(t, re1.Match(out.Bytes()))
			out.Reset()
		})
	}
}

func Test_zeroLogger_Info(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "simple word",
			args: []string{"word"},
		},
		{
			name: "many words",
			args: []string{"word1", "word2"},
		},
		{
			name: "many ints",
			args: []int{123, 12},
		},
	}
	re := regexp.MustCompile(`\[.+\]`)
	re1 := regexp.MustCompile(`INF`)
	for _, tt := range tests {
		out := &bytes.Buffer{}
		t.Run(tt.name, func(t *testing.T) {
			log := new(zeroLogger)
			console := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
			log.logger = zerolog.New(console)
			log.Info(nil, tt.args)
			// t.Log(out.String())
			require.True(t, re.Match(out.Bytes()))
			require.True(t, re1.Match(out.Bytes()))
			out.Reset()
		})
	}
}

func Test_zeroLogger_Warnf(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "simple word",
			args: []string{"word"},
		},
		{
			name: "many words",
			args: []string{"word1", "word2"},
		},
		{
			name: "many ints",
			args: []int{123, 12},
		},
	}
	re := regexp.MustCompile(`format.\[.+\]`)
	re1 := regexp.MustCompile(`WRN`)
	for _, tt := range tests {
		out := &bytes.Buffer{}
		t.Run(tt.name, func(t *testing.T) {
			log := new(zeroLogger)
			console := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
			log.logger = zerolog.New(console)
			log.Warnf("format=%v", nil, tt.args)
			// t.Log(out.String())
			require.True(t, re.Match(out.Bytes()))
			require.True(t, re1.Match(out.Bytes()))
			out.Reset()
		})
	}
}

func Test_zeroLogger_Warn(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "simple word",
			args: []string{"word"},
		},
		{
			name: "many words",
			args: []string{"word1", "word2"},
		},
		{
			name: "many ints",
			args: []int{123, 12},
		},
	}
	re := regexp.MustCompile(`\[.+\]`)
	re1 := regexp.MustCompile(`WRN`)
	for _, tt := range tests {
		out := &bytes.Buffer{}
		t.Run(tt.name, func(t *testing.T) {
			log := new(zeroLogger)
			console := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
			log.logger = zerolog.New(console)
			log.Warn(nil, tt.args)
			// t.Log(out.String())
			require.True(t, re.Match(out.Bytes()))
			require.True(t, re1.Match(out.Bytes()))
			out.Reset()
		})
	}
}

func Test_zeroLogger_Errorf(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "simple word",
			args: []string{"word"},
		},
		{
			name: "many words",
			args: []string{"word1", "word2"},
		},
		{
			name: "many ints",
			args: []int{123, 12},
		},
	}
	re := regexp.MustCompile(`format.\[.+\]`)
	re1 := regexp.MustCompile(`ERR`)
	for _, tt := range tests {
		out := &bytes.Buffer{}
		t.Run(tt.name, func(t *testing.T) {
			log := new(zeroLogger)
			console := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
			log.logger = zerolog.New(console)
			err := errors.New("test error")
			log.Errorf("format=%v", err, tt.args)
			t.Log(out.String())
			require.True(t, re.Match(out.Bytes()))
			require.True(t, re1.Match(out.Bytes()))
			out.Reset()
		})
	}
}

func Test_zeroLogger_Error(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "simple word",
			args: []string{"word"},
		},
		{
			name: "many words",
			args: []string{"word1", "word2"},
		},
		{
			name: "many ints",
			args: []int{123, 12},
		},
	}
	re := regexp.MustCompile(`\[.+\]`)
	re1 := regexp.MustCompile(`ERR`)
	for _, tt := range tests {
		out := &bytes.Buffer{}
		t.Run(tt.name, func(t *testing.T) {
			log := new(zeroLogger)
			console := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
			log.logger = zerolog.New(console)
			err := errors.New("test error")
			log.Error(err, tt.args)
			t.Log(out.String())
			require.True(t, re.Match(out.Bytes()))
			require.True(t, re1.Match(out.Bytes()))
			out.Reset()
		})
	}
}

func Test_zeroLogger_WithPrefix(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "simple word",
			args: []string{"word"},
		},
		{
			name: "many words",
			args: []string{"word1", "word2"},
		},
		{
			name: "many ints",
			args: []int{123, 12},
		},
	}
	for _, tt := range tests {
		out := &bytes.Buffer{}
		t.Run(tt.name, func(t *testing.T) {
			log := new(zeroLogger)
			console := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
			log.logger = zerolog.New(console)
			testLogger := log.WithPrefix("testing")
			testLogger.Print(nil, tt.args)
			// t.Log(out.String())
			require.True(t, strings.Contains(out.String(), `component=`))
			require.True(t, strings.Contains(out.String(), `testing`))
			out.Reset()
		})
	}
}

func Test_zeroLogger_WithFields(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "simple word",
			args: []string{"word"},
		},
		{
			name: "many words",
			args: []string{"word1", "word2"},
		},
		{
			name: "many ints",
			args: []int{123, 12},
		},
	}
	for _, tt := range tests {
		out := &bytes.Buffer{}
		t.Run(tt.name, func(t *testing.T) {
			log := new(zeroLogger)
			console := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
			log.logger = zerolog.New(console)
			fields := logging.Fields{"key1": "value", "key2": 123}
			testLogger := log.WithFields(fields)
			testLogger.Print(nil, tt.args)
			// t.Log(out.String())
			require.True(t, strings.Contains(out.String(), `key1=`))
			require.True(t, strings.Contains(out.String(), `value`))
			require.True(t, strings.Contains(out.String(), `key2=`))
			require.True(t, strings.Contains(out.String(), `123`))
			out.Reset()
		})
	}
}

func Test_zeroLogger_SetLevel(t *testing.T) {
	tests := []struct {
		name string
		arg  logging.Level
	}{
		{
			name: "test Trace Level",
			arg:  logging.TraceLevel,
		},
		{
			name: "test DebugLevel",
			arg:  logging.DebugLevel,
		},
		{
			name: "test InfoLevel",
			arg:  logging.InfoLevel,
		},
		{
			name: "test ErrorLevel",
			arg:  logging.ErrorLevel,
		},
		{
			name: "test WarnLevel",
			arg:  logging.WarnLevel,
		},
		{
			name: "test FatalLevel",
			arg:  logging.FatalLevel,
		},
		{
			name: "test PanicLevel",
			arg:  logging.PanicLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New()
			logger.SetLevel(tt.arg)
			require.Equal(t, tt.arg, logger.GetLevel())
		})
	}
}
