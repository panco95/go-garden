package main

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
)

var service *core.Garden

func main() {
	service = core.New()
	service.Run(route, nil)
}

func route(r *gin.Engine) {
	r.Use(service.CheckCallSafeMiddleware())
	r.POST("test", test)
}

func test(c *gin.Context) {
}
