package main

import (
	"flag"
	"fmt"
	"go-ms/pkg/base"
	"go-ms/pkg/cluster"
	"log"
	"os"
)

var (
	httpPort = flag.String("http_port", "8080", "Http listen port")
	rpcPort  = flag.String("rpc_port", "8180", "Rpc listen port")
	etcdAddr = flag.String("etcd_addr", "127.0.0.1:2379", "Etcd address, cluster format: 127.0.0.1:2379|127.0.0.1:2389")
	version  = flag.Bool("version", false, "Show version info")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Print("developing")
		os.Exit(0)
	}

	base.Init()

	err := cluster.EtcdRegister(*etcdAddr, *rpcPort, *httpPort, "gateway")
	if err != nil {
		log.Fatal("[Etcd register] ", err)
	}

	go base.HttpServer(*httpPort, "gateway", cluster.GatewayRoute)

	forever := make(chan bool)
	<-forever
}
