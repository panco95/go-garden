package global

import (
	"github.com/panco95/go-garden/core"
	"sync"
)

var (
	Garden *core.Garden
	Users   sync.Map
)
