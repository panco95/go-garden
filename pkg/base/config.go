package base

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go-ms/pkg/base/global"
	"log"
)

func LoadServices() {
	LoadConfig("config/services.yml", "yml")
}

func LoadConfig(filePath, fileType string) {
	viper.SetConfigType(fileType)
	viper.SetConfigFile(filePath)
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		global.Logger.Debugf("[Config] %s has changed", filePath)
	})
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("[Config] " + err.Error())
	}
}
