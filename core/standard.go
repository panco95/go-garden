package core

import (
	"net/http"
	"sync"
)

type (
	logLevel int8
	// MapData map like any value datatype
	MapData map[string]interface{}
	// Garden go garden framework class
	Garden struct {
		container      sync.Map
		cfg            cfg
		logBoot        uint
		serviceType    uint //0:service 1:gateway
		services       map[string]*service
		serviceManager chan serviceOperate
		syncCache      []byte
		fusingMap      sync.Map
		limiterMap     sync.Map

		metrics        sync.Map
		requestProcess int64
		requestFinish  int64
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
	httpOk       = http.StatusOK
	httpFail     = http.StatusInternalServerError
	httpNotFound = http.StatusNotFound

	infoSuccess       = "Success"
	infoServerError   = "Server Error"
	infoServerLimiter = "Server limit flow"
	infoServerFusing  = "Server fusing flow"
	infoNoAuth        = "No access permission"
	infoNotFound      = "The resource could not be found"
	infoTimeout       = "Request timeout"
)
