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

func (g *Garden) initLog() {
	encoder := getEncoder()

	var cores []zapcore.Core

	writeSyncer := getLogWriter(g.cfg.runtimePath)
	fileCore := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	cores = append(cores, fileCore)

	if g.cfg.Service.Debug {
		consoleDebug := zapcore.Lock(os.Stdout)
		consoleCore := zapcore.NewCore(encoder, consoleDebug, zapcore.DebugLevel)
		cores = append(cores, consoleCore)
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	g.log = logger.Sugar()

	g.isLogBootstrap = 1
}

// Log write log file and print if debug is true
func (g *Garden) Log(level logLevel, label string, data interface{}) {
	format := logFormat(label, data)
	switch level {
	case DebugLevel:
		g.log.Debug(format)
	case InfoLevel:
		g.log.Info(format)
	case WarnLevel:
		g.log.Warn(format)
	case ErrorLevel:
		g.log.Errorf(format)
	case DPanicLevel:
		g.log.DPanic(format)
	case PanicLevel:
		g.log.Panic(format)
	case FatalLevel:
		if g.isLogBootstrap == 1 {
			g.log.Fatal(format)
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
	fmt.Printf(runtimePath)
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
