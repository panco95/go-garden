package core

import (
	"sync/atomic"
	"testing"
)

func TestRegistryDedupReconcileAndSelect(t *testing.T) {
	g := &Garden{services: map[string]*service{}}
	const keep = "10.0.0.1:8080:9000"
	const other = "10.0.0.2:8080:9000"

	g.servicesMu.Lock()
	if !g.addNodeLocked("svc", keep) {
		t.Fatal("first add failed")
	}
	if g.addNodeLocked("svc", keep) {
		t.Fatal("duplicate address was added")
	}
	g.servicesMu.Unlock()
	if len(g.services["svc"].Nodes) != 1 {
		t.Fatalf("nodes=%d", len(g.services["svc"].Nodes))
	}

	existing := g.services["svc"].Nodes[keep]
	atomic.AddInt64(&existing.Waiting, 3)

	g.servicesMu.Lock()
	g.reconcileServices(map[string]map[string]struct{}{
		"svc": {keep: {}, other: {}},
	})
	g.servicesMu.Unlock()

	if g.services["svc"].Nodes[keep] != existing {
		t.Fatal("resync replaced the live node")
	}
	if atomic.LoadInt64(&existing.Waiting) != 3 {
		t.Fatal("resync cleared Waiting")
	}

	busy, err := g.selectService("svc")
	if err != nil {
		t.Fatal(err)
	}
	if busy.Addr != other {
		t.Fatalf("selected %s", busy.Addr)
	}

	atomic.StoreInt64(&existing.Waiting, 0)
	seen := map[string]int{}
	for i := 0; i < 40; i++ {
		n, err := g.selectService("svc")
		if err != nil {
			t.Fatal(err)
		}
		seen[n.Addr]++
	}
	if seen[keep] == 0 || seen[other] == 0 {
		t.Fatalf("tie picks=%v", seen)
	}
}
