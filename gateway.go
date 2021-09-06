package goms

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"time"
)

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


func SuccessRes(data Any) Any {
	response := Any{
		"status": true,
	}
	for k, v := range data {
		response[k] = v
	}
	return response
}

func FailRes() Any {
	response := Any{
		"status": false,
	}
	return response
}
