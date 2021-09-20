package core

import "go.uber.org/zap"

type (
	logLevel int8
	MapData  map[string]interface{}
	Garden   struct {
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
	DebugLevel logLevel = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)

const (
	ServerError   = "Server Error"
	ServerLimiter = "Service limit flow"
	NoAuth        = "No access permission"
	NotFound      = "The resource could not be found"
)
