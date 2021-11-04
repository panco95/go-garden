package core

import (
	"github.com/gin-gonic/gin"
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
	HttpOk       = http.StatusOK
	HttpFail     = http.StatusInternalServerError
	HttpNotFound = http.StatusNotFound

	InfoSuccess       = "Success"
	InfoInvalidParam  = "Invalid param"
	InfoServerError   = "Server Error"
	InfoServerLimiter = "Server limit flow"
	InfoServerFusing  = "Server fusing flow"
	InfoNoAuth        = "No access permission"
	InfoNotFound      = "The resource could not be found"
	InfoTimeout       = "Request timeout"
)

func Resp(c *gin.Context, code int, dataCode int, msg string, data interface{}) {
	c.JSON(code, MapData{
		"code": dataCode,
		"msg":  msg,
		"data": data,
	})
}

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
