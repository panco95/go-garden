package core

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

type limiterData struct {
	Lock           sync.Mutex
	StartTimestamp int64
	Quantity       int
}

func limiterAnalyze(limiter string) (int, int, error) {
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
	return second, quantity, nil
}

func (g *Garden)limiterInspect(path string, second, quantity int) bool {
	l, ok := g.limiterMap.Load(path)
	if !ok {
		l = g.resetLimiterIndex(path)
	}
	ld := l.(*limiterData)

	now := time.Now().Unix()
	lost := int(now) - int(ld.StartTimestamp)
	if lost >= second {
		ld = g.resetLimiterIndex(path)
	}

	if ld.Quantity >= quantity {
		return false
	}

	ld.Lock.Lock()
	ld.Quantity++
	ld.Lock.Unlock()

	return true
}

func (g *Garden)resetLimiterIndex(index string) *limiterData {
	ld := limiterData{
		StartTimestamp: time.Now().Unix(),
		Quantity:       0,
	}
	g.limiterMap.Store(index, &ld)
	return &ld
}
