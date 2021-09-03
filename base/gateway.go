package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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
		// 从中间件获取相关请求报文
		reqContext, _ := c.Get("reqContext")
		method := reqContext.(ReqContext).Method
		headers := reqContext.(ReqContext).Headers
		urlParam := reqContext.(ReqContext).UrlParam
		body := reqContext.(ReqContext).Body

		// 请求下游服务
		data, err := CallService(c, service, action, method, urlParam, body, headers)
		if err != nil {
			Logger.Error(ErrorLog("call " + service + "/" + action + " error: " + err.Error()))
			c.JSON(http.StatusInternalServerError, MakeFailResponse())
			return
		}
		var result Any
		err = json.Unmarshal([]byte(data), &result)
		if err != nil {
			Logger.Error(ErrorLog(service + "/" + action + " return invalid format: " + data))
			c.JSON(http.StatusInternalServerError, MakeFailResponse())
			return
		}
		c.JSON(http.StatusOK, MakeSuccessResponse(result))
	})

	// 集群信息查询接口
	r.Any("cluster", func(c *gin.Context) {
		c.JSON(http.StatusOK, MakeSuccessResponse(Any{
			"servers": Servers,
		}))
	})
}

// CallService 调用下游服务
// 服务重试：3次
// 失败依次等待0.1s、0.2s
func CallService(c *gin.Context, service, action, method, urlParam string, body, headers Any) (string, error) {
	route := viper.GetString("services." + service + "." + action)
	if len(route) == 0 {
		return "", errors.New("service route config not found")
	}
	serviceAddr, err := chooseServiceNode(service)
	if err != nil {
		return "", err
	}

	var result string
	for retry := 1; retry <= 3; retry++ {
		url := "http://" + serviceAddr + route + urlParam
		result, err = httpReq(c, url, method, body, headers)
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
func chooseServiceNode(service string) (string, error) {
	if _, ok := Servers[service]; !ok {
		return "", errors.New("service key not found")
	}
	serviceHttpAddr, err := AnalyzeHttpAddr(service, Servers[service].PollNext)
	if err != nil {
		return "", err
	}
	go func() {
		serverNum := len(Servers[service].Nodes)
		index := Servers[service].PollNext
		ServersLock.Lock()
		if index >= serverNum-1 {
			Servers[service].PollNext = 0
		} else {
			Servers[service].PollNext = index + 1
		}
		Servers[service].RequestFinish++
		ServersLock.Unlock()
	}()
	return serviceHttpAddr, nil
}

// 请求下游服务
// 一致封装为application/json格式报文进行请求
func httpReq(c *gin.Context, url, method string, body, headers Any) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// 封装请求body
	var s string
	for k, v := range body {
		s += fmt.Sprintf("%v=%v&", k, v)
	}
	s = strings.Trim(s, "&")
	request, err := http.NewRequest(method, url, strings.NewReader(s))
	if err != nil {
		return "", err
	}

	// 新建request请求
	for k, v := range headers {
		request.Header.Add(k, v.(string))
	}
	// 增加body格式头
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// 增加调用下游服务安全验证key
	request.Header.Set("Call-Service-Key", viper.GetString("callServiceKey"))
	// 增加请求ID，为了存储多服务调用链路日志
	rc, _ := c.Get("reqContext")
	request.Header.Set("X-Request-Id", rc.(ReqContext).RequestId)

	res, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	// 请求失败
	if res.StatusCode != http.StatusOK {
		return "", errors.New("http status " + strconv.Itoa(res.StatusCode))
	}
	body2, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body2), nil
}
