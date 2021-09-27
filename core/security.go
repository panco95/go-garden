package core

import (
	"strings"
)

func (g *Garden) checkCallSafe(key string) bool {
	if strings.Compare(key, g.cfg.Service.CallKey) != 0 {
		return false
	}
	return true
}
