package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"goms"
	"goms/example/services/user/validate"
	"goms/pkg/redis"
	"net/http"
)

// Login 登录接口
// @param username 用户名
func Login(c *gin.Context) {
	var Validate validate.Login
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(http.StatusOK, goms.ApiResponse(1000, "参数非法", nil))
		return
	}
	username := c.DefaultPostForm("username", "")
	err := redis.Client().Set(context.Background(), "user."+username, 0, 0).Err()
	if err != nil {
		goms.Logger.Error("redis set error：" + err.Error())
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.JSON(http.StatusOK, goms.ApiResponse(0, "登录成功", nil))
}

// Exists 查询用户是否存在接口
// @param username 用户名
// @return data.exists true || false
func Exists(c *gin.Context) {
	var Validate validate.Login
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(http.StatusOK, goms.ApiResponse(1000, "参数非法", nil))
		return
	}
	username := c.DefaultPostForm("username", "")
	exists := true
	_, err := redis.Client().Get(context.Background(), "user."+username).Result()
	if err != nil {
		exists = false
	}
	c.JSON(http.StatusOK, goms.ApiResponse(0, "", goms.Any{
		"exists": exists,
	}))
}
