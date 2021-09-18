package core

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

type Cfg struct {
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

func (g *Garden) initConfig(path, fileType string) {
	viper.AddConfigPath(path)
	viper.SetConfigType(fileType)

	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		g.Log(FatalLevel, "Config", err)
	}

	viper.SetConfigName("routes")
	if err := viper.MergeInConfig(); err != nil {
		g.Log(FatalLevel, "Config", err)
	}

	g.unmarshalConfig()

	// watch config file 'routes.yml' change
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		filename := filepath.Base(e.Name)
		if strings.Compare(filename, "routes.yml") == 0 {
			g.unmarshalConfig()
			go g.syncRoutes()
		}
	})
}

func (g *Garden) unmarshalConfig() {
	if err := viper.Unmarshal(&g.Cfg); err != nil {
		g.Log(ErrorLevel, "Config", err)
	}
}
