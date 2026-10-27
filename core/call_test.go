package core

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/smallnest/rpcx/server"
)

func TestRetrySwitchesNode(t *testing.T) {
	var badHits, goodHits int32
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&badHits, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bad.Close()
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&goodHits, 1)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer good.Close()

	badAddr := nodeAddrFromURL(t, bad.URL)
	goodAddr := nodeAddrFromURL(t, good.URL)
	g := &Garden{
		cfg:      cfg{Service: serviceCfg{CallKey: "secret"}},
		services: map[string]*service{},
	}
	g.addNodeLocked("svc", badAddr)
	g.addNodeLocked("svc", goodAddr)

	span := opentracing.StartSpan("test")
	defer span.Finish()
	code, body, _, err := g.retryGo("svc", "ping", []int{0, 0}, g.services["svc"].Nodes[badAddr], span, routeCfg{
		Type:    "http",
		Path:    "/ping",
		Timeout: 1000,
	}, &req{Method: http.MethodGet}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if code != http.StatusOK || body != `{"ok":true}` {
		t.Fatalf("code=%d body=%s", code, body)
	}
	if atomic.LoadInt32(&badHits) != 1 || atomic.LoadInt32(&goodHits) != 1 {
		t.Fatalf("bad=%d good=%d", badHits, goodHits)
	}
	if _, ok := g.fusingMap.Load(breakerKey(badAddr, "svc", "ping")); !ok {
		t.Fatal("failed node was not counted")
	}
}

func TestFusingKeyKeepsRouteAction(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatal(err)
	}
	nodeAddr := "127.0.0.1:9:" + port

	g := &Garden{services: map[string]*service{}}
	g.addNodeLocked("svc", nodeAddr)
	span := opentracing.StartSpan("test")
	defer span.Finish()
	g.retryGo("svc", "ping", []int{0}, g.services["svc"].Nodes[nodeAddr], span, routeCfg{
		Type:    "rpc",
		Timeout: 300,
	}, nil, &EchoArgs{}, &EchoReply{})

	if _, ok := g.fusingMap.Load(breakerKey(nodeAddr, "svc", "ping")); !ok {
		t.Fatal("missing route action key")
	}
	if _, ok := g.fusingMap.Load(breakerKey(nodeAddr, "svc", "Ping")); ok {
		t.Fatal("fusing key used the rpc method name")
	}
}

func TestCallRpcReusesClient(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := server.NewServer()
	if err := srv.RegisterName("svc", new(echoRPC), ""); err != nil {
		t.Fatal(err)
	}
	go srv.ServeListener("tcp", ln)
	defer srv.Close()
	waitForTCP(t, ln.Addr().String())

	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	nodeAddr := "127.0.0.1:9:" + port
	g := &Garden{
		cfg: cfg{
			Service: serviceCfg{CallRetry: "0"},
			Routes: map[string]map[string]routeCfg{
				"svc": {"ping": {Type: "rpc", Timeout: 2000}},
			},
		},
		services: map[string]*service{},
	}
	g.addNodeLocked("svc", nodeAddr)

	span := opentracing.StartSpan("test")
	defer span.Finish()
	for i := 0; i < 2; i++ {
		args := EchoArgs{Ping: "hi"}
		reply := EchoReply{}
		if err := g.CallRpc(span, "svc", "ping", &args, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.Pong != "hi" {
			t.Fatalf("pong=%s", reply.Pong)
		}
	}
	g.rpcPool.mu.Lock()
	clients := len(g.rpcPool.clients)
	g.rpcPool.mu.Unlock()
	if clients != 1 {
		t.Fatalf("clients=%d", clients)
	}

	g.closeRPCClients([]routedNode{{serviceName: "svc", node: &node{Addr: nodeAddr}}})
	g.rpcPool.mu.Lock()
	clients = len(g.rpcPool.clients)
	g.rpcPool.mu.Unlock()
	if clients != 0 {
		t.Fatalf("clients after close=%d", clients)
	}
}

type EchoArgs struct {
	Ping string
}

type EchoReply struct {
	Pong string
}

type echoRPC struct{}

func (e *echoRPC) Ping(ctx context.Context, args *EchoArgs, reply *EchoReply) error {
	reply.Pong = args.Ping
	return nil
}

func nodeAddrFromURL(t *testing.T, rawURL string) string {
	t.Helper()
	host := rawURL
	if len(host) > 7 && host[:7] == "http://" {
		host = host[7:]
	}
	hostname, port, err := net.SplitHostPort(host)
	if err != nil {
		t.Fatal(err)
	}
	return hostname + ":" + port + ":9"
}

func waitForTCP(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatal(err)
		}
	}
}
