package main

import (
	"flag"
	"goms"
)

var (
	httpPort = flag.String("http_port", "8080", "Http listen port")
	rpcPort  = flag.String("rpc_port", "8180", "Rpc listen port")
)

func main() {
	flag.Parse()
	goms.Init(*rpcPort, *httpPort, "gateway", "goms")
	goms.GinServer(*httpPort, "gateway", goms.GatewayRoute)
}
