package api

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/examples/user/global"
)

type loginValidate struct {
	Username string `form:"username" binding:"required,max=20,min=1"`
}

func Login(c *gin.Context) {
	var Validate loginValidate
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(200, Response(1000, "参数非法", nil))
		return
	}
	username := c.DefaultPostForm("username", "")
	global.Users.Store(username, 1)
	c.JSON(200, Response(0, "登录成功", nil))
}