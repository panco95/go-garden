package main

import (
	"flag"
	"goms"
	"goms/example/gateway/middleware"
	"log"
)

var (
	httpPort = flag.String("http_port", "8080", "Http listen port")
	rpcPort  = flag.String("rpc_port", "8180", "Rpc listen port")
)

func main() {
	flag.Parse()
	goms.Init(*rpcPort, *httpPort, "gateway")
	log.Fatal(goms.GinServer(*httpPort, goms.GatewayRoute, middleware.Auth))
}
