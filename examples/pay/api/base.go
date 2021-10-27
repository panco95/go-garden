package api

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/examples/pay/global"
)

func Routes(r *gin.Engine) {
	r.Use(global.Service.CheckCallSafeMiddleware())
	r.POST("order", Order)
	r.POST("test", Test)
}
