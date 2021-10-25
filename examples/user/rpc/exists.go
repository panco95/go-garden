package rpc

import (
	"context"
	"github.com/panco95/go-garden/core"
	"github.com/panco95/go-garden/examples/user/global"
	"github.com/panco95/go-garden/examples/user/rpc/define"
)

func (r *Rpc) Exists(ctx context.Context, args *define.ExistsArgs, reply *define.ExistsReply) error {
	span := core.StartSpanFormRpc(ctx, "exists")
	span.Finish()

	reply.Exists = false
	if _, ok := global.Users.Load(args.Username); ok {
		reply.Exists = true
	}
	return nil
}
