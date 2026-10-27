package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/panco95/go-garden/core/log"
)

func (g *Garden) ginListen(listenAddress string, route func(r *gin.Engine), auth func() gin.HandlerFunc) error {
	engine, err := g.httpEngine(route, auth)
	if err != nil {
		return err
	}
	log.Infof("http", "listen on: %s", listenAddress)
	return engine.Run(listenAddress)
}

func (g *Garden) httpEngine(route func(r *gin.Engine), auth func() gin.HandlerFunc) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	if g.cfg.Service.Debug {
		engine.Use(gin.Logger())
		if err := createDir(g.cfg.RuntimePath); err != nil {
			return nil, err
		}
		file, err := os.Create(fmt.Sprintf("%s/gin.log", g.cfg.RuntimePath))
		if err != nil {
			return nil, err
		}
		gin.DefaultWriter = file
		engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
				param.ClientIP,
				param.TimeStamp.Format(time.RFC1123),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Latency,
				param.Request.UserAgent(),
				param.ErrorMessage)
		}))
	} else {
		gin.DefaultWriter = ioutil.Discard
	}

	engine.Use(g.openTracingMiddleware())
	if g.cfg.Service.AllowCors {
		engine.Use(cors)
	}
	engine.GET("/metrics", g.prometheus())
	if auth != nil {
		engine.Use(auth())
	}
	if g.cfg.Service.Debug {
		pprof.Register(engine)
	}
	notFound(engine)
	route(engine)
	return engine, nil
}

// GatewayRoute create gateway service type to use this
func (g *Garden) GatewayRoute(r *gin.Engine) {
	g.serviceType = 1
	r.Any("api/:service/:action", func(c *gin.Context) {
		g.gateway(c)
	})
}

func notFound(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		c.JSON(httpNotFound, gatewayFail(infoNotFound))
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(httpNotFound, gatewayFail(infoNotFound))
	})
}

func cors(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Headers", "*")
	ctx.Header("Access-Control-Allow-Methods", "*")
	ctx.Header("Access-Control-Expose-Headers", "*")
	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}
	ctx.Next()
}

func (g *Garden) openTracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		atomic.AddInt64(&g.requestProcess, 1)

		span := StartSpanFromHeader(c.Request.Header, c.Request.RequestURI)
		span.SetTag("CallType", "Http")
		span.SetTag("ServiceIp", g.GetServiceIp())
		span.SetTag("ServiceId", g.GetServiceId())
		span.SetTag("Status", "unfinished")

		request := req{
			Method:   getMethod(c),
			Url:      getUrl(c),
			UrlParam: getUrlParam(c),
			ClientIp: getClientIp(c),
			RemoteIp: c.RemoteIP(),
			Headers:  getHeaders(c),
			RawBody:  readRawBody(c),
			Body:     getBody(c),
		}
		traced := spanRequest(request)
		s, _ := json.Marshal(&traced)
		span.SetTag("Request", string(s))

		c.Set("span", span)
		c.Set("request", &request)

		c.Next()

		span.SetTag("Status", "finished")
		span.Finish()

		atomic.AddInt64(&g.requestProcess, -1)
		atomic.AddInt64(&g.requestFinish, 1)
	}
}

// CheckCallSafeMiddleware from call service safe check
func (g *Garden) CheckCallSafeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !g.checkCallSafe(c.GetHeader("Call-Key")) {
			c.JSON(httpUnauthorized, gatewayFail(infoNoAuth))
			c.Abort()
			return
		}
		c.Next()
	}
}

// SetContext set custom context
func SetContext(c *gin.Context, name string, val interface{}) {
	c.Set(name, val)
}

// GetContext get custom context
func GetContext(c *gin.Context, name string) (interface{}, error) {
	t, success := c.Get(name)
	if !success {
		return nil, errors.New(name + " is nil")
	}
	return t, nil
}

// GetRequest get request datatype from context
func GetRequest(c *gin.Context) *req {
	t, _ := GetContext(c, "request")
	r := t.(*req)
	return r
}

// GetSpan get opentracing span from context
func GetSpan(c *gin.Context) opentracing.Span {
	t, _ := GetContext(c, "span")
	r := t.(opentracing.Span)
	return r
}

func getMethod(c *gin.Context) string {
	return strings.ToUpper(c.Request.Method)
}

func getClientIp(c *gin.Context) string {
	return c.ClientIP()
}

func spanRequest(r req) req {
	traced := r
	traced.Body = nil
	traced.RawBody = nil
	if len(r.Headers) == 0 {
		return traced
	}
	headers := make(MapData, len(r.Headers))
	for k, v := range r.Headers {
		if strings.EqualFold(k, "Call-Key") {
			continue
		}
		headers[k] = v
	}
	traced.Headers = headers
	return traced
}

func readRawBody(c *gin.Context) []byte {
	if c.Request == nil || c.Request.Body == nil {
		return nil
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Request.Body = io.NopCloser(bytes.NewReader(nil))
		return nil
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return body
}

func getBody(c *gin.Context) MapData {
	body := MapData{}
	h := c.GetHeader("Content-Type")
	// 获取表单格式请求参数
	if strings.Contains(h, "multipart/form-data") || strings.Contains(h, "application/x-www-form-urlencoded") {
		c.PostForm("get_params_bug_fix")
		for k, v := range c.Request.PostForm {
			body[k] = v[0]
		}
	} else if strings.Contains(h, "application/json") {
		c.BindJSON(&body)
	}
	return body
}

func getUrl(c *gin.Context) string {
	return c.Request.URL.Path
}

func getUrlParam(c *gin.Context) string {
	requestUrl := c.Request.RequestURI
	urlSplit := strings.Split(requestUrl, "?")
	if len(urlSplit) > 1 {
		requestUrl = "?" + urlSplit[1]
	} else {
		requestUrl = ""
	}
	return requestUrl
}

func getHeaders(c *gin.Context) MapData {
	headers := MapData{}
	for k, v := range c.Request.Header {
		headers[k] = v[0]
	}
	return headers
}
