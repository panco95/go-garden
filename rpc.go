package garden

import (
	"fmt"
	"github.com/panco95/go-garden/sync"
	"google.golang.org/grpc"
	"net"
)

func runRpc(port string) {
	listenAddress := ":" + port
	listen, err := net.Listen("tcp", listenAddress)
	if err != nil {
		Log(FatalLevel, "Rpc", err)
	}

	s := grpc.NewServer()
	sync.RegisterSyncServer(s, sync.Server)

	Log(InfoLevel, Config.ServiceName, fmt.Sprintf("Rpc listen on port: %s", port))
	Log(FatalLevel, "Rpc", s.Serve(listen).Error())
}
