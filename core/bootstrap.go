package core

import (
	"github.com/panco95/go-garden/drives/db"
	"github.com/panco95/go-garden/drives/etcd"
	"github.com/panco95/go-garden/drives/redis"
)

func (g *Garden) bootstrap(configPath, runtimePath string) {
	g.cfg.configsPath = configPath
	g.cfg.runtimePath = runtimePath
	g.initConfig("yml")
	g.checkConfig()
	g.initLog()
	g.Log(InfoLevel, "bootstrap", g.cfg.Service.ServiceName+" service starting now...")
	g.bootEtcd()
	g.bootService(g.cfg.Service.ServiceName, g.cfg.Service.HttpPort, g.cfg.Service.RpcPort)
	g.bootOpenTracing()
	g.bootDb()
	g.bootRedis()
}

func (g *Garden) bootService(serviceName, httpPort, rpcPort string) {
	var err error
	g.Services = map[string]*service{}
	g.ServiceIp, err = getOutboundIP()
	if err != nil {
		g.Log(FatalLevel, "bootService", err)
	}
	g.ServiceId = g.cfg.Service.EtcdKey + "_" + serviceName + "_" + g.ServiceIp + ":" + httpPort + ":" + rpcPort

	g.serviceManager = make(chan serviceOperate, 0)
	go g.RebootFunc("serviceManageWatchReboot", func() {
		g.serviceManageWatch(g.serviceManager)
	})

	if err = g.serviceRegister(); err != nil {
		g.Log(FatalLevel, "bootService", err)
	}
}

func (g *Garden) bootOpenTracing() {
	var err error
	switch g.cfg.Service.TracerDrive {
	case "jaeger":
		err = connJaeger(g.cfg.Service.ServiceName, g.cfg.Service.JaegerAddress)
		break
	case "zipkin":
		err = connZipkin(g.cfg.Service.ServiceName, g.cfg.Service.ZipkinAddress, g.ServiceIp)
		break
	default:
		err = connZipkin(g.cfg.Service.ServiceName, g.cfg.Service.ZipkinAddress, g.ServiceIp)
		break
	}
	if err != nil {
		g.Log(FatalLevel, "openTracing", err)
	}
}

func (g *Garden) bootEtcd() {
	etcdC, err := etcd.Connect(g.cfg.Service.EtcdAddress)
	if err != nil {
		g.Log(FatalLevel, "etcd", err)
	}
	if err := g.Set("etcd", etcdC); err != nil {
		g.Log(FatalLevel, "etcd", err)
	}
}

func (g *Garden) bootDb() {
	dbConf := g.GetConfigValueMap("db")
	if dbConf != nil {
		dbC, err := db.Connect(dbConf)
		if err != nil {
			g.Log(FatalLevel, "db", err)
		}
		g.Log(InfoLevel, "db", "Connect success")
		if err := g.Set("db", dbC); err != nil {
			g.Log(FatalLevel, "db", err)
		}
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
		if err := g.Set("redis", redisC); err != nil {
			g.Log(FatalLevel, "redis", err)
		}
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
