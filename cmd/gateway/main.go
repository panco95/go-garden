package main

import (
	"flag"
	"go-ms/base"
)

var (
	httpPort = flag.String("http_port", "8080", "Http listen port")
	rpcPort  = flag.String("rpc_port", "8180", "Rpc listen port")
	etcdAddr = flag.String("etcd_addr", "127.0.0.1:2379", "Etcd address, cluster format: 127.0.0.1:2379|127.0.0.1:2389")
)

func main() {
	flag.Parse()
	base.Init(*etcdAddr, *rpcPort, *httpPort, "gateway")
	base.GinServer(*httpPort, "gateway", base.GatewayRoute)
}
