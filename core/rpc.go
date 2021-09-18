package core

import (
	"fmt"
	"github.com/panco95/go-garden/sync"
	"google.golang.org/grpc"
	"net"
)

func (g *Garden) runRpc(port string) {
	listenAddress := ":" + port
	listen, err := net.Listen("tcp", listenAddress)
	if err != nil {
		g.Log(FatalLevel, "Rpc", err)
	}

	s := grpc.NewServer()
	sync.RegisterSyncServer(s, sync.Server)

	g.Log(InfoLevel, g.Cfg.ServiceName, fmt.Sprintf("Rpc listen on port: %s", port))
	g.Log(FatalLevel, "Rpc", s.Serve(listen).Error())
}
