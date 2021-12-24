package api

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/examples/user2/global"
	"github.com/panco95/go-garden/examples/user2/model"
)

func Register(c *gin.Context) {
	var validate struct {
		Username string `form:"username" binding:"required,max=20,min=1"`
	}
	if err := c.ShouldBind(&validate); err != nil {
		Fail(c, MsgInvalidParams)
		return
	}
	username := c.PostForm("username")

	db := global.Garden.GetDb()
	user := model.User{}
	result := db.Where("username = ?", username).First(&user)
	if result.RowsAffected > 0 {
		Fail(c, "username exists!")
		return
	}
	user.Username = username
	if db.Create(&user).RowsAffected < 1 {
		Fail(c, "register fail!")
		return
	}

	Success(c, "register success!", nil)
}
