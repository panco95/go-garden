package core

import (
	"context"
)

type Rpc int

type SyncRoutesArgs struct {
	Yml []byte
}

type SyncRoutesReply struct {
	Result bool
}

func (r *Rpc) SyncRoutes(ctx context.Context, args *SyncRoutesArgs, reply *SyncRoutesReply) error {
	reply.Result = true
	if err := writeFile("./configs/routes.yml", args.Yml); err != nil {
		reply.Result = false
	}
	return nil
}
