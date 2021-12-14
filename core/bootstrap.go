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

func (g *Garden) bootstrap(configPath, runtimePath string) {
	g.cfg.configsPath = configPath
	g.cfg.runtimePath = runtimePath
	g.initConfig("yml")
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

	if err := g.initOpenTracing(); err != nil {
		g.Log(FatalLevel, "openTracing", err)
	}

	dbConf := g.GetConfigValueMap("db")
	if dbConf != nil && dbConf["open"].(bool) {
		g.Db, err = db.Connect(dbConf, func(err interface{}) {
			g.Log(FatalLevel, "db", err)
		})
		if err != nil {
			g.Log(FatalLevel, "db", err)
		}
		g.Log(InfoLevel, "db", "Connect success")
	}

	redisConf := g.GetConfigValueMap("redis")
	if redisConf != nil && redisConf["open"].(bool) {
		g.Redis, err = redis.Connect(redisConf, func(err interface{}) {
			g.Log(FatalLevel, "redis", err)
		})
		if err != nil {
			g.Log(FatalLevel, "database", err)
		}
		g.Log(InfoLevel, "redis", "Connect success")
	}
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
	if g.cfg.Service.TracerDrive != "zipkin" && g.cfg.Service.TracerDrive != "jaeger" {
		g.Log(FatalLevel, "Config", "traceDrive just support zipkin or jaeger")
	}
	if g.cfg.Service.TracerDrive == "zipkin" && g.cfg.Service.ZipkinAddress == "" {
		g.Log(FatalLevel, "Config", "empty option zipkinAddress")
	}
	if g.cfg.Service.TracerDrive == "jaeger" && g.cfg.Service.JaegerAddress == "" {
		g.Log(FatalLevel, "Config", "empty option jaegerAddress")
	}
}
