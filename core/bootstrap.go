package core

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core/drives/etcd"
)

func (g *Garden) Run(route func(r *gin.Engine), auth func() gin.HandlerFunc) {
	go g.runRpc(g.cfg.Service.RpcPort)
	g.Log(FatalLevel, "Run", g.runGin(g.cfg.Service.HttpPort, route, auth).Error())
}

func (g *Garden) bootstrap() {
	if g.isBootstrap == 1 {
		return
	}

	g.initConfig("configs", "yml")
	g.checkConfig()
	g.initLog()

	if err := etcd.Connect(g.cfg.Service.EtcdAddress); err != nil {
		g.Log(FatalLevel, "Etcd", err)
	}

	if err := g.initService(g.cfg.Service.ServiceName, g.cfg.Service.HttpPort, g.cfg.Service.RpcPort); err != nil {
		g.Log(FatalLevel, "Init", err)
	}

	if err := g.initOpenTracing(g.cfg.Service.ServiceName, g.cfg.Service.ZipkinAddress, g.serviceIp+":"+g.cfg.Service.HttpPort); err != nil {
		g.Log(FatalLevel, "OpenTracing", err)
	}

	g.isBootstrap = 1
}

func (g *Garden) checkConfig() {
	if g.cfg.Service.ServiceName == "" {
		g.Log(FatalLevel, "Config", "empty option ServiceName")
	}
	if g.cfg.Service.HttpPort == "" {
		g.Log(FatalLevel, "Config", "empty option HttpPort")
	}
	if g.cfg.Service.RpcPort == "" {
		g.Log(FatalLevel, "Config", "empty option RpcPort")
	}
	if g.cfg.Service.CallServiceKey == "" {
		g.Log(FatalLevel, "Config", "empty option CallServiceKey")
	}
	if len(g.cfg.Service.EtcdAddress) == 0 {
		g.Log(FatalLevel, "Config", "empty option EtcdAddress")
	}
	if g.cfg.Service.ZipkinAddress == "" {
		g.Log(FatalLevel, "Config", "empty option ZipkinAddress")
	}
}
