package logger

import (
	"time"

	"lead_scrapper_be/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Panic(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	With(fields ...interface{}) Logger
	Sync() error
}

type zapLogger struct {
	logger *zap.SugaredLogger
}

// CustomTimeEncoder formats time in a human-readable format.
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339)) // You can change the format if needed
}

func NewZapLogger(c *config.Config) (Logger, error) {
	config := zap.NewProductionConfig()
	// if c.Log.Level != nil && *c.Log.Level != "" {
	// 	logLevel, err := zapcore.ParseLevel(*c.Log.Level)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	config.Level = zap.NewAtomicLevelAt(logLevel)
	// }
	config.Encoding = "console"
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = CustomTimeEncoder

	// if c.Log.Location != nil && *c.Log.Location != "" {
	// 	// Set the log file path
	// 	config.OutputPaths = []string{"stdout", *c.Log.Location}
	// 	config.ErrorOutputPaths = []string{"stderr", *c.Log.Location}
	// }

	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	sugar := logger.Sugar()

	return &zapLogger{
		logger: sugar,
	}, nil
}

func (l *zapLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debugw(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...interface{}) {
	l.logger.Infow(msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...interface{}) {
	l.logger.Warnw(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...interface{}) {
	l.logger.Errorw(msg, fields...)
}

func (l *zapLogger) Panic(msg string, fields ...interface{}) {
	l.logger.Panicw(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...interface{}) {
	l.logger.Fatalw(msg, fields...)
}

func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

func (l *zapLogger) With(fields ...interface{}) Logger {
	logger := l.logger.With(fields...)

	return &zapLogger{
		logger: logger,
	}
}
