package garden

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

var Config config

type config struct {
	Debug                bool
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

func initConfig(path, fileType string) {
	viper.AddConfigPath(path)
	viper.SetConfigType(fileType)

	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		Log(FatalLevel, "Config", err)
	}

	viper.SetConfigName("routes")
	if err := viper.MergeInConfig(); err != nil {
		Log(FatalLevel, "Config", err)
	}

	unmarshalConfig()

	// watch config file 'routes.yml' change
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		filename := filepath.Base(e.Name)
		if strings.Compare(filename, "routes.yml") == 0 {
			unmarshalConfig()
			go syncRoutes()
		}
	})
}

func unmarshalConfig() {
	if err := viper.Unmarshal(&Config); err != nil {
		Log(ErrorLevel, "Config", err)
	}
}
