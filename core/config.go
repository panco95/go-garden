package core

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

type routeCfg struct {
	Path    string
	Limiter string
	Fusing  string
}

type serviceCfg struct {
	Debug                bool
	ServiceName          string
	HttpPort             string
	RpcPort              string
	CallServiceKey       string
	EtcdAddress          []string
	ZipkinAddress        string
}

type cfg struct {
	Service serviceCfg
	Routes  map[string]map[string]routeCfg
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
			g.syncRoutes()
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
	val, ok := config[strings.ToLower(key)]
	if !ok {
		return nil
	}
	return val
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
