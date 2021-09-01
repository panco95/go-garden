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
	route(server)

	log.Printf("[%s] Http Listen on port: %s", serverName, port)
	log.Fatal(server.Run(":" + port))
}

// 服务调用验证
func CheckCallServiceKey(c *gin.Context) {
	requestKey := c.GetHeader("Call-Service-Key")
	if strings.Compare(requestKey, viper.GetString("callServiceKey")) != 0 {
		c.JSON(http.StatusForbidden, MakeFailResponse())
		c.Abort()
		return
	}
}

func GetMethod(c *gin.Context) string {
	return strings.ToUpper(c.Request.Method)
}

func GetBody(c *gin.Context) Any {
	body := Any{}
	c.PostForm("get_params_bug_fix")
	for k, v := range c.Request.PostForm {
		body[k] = v[0]
	}
	if len(body) < 1 {
		c.BindJSON(&body)
	}
	return body
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
