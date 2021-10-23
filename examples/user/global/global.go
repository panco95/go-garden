package global

import (
	"github.com/panco95/go-garden/core"
	"sync"
)

var (
	Service *core.Garden
	Users   sync.Map
)
