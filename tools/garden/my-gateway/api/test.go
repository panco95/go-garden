package api

import (
	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core"
	"my-gateway/global"
	"my-gateway/rpc/define"
)

func Test(c *gin.Context) {
	span, _ := core.GetSpan(c)

	// rpc call test
	args := define.TestrpcArgs{
		Ping: "ping",
	}
	reply := define.TestrpcReply{}
	err := global.Garden.CallRpc(span, "my-gateway", "testrpc", &args, &reply)
	if err != nil {
		global.Garden.Log(core.ErrorLevel, "rpcCall", err)
		span.SetTag("CallService", err)
		Fail(c, MsgFail)
		return
	}

	Success(c, MsgOk, core.MapData{
		"pong": reply.Pong,
	})
}
