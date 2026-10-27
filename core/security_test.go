package core

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/share"
)

func TestSyncRoutesSameClusterOnly(t *testing.T) {
	dir := t.TempDir()
	g := &Garden{cfg: cfg{
		ConfigsPath: dir,
		Service:     serviceCfg{EtcdKey: "garden"},
	}}
	g.bindRouteSync()
	defer g.clearRouteSync()

	var rpc Rpc
	reply := SyncRoutesReply{}
	err := rpc.SyncRoutes(context.Background(), &SyncRoutesArgs{
		EtcdKey: "other",
		Yml:     []byte("routes: {}\n"),
	}, &reply)
	if err == nil || reply.Result {
		t.Fatalf("err=%v result=%v", err, reply.Result)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "routes.yml")); !os.IsNotExist(statErr) {
		t.Fatal("foreign cluster wrote routes.yml")
	}

	body := []byte("routes:\n  svc: {}\n")
	reply = SyncRoutesReply{}
	if err := rpc.SyncRoutes(context.Background(), &SyncRoutesArgs{EtcdKey: "garden", Yml: body}, &reply); err != nil {
		t.Fatal(err)
	}
	if !reply.Result {
		t.Fatal("same cluster was rejected")
	}
	got, err := os.ReadFile(g.routesFile())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(body) {
		t.Fatalf("file=%s", got)
	}
}

func (g *Garden) clearRouteSync() {
	routeSync.mu.Lock()
	routeSync.file = ""
	routeSync.etcdKey = ""
	routeSync.mu.Unlock()
}

func TestPprofRequiresCallKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	g := &Garden{cfg: cfg{
		RuntimePath: t.TempDir(),
		Service:     serviceCfg{Debug: true, CallKey: "secret"},
	}}
	engine, err := g.httpEngine(func(r *gin.Engine) {}, g.CheckCallSafeMiddleware)
	if err != nil {
		t.Fatal(err)
	}

	denied := httptest.NewRecorder()
	engine.ServeHTTP(denied, httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil))
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", denied.Code)
	}

	allowed := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	req.Header.Set("Call-Key", "secret")
	engine.ServeHTTP(allowed, req)
	if allowed.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", allowed.Code, allowed.Body.String())
	}
}

func TestSpanOmitsCallKeyAndBody(t *testing.T) {
	tracer := mocktracer.New()
	prev := opentracing.GlobalTracer()
	opentracing.SetGlobalTracer(tracer)
	defer opentracing.SetGlobalTracer(prev)

	gin.SetMode(gin.TestMode)
	g := &Garden{}
	r := gin.New()
	r.Use(g.openTracingMiddleware())
	r.POST("/in", func(c *gin.Context) {
		got := GetRequest(c)
		if string(got.RawBody) != `{"password":"secret"}` || got.Headers["Call-Key"] != "secret" {
			t.Fatalf("handler request=%+v", got)
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/in", strings.NewReader(`{"password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Call-Key", "secret")
	r.ServeHTTP(httptest.NewRecorder(), req)

	spans := tracer.FinishedSpans()
	if len(spans) != 1 {
		t.Fatalf("spans=%d", len(spans))
	}
	tag, _ := spans[0].Tags()["Request"].(string)
	if strings.Contains(tag, "secret") || strings.Contains(tag, "password") {
		t.Fatalf("span=%s", tag)
	}

	rpcSpan := g.StartRpcTrace(context.WithValue(context.Background(), share.ReqMetaDataKey, map[string]string{}), map[string]string{"ping": "pong"}, "Ping")
	rpcSpan.Finish()
	finished := tracer.FinishedSpans()
	if len(finished) != 2 {
		t.Fatalf("spans=%d", len(finished))
	}
	args, _ := finished[1].Tags()["Args"].(string)
	if !strings.Contains(args, "pong") {
		t.Fatalf("args=%s", args)
	}
}

func TestRPCRequiresCallKey(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := server.NewServer()
	guard := &Garden{cfg: cfg{Service: serviceCfg{CallKey: "secret"}}}
	srv.AuthFunc = guard.rpcAuth
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
	newGarden := func(key string) *Garden {
		g := &Garden{
			cfg: cfg{
				Service: serviceCfg{CallKey: key, CallRetry: "0"},
				Routes: map[string]map[string]routeCfg{
					"svc": {"ping": {Type: "rpc", Timeout: 2000}},
				},
			},
			services: map[string]*service{},
		}
		g.addNodeLocked("svc", nodeAddr)
		return g
	}

	span := opentracing.StartSpan("test")
	defer span.Finish()
	args := EchoArgs{Ping: "hi"}
	reply := EchoReply{}
	if err := newGarden("nope").CallRpc(span, "svc", "ping", &args, &reply); err == nil {
		t.Fatal("accepted a bad call key")
	}
	reply = EchoReply{}
	if err := newGarden("secret").CallRpc(span, "svc", "ping", &args, &reply); err != nil {
		t.Fatal(err)
	}
	if reply.Pong != "hi" {
		t.Fatalf("pong=%s", reply.Pong)
	}
}
