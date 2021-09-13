package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"goms"
	"log"
)

var (
	httpPort = flag.String("http_port", "8080", "Http listen port")
	rpcPort  = flag.String("rpc_port", "8180", "Rpc listen port")
)

func main() {
	flag.Parse()
	goms.Init(*rpcPort, *httpPort, "gateway")
	log.Fatal(goms.GinServer(*httpPort, goms.GatewayRoute, auth))
}

func auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在这里写网关统一鉴权逻辑
		c.Next()
		log.Printf(c.Request.RequestURI)
	}
}
