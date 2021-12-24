package core

import (
	"github.com/panco95/go-garden/drives/db"
	"github.com/panco95/go-garden/drives/etcd"
	"github.com/panco95/go-garden/drives/redis"
)

func (g *Garden) bootstrap(configPath, runtimePath string) {
	g.cfg.ConfigsPath = configPath
	g.cfg.RuntimePath = runtimePath
	g.bootConfig("yml")
	g.checkConfig()
	g.bootLog()
	g.Log(InfoLevel, "bootstrap", g.cfg.Service.ServiceName+" running")
	g.bootEtcd()
	g.bootService()
	g.bootOpenTracing()
	g.bootDb()
	g.bootRedis()
}

func (g *Garden) bootEtcd() {
	etcdC, err := etcd.Connect(g.cfg.Service.EtcdAddress)
	if err != nil {
		g.Log(FatalLevel, "etcd", err)
	}
	g.setSafe("etcd", etcdC)
}

func (g *Garden) bootDb() {
	dbConf := g.GetConfigValueMap("db")
	if dbConf != nil {
		dbC, err := db.Connect(dbConf)
		if err != nil {
			g.Log(FatalLevel, "db", err)
		}
		g.Log(InfoLevel, "db", "Connect success")
		g.setSafe("db", dbC)
	}
}

func (g *Garden) bootRedis() {
	redisConf := g.GetConfigValueMap("redis")
	if redisConf != nil {
		redisC, err := redis.Connect(redisConf, func(err interface{}) {
			g.Log(FatalLevel, "redis", err)
		})
		if err != nil {
			g.Log(FatalLevel, "database", err)
		}
		g.Log(InfoLevel, "redis", "Connect success")
		g.setSafe("redis", redisC)
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
