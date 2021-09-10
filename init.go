package goms

import (
	"github.com/spf13/viper"
	"goms/pkg/etcd"
	"log"
)

// Init 启动一个服务的组件初始化封装
func Init(rpcPort, httpPort, serviceName, projectName string) {
	InitLog()
	InitConfig("configs/config.yml", "yml")

	etcdAddr := viper.GetStringSlice("etcdAddr")
	err := etcd.EtcdConnect(etcdAddr)
	if err != nil {
		log.Fatal("[etcd] " + err.Error())
	}

	err = InitService(projectName, serviceName, httpPort, rpcPort)
	if err != nil {
		log.Fatal("[etcd] " + err.Error())
	}

	zipkinAddr := viper.GetString("zipkinAddr")
	err = InitOpenTracing(serviceName, zipkinAddr, GetOutboundIP()+":"+httpPort)
	if err != nil {
		log.Fatal("[openTracing] " + err.Error())
	}
}
