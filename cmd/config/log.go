package config

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

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
	LogLevelDevPanic = "dpanic"
	LogLevelPanic    = "panic"
	LogLevelFatal    = "fatal"
)

type Log struct {
	Type  string `validate:"optional,in(dev|prod)"`
	Level string `validate:"-"`
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

func InitLog() error {
	logConfig := LoadLogConfig()
	if logConfig.Level == "" {
		logConfig.Level = LogLevelInfo
	}

	var zapLoglevel zapcore.Level
	if errLogLevel := zapLoglevel.Set(logConfig.Level); errLogLevel != nil {
		return errors.Wrapf(errLogLevel, "cannot parse log level '%s'", logConfig.Level)
	}

	var zapLogConfig zap.Config
	switch logConfig.Type {
	case LogTypeDev:
		zapLogConfig = zap.NewDevelopmentConfig()
	case LogTypeProd:
		zapLogConfig = zap.NewProductionConfig()
	default:
		logConfig.Type = LogTypeProd
		zapLogConfig = zap.NewProductionConfig()
	}

	zapLogConfig.Level.SetLevel(zapLoglevel)

	logger, err := zapLogConfig.Build()
	if err != nil {
		return errors.Wrapf(err, "cannot build '%s' logger", logConfig.Type)
	}
	zap.ReplaceGlobals(logger)

	return nil
}
