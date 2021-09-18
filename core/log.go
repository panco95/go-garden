package core

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func (g *Garden) initLog() {
	encoder := getEncoder()

	var cores []zapcore.Core

	writeSyncer := getLogWriter()
	fileCore := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	cores = append(cores, fileCore)

	if g.Cfg.Debug {
		consoleDebug := zapcore.Lock(os.Stdout)
		consoleCore := zapcore.NewCore(encoder, consoleDebug, zapcore.DebugLevel)
		cores = append(cores, consoleCore)
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	g.log = logger.Sugar()
}

func (g *Garden) Log(level Level, label string, data interface{}) {
	format := logFormat(label, data)
	switch level {
	case DebugLevel:
		g.log.Debug(format)
	case InfoLevel:
		g.log.Info(format)
	case WarnLevel:
		g.log.Debug(format)
	case ErrorLevel:
		g.log.Debug(format)
	case DPanicLevel:
		g.log.Debug(format)
	case PanicLevel:
		g.log.Debug(format)
	case FatalLevel:
		g.log.Debug(format)
	}
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./runtime/logs/log.log",
		MaxSize:    2,
		MaxBackups: 10000,
		MaxAge:     180,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func logFormat(label string, log interface{}) string {
	e := fmt.Sprintf("[%s] %s", label, log)
	return e
}
