package main

import (
	"context"
	"garden"
	"garden/drives/redis"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	// 服务初始化
	garden.Init()
	// 服务启动
	garden.Run(Route, nil)
}

// Route Gin路由
func Route(r *gin.Engine) {
	r.Use(garden.CheckCallSafeMiddleware()) // 调用接口安全验证
	r.POST("login", Login)
	r.POST("exists", Exists)
}

// Login 登录接口
// @param username 用户名
func Login(c *gin.Context) {
	span, err := garden.GetSpan(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		garden.Logger.Errorf("[%s] %s", "GetSpan", err)
		return
	}

	var Validate VLogin
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(http.StatusOK, garden.ApiResponse(1000, "参数非法", nil))
		return
	}

	username := c.DefaultPostForm("username", "")
	if err := redis.Client().Set(context.Background(), "user."+username, 0, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		garden.Logger.Errorf("[%s] %s", "RedisSet", err)
		span.SetTag("RedisSet", err)
		return
	}
	c.JSON(http.StatusOK, garden.ApiResponse(0, "登录成功", nil))
}

// Exists 查询用户是否存在接口
// @param username 用户名
// @return data.exists true || false
func Exists(c *gin.Context) {
	var Validate VExists
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(http.StatusOK, garden.ApiResponse(1000, "参数非法", nil))
		return
	}
	username := c.DefaultPostForm("username", "")
	exists := true
	_, err := redis.Client().Get(context.Background(), "user."+username).Result()
	if err != nil {
		exists = false
	}
	c.JSON(http.StatusOK, garden.ApiResponse(0, "", garden.Any{
		"exists": exists,
	}))
}

// VLogin login接口验证器
type VLogin struct {
	Username string `form:"username" binding:"required,max=20,min=1"`
}

// VExists exists接口验证器
type VExists struct {
	Username string `form:"username" binding:"required,max=20,min=1"`
}
