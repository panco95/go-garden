package base

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
)

func InitConfig(filePath, fileType string) {
	viper.SetConfigType(fileType)
	viper.SetConfigFile(filePath)
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		Logger.Debugf("[Config] %s has changed", filePath)
	})
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("[Config] " + err.Error())
	}
}
