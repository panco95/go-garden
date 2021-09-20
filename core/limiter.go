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

var (
	limiterMap     map[string]*limiterData
	limiterMapLock sync.Mutex
)

func init() {
	limiterMap = make(map[string]*limiterData)
}

func limiterAnalyze(limiter string) (int, int, error) {
	arr := strings.Split(limiter, "/")
	if len(arr) != 2 {
		return 0, 0, errors.New("route limit format error")
	}
	second, err := strconv.Atoi(arr[0])
	if err != nil {
		return 0, 0, errors.New("route limit format error")
	}
	quantity, err := strconv.Atoi(arr[1])
	if err != nil {
		return 0, 0, errors.New("route limit format error")
	}
	return second, quantity, nil
}

func limiterInspect(path string, second, quantity int) bool {
	ld, ok := limiterMap[path]
	if !ok {
		ld = resetLimiterIndex(path)
	}

	now := time.Now().Unix()
	lost := int(now) - int(ld.StartTimestamp)
	if lost >= second {
		ld = resetLimiterIndex(path)
	}

	if ld.Quantity >= quantity {
		return false
	}

	ld.Lock.Lock()
	ld.Quantity++
	ld.Lock.Unlock()

	return true
}

func resetLimiterIndex(index string) *limiterData {
	limiterMapLock.Lock()
	limiterMap[index] = &limiterData{
		StartTimestamp: time.Now().Unix(),
		Quantity:       0,
	}
	limiterMapLock.Unlock()
	return limiterMap[index]
}
