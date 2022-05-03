package core

import (
	"github.com/panco95/go-garden/core/drives/etcd"
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
	g.Log(1, "a", "a")
}

func (g *Garden) bootEtcd() {
	l, err := g.GetLog()
	if err != nil {
		g.Log(FatalLevel, "etcd GetLog", err)
	}
	etcdC, err := etcd.Connect(g.cfg.Service.EtcdAddress, l.Desugar())
	if err != nil {
		g.Log(FatalLevel, "etcd", err)
	}
	g.setSafe("etcd", etcdC)
}

func (g *Garden) checkConfig() {
	if g.cfg.Service.ServiceName == "" {
		g.Log(FatalLevel, "config", "empty option serviceName")
	}
	if g.cfg.Service.HttpPort == "" {
		g.Log(FatalLevel, "config", "empty option httpPort")
	}
	if g.cfg.Service.RpcPort == "" {
		g.Log(FatalLevel, "config", "empty option httpPort")
	}
	if g.cfg.Service.CallKey == "" {
		g.Log(FatalLevel, "config", "empty option callKey")
	}
	if g.cfg.Service.CallRetry == "" {
		g.Log(FatalLevel, "config", "empty option callRetry")
	}
	if g.cfg.Service.EtcdKey == "" {
		g.Log(FatalLevel, "config", "empty option etcdKey")
	}
	if len(g.cfg.Service.EtcdAddress) == 0 {
		g.Log(FatalLevel, "config", "empty option etcdAddress")
	}
	if g.cfg.Service.TracerDrive == "zipkin" && g.cfg.Service.ZipkinAddress == "" {
		g.Log(FatalLevel, "config", "empty option zipkinAddress")
	}
	if g.cfg.Service.TracerDrive == "jaeger" && g.cfg.Service.JaegerAddress == "" {
		g.Log(FatalLevel, "config", "empty option jaegerAddress")
	}
}
