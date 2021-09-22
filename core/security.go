package core

import (
	"strings"
)

func (g *Garden) checkCallSafe(key string) bool {
	if strings.Compare(key, g.cfg.Service.CallServiceKey) != 0 {
		return false
	}
	return true
}
