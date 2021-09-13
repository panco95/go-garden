package main

import (
	"flag"
	"goms"
	"goms/example/services/order/route"
	"log"
)

var (
	rpcPort  = flag.String("rpc_port", "9010", "Rpc listen port")
	httpPort = flag.String("http_port", "9510", "Http listen port")
)

func main() {
	flag.Parse()
	goms.Init(*rpcPort, *httpPort, "order")
	log.Fatal(goms.GinServer(*httpPort, route.Route, nil))
}
