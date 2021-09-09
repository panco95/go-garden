package goms

import (
	"errors"
	"github.com/spf13/viper"
	"time"
)

// CallService 调用Http服务
// @Description     服务重试：3次，失败依次等待0.1s、0.2s
// @param service   服务名称
// @param action    服务行为
// @param method    请求方式：GET || POST
// @param urlParam  url请求参数
// @param body      请求body结构体
// @param headers   请求头结构体
// @param requestId 请求id
func CallService(service, action, method, urlParam string, body, headers Any, requestId string) (string, error) {
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
