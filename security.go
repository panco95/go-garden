package garden

import (
	"strings"
)

func checkCallSafe(key string) bool {
	if strings.Compare(key, Config.CallServiceKey) != 0 {
		return false
	}
	return true
}
