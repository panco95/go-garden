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

// CallService call the service api
func (g *Garden) CallService(span opentracing.Span, service, action string, request *Request, args, reply interface{}) (int, string, error) {
	s := g.cfg.Routes[service]
	if len(s) == 0 {
		return 404, NotFound, errors.New("service not found")
	}
	route := s[action]
	if route.Type == "api" && len(route.Path) == 0 {
		return 404, NotFound, errors.New("service route not found")
	}

	serviceAddr, nodeIndex, err := g.selectService(service)
	if err != nil {
		return 404, NotFound, err
	}

	// service limiter
	if route.Limiter != "" {
		second, quantity, err := limiterAnalyze(route.Limiter)
		if err != nil {
			g.Log(DebugLevel, "Limiter", err)
		} else if !g.limiterInspect(serviceAddr+"/"+service+"/"+action, second, quantity) {
			span.SetTag("break", "service limiter")
			return 403, ServerLimiter, errors.New("server limiter")
		}
	}

	// service fusing
	if route.Fusing != "" {
		second, quantity, err := g.fusingAnalyze(route.Fusing)
		if err != nil {
			g.Log(DebugLevel, "Fusing", err)
		} else if !g.fusingInspect(serviceAddr+"/"+service+"/"+action, second, quantity) {
			span.SetTag("break", "service fusing")
			return 403, ServerFusing, errors.New("server fusing")
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
		return 500, "", err
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
		return 500, "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return res.StatusCode, "", errors.New("http status " + strconv.Itoa(res.StatusCode))
	}
	body2, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 500, "", err
	}
	return 200, string(body2), nil
}
