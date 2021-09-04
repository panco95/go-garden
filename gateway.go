package goms

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

// GatewayRoute 网关路由解析
// 第一个参数：下游服务名称
// 第二个参数：下游服务接口路由
func GatewayRoute(r *gin.Engine) {
	r.Any("api/:service/:action", func(c *gin.Context) {
		// 服务名称和服务路由
		service := c.Param("service")
		action := c.Param("action")
		// 从reqTrace获取相关请求报文
		traceLog, err := GetTraceLog(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, FailRes())
			Logger.Error()
			return
		}
		method := traceLog.Request.Method
		headers := traceLog.Request.Headers
		urlParam := traceLog.Request.UrlParam
		body := traceLog.Request.Body
		requestId := traceLog.RequestId

		// 请求下游服务
		data, err := CallService(c, service, action, method, urlParam, body, headers, requestId)
		if err != nil {
			Logger.Error(ErrorLog("call " + service + "/" + action + " error: " + err.Error()))
			c.JSON(http.StatusInternalServerError, FailRes())
			return
		}
		var result Any
		err = json.Unmarshal([]byte(data), &result)
		if err != nil {
			Logger.Error(ErrorLog(service + "/" + action + " return invalid format: " + data))
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

// CallService 调用下游服务
// 服务重试：3次
// 失败依次等待0.1s、0.2s
func CallService(c *gin.Context, service, action, method, urlParam string, body, headers Any, requestId string) (string, error) {
	route := viper.GetString("services." + service + "." + action)
	if len(route) == 0 {
		return "", errors.New("service route config not found")
	}
	serviceAddr, err := selectServiceNode(service)
	if err != nil {
		return "", err
	}

	var result string
	for retry := 1; retry <= 3; retry++ {
		url := "http://" + serviceAddr + route + urlParam
		result, err = RequestService(url, method, body, headers, requestId)
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

// 根据服务名称选择下游服务node
// 负载均衡轮询+1
func selectServiceNode(name string) (string, error) {
	if _, ok := Services[name]; !ok {
		return "", errors.New("service key not found")
	}
	serviceHttpAddr, err := AnalyzeHttpAddr(name, Services[name].PollNext)
	if err != nil {
		return "", err
	}
	go func() {
		serviceNum := len(Services[name].Nodes)
		index := Services[name].PollNext
		ServicesLock.Lock()
		if index >= serviceNum-1 {
			Services[name].PollNext = 0
		} else {
			Services[name].PollNext = index + 1
		}
		Services[name].RequestFinish++
		ServicesLock.Unlock()
	}()
	return serviceHttpAddr, nil
}
