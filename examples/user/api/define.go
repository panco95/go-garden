package api

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
)

const (
	CodeOk   = 1000
	CodeFail = 1001

	MsgOk            = "Success"
	MsgFail          = "Server error"
	MsgInvalidParams = "Invalid params"
)

func Success(c *gin.Context, msg string, data core.MapData) {
	c.JSON(200, core.MapData{
		"code": CodeOk,
		"msg":  msg,
		"data": data,
	})
}

func Fail(c *gin.Context, msg string) {
	c.JSON(200, core.MapData{
		"code": CodeFail,
		"msg":  msg,
		"data": nil,
	})
}
