package core

import (
	"github.com/gin-gonic/gin"
)

// Run amqp and gin http server
func (g *Garden) Run(route func(r *gin.Engine), auth func() gin.HandlerFunc) {
	go func() {
		if err := g.amqpConsumer("fanout", "sync", "", "", g.syncAmqp); err != nil {
			g.Log(FatalLevel, "amqpConsumeRun", err)
		}
	}()

	address := g.serviceIp
	if g.cfg.Service.ListenOut {
		address = "0.0.0.0"
	}
	listenAddress := address + ":" + g.cfg.Service.ListenPort
	g.Log(FatalLevel, "Run", g.runGin(listenAddress, route, auth).Error())
}

func (g *Garden) bootstrap() {
	if g.isBootstrap == 1 {
		return
	}

	g.initConfig("configs", "yml")
	g.checkConfig()
	g.initLog()

	if err := g.connEtcd(g.cfg.Service.EtcdAddress); err != nil {
		g.Log(FatalLevel, "Etcd", err)
	}

	if err := g.connAmqp(g.cfg.Service.AmqpAddress); err != nil {
		g.Log(FatalLevel, "Amqp", err)
	}

	if err := g.initService(g.cfg.Service.ServiceName, g.cfg.Service.ListenPort); err != nil {
		g.Log(FatalLevel, "Init", err)
	}

	if err := g.initOpenTracing(g.cfg.Service.ServiceName, g.cfg.Service.ZipkinAddress, g.serviceIp+":"+g.cfg.Service.ListenPort); err != nil {
		g.Log(FatalLevel, "OpenTracing", err)
	}

	g.isBootstrap = 1
}

func (g *Garden) checkConfig() {
	if g.cfg.Service.ServiceName == "" {
		g.Log(FatalLevel, "Config", "empty option serviceName")
	}
	if g.cfg.Service.ListenPort == "" {
		g.Log(FatalLevel, "Config", "empty option listenPort")
	}
	if g.cfg.Service.CallKey == "" {
		g.Log(FatalLevel, "Config", "empty option callKey")
	}
	if g.cfg.Service.CallRetry == "" {
		g.Log(FatalLevel, "Config", "empty option callRetry")
	}
	if len(g.cfg.Service.EtcdAddress) == 0 {
		g.Log(FatalLevel, "Config", "empty option etcdAddress")
	}
	if g.cfg.Service.ZipkinAddress == "" {
		g.Log(FatalLevel, "Config", "empty option zipkinAddress")
	}
	if g.cfg.Service.AmqpAddress == "" {
		g.Log(FatalLevel, "Config", "empty option amqpAddress")
	}
}
