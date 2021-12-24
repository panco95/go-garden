package core

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"time"
)

func (g *Garden) bootLog() {
	encoder := getEncoder()

	var cores []zapcore.Core

	writeSyncer := getLogWriter(g.cfg.RuntimePath)
	fileCore := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	cores = append(cores, fileCore)

	if g.cfg.Service.Debug {
		consoleDebug := zapcore.Lock(os.Stdout)
		consoleCore := zapcore.NewCore(encoder, consoleDebug, zapcore.DebugLevel)
		cores = append(cores, consoleCore)
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	l := logger.Sugar()
	g.setSafe("log", l)

	g.logBoot = 1
}

// Log format to write log file and print to shell if debug set true
func (g *Garden) Log(level logLevel, label string, data interface{}) {
	l := g.GetLog()
	format := logFormat(label, data)
	switch level {
	case DebugLevel:
		l.Debug(format)
	case InfoLevel:
		l.Info(format)
	case WarnLevel:
		l.Warn(format)
	case ErrorLevel:
		l.Errorf(format)
	case DPanicLevel:
		l.DPanic(format)
	case PanicLevel:
		l.Panic(format)
	case FatalLevel:
		if g.logBoot == 1 {
			l.Fatal(format)
		} else {
			log.Fatal(format)
		}
	}
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		TimeKey:      "time",
		LevelKey:     "level",
		NameKey:      "logger",
		CallerKey:    "caller",
		MessageKey:   "msg",
		EncodeLevel:  zapcore.LowercaseLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(runtimePath string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   runtimePath + "/logs/log.log",
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
