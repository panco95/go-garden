package rpc

import (
	"context"
	"github.com/panco95/go-garden/examples/user2/global"
	"github.com/panco95/go-garden/examples/user2/rpc/define"
)

func (r *Rpc) Exists(ctx context.Context, args *define.ExistsArgs, reply *define.ExistsReply) error {
	span := global.Garden.StartRpcTrace(ctx, args, "Exists")

	reply.Exists = false
	if _, ok := global.Users.Load(args.Username); ok {
		reply.Exists = true
	}

	global.Garden.FinishRpcTrace(span)
	return nil
}
