package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	DefaultLogger Logger = newDefaultLogger()
)

type Logger interface {
	Debugw(info string, kvs ...interface{})
	Infow(info string, kvs ...interface{})
	Warnw(info string, kvs ...interface{})
	Errorw(info string, kvs ...interface{})
	Fatalw(info string, kvs ...interface{})
}

func newDefaultLogger() *zap.SugaredLogger {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewConsoleEncoder(cfg)
	core := zapcore.NewCore(encoder, os.Stdout, zap.InfoLevel)
	return zap.New(core).Sugar()
}
