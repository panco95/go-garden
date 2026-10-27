package core

import (
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/panco95/go-garden/core/log"
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
	RuntimePath string
	ConfigsPath string
}

func (g *Garden) GetCfg() cfg {
	return g.cfg
}

func (g *Garden) bootConfig(fileType string) error {
	viper.AddConfigPath(g.cfg.ConfigsPath)
	viper.SetConfigType(fileType)

	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	viper.SetConfigName("routes")
	if err := viper.MergeInConfig(); err != nil {
		return err
	}

	if err := g.unmarshalConfig(); err != nil {
		return err
	}

	// watch config file 'routes.yml' change
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		filename := filepath.Base(e.Name)
		if strings.Compare(filename, "routes.yml") == 0 {
			if err := g.unmarshalConfig(); err != nil {
				log.Error("config", err)
				return
			}
			g.sendRoutes()
		}
	})
	return nil
}

func (g *Garden) unmarshalConfig() error {
	return viper.Unmarshal(&g.cfg)
}
