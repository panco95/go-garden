package core

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

type route struct {
	Path    string
	Limiter string
	Fusing  string
}

type Service struct {
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
}

type cfg struct {
	Service Service
	Routes  map[string]map[string]route
	Config  map[string]interface{}
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
	if err := viper.Unmarshal(&g.cfg); err != nil {
		g.Log(ErrorLevel, "Config", err)
	}
}

func (g *Garden) GetConfigValue(key string) interface{} {
	config := g.cfg.Config
	return config[key]
}

func (g *Garden) GetConfigValueMap(key string) map[string]interface{} {
	config := g.cfg.Config
	val, ok := config[strings.ToLower(key)]
	if ok {
		return val.(map[string]interface{})
	} else {
		return nil
	}
}

func (g *Garden) GetConfigValueString(key string) string {
	config := g.cfg.Config
	val, ok := config[strings.ToLower(key)]
	if ok {
		return val.(string)
	} else {
		return ""
	}
}

func (g *Garden) GetConfigValueStringSlice(key string) []string {
	config := g.cfg.Config
	val, ok := config[strings.ToLower(key)]
	if ok {
		return val.([]string)
	} else {
		return []string{}
	}
}

func (g *Garden) GetConfigValueInt(key string) int {
	config := g.cfg.Config
	val, ok := config[strings.ToLower(key)]
	if ok {
		return val.(int)
	} else {
		return 0
	}
}

func (g *Garden) GetConfigValueIntSlice(key string) []int {
	config := g.cfg.Config
	val, ok := config[strings.ToLower(key)]
	if ok {
		return val.([]int)
	} else {
		return []int{}
	}
}
