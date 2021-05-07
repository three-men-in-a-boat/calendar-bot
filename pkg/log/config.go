package log

import "os"

const (
	EnvLogType  = "LOG_TYPE"
	EnvLogLevel = "LOG_LEVEL"
)

const (
	DevLogMode  = "dev"
	ProdLogMode = "prod"
)

const (
	LevelDebug    = "debug"
	LevelInfo     = "info"
	LevelWarn     = "warn"
	LevelError    = "error"
	LevelDevPanic = "dpanic"
	LevelPanic    = "panic"
	LevelFatal    = "fatal"
)

type Config struct {
	Type  string `validate:"optional,in(dev|prod)"`
	Level string `validate:"-"`
}

func LoadLogConfig() Config {
	t := os.Getenv(EnvLogType)
	level := os.Getenv(EnvLogLevel)

	// TODO(nickeskov): validate struct

	return Config{
		Type:  t,
		Level: level,
	}
}

func (l *Config) ToEnv() map[string]string {
	return map[string]string{
		EnvLogType:  l.Type,
		EnvLogLevel: l.Level,
	}
}
