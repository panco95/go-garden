package core

import (
	"bytes"
	"context"
	"strings"
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

func (g *Garden) sendRoutes() {
	fileData, err := readFile("configs/routes.yml")
	if err != nil {
		g.Log(ErrorLevel, "SyncRoutes", err)
		return
	}

	if len(fileData) == 0 {
		return
	}

	if bytes.Compare(g.syncCache, fileData) == 0 {
		return
	}

	g.syncCache = fileData

	args := SyncRoutesArgs{
		Yml: fileData,
	}
	reply := SyncRoutesReply{}
	for k1, v1 := range g.Services {
		for k2, v2 := range v1.Nodes {
			if strings.Compare(v2.Addr, g.ServiceId) == 0 {
				continue
			}
			addr, err := g.getServiceRpcAddr(k1, k2)
			if err != nil {
				g.Log(ErrorLevel, "getServiceRpcAddr", err)
				continue
			}
			if err := rpcCall(nil, addr, k1, "SyncRoutes", &args, &reply); err != nil {
				g.Log(ErrorLevel, "SyncRoutes", err)
				return
			}
			if !reply.Result {
				g.Log(ErrorLevel, "SyncRoutes", "fail")
			}
			g.Log(InfoLevel, "SyncRoutes", "success")
		}
	}
}
