package garden

import (
	"garden/drives/amqp"
	"garden/drives/etcd"
	"garden/drives/redis"
)

// Init 启动一个服务的组件初始化封装
func Init() {
	InitLog()
	InitConfig("configs", "yml")

	if err := etcd.Connect(Config.EtcdAddress); err != nil {
		Fatal("Etcd", err)
	}

	if err := InitService(Config.ProjectName, Config.ServiceName, Config.HttpPort, Config.RpcPort); err != nil {
		Fatal("InitService", err)
	}

	if err := InitOpenTracing(Config.ServiceName, Config.ZipkinAddress, GetOutboundIP()+":"+Config.HttpPort); err != nil {
		Fatal("Zipkin", err)
	}

	if Config.RedisAddress != "" {
		if err := redis.Connect(Config.RedisAddress); err != nil {
			Fatal("Redis", err)
		}
	}

	if Config.AmqpAddress != "" {
		if err := amqp.Connect(Config.AmqpAddress); err != nil {
			Fatal("Amqp", err)
		}
	}

	if Config.ElasticsearchAddress != "" {
		if err := redis.Connect(Config.ElasticsearchAddress); err != nil {
			Fatal("Elasticsearch", err)
		}
	}
}
