package goms

import (
	"github.com/spf13/viper"
	"log"
)

func Init(rpcPort, httpPort, serviceName, projectName string) {
	InitLog()
	InitConfig("config/config.yml", "yml")

	etcdAddr := viper.GetStringSlice("etcdAddr")
	err := EtcdConnect(etcdAddr)
	if err != nil {
		log.Fatal("[etcd] " + err.Error())
	}

	err = InitService(projectName, serviceName, httpPort, rpcPort)
	if err != nil {
		log.Fatal("[etcd] " + err.Error())
	}

	esAddr := viper.GetString("esAddr")
	err = EsConnect(esAddr)
	if err != nil {
		log.Fatal("[elasticsearch] " + err.Error())
	}

	amqpAddr := viper.GetString("amqpAddr")
	err = AmqpConnect(amqpAddr)
	if err != nil {
		log.Fatal("[amqp] " + err.Error())
	}
}
