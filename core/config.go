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
	Debug              bool
	ServiceName        string
	HttpOut            bool
	HttpPort           string
	AllowCors          bool
	RpcOut             bool
	RpcPort            string
	CallKey            string
	CallRetry          string
	EtcdKey            string
	EtcdAddress        []string
	TracerDrive        string
	ZipkinAddress      string
	JaegerAddress      string
	PushGatewayAddress string
}

type cfg struct {
	Service     serviceCfg
	Routes      map[string]map[string]routeCfg
	Config      map[string]interface{}
	runtimePath string
	configsPath string
}

func (g *Garden) initConfig(fileType string) {
	viper.AddConfigPath(g.cfg.configsPath)
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

// GetConfigValueInterface in config.yml: configs key
func (g *Garden) GetConfigValueInterface(key string) interface{} {
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

// GetConfigValueInt in config.yml: configs key
func (g *Garden) GetConfigValueInt(key string) int {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.(int)
	}
	return 0
}

// GetConfigValueFloat32 in config.yml: configs key
func (g *Garden) GetConfigValueFloat32(key string) float32 {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.(float32)
	}
	return 0
}

// GetConfigValueFloat64 in config.yml: configs key
func (g *Garden) GetConfigValueFloat64(key string) float64 {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.(float64)
	}
	return 0
}

// GetConfigValueBool in config.yml: configs key
func (g *Garden) GetConfigValueBool(key string) bool {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.(bool)
	}
	return false
}
