package core

import "crypto/subtle"

func (g *Garden) checkCallSafe(key string) bool {
	got := []byte(key)
	want := []byte(g.cfg.Service.CallKey)
	return subtle.ConstantTimeCompare(got, want) == 1
}
