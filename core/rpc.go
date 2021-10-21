package core

import (
	"context"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
)

// RpcListen core.RpcServer("Test", "tcp", ":9999", new(Test), "")
func (g *Garden) RpcListen(name, network, address string, obj interface{}, metadata string) error {
	s := server.NewServer()
	if err := s.RegisterName(name, obj, metadata); err != nil {
		return err
	}
	g.Log(InfoLevel, "rpc", "listen on: "+address)
	if err := s.Serve(network, address); err != nil {
		return err
	}
	return nil
}

// RpcCall
// 	args := core.PingArgs{
//		Msg: "asd",
//	}
//	reply := core.PingReply{}
//	service.RpcCall("192.168.8.98:9001", "gateway", "Ping", &args, &reply)
//	log.Println(reply.Result)
func (g *Garden) RpcCall(addr, service, method string, args, reply interface{}) error {
	d, err := client.NewPeer2PeerDiscovery("tcp@"+addr, "")
	if err != nil {
		return err
	}
	xClient := client.NewXClient(service, client.Failtry, client.RandomSelect, d, client.DefaultOption)
	defer xClient.Close()
	err = xClient.Call(context.Background(), method, args, reply)
	if err != nil {
		return err
	}
	return nil
}
