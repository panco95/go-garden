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

	args := user.ExistsArgs{
		Username: username,
	}
	reply := user.ExistsReply{}
	_, _, err = global.Service.CallService(span, "user", "exists", nil, &args, &reply)
	if err != nil {
		global.Service.Log(core.ErrorLevel, "rpcCall", err)
		c.JSON(500, nil)
		span.SetTag("callRpc", err)
	}
	fmt.Print(reply)
	if !reply.Exists {
		c.JSON(500, Response(1000, "下单失败", nil))
		return
	}

	orderId := fmt.Sprintf("%d%d", time.Now().Unix(), rand.Intn(10000))
	c.JSON(200, Response(0, "下单成功", core.MapData{
		"orderId": orderId,
	}))
}
