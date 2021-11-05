package core

import (
	"errors"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type limiterData struct {
	StartTimestamp int64
	Quantity       int64
}

func limiterAnalyze(limiter string) (int64, int64, error) {
	arr := strings.Split(limiter, "/")
	if len(arr) != 2 {
		return 0, 0, errors.New("route limiter format error")
	}
	second, err := strconv.Atoi(arr[0])
	if err != nil {
		return 0, 0, errors.New("route limiter format error")
	}
	quantity, err := strconv.Atoi(arr[1])
	if err != nil {
		return 0, 0, errors.New("route limiter format error")
	}
	return int64(second), int64(quantity), nil
}

func (g *Garden) limiterInspect(path string, second, quantity int64) bool {
	l, ok := g.limiterMap.Load(path)
	if !ok {
		l = g.resetLimiterIndex(path)
	}
	ld := l.(*limiterData)

	now := time.Now().Unix()
	lost := now - ld.StartTimestamp
	if lost >= second {
		ld = g.resetLimiterIndex(path)
	}

	if atomic.LoadInt64(&ld.Quantity) >= quantity {
		return false
	}

	atomic.AddInt64(&ld.Quantity, 1)

	return true
}

func (g *Garden) resetLimiterIndex(index string) *limiterData {
	ld := limiterData{
		StartTimestamp: time.Now().Unix(),
		Quantity:       0,
	}
	g.limiterMap.Store(index, &ld)
	return &ld
}
