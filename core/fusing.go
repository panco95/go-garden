package core

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

type fusingData struct {
	Lock           sync.Mutex
	StartTimestamp int64
	Quantity       int
}

var (
	fusingMap sync.Map
)

func fusingAnalyze(limiter string) (int, int, error) {
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
	return second, quantity, nil
}

func fusingInspect(path string, second, quantity int) bool {
	f, ok := fusingMap.Load(path)
	if !ok {
		f = resetFusingIndex(path)
	}
	fd := f.(*fusingData)

	now := time.Now().Unix()
	lost := int(now) - int(fd.StartTimestamp)
	if lost >= second {
		fd = resetFusingIndex(path)
	}

	if fd.Quantity >= quantity {
		return false
	}

	return true
}

func resetFusingIndex(index string) *fusingData {
	fd := fusingData{
		StartTimestamp: time.Now().Unix(),
		Quantity:       0,
	}
	fusingMap.Store(index, &fd)
	return &fd
}

func addFusingQuantity(index string) {
	f, ok := fusingMap.Load(index)
	if !ok {
		f = resetFusingIndex(index)
	}
	fd := f.(*fusingData)

	fd.Lock.Lock()
	fd.Quantity++
	fd.Lock.Unlock()
}
