package core

import (
	"context"
)

type Rpc int

type PingArgs struct {
	Msg string
}

type PingReply struct {
	Result string
}

func (r *Rpc) Ping(ctx context.Context, args *PingArgs, reply *PingReply) error {
	reply.Result = "pong"
	return nil
}