package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"github.com/panco95/go-garden/examples/pay/global"
	"github.com/panco95/go-garden/examples/pay/rpc/user"
	"math/rand"
	"time"
)

func Order(c *gin.Context) {
	var validate struct {
		Username string `form:"username" binding:"required,max=20,min=1" `
	}
	if err := c.ShouldBind(&validate); err != nil {
		Fail(c, MsgInvalidParams)
		return
	}
	username := c.DefaultPostForm("username", "")

	span, err := core.GetSpan(c)
	if err != nil {
		Fail(c, MsgFail)
		global.Garden.Log(core.ErrorLevel, "GetSpan", err)
		return
	}

	args := user.ExistsArgs{
		Username: username,
	}
	reply := user.ExistsReply{}
	err = global.Garden.CallRpc(span, "user", "exists", &args, &reply)
	if err != nil {
		Fail(c, MsgFail)
		global.Garden.Log(core.ErrorLevel, "rpcCall", err)
		span.SetTag("callRpc", err)
		return
	}
	if !reply.Exists {
		Fail(c, MsgFail)
		return
	}

	orderId := fmt.Sprintf("%d%d", time.Now().Unix(), rand.Intn(10000))
	Success(c, MsgOk, core.MapData{
		"orderId": orderId,
	})
}
