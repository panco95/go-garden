package core

import (
	"github.com/panco95/go-garden/core/drives/etcd"
	"github.com/panco95/go-garden/core/log"
)

func (g *Garden) bootstrap(configPath, runtimePath string) {
	g.cfg.ConfigsPath = configPath
	g.cfg.RuntimePath = runtimePath
	g.bootConfig("yml")
	g.checkConfig()
	log.Setup(g.cfg.RuntimePath, g.cfg.Service.Debug)
	log.Info("bootstrap", g.cfg.Service.ServiceName+" running")
	g.bootEtcd()
	g.bootService()
	g.bootOpenTracing()
}

func (g *Garden) bootEtcd() {
	etcdC, err := etcd.Connect(g.cfg.Service.EtcdAddress, log.GetLogger().Desugar())
	if err != nil {
		log.Fatal("etcd", err)
	}
	g.setSafe("etcd", etcdC)
}

func (g *Garden) checkConfig() {
	if g.cfg.Service.ServiceName == "" {
		log.Fatal("config", "empty option serviceName")
	}
	if g.cfg.Service.HttpPort == "" {
		log.Fatal("config", "empty option httpPort")
	}
	if g.cfg.Service.RpcPort == "" {
		log.Fatal("config", "empty option httpPort")
	}
	if g.cfg.Service.CallKey == "" {
		log.Fatal("config", "empty option callKey")
	}
	if g.cfg.Service.CallRetry == "" {
		log.Fatal("config", "empty option callRetry")
	}
	if g.cfg.Service.EtcdKey == "" {
		log.Fatal("config", "empty option etcdKey")
	}
	if len(g.cfg.Service.EtcdAddress) == 0 {
		log.Fatal("config", "empty option etcdAddress")
	}
	if g.cfg.Service.TracerDrive == "zipkin" && g.cfg.Service.ZipkinAddress == "" {
		log.Fatal("config", "empty option zipkinAddress")
	}
	if g.cfg.Service.TracerDrive == "jaeger" && g.cfg.Service.JaegerAddress == "" {
		log.Fatal("config", "empty option jaegerAddress")
	}
}
