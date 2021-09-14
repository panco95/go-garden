package garden

import (
	"garden/drives/ping"
	"google.golang.org/grpc"
	"log"
	"net"
)

// InitRpc 启动RPC服务
// @param port 监听端口
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
