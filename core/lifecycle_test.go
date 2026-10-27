package core

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewReturnsConfigError(t *testing.T) {
	dir := t.TempDir()
	g, err := New(dir, dir)
	if err == nil || g != nil {
		t.Fatalf("garden=%v err=%v", g, err)
	}
}

func TestMetricsExport(t *testing.T) {
	g := &Garden{}
	atomic.StoreInt64(&g.requestFinish, 3)
	g.SetMetric("orders_total", 7)
	g.SetMetric("not a metric", 1)
	g.SetMetric("note", "text")

	engine, err := g.httpEngine(func(r *gin.Engine) {}, nil)
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: engine}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(ln)
	defer srv.Close()

	resp, err := http.Get("http://" + ln.Addr().String() + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range []string{
		"garden_requests_in_flight",
		"garden_requests_finished_total",
		"orders_total",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %s in %s", want, text)
		}
	}
	if strings.Contains(text, "not a metric") || strings.Contains(text, "note ") {
		t.Fatalf("invalid metric exported: %s", text)
	}
}

func TestServeStopsOnCancel(t *testing.T) {
	httpPort := freePort(t)
	rpcPort := freePort(t)
	g := &Garden{cfg: cfg{Service: serviceCfg{
		ServiceName: "svc",
		ServiceIp:   "127.0.0.1",
		HttpPort:    httpPort,
		RpcPort:     rpcPort,
		CallKey:     "k",
	}}}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- g.serve(ctx, func(r *gin.Engine) {
			r.GET("/health", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
		}, new(Rpc), nil)
	}()

	base := "127.0.0.1:" + httpPort
	waitForTCP(t, base)
	resp, err := http.Get("http://" + base + "/health")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	waitForTCP(t, "127.0.0.1:"+rpcPort)

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return")
	}
	if portOpen("127.0.0.1:" + httpPort) {
		t.Fatal("http still listening")
	}
	if portOpen("127.0.0.1:" + rpcPort) {
		t.Fatal("rpc still listening")
	}
}

func freePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	return port
}

func portOpen(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
