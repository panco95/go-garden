package main

import (
	"github.com/panco95/go-garden/core"
	"github.com/panco95/go-garden/examples/pay/api"
	"github.com/panco95/go-garden/examples/pay/global"
	"github.com/panco95/go-garden/examples/pay/rpc"
)

func main() {
	global.Garden = core.New()
	global.Garden.Run(api.Routes, new(rpc.Rpc), global.Garden.CheckCallSafeMiddleware)
}
