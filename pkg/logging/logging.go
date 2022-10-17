package logging

// Logger interface used as base logger throughout the library
type Logger interface {
	Print(err error, args ...any)
	Printf(format string, err error, args ...any)

	Trace(err error, args ...any)
	Tracef(format string, err error, args ...any)

	Debug(err error, args ...any)
	Debugf(format string, err error, args ...any)

	Info(err error, args ...any)
	Infof(format string, err error, args ...any)

	Warn(err error, args ...any)
	Warnf(format string, err error, args ...any)

	Error(err error, args ...any)
	Errorf(format string, err error, args ...any)

	Fatal(err error, args ...any)
	Fatalf(format string, err error, args ...any)

	Panic(err error, args ...any)
	Panicf(format string, err error, args ...any)

	WithPrefix(prefix string) Logger
	WithFields(fields Fields) Logger

	SetLevel(level Level)
	GetLevel() Level
}

type Loggable interface {
	Log() Logger
}

type Fields map[string]interface{}

// Level type
type Level uint8

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel = Level(iota)
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)
