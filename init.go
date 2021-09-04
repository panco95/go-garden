package goms

import (
	"github.com/spf13/viper"
	"log"
)

func Init(rpcPort, httpPort, serviceName, projectName string) {
	InitLog()
	InitConfig("config/config.yml", "yml")
	InitProjectName(projectName)
	InitServiceId(ProjectName, rpcPort, httpPort, serviceName)

	etcdAddr := viper.GetStringSlice("etcdAddr")
	err := EtcdConnect(etcdAddr)
	if err != nil {
		log.Fatal("[etcd] " + err.Error())
	}
	err = ServiceRegister()
	if err != nil {
		log.Fatal("[service register] " + err.Error())
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
