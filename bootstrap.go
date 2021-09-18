package garden

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/drives/amqp"
	"github.com/panco95/go-garden/drives/etcd"
	"github.com/panco95/go-garden/drives/redis"
)

func Run(route func(r *gin.Engine), auth func() gin.HandlerFunc) {
	go runRpc(Config.RpcPort)
	Log(FatalLevel, "Run", runGin(Config.HttpPort, route, auth).Error())
}

func Init() {
	initConfig("configs", "yml")
	checkConfig()

	initLog()

	if err := etcd.Connect(Config.EtcdAddress); err != nil {
		Log(FatalLevel, "Etcd", err)
	}

	if err := initService(Config.ServiceName, Config.HttpPort, Config.RpcPort); err != nil {
		Log(FatalLevel, "Init", err)
	}

	if err := initOpenTracing(Config.ServiceName, Config.ZipkinAddress, serviceIp+":"+Config.HttpPort); err != nil {
		Log(FatalLevel, "OpenTracing", err)
	}

	if Config.RedisAddress != "" {
		if err := redis.Connect(Config.RedisAddress); err != nil {
			Log(FatalLevel, "Redis", err)
		}
	}

	if Config.AmqpAddress != "" {
		if err := amqp.Connect(Config.AmqpAddress); err != nil {
			Log(FatalLevel, "Amqp", err)
		}
	}

	if Config.ElasticsearchAddress != "" {
		if err := redis.Connect(Config.ElasticsearchAddress); err != nil {
			Log(FatalLevel, "Elasticsearch", err)
		}
	}
}

func checkConfig() {
	if Config.ServiceName == "" {
		Log(FatalLevel, "Config", "empty option ServiceName")
	}
	if Config.HttpPort == "" {
		Log(FatalLevel, "Config", "empty option HttpPort")
	}
	if Config.RpcPort == "" {
		Log(FatalLevel, "Config", "empty option RpcPort")
	}
	if Config.CallServiceKey == "" {
		Log(FatalLevel, "Config", "empty option CallServiceKey")
	}
	if len(Config.EtcdAddress) == 0 {
		Log(FatalLevel, "Config", "empty option EtcdAddress")
	}
	if Config.ZipkinAddress == "" {
		Log(FatalLevel, "Config", "empty option ZipkinAddress")
	}
}
