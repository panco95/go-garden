package main

import (
	"flag"
	"goms"
	"log"
)

var (
	httpPort = flag.String("http_port", "8080", "Http listen port")
	rpcPort  = flag.String("rpc_port", "8180", "Rpc listen port")
)

func main() {
	flag.Parse()
	serviceName := "gateway"
	projectName := "goms"
	goms.Init(*rpcPort, *httpPort, serviceName, projectName)
	log.Fatal(goms.GinServer(*httpPort, serviceName, goms.GatewayRoute))
}
