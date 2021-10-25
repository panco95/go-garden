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
		core.Resp(c, core.HttpOk, core.CodeFail, core.InfoInvalidParam, nil)
		return
	}
	username := c.DefaultPostForm("username", "")

	span, err := core.GetSpan(c)
	if err != nil {
		core.Resp(c, core.HttpFail, core.CodeFail, core.InfoServerError, nil)
		global.Service.Log(core.ErrorLevel, "GetSpan", err)
		return
	}

	args := user.ExistsArgs{
 		Username: username,
	}
	reply := user.ExistsReply{}
	_, _, err = global.Service.CallService(span, "user", "exists", nil, &args, &reply)
	if err != nil {
		core.Resp(c, core.HttpFail, core.CodeFail, core.InfoServerError, nil)
		global.Service.Log(core.ErrorLevel, "rpcCall", err)
		span.SetTag("callRpc", err)
		return
	}
	if !reply.Exists {
		core.Resp(c, core.HttpOk, core.CodeFail, "下单失败", nil)
		return
	}

	orderId := fmt.Sprintf("%d%d", time.Now().Unix(), rand.Intn(10000))
	core.Resp(c, core.HttpOk, core.CodeSuccess, "下单成功", core.MapData{
		"orderId": orderId,
	})
}
