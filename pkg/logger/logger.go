package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type Logger interface {
    Info(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
    Sync() error
}

type Field = zapcore.Field

var (
    String   = zap.String
    Int      = zap.Int
    Duration = zap.Duration
    Error    = zap.Error
)

type logger struct {
    *zap.Logger
}

func New() Logger {
    config := zap.NewProductionConfig()
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

    zapLogger, err := config.Build()
    if err != nil {
        panic(err)
    }

    return &logger{zapLogger}
}