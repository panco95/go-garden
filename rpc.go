package garden

import (
	"github.com/panco95/go-garden/drives/ping"
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
	ping.RegisterPingServer(s, ping.Service)

	log.Printf("[%s] Rpc listen on port: %s", Config.ServiceName, port)
	Fatal("Rpc", s.Serve(listen))
}
