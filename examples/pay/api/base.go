package api

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"github.com/panco95/go-garden/examples/pay/global"
)

func Routes(r *gin.Engine) {
	r.Use(global.Service.CheckCallSafeMiddleware())
	r.POST("order", Order)
}

func Response(code int, msg string, data interface{}) core.MapData {
	return core.MapData{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}
