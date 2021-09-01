package base

import (
	"log"
)

func Init(etcdAddr, rpcPort, httpPort, serverName string) {
	InitLog()
	InitConfig("config/services.yml", "yml")
	InitServerId(ProjectName, rpcPort, httpPort, serverName)
	err := EtcdRegister(etcdAddr, rpcPort, httpPort, serverName)
	if err != nil {
		log.Fatal("[etcd register] " + err.Error())
	}
}
