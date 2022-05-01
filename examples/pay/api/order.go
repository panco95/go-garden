package api

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"github.com/panco95/go-garden/examples/pay/global"
	"github.com/panco95/go-garden/examples/pay/model"
	"github.com/panco95/go-garden/examples/pay/rpc/user"
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

	span := core.GetSpan(c)
	args := user.ExistsArgs{
		Username: username,
	}
	reply := user.ExistsReply{}
	err := global.Garden.CallRpc(span, "user", "exists", &args, &reply)
	if err != nil {
		Fail(c, MsgFail)
		global.Garden.Log(core.ErrorLevel, "rpcCall", err)
		span.SetTag("callRpc", err)
		return
	}
	if !reply.Exists {
		Fail(c, "user not exists")
		return
	}

	orderId := fmt.Sprintf("%d%d", time.Now().Unix(), rand.Intn(10000))
	order := model.Order{
		OrderId: orderId,
	}
	db, _ := global.Garden.GetDb()
	if db.Create(&order).RowsAffected < 1 {
		Fail(c, "order fail!")
		return
	}

	Success(c, "order success!", core.MapData{
		"orderId": orderId,
	})
}
