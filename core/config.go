package core

import (
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
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
	ServiceIp          string
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
	RuntimePath string
	ConfigsPath string
}

//GetCfg instance to read
func (g *Garden) GetCfg() cfg {
	return g.cfg
}

func (g *Garden) bootConfig(fileType string) {
	viper.AddConfigPath(g.cfg.ConfigsPath)
	viper.SetConfigType(fileType)

	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		g.Log(FatalLevel, "config", err)
	}

	viper.SetConfigName("routes")
	if err := viper.MergeInConfig(); err != nil {
		g.Log(FatalLevel, "config", err)
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
		g.Log(ErrorLevel, "config", err)
	}
}

// GetConfigValueInterface to read as interface{} datatype
func (g *Garden) GetConfigValueInterface(key string) interface{} {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val
	}
	return nil
}

// GetConfigValueMap to read as map[string]interface{} datatype
func (g *Garden) GetConfigValueMap(key string) map[string]interface{} {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.(map[string]interface{})
	}
	return nil
}

// GetConfigValueString to read as string datatype
func (g *Garden) GetConfigValueString(key string) string {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.(string)
	}
	return ""
}

// GetConfigValueInt to read as int datatype
func (g *Garden) GetConfigValueInt(key string) int {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.(int)
	}
	return 0
}

// GetConfigValueFloat32 to read as float32 datatype
func (g *Garden) GetConfigValueFloat32(key string) float32 {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.(float32)
	}
	return 0
}

// GetConfigValueFloat64 to read as float64 datatype
func (g *Garden) GetConfigValueFloat64(key string) float64 {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.(float64)
	}
	return 0
}

// GetConfigValueBool to read as bool datatype
func (g *Garden) GetConfigValueBool(key string) bool {
	config := g.cfg.Config
	if val, ok := config[strings.ToLower(key)]; ok {
		return val.(bool)
	}
	return false
}
