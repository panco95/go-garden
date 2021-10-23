package api

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"github.com/panco95/go-garden/examples/user2/global"
)

func Routes(r *gin.Engine) {
	r.Use(global.Service.CheckCallSafeMiddleware()) // 调用接口安全验证
	r.POST("login", Login)
}

func Response(code int, msg string, data interface{}) core.MapData {
	return core.MapData{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}
