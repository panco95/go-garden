package main

import (
	"github.com/panco95/go-garden/core"
	"github.com/panco95/go-garden/examples/gateway/auth"
	"github.com/panco95/go-garden/examples/gateway/global"
	"github.com/panco95/go-garden/examples/gateway/rpc"
)

func main() {
	global.Service = core.New()
	global.Service.Run(global.Service.GatewayRoute, new(rpc.Rpc), auth.Auth)
}
