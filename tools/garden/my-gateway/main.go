package main

import (
	"github.com/panco95/go-garden/core"
	"my-gateway/api"
	"my-gateway/global"
	"my-gateway/rpc"
)

func main() {
	global.Garden = core.New()
	global.Garden.Run(api.Routes, new(rpc.Rpc), global.Garden.CheckCallSafeMiddleware)
}
