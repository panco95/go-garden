package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-ms/pkg/base"
	"go-ms/pkg/base/global"
	"go-ms/pkg/cluster"
	"log"
	"net/http"
	"os"
)

var (
	rpcPort  = flag.String("rpc_port", "9010", "Rpc listen port")
	httpPort = flag.String("http_port", "9510", "Http listen port")
	etcdAddr = flag.String("etcd_addr", "127.0.0.1:2379", "Etcd address, cluster format: 127.0.0.1:2379|127.0.0.1:2389")
	version  = flag.Bool("version", false, "Show version info")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println("developing")
		os.Exit(0)
	}

	var err error
	base.LogInit()
	err = cluster.EtcdRegister(*etcdAddr, *rpcPort, *httpPort, "user")
	if err != nil {
		log.Fatal("[Etcd register] ", err)
	}

	go base.HttpServer(*httpPort, "user", route)

	forever := make(chan bool)
	<-forever
}

func route(r *gin.Engine) {
	r.POST("login", func(c *gin.Context) {
		jsonBody := global.Any{}
		c.Bind(&jsonBody)
		c.JSON(http.StatusOK, global.Any{
			"data": jsonBody,
		})
	})
	r.POST("register", func(c *gin.Context) {
		jsonBody := global.Any{}
		c.BindJSON(&jsonBody)
		c.JSON(http.StatusOK, global.Any{
			"data": jsonBody,
		})
	})
}
