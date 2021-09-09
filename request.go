package goms

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RequestService 请求下游服务封装
// @param url 服务http地址
// @param method 请求方式
// @param body 请求body结构体
// @param header 请求头结构体
// @param requestId 请求id
// @return string 响应内容
func RequestService(url, method string, body, headers Any, requestId string) (string, error) {
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
	request.Header.Set("X-Request-Id", requestId)

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
