package core

import (
	"github.com/gin-gonic/gin"
)

// Run  http(gin) && rpc(rpcx)
func (g *Garden) Run(route func(r *gin.Engine), rpc interface{}, auth func() gin.HandlerFunc) {
	g.Log(InfoLevel, "bootstrap", g.cfg.Service.ServiceName+" service starting now...")

	go func() {
		address := g.ServiceIp
		if g.cfg.Service.HttpOut {
			address = "0.0.0.0"
		}
		listenAddress := address + ":" + g.cfg.Service.HttpPort
		if err := g.runGin(listenAddress, route, auth); err != nil {
			g.Log(FatalLevel, "ginRun", err)
		}
	}()

	go func() {
		address := g.ServiceIp
		if g.cfg.Service.RpcOut {
			address = "0.0.0.0"
		}
		rpcAddress := address + ":" + g.cfg.Service.RpcPort
		if err := g.RpcListen(g.cfg.Service.ServiceName, "tcp", rpcAddress, rpc, ""); err != nil {
			g.Log(FatalLevel, "rpcRun", err)
		}
	}()

	forever := make(chan int, 0)
	<-forever
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

	if err := g.initService(g.cfg.Service.ServiceName, g.cfg.Service.HttpPort, g.cfg.Service.RpcPort); err != nil {
		g.Log(FatalLevel, "Init", err)
	}

	if err := g.initOpenTracing(g.cfg.Service.ServiceName, g.cfg.Service.ZipkinAddress, g.ServiceIp); err != nil {
		g.Log(FatalLevel, "OpenTracing", err)
	}

	g.isBootstrap = 1
}

func (g *Garden) checkConfig() {
	if g.cfg.Service.ServiceName == "" {
		g.Log(FatalLevel, "Config", "empty option serviceName")
	}
	if g.cfg.Service.HttpPort == "" {
		g.Log(FatalLevel, "Config", "empty option httpPort")
	}
	if g.cfg.Service.RpcPort == "" {
		g.Log(FatalLevel, "Config", "empty option httpPort")
	}
	if g.cfg.Service.CallKey == "" {
		g.Log(FatalLevel, "Config", "empty option callKey")
	}
	if g.cfg.Service.CallRetry == "" {
		g.Log(FatalLevel, "Config", "empty option callRetry")
	}
	if g.cfg.Service.EtcdKey == "" {
		g.Log(FatalLevel, "Config", "empty option etcdKey")
	}
	if len(g.cfg.Service.EtcdAddress) == 0 {
		g.Log(FatalLevel, "Config", "empty option etcdAddress")
	}
	if g.cfg.Service.ZipkinAddress == "" {
		g.Log(FatalLevel, "Config", "empty option zipkinAddress")
	}
}
