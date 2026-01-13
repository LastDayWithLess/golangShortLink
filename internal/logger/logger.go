package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	log *zap.Logger
}

func NewLogger(logFile string) (*Logger, error) {
	config := zap.NewProductionConfig()

	if logFile == "" {
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
	} else {
		config.OutputPaths = []string{logFile}
		config.ErrorOutputPaths = []string{logFile}
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	cfg, err := config.Build()

	return &Logger{log: cfg}, err
}

func String(key string, value string) zap.Field {
	return zap.String(key, value)
}

func ErrorField(err error) zap.Field {
	return zap.Error(err)
}

func (l *Logger) Error(err string, field ...zap.Field) {
	l.log.Error(err, field...)
}

func (l *Logger) Info(msg string, field ...zap.Field) {
	l.log.Info(msg, field...)
}

func (l *Logger) Close() {
	_ = l.log.Sync()
}
