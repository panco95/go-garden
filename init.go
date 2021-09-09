package goms

import (
	"github.com/spf13/viper"
	"goms/drives"
	"log"
)

// Init 启动一个服务的组件初始化封装
func Init(rpcPort, httpPort, serviceName, projectName string) {
	InitLog()
	InitConfig("configs/config.yml", "yml")

	etcdAddr := viper.GetStringSlice("etcdAddr")
	err := drives.EtcdConnect(etcdAddr)
	if err != nil {
		log.Fatal("[etcd] " + err.Error())
	}

	err = InitService(projectName, serviceName, httpPort, rpcPort)
	if err != nil {
		log.Fatal("[etcd] " + err.Error())
	}

	esAddr := viper.GetString("esAddr")
	err = drives.EsConnect(esAddr)
	if err != nil {
		log.Fatal("[elasticsearch] " + err.Error())
	}

	amqpAddr := viper.GetString("amqpAddr")
	err = drives.AmqpConnect(amqpAddr)
	if err != nil {
		log.Fatal("[amqp] " + err.Error())
	}
}
