package core

import (
	"errors"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	Method   string  `json:"method"`
	Url      string  `json:"url"`
	UrlParam string  `json:"urlParam"`
	ClientIp string  `json:"clientIp"`
	Headers  MapData `json:"headers"`
	Body     MapData `json:"body"`
}

func (g *Garden) CallService(span opentracing.Span, service, action string, request *Request) (string, error) {
	route := g.Cfg.Routes[service][action]
	if len(route) == 0 {
		return "", errors.New("service route config not found")
	}
	serviceAddr, nodeIndex, err := g.selectServiceHttpAddr(service)
	if err != nil {
		return "", err
	}

	var result string
	url := "http://" + serviceAddr + route + result
	for retry := 1; retry <= 3; retry++ {
		sm := serviceOperate{
			operate: "incWaiting",
			serviceName: service,
			nodeIndex: nodeIndex,
		}
		g.serviceManager <- sm
		result, err = g.requestService(span, url, request)
		sm.operate = "decWaiting"
		g.serviceManager <- sm
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

func (g *Garden) requestService(span opentracing.Span, url string, request *Request) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// encapsulation request body
	var s string
	for k, v := range request.Body {
		s += fmt.Sprintf("%v=%v&", k, v)
	}
	s = strings.Trim(s, "&")
	r, err := http.NewRequest(request.Method, url, strings.NewReader(s))
	if err != nil {
		return "", err
	}

	// New request request
	for k, v := range request.Headers {
		r.Header.Add(k, v.(string))
	}
	// Add the body format header
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Increase calls to the downstream service security validation key
	r.Header.Set("Call-Service-Key", g.Cfg.CallServiceKey)

	// add request opentracing span header
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))

	res, err := client.Do(r)
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
