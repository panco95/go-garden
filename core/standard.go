package core

import (
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

type (
	logLevel int8
	// MapData map like any value datatype
	MapData map[string]interface{}
	// Garden go garden framework class
	Garden struct {
		isLogBootstrap uint
		serviceType    uint //0:service 1:gateway
		cfg            cfg
		Services       map[string]*service
		serviceManager chan serviceOperate
		syncCache      []byte
		log            *zap.SugaredLogger
		fusingMap      sync.Map
		limiterMap     sync.Map
		ServiceIp      string
		ServiceId      string

		//Etcd           *clientV3.Client
		//Db             *gorm.DB
		//Redis          *redis.Client
		Metrics        sync.Map
		RequestProcess atomic.Int64
		RequestFinish  atomic.Int64

		container sync.Map
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
