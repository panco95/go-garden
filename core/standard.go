package core

import (
	"github.com/streadway/amqp"
	clientV3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"sync"
)

type (
	logLevel int8
	// MapData map like any value datatype
	MapData map[string]interface{}
	// Garden go garden framework class
	Garden struct {
		isBootstrap    uint
		cfg            cfg
		services       map[string]*service
		serviceManager chan serviceOperate
		syncCache      []byte
		serviceId      string
		serviceIp      string
		log            *zap.SugaredLogger
		amqp           *amqp.Connection
		etcd           *clientV3.Client
		fusingMap      sync.Map
		limiterMap     sync.Map
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
	Timeout       = "Request timeout"
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
