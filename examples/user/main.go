package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden"
	"github.com/panco95/go-garden/core"
	"github.com/panco95/go-garden/core/drives/redis"
	"net/http"
)

var service core.Garden

func main() {
	service = garden.NewService()
	service.Run(Route, nil)
}

func Route(r *gin.Engine) {
	r.Use(service.CheckCallSafeMiddleware()) // 调用接口安全验证
	r.POST("login", Login)
	r.POST("exists", Exists)
}

func Login(c *gin.Context) {
	span, err := core.GetSpan(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		service.Log(core.ErrorLevel, "GetSpan", err)
		return
	}

	var Validate VLogin
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(http.StatusOK, ApiResponse(1000, "参数非法", nil))
		return
	}

	username := c.DefaultPostForm("username", "")
	if err := redis.Client().Set(context.Background(), "user."+username, 0, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		service.Log(core.ErrorLevel, "RedisSet", err)
		span.SetTag("RedisSet", err)
		return
	}
	c.JSON(http.StatusOK, ApiResponse(0, "登录成功", nil))
}

// Exists Query if the user exists
func Exists(c *gin.Context) {
	var Validate VExists
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(http.StatusOK, ApiResponse(1000, "参数非法", nil))
		return
	}
	username := c.DefaultPostForm("username", "")
	exists := true
	_, err := redis.Client().Get(context.Background(), "user."+username).Result()
	if err != nil {
		exists = false
	}
	c.JSON(http.StatusOK, ApiResponse(0, "", core.MapData{
		"exists": exists,
	}))
}

// VLogin The login api validator
type VLogin struct {
	Username string `form:"username" binding:"required,max=20,min=1"`
}

// VExists The exists interface validator
type VExists struct {
	Username string `form:"username" binding:"required,max=20,min=1"`
}

func ApiResponse(code int, msg string, data interface{}) core.MapData {
	return core.MapData{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}