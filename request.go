package goms

import (
	"errors"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Request HTTP请求 调试结构体
// Method 请求方式
// Url 请求地址
// UrlParam 请求query参数
// ClientIp 请求客户端IP
// Headers 请求头map
// Body 请求体map
type Request struct {
	Method   string `json:"method"`
	Url      string `json:"url"`
	UrlParam string `json:"urlParam"`
	ClientIp string `json:"clientIp"`
	Headers  Any    `json:"headers"`
	Body     Any    `json:"body"`
}

// RequestService 请求下游服务封装
// @param span opentracing span
// @param url 服务http地址
// @param method 请求方式
// @param body 请求body结构体
// @param header 请求头结构体
// @param requestId 请求id
// @return string 响应内容
func RequestService(span opentracing.Span, url string, request *Request) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// 封装请求body
	var s string
	for k, v := range request.Body {
		s += fmt.Sprintf("%v=%v&", k, v)
	}
	s = strings.Trim(s, "&")
	r, err := http.NewRequest(request.Method, url, strings.NewReader(s))
	if err != nil {
		return "", err
	}

	// 新建request请求
	for k, v := range request.Headers {
		r.Header.Add(k, v.(string))
	}
	// 增加body格式头
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// 增加调用下游服务安全验证key
	r.Header.Set("Call-Service-Key", viper.GetString("callServiceKey"))

	// 给请求封装opentracing-span header头
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))

	res, err := client.Do(r)
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
