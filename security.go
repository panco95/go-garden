package goms

import (
	"github.com/spf13/viper"
	"strings"
)

func CheckCallSafe(key string) bool {
	if strings.Compare(key, viper.GetString("callServiceKey")) != 0 {
		return false
	}
	return true
}
