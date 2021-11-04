package api

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"github.com/panco95/go-garden/examples/user/global"
)

func Login(c *gin.Context) {
	//db := global.Service.Db
	//result := make(map[string]interface{})
	//db.Raw("SELECT * FROM test").Scan(&result)
	//global.Service.Log(core.InfoLevel, "result", result)

	//redis := global.Service.Redis
	//err := redis.Set(context.Background(), "key", "value", 0).Err()
	//if err != nil {
	//	global.Service.Log(core.InfoLevel, "redis", err)
	//}

	var validate struct {
		Username string `form:"username" binding:"required,max=20,min=1"`
	}
	if err := c.ShouldBind(&validate); err != nil {
		core.Resp(c, core.HttpOk, -1, core.InfoInvalidParam, nil)
		return
	}
	username := c.DefaultPostForm("username", "")
	global.Users.Store(username, 1)
	core.Resp(c, core.HttpOk, 0, "登陆成功", nil)
}
