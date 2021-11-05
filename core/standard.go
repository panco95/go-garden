package core

import (
	"github.com/go-redis/redis/v8"
	clientV3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"sync"
)

type (
	logLevel int8
	// MapData map like any value datatype
	MapData map[string]interface{}
	// Garden go garden framework class
	Garden struct {
		isLogBootstrap    uint
		serviceType    uint //0:service 1:gateway
		cfg            cfg
		Services       map[string]*service
		serviceManager chan serviceOperate
		syncCache      []byte
		ServiceId      string
		ServiceIp      string
		log            *zap.SugaredLogger
		fusingMap      sync.Map
		limiterMap     sync.Map
		Etcd           *clientV3.Client
		Db             *gorm.DB
		Redis          *redis.Client
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

// RebootFunc if func panic
func (g *Garden) RebootFunc(label string, f func()) {
	defer func() {
		if err := recover(); err != nil {
			g.Log(ErrorLevel, label, err)
			f()
		}
	}()
	f()
}
