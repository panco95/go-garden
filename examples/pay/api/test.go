package api

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"github.com/panco95/go-garden/examples/pay/global"
)

func Test(c *gin.Context) {
	mysql := global.Service.GetConfigValueMap("mysql")
	number := global.Service.GetConfigValueInt("number")
	str := global.Service.GetConfigValueString("str")
	core.Resp(c, core.HttpOk, 0, "", core.MapData{
		"mysql":   mysql,
		"number":  number,
		"str":     str,
	})
}
