package core

import (
	clientV3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type (
	logLevel int8
	// MapData map like any value datatype
	MapData map[string]interface{}
	// Garden go garden framework class
	Garden struct {
		cfg            cfg
		services       map[string]*service
		serviceManager chan serviceOperate
		syncCache      []byte
		serviceId      string
		serviceIp      string
		log            *zap.SugaredLogger
		isBootstrap    uint
		etcd           *clientV3.Client
		remoteServer   remoteServer
	}
)

// log level
const (
	DebugLevel logLevel = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)

// error message
const (
	ServerError   = "Server Error"
	ServerLimiter = "Server limit flow"
	ServerFusing  = "Server fusing flow"
	NoAuth        = "No access permission"
	NotFound      = "The resource could not be found"
)
