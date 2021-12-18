package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"net/http"
	"os"
	"strings"
	"time"
)

func (g *Garden) ginListen(listenAddress string, route func(r *gin.Engine), auth func() gin.HandlerFunc) error {
	// init
	gin.SetMode("release")
	server := gin.Default()

	// log
	if err := createDir(g.cfg.runtimePath); err != nil {
		return err
	}
	file, err := os.Create(fmt.Sprintf("%s/gin.log", g.cfg.runtimePath))
	if err != nil {
		return err
	}
	gin.DefaultWriter = file

	// middlewares
	server.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
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
	server.Use(gin.Recovery())
	server.Use(g.openTracingMiddleware())
	if g.cfg.Service.AllowCors {
		server.Use(g.cors)
	}

	// monitoring
	g.prometheus(server)
	pprof.Register(server)

	if auth != nil {
		server.Use(auth())
	}

	// routes
	notFound(server)
	route(server)

	// run
	g.Log(InfoLevel, "http", fmt.Sprintf("listen on: %s", listenAddress))
	return server.Run(listenAddress)
}

// GatewayRoute create gateway service, use this gin route
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

func (g *Garden) prometheus(r *gin.Engine) {
	r.GET("/metrics", func(c *gin.Context) {
		data := MapData{
			"RequestProcess": g.RequestProcess.String(),
			"RequestFinish":  g.RequestFinish.String(),
		}
		g.Metrics.Range(func(k, v interface{}) bool {
			data[k.(string)] = v
			return true
		})
		c.String(200, GenMetricsData(data))
	})
}

func (g *Garden) cors(ctx *gin.Context) {
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
		g.RequestProcess.Inc()

		span := StartSpanFromHeader(c.Request.Header, c.Request.RequestURI)
		span.SetTag("CallType", "Http")
		span.SetTag("ServiceIp", g.ServiceIp)
		span.SetTag("ServiceId", g.ServiceId)
		span.SetTag("Status", "unfinished")

		request := Request{
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

		g.RequestProcess.Dec()
		g.RequestFinish.Inc()
	}
}

// CheckCallSafeMiddleware create service use this middleware
func (g *Garden) CheckCallSafeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !g.checkCallSafe(c.GetHeader("Call-Key")) {
			c.JSON(httpNotFound, gatewayFail(infoNoAuth))
			c.Abort()
		}
	}
}

func getContext(c *gin.Context, name string) (interface{}, error) {
	t, success := c.Get(name)
	if !success {
		return nil, errors.New(name + " is nil")
	}
	return t, nil
}

func getRequest(c *gin.Context) (*Request, error) {
	t, err := getContext(c, "request")
	if err != nil {
		return nil, err
	}
	r := t.(*Request)
	return r, nil
}

// GetSpan service get opentracing span at gin context
func GetSpan(c *gin.Context) (opentracing.Span, error) {
	t, err := getContext(c, "span")
	if err != nil {
		return nil, err
	}
	r := t.(opentracing.Span)
	return r, nil
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
