package api

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"github.com/panco95/go-garden/examples/user/global"
)

func Login(c *gin.Context) {
	var validate struct {
		Username string `form:"username" binding:"required,max=20,min=1"`
	}
	if err := c.ShouldBind(&validate); err != nil {
		core.Resp(c, core.HttpOk, core.CodeFail, core.InfoInvalidParam, nil)
		return
	}
	username := c.DefaultPostForm("username", "")
	global.Users.Store(username, 1)
	core.Resp(c, core.HttpOk, core.CodeSuccess, "登陆成功", nil)
}
