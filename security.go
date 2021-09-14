package garden

import (
	"strings"
)

func CheckCallSafe(key string) bool {
	if strings.Compare(key, Config.CallServiceKey) != 0 {
		return false
	}
	return true
}
