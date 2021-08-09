package main

import (
	"flag"
	"fmt"
	"go-ms/pkg/base"
	"go-ms/pkg/cluster"
	"log"
	"os"
	"runtime"
)

var (
	rpcPort  = flag.String("rpc_port", "9010", "Rpc listen port")
	etcdAddr = flag.String("etcd_addr", "127.0.0.1:2379", "Etcd address, cluster format: 127.0.0.1:2379|127.0.0.1:2389")
	version  = flag.Bool("version", false, "Show version info")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println("developing")
		os.Exit(0)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	var err error
	base.LogInit()
	err = cluster.EtcdRegister(*etcdAddr, *rpcPort, "user")
	if err != nil {
		log.Fatal("[Etcd register] ", err)
	}

	go base.RpcServer(*rpcPort, "user")

	forever := make(chan bool)
	<-forever
}
