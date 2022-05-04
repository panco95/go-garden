package core

import (
	"bytes"
	"context"
	"strings"

	"github.com/panco95/go-garden/core/log"
)

type Rpc int

type SyncRoutesArgs struct {
	Yml []byte
}

type SyncRoutesReply struct {
	Result bool
}

//SyncRoutes sync routes.yml method
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
		log.Error("syncRoutes", err)
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
	for k1, v1 := range g.services {
		for k2, v2 := range v1.Nodes {
			if strings.Compare(v2.Addr, g.GetServiceId()) == 0 {
				continue
			}
			addr, err := g.getServiceRpcAddr(k1, k2)
			if err != nil {
				log.Error("getServiceRpcAddr", err)
				continue
			}
			if err := rpcCall(nil, addr, k1, "SyncRoutes", &args, &reply, 10000); err != nil {
				log.Error("syncRoutes", err)
				return
			}
			if !reply.Result {
				log.Error("syncRoutes", "fail")
			}
			log.Info("syncRoutes", "success")
		}
	}
}
