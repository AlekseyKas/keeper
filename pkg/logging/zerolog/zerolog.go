package zerolog

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/AlekseyKas/keeper/pkg/logging"
)

var (
	once sync.Once
	_zl  *zeroLogger
)

type zeroLogger struct {
	logger zerolog.Logger
}

// New - new logger with custom destination(s). logger print pretty string to os.Stderr and json to out
//   out... may be any implemets method io.Write. To out log will be written in json format
func New(out ...io.Writer) logging.Logger {
	once.Do(func() {
		_zl = new(zeroLogger)
		pretty := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
		out = append(out, pretty)
		output := io.MultiWriter(out...)
		_zl.logger = zerolog.New(output).With().Timestamp().Logger()
		_zl.Info(nil, "logger successfully initialized")
	})
	return _zl
}

// Print - print log without logging level
func (log *zeroLogger) Print(err error, args ...any) {
	msg := fmt.Sprint(args...)
	log.logger.Log().Err(err).Msg(msg)
}

// Printf - print log without logging level in custom format
func (log *zeroLogger) Printf(format string, err error, args ...any) {
	log.logger.Log().Err(err).Msgf(format, args...)
}

// Trace - print log with TRACE logging level
func (log *zeroLogger) Trace(err error, args ...any) {
	msg := fmt.Sprint(args...)
	log.logger.Trace().Err(err).Msg(msg)
}

func (log *zeroLogger) Tracef(format string, err error, args ...any) {
	log.logger.Trace().Err(err).Msgf(format, args...)
}

func (log *zeroLogger) Debug(err error, args ...any) {
	msg := fmt.Sprint(args...)
	log.logger.Debug().Err(err).Msg(msg)
}

func (log *zeroLogger) Debugf(format string, err error, args ...any) {
	log.logger.Debug().Err(err).Msgf(format, args...)
}

func (log *zeroLogger) Info(err error, args ...any) {
	msg := fmt.Sprint(args...)
	log.logger.Info().Err(err).Msg(msg)
}

func (log *zeroLogger) Infof(format string, err error, args ...any) {
	log.logger.Info().Err(err).Msgf(format, args...)
}

func (log *zeroLogger) Warn(err error, args ...any) {
	msg := fmt.Sprint(args...)
	log.logger.Warn().Err(err).Msg(msg)
}

func (log *zeroLogger) Warnf(format string, err error, args ...any) {
	log.logger.Warn().Err(err).Msgf(format, args...)
}

func (log *zeroLogger) Error(err error, args ...any) {
	msg := fmt.Sprint(args...)
	log.logger.Error().Err(err).Msg(msg)
}

func (log *zeroLogger) Errorf(format string, err error, args ...any) {
	log.logger.Error().Err(err).Msgf(format, args...)
}

func (log *zeroLogger) Fatal(err error, args ...any) {
	msg := fmt.Sprint(args...)
	log.logger.Fatal().Err(err).Msg(msg)
}

func (log *zeroLogger) Fatalf(format string, err error, args ...any) {
	log.logger.Fatal().Err(err).Msgf(format, args...)
}

func (log *zeroLogger) Panic(err error, args ...any) {
	msg := fmt.Sprint(args...)
	log.logger.Panic().Err(err).Msg(msg)
}

func (log *zeroLogger) Panicf(format string, err error, args ...any) {
	log.logger.Panic().Err(err).Msgf(format, args...)
}

func (log *zeroLogger) WithPrefix(prefix string) logging.Logger {
	return &zeroLogger{logger: log.logger.With().Str("component", prefix).Logger()}
}

func (log *zeroLogger) WithFields(fields logging.Fields) logging.Logger {
	return &zeroLogger{logger: log.logger.With().Fields(map[string]interface{}(fields)).Logger()}
}

func (log *zeroLogger) SetLevel(level logging.Level) {
	switch level {
	case 0: //PanicLevel
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case 1: //FatalLevel
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case 2: //ErrorLevel
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case 3: //WarnLevel
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case 4: //InfoLevel
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case 5: //DebugLevel
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case 6: //TraceLevel
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}
}

func (log *zeroLogger) GetLevel() logging.Level {
	level := zerolog.GlobalLevel()
	var l logging.Level
	switch level {
	case -1:
		l = logging.TraceLevel
	case 0:
		l = logging.DebugLevel
	case 1:
		l = logging.InfoLevel
	case 2:
		l = logging.WarnLevel
	case 3:
		l = logging.ErrorLevel
	case 4:
		l = logging.FatalLevel
	case 5:
		l = logging.PanicLevel
	}
	return l
}
