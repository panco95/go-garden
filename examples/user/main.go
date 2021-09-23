package main

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"sync"
)

var service *core.Garden
var users sync.Map

func main() {
	service = core.New()
	service.Run(route, nil)
}

func route(r *gin.Engine) {
	r.Use(service.CheckCallSafeMiddleware()) // 调用接口安全验证
	r.POST("login", login)
	r.POST("exists", exists)
}

func login(c *gin.Context) {
	var Validate vLogin
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(200, apiResponse(1000, "参数非法", nil))
		return
	}
	username := c.DefaultPostForm("username", "")
	users.Store(username, 1)
	c.JSON(200, apiResponse(0, "登录成功", nil))
}

func exists(c *gin.Context) {
	var Validate vExists
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(200, apiResponse(1000, "参数非法", nil))
		return
	}
	username := c.DefaultPostForm("username", "")
	exists := true
	if _, ok := users.Load(username); !ok {
		exists = false
	}
	c.JSON(200, apiResponse(0, "", core.MapData{
		"exists": exists,
	}))
}

type vLogin struct {
	Username string `form:"username" binding:"required,max=20,min=1"`
}

type vExists struct {
	Username string `form:"username" binding:"required,max=20,min=1"`
}

func apiResponse(code int, msg string, data interface{}) core.MapData {
	return core.MapData{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}
