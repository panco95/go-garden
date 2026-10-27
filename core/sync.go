package core

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"sync"

	"github.com/panco95/go-garden/core/log"
)

type Rpc int

type SyncRoutesArgs struct {
	EtcdKey string
	Yml     []byte
}

type SyncRoutesReply struct {
	Result bool
}

type clusterRouteSync struct {
	mu      sync.RWMutex
	file    string
	etcdKey string
}

var routeSync clusterRouteSync

func (g *Garden) routesFile() string {
	return filepath.Join(g.cfg.ConfigsPath, "routes.yml")
}

func (g *Garden) bindRouteSync() {
	routeSync.mu.Lock()
	routeSync.file = g.routesFile()
	routeSync.etcdKey = g.cfg.Service.EtcdKey
	routeSync.mu.Unlock()
}

// SyncRoutes writes this process's routes file when the caller belongs to the same cluster.
func (r *Rpc) SyncRoutes(ctx context.Context, args *SyncRoutesArgs, reply *SyncRoutesReply) error {
	reply.Result = false
	if args == nil {
		return errors.New("empty route sync")
	}
	routeSync.mu.RLock()
	file, key := routeSync.file, routeSync.etcdKey
	routeSync.mu.RUnlock()
	if file == "" || key == "" || args.EtcdKey != key {
		return errors.New("route sync rejected")
	}
	if err := writeFile(file, args.Yml); err != nil {
		return err
	}
	reply.Result = true
	return nil
}

func (g *Garden) sendRoutes() {
	fileData, err := readFile(g.routesFile())
	if err != nil {
		log.Error("syncRoutes", err)
		return
	}

	if len(fileData) == 0 {
		return
	}

	if bytes.Equal(g.syncCache, fileData) {
		return
	}

	g.syncCache = fileData

	args := SyncRoutesArgs{
		EtcdKey: g.cfg.Service.EtcdKey,
		Yml:     fileData,
	}
	reply := SyncRoutesReply{}
	self := g.localNodeAddr()
	for _, item := range g.listServiceNodes() {
		if item.node.Addr == self {
			continue
		}
		addr, err := nodeRPCAddr(item.node.Addr)
		if err != nil {
			log.Error("getServiceRpcAddr", err)
			continue
		}
		if err := g.rpcCall(nil, addr, item.serviceName, "SyncRoutes", &args, &reply, 10000); err != nil {
			log.Error("syncRoutes", err)
			return
		}
		if !reply.Result {
			log.Error("syncRoutes", "fail")
		}
		log.Info("syncRoutes", "success")
	}
}
