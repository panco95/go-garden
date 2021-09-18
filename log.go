package garden

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var logging *zap.SugaredLogger

type Level int8

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)

func initLog() {
	encoder := getEncoder()

	var cores []zapcore.Core

	writeSyncer := getLogWriter()
	fileCore := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	cores = append(cores, fileCore)

	if Config.Debug {
		consoleDebug := zapcore.Lock(os.Stdout)
		consoleCore := zapcore.NewCore(encoder, consoleDebug, zapcore.DebugLevel)
		cores = append(cores, consoleCore)
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	logging = logger.Sugar()
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

func Log(level Level, label string, data interface{}) {
	format := logFormat(label, data)
	switch level {
	case DebugLevel:
		logging.Debug(format)
	case InfoLevel:
		logging.Info(format)
	case WarnLevel:
		logging.Warn(format)
	case ErrorLevel:
		logging.Error(format)
	case DPanicLevel:
		logging.DPanic(format)
	case PanicLevel:
		logging.Panic(format)
	case FatalLevel:
		logging.Fatal(format)
	}
}
