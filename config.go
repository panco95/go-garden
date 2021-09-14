package garden

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
)

var Config config

type config struct {
	ProjectName          string
	ServiceName          string
	HttpPort             string
	RpcPort              string
	CallServiceKey       string
	EtcdAddress          []string
	ZipkinAddress        string
	RedisAddress         string
	ElasticsearchAddress string
	AmqpAddress          string
	Services             map[string]map[string]string
}

// InitConfig 初始化配置文件
func InitConfig(path, fileType string) {
	viper.AddConfigPath(path)
	viper.SetConfigType(fileType)

	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("[Config] %s", err)
	}

	viper.SetConfigName("services")
	if err := viper.MergeInConfig(); err != nil {
		log.Fatalf("[Config] %s", err)
	}

	//配置文件变化监听
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		UnmarshalConfig()
	})
	UnmarshalConfig()
}

// UnmarshalConfig 解析配置文件到结构体
func UnmarshalConfig() {
	if err := viper.Unmarshal(&Config); err != nil {
		e := fmt.Sprintf("unmarshal config error: %s", err)
		log.Print(e)
		Logger.Error(e)
	}
}
