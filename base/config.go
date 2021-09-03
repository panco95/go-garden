package base

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
)

// InitConfig 初始化配置文件
func InitConfig(filePath, fileType string) {
	viper.SetConfigType(fileType)
	viper.SetConfigFile(filePath)
	viper.WatchConfig()
	//配置文件变化监听
	viper.OnConfigChange(func(e fsnotify.Event) {
		Logger.Debugf("[Config] %s has changed", filePath)
	})
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("[Config] " + err.Error())
	}
}
