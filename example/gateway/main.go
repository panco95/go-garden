package main

import (
	"flag"
	"go-ms/base"
)

var (
	httpPort = flag.String("http_port", "8080", "Http listen port")
	rpcPort  = flag.String("rpc_port", "8180", "Rpc listen port")
)

func main() {
	flag.Parse()
	base.Init(*rpcPort, *httpPort, "gateway")
	base.GinServer(*httpPort, "gateway", base.GatewayRoute)
}
