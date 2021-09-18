package core

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (g *Garden) gateway(c *gin.Context) {
	// openTracing span
	span, err := GetSpan(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gatewayFail())
		g.Log(ErrorLevel, "GetSpan", err)
		return
	}
	// request struct
	request, err := getRequest(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gatewayFail())
		g.Log(ErrorLevel, "GetRequestContext", err)
		span.SetTag("GetRequestContext", err)
		return
	}

	service := c.Param("service")
	action := c.Param("action")

	// request service
	data, err := g.CallService(span, service, action, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gatewayFail())
		g.Log(ErrorLevel, "CallService", err)
		span.SetTag("CallService", err)
		return
	}
	var result MapData
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		c.JSON(http.StatusInternalServerError, gatewayFail())
		g.Log(ErrorLevel, "ReturnInvalidFormat", err)
		span.SetTag("ReturnInvalidFormat", err)
		return
	}
	c.JSON(http.StatusOK, gatewaySuccess(result))
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

func gatewayFail() MapData {
	response := MapData{
		"status": false,
	}
	return response
}
