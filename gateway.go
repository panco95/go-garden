package goms

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"net/http"
	"reflect"
	"time"
)

// Gateway 网关，http服务判断
func Gateway(ctx interface{}) {
	t := reflect.TypeOf(ctx)
	switch t.String() {
	case "*gin.Context":
		c := ctx.(*gin.Context)
		ginSupport(c)
		break
	default:
		break
	}
}

// gin框架支持
func ginSupport(c *gin.Context) {
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
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		Logger.Error(service + "/" + action + " return invalid format: " + data)
		c.JSON(http.StatusInternalServerError, FailRes())
		return
	}
	c.JSON(http.StatusOK, SuccessRes(result))
}

// CallService 调用Http服务
// @Description     服务重试：3次，失败依次等待0.1s、0.2s
// @param service   服务名称
// @param action    服务行为
// @param method    请求方式：GET || POST
// @param urlParam  url请求参数
// @param body      请求body结构体
// @param headers   请求头结构体
// @param requestId 请求id
func CallService(span opentracing.Span, service, action string, request *Request) (string, error) {
	route := viper.GetString("services." + service + "." + action)
	if len(route) == 0 {
		return "", errors.New("service route config not found")
	}
	serviceAddr, err := SelectServiceHttpAddr(service)
	if err != nil {
		return "", err
	}

	var result string
	// 服务重试3次
	for retry := 1; retry <= 3; retry++ {
		url := "http://" + serviceAddr + route + result
		result, err = RequestService(span, url, request)
		if err != nil {
			if retry >= 3 {
				return "", err
			} else {
				time.Sleep(time.Millisecond * time.Duration(retry*100))
				continue
			}
		}
		break
	}

	return result, nil
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
