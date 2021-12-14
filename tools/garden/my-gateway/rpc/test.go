package rpc

import (
	"context"
	"github.com/panco95/go-garden/core"
	"my-gateway/global"
	"my-gateway/rpc/define"
)

func (r *Rpc) Testrpc(ctx context.Context, args *define.TestrpcArgs, reply *define.TestrpcReply) error {
	span := global.Garden.StartRpcTrace(ctx, args, "testrpc")

	global.Garden.Log(core.InfoLevel, "Test", "Receive a rpc message")
	reply.Pong = "pong"

	global.Garden.FinishRpcTrace(span)
	return nil
}
