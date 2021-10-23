package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"github.com/panco95/go-garden/examples/pay/global"
	"math/rand"
	"time"
)

func Order(c *gin.Context) {
	span, err := core.GetSpan(c)
	if err != nil {
		c.JSON(500, nil)
		global.Service.Log(core.ErrorLevel, "GetSpan", err)
		return
	}
	var Validate struct {
		Username string `form:"username" binding:"required,max=20,min=1" `
	}
	if err := c.ShouldBind(&Validate); err != nil {
		c.JSON(200, Response(1000, "非法参数", nil))
		return
	}
	username := c.DefaultPostForm("username", "")

	type ExistsArgs struct {
		Username string
	}
	type ExistsReply struct {
		Exists bool
	}
	args := ExistsArgs{
		Username: username,
	}
	reply := ExistsReply{}
	global.Service.CallService(span, "user", "Exists", nil, &args, &reply)
	err = global.Service.RpcCall("192.168.8.98:9001", "user", "Exists", &args, &reply)
	if err != nil {
		global.Service.Log(core.ErrorLevel, "rpcCall", err)
		c.JSON(500, nil)
		span.SetTag("callRpc", err)
	}
	if !reply.Exists {
		c.JSON(500, Response(1000, "下单失败", nil))
		return
	}

	orderId := fmt.Sprintf("%d%d", time.Now().Unix(), rand.Intn(10000))
	c.JSON(200, Response(0, "下单成功", core.MapData{
		"orderId": orderId,
	}))
}
