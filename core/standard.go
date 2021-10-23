package core

import (
	"github.com/gin-gonic/gin"
	clientV3 "go.etcd.io/etcd/client/v3"
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
		isBootstrap    uint
		serviceType    uint //0:service 1:gateway
		cfg            cfg
		Services       map[string]*service
		serviceManager chan serviceOperate
		syncCache      []byte
		serviceId      string
		ServiceIp      string
		log            *zap.SugaredLogger
		etcd           *clientV3.Client
		fusingMap      sync.Map
		limiterMap     sync.Map
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

	CodeSuccess      = 0
	CodeFail         = 10001

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
