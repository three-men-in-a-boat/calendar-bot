package config

import "os"

const (
	EnvLogType  = "LOG_TYPE"
	EnvLogLevel = "LOG_LEVEL"
)

const (
	LogTypeDev  = "dev"
	LogTypeProd = "prod"
)

const (
	LogLevelDebug    = "debug"
	LogLevelInfo     = "info"
	LogLevelWarn     = "warn"
	LogLevelError    = "error"
	LogLevelDevPanic = "dev-panic"
	LogLevelPanic    = "panic"
	LogLevelFatal    = "fatal"
)

type Log struct {
	Type  string `validate:"optional,in(dev|prod)"`
	Level string `validate:"optional,in(debug|info|warn|error|dev-panic|panic|fatal)"`
}

func LoadLogConfig() Log {
	t := os.Getenv(EnvLogType)
	level := os.Getenv(EnvLogLevel)

	// TODO(nickeskov): validate struct

	return Log{
		Type:  t,
		Level: level,
	}
}

func (l *Log) ToEnv() map[string]string {
	return map[string]string{
		EnvLogType:  l.Type,
		EnvLogLevel: l.Level,
	}
}
