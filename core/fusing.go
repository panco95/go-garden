package core

import (
	"errors"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type fusingData struct {
	StartTimestamp int64
	Quantity       int64
}

func (g *Garden) fusingAnalyze(limiter string) (int64, int64, error) {
	arr := strings.Split(limiter, "/")
	if len(arr) != 2 {
		return 0, 0, errors.New("route fusing format error")
	}
	second, err := strconv.Atoi(arr[0])
	if err != nil {
		return 0, 0, errors.New("route fusing format error")
	}
	quantity, err := strconv.Atoi(arr[1])
	if err != nil {
		return 0, 0, errors.New("route fusing format error")
	}
	return int64(second), int64(quantity), nil
}

func (g *Garden) fusingInspect(path string, second, quantity int64) bool {
	f, ok := g.fusingMap.Load(path)
	if !ok {
		f = g.resetFusingIndex(path)
	}
	fd := f.(*fusingData)

	now := time.Now().Unix()
	lost := now - fd.StartTimestamp
	if lost >= second {
		fd = g.resetFusingIndex(path)
	}

	q := atomic.LoadInt64(&fd.Quantity)
	if q >= quantity {
		return false
	}

	return true
}

func (g *Garden) resetFusingIndex(index string) *fusingData {
	fd := fusingData{
		StartTimestamp: time.Now().Unix(),
		Quantity:       0,
	}
	g.fusingMap.Store(index, &fd)
	return &fd
}

func (g *Garden) addFusingQuantity(index string) {
	f, ok := g.fusingMap.Load(index)
	if !ok {
		f = g.resetFusingIndex(index)
	}
	fd := f.(*fusingData)

	atomic.AddInt64(&fd.Quantity, 1)
}
