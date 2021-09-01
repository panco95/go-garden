package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"go-ms/base"
	"net/http"
)

var (
	rpcPort  = flag.String("rpc_port", "9010", "Rpc listen port")
	httpPort = flag.String("http_port", "9510", "Http listen port")
	etcdAddr = flag.String("etcd_addr", "127.0.0.1:2379", "Etcd address, cluster format: 127.0.0.1:2379|127.0.0.1:2389")
)

func main() {
	flag.Parse()
	base.Init(*etcdAddr, *rpcPort, *httpPort, "user")
	base.GinServer(*httpPort, "user", route)
}

func route(r *gin.Engine) {
	c := r.Use(base.CheckCallServiceKey)
	c.Any("login", func(c *gin.Context) {
		c.JSON(http.StatusOK, base.Any{
			"code": 0,
			"msg":  "success",
			"data": base.Any{
				"method":   base.GetMethod(c),
				"urlParam": base.GetUrlParam(c),
				"headers":  base.GetHeaders(c),
				"body":     base.GetBody(c),
			},
		})
	})
	c.Any("register", func(c *gin.Context) {
		c.JSON(http.StatusOK, base.Any{
			"code": 0,
			"msg":  "success",
			"data": base.Any{
				"method":   base.GetMethod(c),
				"urlParam": base.GetUrlParam(c),
				"headers":  base.GetHeaders(c),
				"body":     base.GetBody(c),
			},
		})
	})
}
