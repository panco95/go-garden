package rpc

import (
	"context"
	"github.com/panco95/go-garden/examples/user2/global"
	"github.com/panco95/go-garden/examples/user2/model"
	"github.com/panco95/go-garden/examples/user2/rpc/define"
)

func (r *Rpc) Exists(ctx context.Context, args *define.ExistsArgs, reply *define.ExistsReply) error {
	span := global.Garden.StartRpcTrace(ctx, args, "Exists")

	db := global.Garden.GetDb()
	user := model.User{}
	result := db.Where("username = ?", args.Username).First(&user)
	if result.RowsAffected > 0 {
		reply.Exists = true
	} else {
		reply.Exists = false
	}

	global.Garden.FinishRpcTrace(span)
	return nil
}
