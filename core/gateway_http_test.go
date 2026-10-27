package core

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
)

func TestCorsReachesRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(cors)
	hits := 0
	r.GET("/ok", func(c *gin.Context) {
		hits++
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ok", nil))
	if hits != 1 || w.Code != http.StatusOK || w.Body.String() != "ok" {
		t.Fatalf("hits=%d status=%d body=%s", hits, w.Code, w.Body.String())
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("origin=%s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatal("wildcard CORS must not advertise credentials")
	}
}

func TestCorsOptionsStops(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(cors)
	r.OPTIONS("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "handler")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodOptions, "/ok", nil))
	if w.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestCallKeyUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	g := &Garden{cfg: cfg{Service: serviceCfg{CallKey: "secret"}}}
	r := gin.New()
	r.Use(g.CheckCallSafeMiddleware())
	hits := 0
	r.GET("/ok", func(c *gin.Context) {
		hits++
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ok", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", w.Code)
	}

	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.Header.Set("Call-Key", "secret")
	r.ServeHTTP(w, req)
	if hits != 1 || w.Code != http.StatusOK {
		t.Fatalf("hits=%d status=%d", hits, w.Code)
	}
}

func TestMiddlewareKeepsRawBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	g := &Garden{}
	r := gin.New()
	r.Use(g.openTracingMiddleware())
	r.POST("/in", func(c *gin.Context) {
		got := GetRequest(c)
		if string(got.RawBody) != `{"a":"b"}` {
			t.Fatalf("raw=%s", got.RawBody)
		}
		if got.Body["a"] != "b" {
			t.Fatalf("body=%v", got.Body)
		}
		if got.UrlParam != "?x=1" {
			t.Fatalf("query=%s", got.UrlParam)
		}
		if got.RemoteIp != "9.9.9.9" {
			t.Fatalf("remote=%s", got.RemoteIp)
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/in?x=1", strings.NewReader(`{"a":"b"}`))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "9.9.9.9:1234"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestRequestServiceHTTPForwardsOriginalRequest(t *testing.T) {
	var got struct {
		body      string
		query     string
		ctype     string
		forwarded string
		realIP    string
		callKey   string
		length    int64
		custom    string
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got.body = string(b)
		got.query = r.URL.RawQuery
		got.ctype = r.Header.Get("Content-Type")
		got.forwarded = r.Header.Get("X-Forwarded-For")
		got.realIP = r.Header.Get("X-Real-Ip")
		got.callKey = r.Header.Get("Call-Key")
		got.length = r.ContentLength
		got.custom = r.Header.Get("X-Drop-Me")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer upstream.Close()

	g := &Garden{cfg: cfg{Service: serviceCfg{CallKey: "secret"}}}
	span := opentracing.StartSpan("test")
	defer span.Finish()
	raw := []byte(`{"name":"a&b=c"}`)
	code, body, _, err := g.requestServiceHttp(span, upstream.URL+"/order", &req{
		Method:   http.MethodPost,
		UrlParam: "?user=ada&q=1",
		ClientIp: "1.1.1.1",
		RemoteIp: "2.2.2.2",
		Headers: MapData{
			"Content-Type":    "application/json",
			"Content-Length":  "999",
			"X-Forwarded-For": "1.1.1.1",
			"Connection":      "close, X-Drop-Me",
			"X-Drop-Me":       "nope",
		},
		RawBody: raw,
	}, 2000)
	if err != nil {
		t.Fatal(err)
	}
	if code != http.StatusOK || body != `{"ok":true}` {
		t.Fatalf("code=%d body=%s", code, body)
	}
	if got.body != string(raw) {
		t.Fatalf("body=%s", got.body)
	}
	if got.query != "user=ada&q=1" {
		t.Fatalf("query=%s", got.query)
	}
	if got.ctype != "application/json" {
		t.Fatalf("content-type=%s", got.ctype)
	}
	if got.forwarded != "1.1.1.1, 2.2.2.2" {
		t.Fatalf("xff=%s", got.forwarded)
	}
	if got.realIP != "1.1.1.1" {
		t.Fatalf("real-ip=%s", got.realIP)
	}
	if got.callKey != "secret" {
		t.Fatalf("call-key=%s", got.callKey)
	}
	if got.length != int64(len(raw)) {
		t.Fatalf("content-length=%d", got.length)
	}
	if got.custom != "" {
		t.Fatalf("hop-by-hop header forwarded: %s", got.custom)
	}
}

func TestLimiterTooManyRequests(t *testing.T) {
	g := &Garden{
		cfg:      cfg{Service: serviceCfg{CallRetry: "bad"}},
		services: map[string]*service{},
	}
	g.addNodeLocked("svc", "127.0.0.1:1:2")
	g.cfg.Routes = map[string]map[string]routeCfg{
		"svc": {
			"ping": {Type: "http", Path: "/ping", Limiter: "60/1", Timeout: 50},
		},
	}
	span := opentracing.StartSpan("test")
	defer span.Finish()
	request := &req{Method: http.MethodGet}

	if _, _, _, err := g.callService(span, "svc", "ping", request, nil, nil); err == nil {
		t.Fatal("expected upstream failure")
	}
	code, msg, _, err := g.callService(span, "svc", "ping", request, nil, nil)
	if err == nil || code != http.StatusTooManyRequests || msg != infoServerLimiter {
		t.Fatalf("code=%d msg=%s err=%v", code, msg, err)
	}
}
