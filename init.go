package goms

import (
	"github.com/spf13/viper"
	"goms/pkg/etcd"
	"goms/pkg/redis"
)

// Init 启动一个服务的组件初始化封装
func Init(rpcPort, httpPort, serviceName string) {
	InitLog()
	InitConfig("configs/config.yml", "yml")

	etcdAddr := viper.GetStringSlice("etcdAddr")
	err := etcd.Connect(etcdAddr)
	if err != nil {
		Fatal("Etcd", err)
	}

	err = InitService(viper.GetString("projectName"), serviceName, httpPort, rpcPort)
	if err != nil {
		Fatal("InitService", err)
	}

	zipkinAddr := viper.GetString("zipkinAddr")
	err = InitOpenTracing(serviceName, zipkinAddr, GetOutboundIP()+":"+httpPort)
	if err != nil {
		Fatal("Zipkin", err)
	}

	redisAddr := viper.GetString("redisAddr")
	err = redis.Connect(redisAddr)
	if err != nil {
		Fatal("Redis", err)
	}
}
