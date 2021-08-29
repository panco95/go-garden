package cluster

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go-ms/pkg/base/global"
	"go-ms/pkg/base/request"
	"io/ioutil"
	"net/http"
)

// 网关路由解析
// 第一个参数：下游服务名称
// 第二个参数：下游服务接口路由
func GatewayRoute(r *gin.Engine) {
	r.Any("api/:server/:action", func(c *gin.Context) {
		// 服务名称和服务路由
		server := c.Param("server")
		action := c.Param("action")
		// 报文
		method := request.GetMethod(c)
		headers := request.GetHeaders(c)
		urlParam := request.GetUrlParam(c)
		body := request.GetBody(c)

		// 请求下游服务
		data, err := CallService(server, action, method, urlParam, body, headers)
		if err != nil {
			c.JSON(http.StatusInternalServerError, request.MakeFailResponse())
			return
		}
		var result global.Any
		err = json.Unmarshal([]byte(data), &result)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, request.MakeFailResponse())
			return
		}
		c.JSON(http.StatusOK, request.MakeSuccessResponse(result))
	})

	// 集群信息查询接口
	r.Any("cluster", func(c *gin.Context) {
		c.JSON(http.StatusOK, request.MakeSuccessResponse(global.Any{
			"servers": Servers,
		}))
	})
}

// 调用下游服务
func CallService(serverName, action, method, urlParam string, body, headers global.Any) (string, error) {
	route := viper.GetString(serverName + "." + action)
	if len(route) == 0 {
		return "", nil
	}
	serverAddr, err := chooseServer(serverName)
	if err != nil {
		return "", err
	}

	url := "http://" + serverAddr + route + urlParam
	result, err := httpReq(url, method, body, headers)
	if err != nil {
		return "", err
	}
	return result, nil
}

// 根据服务名称选择下游服务
// 负载均衡轮询+1
func chooseServer(serverName string) (string, error) {
	if _, ok := Servers[serverName]; !ok {
		return "", errors.New("Server not found")
	}
	serverHttpAddr, err := AnalyzeHttpAddr(serverName, Servers[serverName].PollNext)
	if err != nil {
		return "", err
	}
	go func() {
		serverNum := len(Servers[serverName].Nodes)
		index := Servers[serverName].PollNext
		ServersLock.Lock()
		if index >= serverNum-1 {
			Servers[serverName].PollNext = 0
		} else {
			Servers[serverName].PollNext = index + 1
		}
		Servers[serverName].RequestFinish++
		ServersLock.Unlock()
	}()
	return serverHttpAddr, nil
}

// 请求下游服务
// 一致封装为application/json格式报文进行请求
func httpReq(url, method string, body, headers global.Any) (string, error) {
	client := &http.Client{}
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
	body2, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body2), nil
}
