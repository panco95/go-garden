package core

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/drives/amqp"
	"github.com/panco95/go-garden/drives/etcd"
	"github.com/panco95/go-garden/drives/redis"
)

func (g *Garden) Run(route func(r *gin.Engine), auth func() gin.HandlerFunc) {
	go g.runRpc(g.Cfg.RpcPort)
	g.Log(FatalLevel, "Run", g.runGin(g.Cfg.HttpPort, route, auth).Error())
}

func (g *Garden) Bootstrap() {
	if g.isBootstrap == 1 {
		return
	}

	g.initConfig("configs", "yml")
	g.checkConfig()
	g.initLog()

	if err := etcd.Connect(g.Cfg.EtcdAddress); err != nil {
		g.Log(FatalLevel, "Etcd", err)
	}

	if err := g.initService(g.Cfg.ServiceName, g.Cfg.HttpPort, g.Cfg.RpcPort); err != nil {
		g.Log(FatalLevel, "Init", err)
	}

	if err := g.initOpenTracing(g.Cfg.ServiceName, g.Cfg.ZipkinAddress, g.serviceIp+":"+g.Cfg.HttpPort); err != nil {
		g.Log(FatalLevel, "OpenTracing", err)
	}

	if g.Cfg.RedisAddress != "" {
		if err := redis.Connect(g.Cfg.RedisAddress); err != nil {
			g.Log(FatalLevel, "Redis", err)
		}
	}

	if g.Cfg.AmqpAddress != "" {
		if err := amqp.Connect(g.Cfg.AmqpAddress); err != nil {
			g.Log(FatalLevel, "Amqp", err)
		}
	}

	if g.Cfg.ElasticsearchAddress != "" {
		if err := redis.Connect(g.Cfg.ElasticsearchAddress); err != nil {
			g.Log(FatalLevel, "Elasticsearch", err)
		}
	}

	g.isBootstrap = 1
}

func (g *Garden) checkConfig() {
	if g.Cfg.ServiceName == "" {
		g.Log(FatalLevel, "Config", "empty option ServiceName")
	}
	if g.Cfg.HttpPort == "" {
		g.Log(FatalLevel, "Config", "empty option HttpPort")
	}
	if g.Cfg.RpcPort == "" {
		g.Log(FatalLevel, "Config", "empty option RpcPort")
	}
	if g.Cfg.CallServiceKey == "" {
		g.Log(FatalLevel, "Config", "empty option CallServiceKey")
	}
	if len(g.Cfg.EtcdAddress) == 0 {
		g.Log(FatalLevel, "Config", "empty option EtcdAddress")
	}
	if g.Cfg.ZipkinAddress == "" {
		g.Log(FatalLevel, "Config", "empty option ZipkinAddress")
	}
}
