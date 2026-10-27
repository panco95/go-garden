package core

import (
	"github.com/panco95/go-garden/core/drives/etcd"
	"github.com/panco95/go-garden/core/log"
)

func (g *Garden) bootstrap(configPath, runtimePath string) error {
	g.cfg.ConfigsPath = configPath
	g.cfg.RuntimePath = runtimePath
	if err := g.bootConfig("yml"); err != nil {
		return err
	}
	if err := g.checkConfig(); err != nil {
		return err
	}
	log.Setup(g.cfg.RuntimePath, g.cfg.Service.Debug)
	g.bindRouteSync()
	log.Info("bootstrap", g.cfg.Service.ServiceName+" running")
	if err := g.bootEtcd(); err != nil {
		return err
	}
	if err := g.bootService(); err != nil {
		return err
	}
	return g.bootOpenTracing()
}

func (g *Garden) bootEtcd() error {
	etcdC, err := etcd.Connect(g.cfg.Service.EtcdAddress, log.GetLogger().Desugar())
	if err != nil {
		return err
	}
	g.setSafe("etcd", etcdC)
	return nil
}

func (g *Garden) checkConfig() error {
	if g.cfg.Service.ServiceName == "" {
		return emptyOption("serviceName")
	}
	if g.cfg.Service.HttpPort == "" {
		return emptyOption("httpPort")
	}
	if g.cfg.Service.RpcPort == "" {
		return emptyOption("rpcPort")
	}
	if g.cfg.Service.CallKey == "" {
		return emptyOption("callKey")
	}
	if g.cfg.Service.CallRetry == "" {
		return emptyOption("callRetry")
	}
	if g.cfg.Service.EtcdKey == "" {
		return emptyOption("etcdKey")
	}
	if len(g.cfg.Service.EtcdAddress) == 0 {
		return emptyOption("etcdAddress")
	}
	if g.cfg.Service.TracerDrive == "zipkin" && g.cfg.Service.ZipkinAddress == "" {
		return emptyOption("zipkinAddress")
	}
	if g.cfg.Service.TracerDrive == "jaeger" && g.cfg.Service.JaegerAddress == "" {
		return emptyOption("jaegerAddress")
	}
	return nil
}
