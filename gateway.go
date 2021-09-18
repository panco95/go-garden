package garden

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
)

func gateway(ctx interface{}) {
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
		c.JSON(http.StatusInternalServerError, gatewayFail())
		Log(ErrorLevel, "GetSpan", err)
		return
	}
	// request struct
	request, err := getRequest(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gatewayFail())
		Log(ErrorLevel, "GetRequestContext", err)
		span.SetTag("GetRequestContext", err)
		return
	}

	service := c.Param("service")
	action := c.Param("action")

	// request service
	data, err := CallService(span, service, action, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gatewayFail())
		Log(ErrorLevel, "CallService", err)
		span.SetTag("CallService", err)
		return
	}
	var result MapData
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		c.JSON(http.StatusInternalServerError, gatewayFail())
		Log(ErrorLevel, "ReturnInvalidFormat", err)
		span.SetTag("ReturnInvalidFormat", err)
		return
	}
	c.JSON(http.StatusOK, gatewaySuccess(result))
}
