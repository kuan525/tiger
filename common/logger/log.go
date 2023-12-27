package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/kuan525/tiger/common/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	// Lumberjack 是一个用于将日志写入滚动文件的 Go 包。
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger = &Logger{}

type Logger struct {
	Options
	logger *zap.Logger
}

func NewLogger(opts ...Option) {
	logger.Options = defaultOptions
	for _, o := range opts {
		o.apply(&logger.Options)
	}

	fileWriteSyncer := logger.getFileLogWriter()
	var core zapcore.Core
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	if config.IsDebug() {
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
			zapcore.NewCore(encoder, fileWriteSyncer, zapcore.DebugLevel),
		)
	} else {
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, fileWriteSyncer, zapcore.InfoLevel),
		)
	}
	logger.logger = zap.New(core, zap.WithCaller(true), zap.AddCallerSkip(logger.callerSkip))
}

func (l *Logger) getFileLogWriter() (writeSyncer zapcore.WriteSyncer) {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/%s", l.logDir, l.fileName),
		MaxSize:    l.maxSize,
		MaxBackups: l.maxBackups,
		MaxAge:     l.maxAge,
		Compress:   l.compress,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func DebugCtx(ctx context.Context, message string, fields ...zap.Field) {
	logger.logger.With(zap.String(traceID, GetTraceID(ctx))).Debug(message, fields...)
}

func InfoCtx(ctx context.Context, message string, fields ...zap.Field) {
	logger.logger.With(zap.String(traceID, GetTraceID(ctx))).Info(message, fields...)
}

func WarnCtx(ctx context.Context, message string, fields ...zap.Field) {
	logger.logger.With(zap.String(traceID, GetTraceID(ctx))).Warn(message, fields...)
}

func ErrorCtx(ctx context.Context, message string, fields ...zap.Field) {
	logger.logger.With(zap.String(traceID, GetTraceID(ctx))).Error(message, fields...)
}

func DPanicCtx(ctx context.Context, message string, fields ...zap.Field) {
	logger.logger.With(zap.String(traceID, GetTraceID(ctx))).DPanic(message, fields...)
}

func PanicCtx(ctx context.Context, message string, fields ...zap.Field) {
	logger.logger.With(zap.String(traceID, GetTraceID(ctx))).Panic(message, fields...)
}

func FatalCtx(ctx context.Context, message string, fields ...zap.Field) {
	logger.logger.With(zap.String(traceID, GetTraceID(ctx))).Fatal(message, fields...)
}
