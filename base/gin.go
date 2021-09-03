package base

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go-ms/utils"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ReqContext 请求上下文封装
type ReqContext struct {
	Time        string `json:"requestTime"`
	RequestId   string `json:"requestId"`
	Event       string `json:"event"`
	Method      string `json:"method"`
	Url         string `json:"url"`
	UrlParam    string `json:"urlParam"`
	ServiceName string `json:"serviceName"`
	ServiceId   string `json:"serviceId"`
	ClientIp    string `json:"clientIp"`
	Headers     Any    `json:"headers"`
	Body        Any    `json:"body"`
	Record      Any    `json:"record"`
}

// GinServer 开启Gin服务
func GinServer(port, serverName string, route func(r *gin.Engine)) {
	gin.SetMode("release")
	server := gin.Default()
	path, _ := os.Getwd()
	err := utils.CreateDir(path + "/runtime")
	if err != nil {
		log.Fatal("[Create runtime folder] ", err)
	}
	file, err := os.Create(fmt.Sprintf("%s/runtime/gin_%s.log", path, serverName))
	if err != nil {
		log.Fatal("[Create gin log file] ", err)
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
	route(server)

	log.Printf("[%s] Http Listen on port: %s", serverName, port)
	log.Fatal(server.Run(":" + port))
}

// Trace 链路追踪调试中间件
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 生成唯一requestId，提供给下游服务获取
		requestId := c.GetHeader("X-Request-Id")
		startEvent := "service.start"
		endEvent := "service.end"
		if requestId == "" || false == utils.ParseUuid(requestId) {
			requestId = utils.NewUuid()
			startEvent = "request.start"
			endEvent = "request.end"
		}

		reqContext := ReqContext{
			RequestId:   requestId,
			Event:       startEvent,
			Time:        utils.ToDatetime(start),
			ServiceName: ServerName,
			ServiceId:   ServerId,
			ClientIp:    c.ClientIP(),
			Method:      GetMethod(c),
			UrlParam:    GetUrlParam(c),
			Headers:     GetHeaders(c),
			Body:        GetBody(c),
			Url:         GetUrl(c),
		}
		// 记录远程调试日志
		RemoteTrace(&reqContext)
		// 封装到gin请求上下文
		c.Set("reqContext", reqContext)

		// 执行请求接口
		c.Next()
		c.Abort()

		// 接口执行完毕后执行
		// 记录远程调试日志，代表当前请求完毕
		end := time.Now()
		timeConsume := utils.TimeGap(start, end)
		reqContext.Event = endEvent
		reqContext.Time = utils.ToDatetime(end)
		reqContext.Record = Any{
			"timeConsume": timeConsume,
		}
		RemoteTrace(&reqContext)
	}
}

// CheckCallServiceKey 服务调用安全验证
func CheckCallServiceKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestKey := c.GetHeader("Call-Service-Key")
		if strings.Compare(requestKey, viper.GetString("callServiceKey")) != 0 {
			c.JSON(http.StatusForbidden, MakeFailResponse())
			c.Abort()
		}
	}
}

func GetMethod(c *gin.Context) string {
	return strings.ToUpper(c.Request.Method)
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
		// 获取json格式请求参数
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

func MakeSuccessResponse(data Any) Any {
	response := Any{
		"status": true,
	}
	for k, v := range data {
		response[k] = v
	}
	return response
}

func MakeFailResponse() Any {
	response := Any{
		"status": false,
	}
	return response
}
