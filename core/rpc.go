package core

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/log"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/share"
)

func (g *Garden) rpcListen(name, network, address string, obj interface{}, metadata string) error {
	s := server.NewServer()

	l, err := g.GetLog()
	if err != nil {
		return err
	}
	log.SetLogger(l)

	if err := s.RegisterName(name, obj, metadata); err != nil {
		return err
	}
	g.Log(InfoLevel, "rpc", "listen on: "+address)
	if err := s.Serve(network, address); err != nil {
		return err
	}
	return nil
}

func rpcCall(span opentracing.Span, addr, service, method string, args, reply interface{}, timeout int) error {
	d, err := client.NewPeer2PeerDiscovery("tcp@"+addr, "")
	if err != nil {
		return err
	}
	xClient := client.NewXClient(service, client.Failtry, client.RandomSelect, d, client.DefaultOption)
	defer xClient.Close()

	textMapString := map[string]string{}
	if span != nil {
		textMap := opentracing.TextMapCarrier{}
		opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.TextMap,
			textMap)
		// write opentracing span to textMap
		textMap.ForeachKey(func(key, val string) error {
			textMapString[key] = val
			return nil
		})
	}

	// rpc timeout and value
	ctx := context.WithValue(context.Background(), share.ReqMetaDataKey, textMapString)
	ctx2, _ := context.WithTimeout(ctx, time.Millisecond*time.Duration(timeout))
	err = xClient.Call(ctx2, method, args, reply)

	if err != nil {
		return err
	}
	return nil
}

// StartSpanFormRpc start and get opentracing span from rpc call
func StartSpanFormRpc(ctx context.Context, operateName string) opentracing.Span {
	reqMeta := ctx.Value(share.ReqMetaDataKey).(map[string]string)
	span := StartSpanFromTextMap(reqMeta, operateName)
	return span
}
