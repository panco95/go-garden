package test

import (
	"sync/atomic"
	"testing"
)

func TestAtomic(t *testing.T) {
	var a int64 = 100
	atomic.AddInt64(&a, -1)
	atomic.AddInt64(&a, -3)
	atomic.AddInt64(&a, 2)
	t.Log(a)
}
