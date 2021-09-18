package core

import "go.uber.org/zap"

type (
	Level   int8
	MapData map[string]interface{}
	Garden  struct {
		Cfg            Cfg
		services       map[string]*service
		serviceManager chan serviceOperate
		syncCache      []byte
		serviceId      string
		serviceIp      string
		log            *zap.SugaredLogger
		isBootstrap    uint
	}
)

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)
