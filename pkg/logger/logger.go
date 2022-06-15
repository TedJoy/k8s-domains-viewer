package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func New(name string, debug bool) *zap.SugaredLogger {

	logger, _ := newDevelopmentLogger()

	if !debug {
		logger, _ = zap.NewProduction()
	}

	logger = logger.Named(name)

	defer logger.Sync()

	return logger.Sugar()
}

// newDevelopmentLogger will setup a new Development Logger
func newDevelopmentLogger() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	return cfg.Build(zap.AddCaller())
}
