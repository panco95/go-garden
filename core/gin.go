package core

import (
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (g *Garden) ginListen(listenAddress string, route func(r *gin.Engine), auth func() gin.HandlerFunc) error {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())
	if g.cfg.Service.Debug {
		if err := createDir(g.cfg.RuntimePath); err != nil {
			return err
		}
		file, err := os.Create(fmt.Sprintf("%s/gin.log", g.cfg.RuntimePath))
		if err != nil {
			return err
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
		pprof.Register(engine)
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
	notFound(engine)
	route(engine)

	log.Infof("http", "listen on: %s", listenAddress)
	return engine.Run(listenAddress)
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

func (g *Garden) prometheus() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func cors(ctx *gin.Context) {
	method := ctx.Request.Method
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Headers", "*")
	ctx.Header("Access-Control-Allow-Methods", "*")
	ctx.Header("Access-Control-Expose-Headers", "*")
	ctx.Header("Access-Control-Allow-Credentials", "true")
	if method == "OPTIONS" {
		ctx.AbortWithStatus(http.StatusNoContent)
	}
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
			getMethod(c),
			getUrl(c),
			getUrlParam(c),
			getClientIp(c),
			getHeaders(c),
			getBody(c)}
		s, _ := json.Marshal(&request)
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
			c.JSON(httpNotFound, gatewayFail(infoNoAuth))
			c.Abort()
		}
	}
}

//SetContext set custom context
func SetContext(c *gin.Context, name string, val interface{}) {
	c.Set(name, val)
}

//GetContext get custom context
func GetContext(c *gin.Context, name string) (interface{}, error) {
	t, success := c.Get(name)
	if !success {
		return nil, errors.New(name + " is nil")
	}
	return t, nil
}

//GetRequest get request datatype from context
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
