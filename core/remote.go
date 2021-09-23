package core

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"net"
)

func (g *Garden) runRemoteRpc(port string) {
	defer func() {
		if err := recover(); err != nil {
			g.Log(DebugLevel, "runRemoteRpcReboot", err)
			g.runRemoteRpc(port)
		}
	}()

	listenAddress := ":" + port
	listen, err := net.Listen("tcp", listenAddress)
	if err != nil {
		g.Log(FatalLevel, "Rpc", err)
	}

	s := grpc.NewServer()
	RegisterRemoteServer(s, g.remoteServer)

	g.Log(InfoLevel, g.cfg.Service.ServiceName, fmt.Sprintf("Rpc listen on port: %s", port))
	g.Log(FatalLevel, "Rpc", s.Serve(listen).Error())
}

// remote rpc server
type remoteServer struct{}

// SyncRoute receive routes.yml and write to file
func (s remoteServer) SyncRoute(ctx context.Context, in *SyncRouteReq) (*SyncRouteRes, error) {
	resp := new(SyncRouteRes)
	resp.Result = true

	if err := writeFile("configs/routes.yml", in.Data); err != nil {
		resp.Result = false
	}

	return resp, nil
}

// send syncRoute routes.yml file to rpc address
func sendSyncRoute(address string, data []byte) (bool, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return false, err
	}
	defer conn.Close()

	c := NewRemoteClient(conn)

	req := &SyncRouteReq{Data: data}
	res, err := c.SyncRoute(context.Background(), req)

	if err != nil {
		return false, err
	}

	return res.Result, nil
}
