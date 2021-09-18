package core

import (
	"errors"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/panco95/go-garden/core/utils"
	"net/http"
	"os"
	"strings"
	"time"
)

func (g *Garden) runGin(port string, route func(r *gin.Engine), auth func() gin.HandlerFunc) error {
	gin.SetMode("release")
	server := gin.Default()
	path, _ := os.Getwd()
	if err := utils.CreateDir(path + "/runtime"); err != nil {
		return err
	}
	file, err := os.Create(fmt.Sprintf("%s/runtime/gin.log", path))
	if err != nil {
		return err
	}
	gin.DefaultWriter = file
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
	if auth != nil {
		server.Use(auth())
	}
	route(server)

	pprof.Register(server)

	g.Log(InfoLevel, g.Cfg.ServiceName, fmt.Sprintf("Http listen on port: %s", port))
	return server.Run(":" + port)
}

func (g *Garden) GatewayRoute(r *gin.Engine) {
	r.Any("api/:service/:action", func(c *gin.Context) {
		g.gateway(c)
	})
}

func (g *Garden) openTracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := startSpanFromHeader(c.Request.Header, c.Request.RequestURI)
		span.SetTag("ServiceIp", g.serviceIp)
		span.SetTag("ServiceId", g.serviceId)
		span.SetTag("Result", "running")
		requestTracing(c, span)

		c.Next()

		span.SetTag("Result", "success")
		span.Finish()
	}
}

func (g *Garden) CheckCallSafeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !g.checkCallSafe(c.GetHeader("Call-Service-Key")) {
			c.JSON(http.StatusForbidden, gatewayFail())
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
