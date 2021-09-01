package base

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// 网关路由解析
// 第一个参数：下游服务名称
// 第二个参数：下游服务接口路由
func GatewayRoute(r *gin.Engine) {
	r.Any("api/:service/:action", func(c *gin.Context) {
		// 服务名称和服务路由
		service := c.Param("service")
		action := c.Param("action")
		// 报文
		method := GetMethod(c)
		headers := GetHeaders(c)
		urlParam := GetUrlParam(c)
		body := GetBody(c)

		// 请求下游服务
		data, err := CallService(service, action, method, urlParam, body, headers)
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

// 调用下游服务
// 服务重试：3次
// 失败依次等待0.1s、0.2s
func CallService(service, action, method, urlParam string, body, headers Any) (string, error) {
	route := viper.GetString(service + "." + action)
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
		result, err = httpReq(url, method, body, headers)
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
func httpReq(url, method string, body, headers Any) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	bodyString, err := json.Marshal(&body)
	reader := bytes.NewReader(bodyString)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequest(method, url, reader)
	if err != nil {
		return "", err
	}

	for k, v := range headers {
		request.Header.Add(k, v.(string))
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Call-Service-Key", viper.GetString("callServiceKey")) //服务调用验证信息

	res, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", errors.New("http status " + strconv.Itoa(res.StatusCode))
	}
	body2, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body2), nil
}
