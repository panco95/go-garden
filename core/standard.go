package core

import "go.uber.org/zap"

type (
	logLevel int8
	MapData  map[string]interface{}
	Garden   struct {
		cfg            cfg
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
	ServerLimiter = "Server limit flow"
	ServerFusing  = "Server fusing flow"
	NoAuth        = "No access permission"
	NotFound      = "The resource could not be found"
)
