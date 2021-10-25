package rpc

import (
	"context"
	"encoding/json"
	"github.com/panco95/go-garden/examples/user/global"
	"github.com/panco95/go-garden/examples/user/rpc/define"
)

func (r *Rpc) Exists(ctx context.Context, args *define.ExistsArgs, reply *define.ExistsReply) error {
	span := global.Service.StartSpanFormRpc(ctx, "Exists")
	span.SetTag("CallType", "Rpc")
	span.SetTag("ServiceIp", global.Service.ServiceIp)
	span.SetTag("ServiceId", global.Service.ServiceId)
	span.SetTag("Result", "running")
	s, _ := json.Marshal(&args)
	span.SetTag("Args", string(s))

	reply.Exists = false
	if _, ok := global.Users.Load(args.Username); ok {
		reply.Exists = true
	}

	span.SetTag("Result", "success")
	span.Finish()
	return nil
}
