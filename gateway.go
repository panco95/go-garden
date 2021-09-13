package goms

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
)

// Gateway 网关，http服务判断
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

// 网关：gin框架支持
func gatewayGin(c *gin.Context) {
	// openTracing span
	span, err := GetSpan(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, FailRes())
		Logger.Error("get span fail")
		return
	}
	// request结构体
	request, err := GetRequest(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, FailRes())
		span.LogKV("get request context fail")
		return
	}
	// 服务名称和服务路由
	service := c.Param("service")
	action := c.Param("action")

	// 请求下游服务
	data, err := CallService(span, service, action, request)
	if err != nil {
		Logger.Error("call " + service + "/" + action + " error: " + err.Error())
		c.JSON(http.StatusInternalServerError, FailRes())
		return
	}
	var result Any
	if err := json.Unmarshal([]byte(data), &result) ;err != nil {
		Logger.Error(service + "/" + action + " return invalid format: " + data)
		c.JSON(http.StatusInternalServerError, FailRes())
		return
	}
	c.JSON(http.StatusOK, SuccessRes(result))
}

// SuccessRes 成功响应
func SuccessRes(data Any) Any {
	response := Any{
		"status": true,
	}
	for k, v := range data {
		response[k] = v
	}
	return response
}

// FailRes 失败响应
func FailRes() Any {
	response := Any{
		"status": false,
	}
	return response
}
