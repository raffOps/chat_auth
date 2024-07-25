package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func init() {
	var err error
	var config zap.Config
	var encoderConfig zapcore.EncoderConfig
	if os.Getenv("ENV") == "PRD" {
		config = zap.NewProductionConfig()
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig = encoderConfig

	Logger, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
}

func Info(message string, fields ...zap.Field) {
	Logger.Info(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	Logger.Debug(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	Logger.Error(message, fields...)
}

func Fatal(message string, fields ...zap.Field) {
	Logger.Fatal(message, fields...)
}

func String(key string, value string) zap.Field {
	return zap.String(key, value)
}

func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}
