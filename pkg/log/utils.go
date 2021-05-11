package log

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLog() error {
	logConfig := LoadLogConfig()
	if logConfig.Level == "" {
		logConfig.Level = LevelInfo
	}

	var zapLoglevel zapcore.Level
	if errLogLevel := zapLoglevel.Set(logConfig.Level); errLogLevel != nil {
		return errors.Wrapf(errLogLevel, "cannot parse log level '%s'", logConfig.Level)
	}

	var zapLogConfig zap.Config
	switch logConfig.Type {
	case DevLogMode:
		zapLogConfig = zap.NewDevelopmentConfig()
	case ProdLogMode:
		zapLogConfig = zap.NewProductionConfig()
	default:
		logConfig.Type = ProdLogMode
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
