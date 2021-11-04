package main

import (
	"github.com/panco95/go-garden/core"
	"github.com/panco95/go-garden/examples/user2/api"
	"github.com/panco95/go-garden/examples/user2/global"
	"github.com/panco95/go-garden/examples/user2/rpc"
)

func main() {
	global.Garden = core.New()
	global.Garden.Run(api.Routes, new(rpc.Rpc), global.Garden.CheckCallSafeMiddleware)
}
