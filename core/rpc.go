package core

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/share"
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

// RpcCall service.RpcCall("192.168.8.98:9001", "gateway", "Ping", &args, &reply)
func (g *Garden) RpcCall(span opentracing.Span, addr, service, method string, args, reply interface{}) error {
	d, err := client.NewPeer2PeerDiscovery("tcp@"+addr, "")
	if err != nil {
		return err
	}
	xClient := client.NewXClient(service, client.Failtry, client.RandomSelect, d, client.DefaultOption)
	defer xClient.Close()

	textMap := opentracing.TextMapCarrier{}
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.TextMap,
		textMap)

	// write opentracing span to textMap
	textMapString := map[string]string{}
	textMap.ForeachKey(func(key, val string) error {
		textMapString[key] = val
		return nil
	})

	ctx := context.WithValue(context.Background(), share.ReqMetaDataKey, textMapString)
	err = xClient.Call(ctx, method, args, reply)

	if err != nil {
		return err
	}
	return nil
}
