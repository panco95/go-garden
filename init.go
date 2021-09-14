package garden

import (
	"errors"
	"garden/drives/amqp"
	"garden/drives/etcd"
	"garden/drives/redis"
)

// Init 启动一个服务的组件初始化封装
func Init() {
	InitLog()
	InitConfig("configs", "yml")
	CheckConfig()

	if err := etcd.Connect(Config.EtcdAddress); err != nil {
		Fatal("Etcd", err)
	}

	if err := InitService(Config.ProjectName, Config.ServiceName, Config.HttpPort, Config.RpcPort); err != nil {
		Fatal("Init", err)
	}

	if err := InitOpenTracing(Config.ServiceName, Config.ZipkinAddress, GetOutboundIP()+":"+Config.HttpPort); err != nil {
		Fatal("OpenTracing", err)
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

// CheckConfig 检测配置是否合法
func CheckConfig() {
	if Config.ProjectName == "" {
		Fatal("Config", errors.New("config empty: ProjectName"))
	}
	if Config.ServiceName == "" {
		Fatal("Config", errors.New("config empty: ServiceName"))
	}
	if Config.HttpPort == "" {
		Fatal("Config", errors.New("config empty: HttpPort"))
	}
	if Config.RpcPort == "" {
		Fatal("Config", errors.New("config empty: RpcPort"))
	}
	if Config.CallServiceKey == "" {
		Fatal("Config", errors.New("config empty: CallServiceKey"))
	}
	if len(Config.EtcdAddress) == 0 {
		Fatal("Config", errors.New("config empty: EtcdAddress"))
	}
	if Config.ZipkinAddress == "" {
		Fatal("Config", errors.New("config empty: ZipkinAddress"))
	}
}
