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

// Request struct
type Request struct {
	Method   string  `json:"method"`
	Url      string  `json:"url"`
	UrlParam string  `json:"urlParam"`
	ClientIp string  `json:"clientIp"`
	Headers  MapData `json:"headers"`
	Body     MapData `json:"body"`
}

func (g *Garden) callService(span opentracing.Span, service, action string, request *Request, args, reply interface{}) (int, string, error) {
	s := g.cfg.Routes[service]
	if len(s) == 0 {
		return httpNotFound, infoNotFound, errors.New("service not found")
	}
	route := s[action]
	if (route.Type != "http" && route.Type != "rpc") ||
		(route.Type == "http" && len(route.Path) == 0) ||
		(route.Type == "rpc" && (args == nil || reply == nil)) {
		return httpNotFound, infoNotFound, errors.New("service route not found")
	}

	serviceAddr, nodeIndex, err := g.selectService(service)
	if err != nil {
		return httpNotFound, infoNotFound, err
	}

	// service limiter
	if route.Limiter != "" {
		second, quantity, err := limiterAnalyze(route.Limiter)
		if err != nil {
			g.Log(DebugLevel, "Limiter", err)
		} else if !g.limiterInspect(serviceAddr+"/"+service+"/"+action, second, quantity) {
			span.SetTag("break", "service limiter")
			return httpNotFound, infoServerLimiter, errors.New("server limiter")
		}
	}

	// service fusing
	if route.Fusing != "" {
		second, quantity, err := g.fusingAnalyze(route.Fusing)
		if err != nil {
			g.Log(ErrorLevel, "Fusing", err)
		} else if !g.fusingInspect(serviceAddr+"/"+service+"/"+action, second, quantity) {
			span.SetTag("break", "service fusing")
			return httpNotFound, infoServerFusing, errors.New("server fusing")
		}
	}

	// service call retry
	retry, err := retryAnalyze(g.cfg.Service.CallRetry)
	if err != nil {
		g.Log(DebugLevel, "Retry", err)
		retry = []int{0}
	}

	code, result, err := g.retryGo(service, action, retry, nodeIndex, span, route, request, args, reply)

	return code, result, err
}

func (g *Garden) requestServiceHttp(span opentracing.Span, url string, request *Request, timeout int) (int, string, error) {
	client := &http.Client{
		Timeout: time.Millisecond * time.Duration(timeout),
	}

	// encapsulation request body
	var s string
	for k, v := range request.Body {
		s += fmt.Sprintf("%v=%v&", k, v)
	}
	s = strings.Trim(s, "&")
	r, err := http.NewRequest(request.Method, url, strings.NewReader(s))
	if err != nil {
		return httpFail, "", err
	}

	// New request request
	for k, v := range request.Headers {
		r.Header.Add(k, v.(string))
	}
	// Add the body format header
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Increase calls to the downstream service security validation key
	r.Header.Set("Call-Key", g.cfg.Service.CallKey)

	// add request opentracing span header
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))

	res, err := client.Do(r)
	if err != nil {
		return httpFail, "", err
	}
	defer res.Body.Close()

	if res.StatusCode != httpOk {
		return res.StatusCode, "", errors.New("http status " + strconv.Itoa(res.StatusCode))
	}
	body2, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return httpFail, "", err
	}
	return httpOk, string(body2), nil
}

// CallRpc call service rpc method
func (g *Garden) CallRpc(span opentracing.Span, service, action string, args, reply interface{}) error {
	_, _, err := g.callService(span, service, action, nil, &args, &reply)
	if err != nil {
		return err
	}
	return nil
}
