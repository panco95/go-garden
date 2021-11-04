package main

import (
	"github.com/panco95/go-garden/core"
	"github.com/panco95/go-garden/examples/user/api"
	"github.com/panco95/go-garden/examples/user/global"
	"github.com/panco95/go-garden/examples/user/rpc"
)

func main() {
	global.Garden = core.New()
	global.Garden.Run(api.Routes, new(rpc.Rpc), global.Garden.CheckCallSafeMiddleware)
}
