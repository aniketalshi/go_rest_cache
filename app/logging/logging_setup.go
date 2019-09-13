package logging

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

type requestIDKeyType int

const requestIDKey requestIDKeyType = iota

// initializes the default logger with custom options
func InitLogger() {

	// options to customize the logging output
	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	var err error
	initalLogger, err := cfg.Build()
	if err != nil {
		fmt.Printf("Unable to build logger : %s\n", err.Error())
		return
	}

	logger = initalLogger
}

// Logger returns the zap logger instance
func GetLogger() *zap.Logger {
	return logger
}

func NewContext(ctx context.Context, fields ...zap.Field) context.Context {

	logger := Logger(ctx).With(fields...)

	return context.WithValue(ctx, requestIDKey, logger)
}

func Logger(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return logger
	}

	if ctxLogger, ok := ctx.Value(requestIDKey).(*zap.Logger); ok {
		return ctxLogger
	}
	return logger
}
