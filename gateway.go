package garden

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
)

func Gateway(ctx interface{}) {
	t := reflect.TypeOf(ctx)
	switch t.String() {
	case "*gin.Context":
		c := ctx.(*gin.Context)
		gatewayGin(c)
		break
	default:
		break
	}
}

func gatewayGin(c *gin.Context) {
	// openTracing span
	span, err := GetSpan(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GatewayFail())
		Logger.Errorf("[%s] %s", "GetSpan", err)
		return
	}
	// request struct
	request, err := GetRequest(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GatewayFail())
		Logger.Errorf("[%s] %s", "GetRequestContext", err)
		span.SetTag("GetRequestContext", err)
		return
	}

	service := c.Param("service")
	action := c.Param("action")

	// request service
	data, err := CallService(span, service, action, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GatewayFail())
		Logger.Errorf("[%s][%s/%s] %s", "CallService", service, action, err)
		span.SetTag("CallService", err)
		return
	}
	var result Any
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		c.JSON(http.StatusInternalServerError, GatewayFail())
		Logger.Errorf("[%s][%s/%s] %s", "ReturnInvalidFormat", service, action, err)
		span.SetTag("ReturnInvalidFormat", err)
		return
	}
	c.JSON(http.StatusOK, GatewaySuccess(result))
}
