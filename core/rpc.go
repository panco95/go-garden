package core

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/panco95/go-garden/core/log"
	"github.com/smallnest/rpcx/client"
	rpcx_logger "github.com/smallnest/rpcx/log"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/share"
)

// rpcPool keeps one rpcx client per service and dial address.
type rpcPool struct {
	mu      sync.Mutex
	clients map[string]client.XClient
}

func (g *Garden) rpcListen(name, network, address string, obj interface{}, metadata string) error {
	s := server.NewServer()
	s.AuthFunc = g.rpcAuth
	if logger := log.GetLogger(); logger != nil {
		rpcx_logger.SetLogger(logger)
	}
	if err := s.RegisterName(name, obj, metadata); err != nil {
		return err
	}
	log.Infof("rpc", "listen on: %s", address)
	if err := s.Serve(network, address); err != nil {
		return err
	}
	return nil
}

func rpcClientKey(service, addr string) string {
	return service + "\x00" + addr
}

func (g *Garden) rpcClient(service, addr string) (client.XClient, error) {
	key := rpcClientKey(service, addr)
	g.rpcPool.mu.Lock()
	defer g.rpcPool.mu.Unlock()
	if g.rpcPool.clients == nil {
		g.rpcPool.clients = map[string]client.XClient{}
	}
	if c, ok := g.rpcPool.clients[key]; ok {
		return c, nil
	}
	discovery, err := client.NewPeer2PeerDiscovery("tcp@"+addr, "")
	if err != nil {
		return nil, err
	}
	c := client.NewXClient(service, client.Failtry, client.RandomSelect, discovery, client.DefaultOption)
	c.Auth(g.cfg.Service.CallKey)
	g.rpcPool.clients[key] = c
	return c, nil
}

func (g *Garden) closeRPCClients(nodes []routedNode) {
	if len(nodes) == 0 {
		return
	}
	closing := make([]client.XClient, 0, len(nodes))
	g.rpcPool.mu.Lock()
	for _, item := range nodes {
		if item.node == nil || g.rpcPool.clients == nil {
			continue
		}
		host, err := nodeRPCAddr(item.node.Addr)
		if err != nil {
			continue
		}
		key := rpcClientKey(item.serviceName, host)
		c, ok := g.rpcPool.clients[key]
		if !ok {
			continue
		}
		delete(g.rpcPool.clients, key)
		closing = append(closing, c)
	}
	g.rpcPool.mu.Unlock()
	for _, c := range closing {
		c.Close()
	}
}

func (g *Garden) rpcAuth(ctx context.Context, req *protocol.Message, token string) error {
	if g.checkCallSafe(token) {
		return nil
	}
	return errors.New(infoNoAuth)
}

func (g *Garden) closeAllRPCClients() {
	g.rpcPool.mu.Lock()
	clients := g.rpcPool.clients
	g.rpcPool.clients = nil
	g.rpcPool.mu.Unlock()
	for _, c := range clients {
		c.Close()
	}
}

func (g *Garden) rpcCall(span opentracing.Span, addr, service, method string, args, reply interface{}, timeout int) error {
	xClient, err := g.rpcClient(service, addr)
	if err != nil {
		return err
	}

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

	base := context.WithValue(context.Background(), share.ReqMetaDataKey, textMapString)
	ctx, cancel := context.WithTimeout(base, time.Millisecond*time.Duration(timeout))
	defer cancel()
	return xClient.Call(ctx, method, args, reply)
}

// StartSpanFormRpc start and get opentracing span from rpc call
func StartSpanFormRpc(ctx context.Context, operateName string) opentracing.Span {
	reqMeta := ctx.Value(share.ReqMetaDataKey).(map[string]string)
	span := StartSpanFromTextMap(reqMeta, operateName)
	return span
}
