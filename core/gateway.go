package core

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
)

func (g *Garden) gateway(c *gin.Context) {
	// openTracing span
	span, err := GetSpan(c)
	if err != nil {
		c.JSON(httpFail, gatewayFail(infoServerError))
		g.Log(ErrorLevel, "GetSpan", err)
		return
	}
	// request struct
	request, err := getRequest(c)
	if err != nil {
		c.JSON(httpFail, gatewayFail(infoServerError))
		g.Log(ErrorLevel, "GetRequestContext", err)
		span.SetTag("GetRequestContext", err)
		return
	}

	service := c.Param("service")
	action := c.Param("action")

	// request service
	code, data, err := g.callService(span, service, action, request, nil, nil)
	if err != nil {
		c.JSON(code, gatewayFail(data))
		g.Log(ErrorLevel, "CallService", err)
		span.SetTag("CallService", err)
		return
	}
	var result MapData
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		c.JSON(httpFail, gatewayFail(infoServerError))
		g.Log(ErrorLevel, "ReturnInvalidFormat", err)
		span.SetTag("ReturnInvalidFormat", err)
		return
	}
	c.JSON(code, gatewaySuccess(result))
}

func gatewaySuccess(data MapData) MapData {
	response := MapData{
		"status": true,
	}
	for k, v := range data {
		response[k] = v
	}
	return response
}

func gatewayFail(message string) MapData {
	response := MapData{
		"status": false,
		"msg":    message,
	}
	return response
}
