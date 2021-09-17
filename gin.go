package garden

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/panco95/go-garden/utils"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func GinServer(port string, route func(r *gin.Engine), auth func() gin.HandlerFunc) error {
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
	server.Use(OpenTracingMiddleware())
	if auth != nil {
		server.Use(auth())
	}
	route(server)

	log.Printf("[%s] Http listen on port: %s", Config.ServiceName, port)
	return server.Run(":" + port)
}

func GatewayRoute(r *gin.Engine) {
	r.Any("api/:service/:action", func(c *gin.Context) {
		Gateway(c)
	})
}

func OpenTracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := StartSpanFromHeader(c.Request.Header, c.Request.RequestURI)
		span.SetTag("Result", "running")
		RequestTracing(c, span)

		c.Next()

		span.SetTag("Result", "success")
		span.Finish()
	}
}

func CheckCallSafeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !CheckCallSafe(c.GetHeader("Call-Service-Key")) {
			c.JSON(http.StatusForbidden, GatewayFail())
			c.Abort()
		}
	}
}

func GetContext(c *gin.Context, name string) (interface{}, error) {
	t, success := c.Get(name)
	if !success {
		return nil, errors.New(name + " is nil")
	}
	return t, nil
}

func GetRequest(c *gin.Context) (*Request, error) {
	t, err := GetContext(c, "request")
	if err != nil {
		return nil, err
	}
	r := t.(*Request)
	return r, nil
}

func GetSpan(c *gin.Context) (opentracing.Span, error) {
	t, err := GetContext(c, "span")
	if err != nil {
		return nil, err
	}
	r := t.(opentracing.Span)
	return r, nil
}

func GetMethod(c *gin.Context) string {
	return strings.ToUpper(c.Request.Method)
}

func GetClientIp(c *gin.Context) string {
	return c.ClientIP()
}

func GetBody(c *gin.Context) Any {
	body := Any{}
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

func GetUrl(c *gin.Context) string {
	return c.Request.URL.Path
}

func GetUrlParam(c *gin.Context) string {
	requestUrl := c.Request.RequestURI
	urlSplit := strings.Split(requestUrl, "?")
	if len(urlSplit) > 1 {
		requestUrl = "?" + urlSplit[1]
	} else {
		requestUrl = ""
	}
	return requestUrl
}

func GetHeaders(c *gin.Context) Any {
	headers := Any{}
	for k, v := range c.Request.Header {
		headers[k] = v[0]
	}
	return headers
}
