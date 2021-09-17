package garden

import (
	"github.com/panco95/go-garden/sync"
	"google.golang.org/grpc"
	"log"
	"net"
)

func InitRpc(port string) {
	listenAddress := ":" + port
	listen, err := net.Listen("tcp", listenAddress)
	if err != nil {
		Fatal("Rpc", err)
	}

	s := grpc.NewServer()
	sync.RegisterSyncServer(s, sync.Server)

	log.Printf("[%s] Rpc listen on port: %s", Config.ServiceName, port)
	Fatal("Rpc", s.Serve(listen))
}
