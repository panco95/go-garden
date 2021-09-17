package garden

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

var Config config

type config struct {
	ServiceName          string
	HttpPort             string
	RpcPort              string
	CallServiceKey       string
	EtcdAddress          []string
	ZipkinAddress        string
	RedisAddress         string
	ElasticsearchAddress string
	AmqpAddress          string
	Routes               map[string]map[string]string
}

func InitConfig(path, fileType string) {
	viper.AddConfigPath(path)
	viper.SetConfigType(fileType)

	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		Fatal("Config", err)
	}

	viper.SetConfigName("routes")
	if err := viper.MergeInConfig(); err != nil {
		Fatal("Config", err)
	}

	UnmarshalConfig()

	// watch config file 'routes.yml' change
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		filename := filepath.Base(e.Name)
		if strings.Compare(filename,"routes.yml") == 0 {
			UnmarshalConfig()
			go SyncRoutes()
		}
	})
}

func UnmarshalConfig() {
	if err := viper.Unmarshal(&Config); err != nil {
		Fatal("Config", err)
	}
}
