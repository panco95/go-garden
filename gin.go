package goms

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// GinServer 开启Gin服务
// @param port 监听端口
// @param serviceName 服务名称
// @param route gin路由
// @param auth 鉴权中间件
func GinServer(port string, route func(r *gin.Engine), auth func() gin.HandlerFunc) error {
	gin.SetMode("release")
	server := gin.Default()
	path, _ := os.Getwd()
	err := CreateDir(path + "/runtime")
	if err != nil {
		return errors.New("[Create runtime folder] " + err.Error())
	}
	file, err := os.Create(fmt.Sprintf("%s/runtime/gin_%s.log", path, ServiceName))
	if err != nil {
		return errors.New("[Create gin log file] " + err.Error())
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
	server.Use(Trace())
	if auth != nil {
		server.Use(auth())
	}
	route(server)

	log.Printf("[%s] Http Listen on port: %s", ServiceName, port)
	return server.Run(":" + port)
}

// GatewayRoute 网关路由解析
// 第一个参数：下游服务名称
// 第二个参数：下游服务接口路由
func GatewayRoute(r *gin.Engine) {
	r.Any("api/:service/:action", func(c *gin.Context) {
		g, _ := c.Get("span")
		span := g.(opentracing.Span)

		// 服务名称和服务路由
		service := c.Param("service")
		action := c.Param("action")

		request, err := GetRequest(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, FailRes())
			span.LogKV("get request context fail")
			return
		}
		method := request.Method
		headers := request.Headers
		urlParam := request.UrlParam
		body := request.Body

		// 请求下游服务
		data, err := CallService(span, service, action, method, urlParam, body, headers)
		if err != nil {
			Logger.Error("call " + service + "/" + action + " error: " + err.Error())
			c.JSON(http.StatusInternalServerError, FailRes())
			return
		}
		var result Any
		err = json.Unmarshal([]byte(data), &result)
		if err != nil {
			Logger.Error(service + "/" + action + " return invalid format: " + data)
			c.JSON(http.StatusInternalServerError, FailRes())
			return
		}
		c.JSON(http.StatusOK, SuccessRes(result))
	})

	// 集群信息查询接口
	r.Any("cluster", func(c *gin.Context) {
		c.JSON(http.StatusOK, SuccessRes(Any{
			"services": Services,
		}))
	})
}

// Trace 链路追踪调试中间件
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		var span opentracing.Span
		// 尝试从header头获取span上下文
		wireContext, _ := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request.Header))
		// 如果header中没有span，会新建root span，如果有，则会新建child span
		span = opentracing.StartSpan(
			"http",
			//ext.RPCServerOption(wireContext),
			opentracing.ChildOf(wireContext),
		)

		request := Request{
			GetMethod(c), GetUrl(c), GetUrlParam(c), GetClientIp(c), GetHeaders(c), GetBody(c)}
		span.SetTag("Request", request)

		// 保存请求上下文
		c.Set("span", span)
		c.Set("request", &request)

		// 执行请求接口
		c.Next()

		span.Finish()
	}
}

// CheckCallServiceKey 服务调用安全验证
func CheckCallServiceKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestKey := c.GetHeader("Call-Service-Key")
		if strings.Compare(requestKey, viper.GetString("callServiceKey")) != 0 {
			c.JSON(http.StatusForbidden, FailRes())
			c.Abort()
		}
	}
}

// GetRequest 获取request上下文
func GetRequest(c *gin.Context) (*Request, error) {
	t, success := c.Get("request")
	if !success {
		return nil, errors.New("traceLog is nil")
	}
	r := t.(*Request)
	return r, nil
}

// GetMethod 获取请求方式
func GetMethod(c *gin.Context) string {
	return strings.ToUpper(c.Request.Method)
}

// GetClientIp 获取请求客户端ip
func GetClientIp(c *gin.Context) string {
	return c.ClientIP()
}

// GetBody 获取请求body
func GetBody(c *gin.Context) Any {
	body := Any{}
	h := c.GetHeader("Content-Type")
	// 获取表单格式请求参数
	if strings.Contains(h, "multipart/form-data") || strings.Contains(h, "application/x-www-form-urlencoded") {
		c.PostForm("get_params_bug_fix")
		for k, v := range c.Request.PostForm {
			body[k] = v[0]
		}
		// 获取json格式请求参数
	} else if strings.Contains(h, "application/json") {
		c.BindJSON(&body)
	}
	return body
}

// GetUrl 获取请求路径
func GetUrl(c *gin.Context) string {
	return c.Request.URL.Path
}

// GetUrlParam 获取请求query参数
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

// GetHeaders 获取请求头map
func GetHeaders(c *gin.Context) Any {
	headers := Any{}
	for k, v := range c.Request.Header {
		headers[k] = v[0]
	}
	return headers
}
