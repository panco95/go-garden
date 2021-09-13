package goms

import (
	"goms/pkg/etcd"
	"goms/pkg/redis"
)

// Init 启动一个服务的组件初始化封装
func Init(rpcPort, httpPort, serviceName string) {
	InitLog()
	InitConfig("configs", "yml")

	err := etcd.Connect(Config.EtcdAddr)
	if err != nil {
		Fatal("Etcd", err)
	}

	err = InitService(Config.ProjectName, serviceName, httpPort, rpcPort)
	if err != nil {
		Fatal("InitService", err)
	}

	err = InitOpenTracing(serviceName, Config.ZipkinAddr, GetOutboundIP()+":"+httpPort)
	if err != nil {
		Fatal("Zipkin", err)
	}

	err = redis.Connect(Config.RedisAddr)
	if err != nil {
		Fatal("Redis", err)
	}
}
