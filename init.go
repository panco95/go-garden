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

	etcdAddr := viper.GetString("etcdAddr")
	if etcdAddr == "" {
		log.Fatal("[config.yml] etcdAddr is nil")
	}
	err := EtcdConnect(etcdAddr)
	if err != nil {
		log.Fatal("[etcd] " + err.Error())
	}

	esAddr := viper.GetString("esAddr")
	if esAddr == "" {
		log.Fatal("[config.yml] esAddr is nil")
	}
	err = EsConnect(esAddr)
	if err != nil {
		log.Fatal("[elasticsearch] " + err.Error())
	}
}