package core

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

type routeCfg struct {
	Type    string
	Path    string
	Limiter string
	Fusing  string
	Timeout int
}

type serviceCfg struct {
	Debug         bool
	ServiceName   string
	HttpOut     bool
	HttpPort    string
	RpcOut        bool
	RpcPort       string
	CallKey       string
	CallRetry     string
	EtcdKey       string
	EtcdAddress   []string
	ZipkinAddress string
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
			g.sendRoutes()
		}
	})
}

func (g *Garden) unmarshalConfig() {
	if err := viper.Unmarshal(&g.cfg); err != nil {
		g.Log(ErrorLevel, "Config", err)
	}
}

// GetConfigValue in config.yml: configs key
func (g *Garden) GetConfigValue(key string) interface{} {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val
	}
	return nil
}

// GetConfigValueMap in config.yml: configs key
func (g *Garden) GetConfigValueMap(key string) map[string]interface{} {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.(map[string]interface{})
	}
	return nil
}

// GetConfigValueString in config.yml: configs key
func (g *Garden) GetConfigValueString(key string) string {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.(string)
	}
	return ""
}

// GetConfigValueStringSlice in config.yml: configs key
func (g *Garden) GetConfigValueStringSlice(key string) []string {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.([]string)
	}
	return []string{}
}

// GetConfigValueInt in config.yml: configs key
func (g *Garden) GetConfigValueInt(key string) int {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.(int)
	}
	return 0
}

// GetConfigValueIntSlice in config.yml: configs key
func (g *Garden) GetConfigValueIntSlice(key string) []int {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.([]int)
	}
	return []int{}
}
