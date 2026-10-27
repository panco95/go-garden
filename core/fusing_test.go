package core

import "testing"

func TestFusingOpensAtQuantity(t *testing.T) {
	g := &Garden{}
	if !g.fusingInspect("svc/ping", 60, 2) {
		t.Fatal("breaker closed an empty window")
	}
	if _, ok := g.fusingMap.Load("svc/ping"); !ok {
		t.Fatal("inspect did not open a window")
	}
	g.addFusingQuantity("svc/ping")
	if !g.fusingInspect("svc/ping", 60, 2) {
		t.Fatal("breaker closed before the limit")
	}
	g.addFusingQuantity("svc/ping")
	if g.fusingInspect("svc/ping", 60, 2) {
		t.Fatal("breaker stayed closed at the limit")
	}
}
