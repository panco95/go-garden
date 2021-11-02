package core

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/drives/db"
	"github.com/panco95/go-garden/drives/etcd"
	"github.com/panco95/go-garden/drives/redis"
)

// Run  http(gin) && rpc(rpcx)
func (g *Garden) Run(route func(r *gin.Engine), rpc interface{}, auth func() gin.HandlerFunc) {
	go func() {
		address := g.ServiceIp
		if g.cfg.Service.HttpOut {
			address = "0.0.0.0"
		}
		listenAddress := address + ":" + g.cfg.Service.HttpPort
		if err := g.ginListen(listenAddress, route, auth); err != nil {
			g.Log(FatalLevel, "ginRun", err)
		}
	}()

	go func() {
		address := g.ServiceIp
		if g.cfg.Service.RpcOut {
			address = "0.0.0.0"
		}
		rpcAddress := address + ":" + g.cfg.Service.RpcPort
		if err := g.rpcListen(g.cfg.Service.ServiceName, "tcp", rpcAddress, rpc, ""); err != nil {
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

	g.Log(InfoLevel, "bootstrap", g.cfg.Service.ServiceName+" service starting now...")

	var err error

	g.Etcd, err = etcd.Connect(g.cfg.Service.EtcdAddress)
	if err != nil {
		g.Log(FatalLevel, "etcd", err)
	}

	if err := g.initService(g.cfg.Service.ServiceName, g.cfg.Service.HttpPort, g.cfg.Service.RpcPort); err != nil {
		g.Log(FatalLevel, "init", err)
	}

	if err := g.initOpenTracing(g.cfg.Service.ServiceName, g.cfg.Service.ZipkinAddress, g.ServiceIp); err != nil {
		g.Log(FatalLevel, "openTracing", err)
	}

	if g.GetConfigValueBool("mysql_open") {
		g.Db, err = db.Connect(
			g.GetConfigValueString("mysql_user"),
			g.GetConfigValueString("mysql_pass"),
			g.GetConfigValueString("mysql_addr"),
			g.GetConfigValueString("mysql_dbname"),
			g.GetConfigValueString("mysql_charset"),
			g.GetConfigValueBool("mysql_parseTime"),
			g.GetConfigValueInt("mysql_connPool"),
		)
		if err != nil {
			g.Log(FatalLevel, "mysql", err)
		} else {
			g.Log(InfoLevel, "mysql", "Connect success")
		}
	}

	if g.GetConfigValueBool("redis_open") {
		g.Redis, err = redis.Connect(
			g.GetConfigValueString("redis_addr"),
			g.GetConfigValueString("redis_pass"),
			g.GetConfigValueInt("redis_db"),
		)
		if err != nil {
			g.Log(FatalLevel, "redis", err)
		} else {
			g.Log(InfoLevel, "redis", "Connect success")
		}
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
