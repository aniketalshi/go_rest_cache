package logging

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"fmt"
)

var logger *zap.Logger

type requestIDKey struct{}

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
    logger, err = cfg.Build()
	if err != nil {
		fmt.Printf("Unable to build logger : %s\n", err.Error())
		return
	}
}

// Logger returns the zap logger instance
func GetLogger() *zap.Logger {
	return logger
}

// WithReqID returns a conext attaching key to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// Logger retruns the logger with context info stored
func Logger(ctx context.Context) *zap.Logger {
	
	if ctx != nil {
		 // retreive the request id from context and attach that info to logger
		 if ctxRequestId, ok := ctx.Value(requestIDKey{}).(string); ok {
                logger = logger.With(zap.String("requestId", ctxRequestId))
         }
	}

	return logger
}
