package base

import (
	"github.com/spf13/viper"
	"log"
)

func Init(rpcPort, httpPort, serverName string) {
	InitLog()

	InitConfig("config.yml", "yml")

	InitServerId(ProjectName, rpcPort, httpPort, serverName)

	etcdAddr := viper.GetString("etcdAddr")
	if etcdAddr == "" {
		log.Fatal("[config.yml] etcdAddr is nil")
	}
	err := InitEtcd(etcdAddr, rpcPort, httpPort, serverName)
	if err != nil {
		log.Fatal("[etcd] " + err.Error())
	}

	esAddr := viper.GetString("esAddr")
	if esAddr == "" {
		log.Fatal("[config.yml] esAddr is nil")
	}
	err = InitEs(esAddr)
	if err != nil {
		log.Fatal("[elasticsearch] " + err.Error())
	}
}
