package base

import (
	"github.com/gin-gonic/gin"
	"go-ms/pkg/base/global"
	"strings"
)

func GetMethod(c *gin.Context) string {
	return strings.ToUpper(c.Request.Method)
}

func GetBody(c *gin.Context) global.Any {
	body := global.Any{}
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

func GetHeaders(c *gin.Context) global.Any {
	headers := global.Any{}
	for k, v := range c.Request.Header {
		headers[k] = v[0]
	}
	return headers
}

func MakeSuccessResponse(data global.Any) global.Any {
	response := global.Any{
		"status": true,
	}
	for k, v := range data {
		response[k] = v
	}
	return response
}

func MakeFailResponse() global.Any {
	response := global.Any{
		"status": false,
	}
	return response
}
