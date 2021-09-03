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
)

func main() {
	flag.Parse()
	base.Init(*rpcPort, *httpPort, "user")
	base.GinServer(*httpPort, "user", route)
}

func route(r *gin.Engine) {
	r.Use(base.CheckCallServiceKey())
	r.Any("login", func(c *gin.Context) {
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
	r.Any("register", func(c *gin.Context) {
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
