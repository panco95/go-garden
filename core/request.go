package core

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/panco95/go-garden/core/log"
)

// Request datatype
type req struct {
	Method   string  `json:"method"`
	Url      string  `json:"url"`
	UrlParam string  `json:"urlParam"`
	ClientIp string  `json:"clientIp"`
	RemoteIp string  `json:"remoteIp"`
	Headers  MapData `json:"headers"`
	Body     MapData `json:"body"`
	RawBody  []byte  `json:"-"`
}

func (g *Garden) callService(span opentracing.Span, service, action string, request *req, args, reply interface{}) (int, string, http.Header, error) {
	s := g.cfg.Routes[service]
	if len(s) == 0 {
		return httpNotFound, infoNotFound, nil, errors.New("service not found")
	}
	route := s[action]
	if (route.Type != "http" && route.Type != "rpc") ||
		(route.Type == "http" && len(route.Path) == 0) ||
		(route.Type == "rpc" && (args == nil || reply == nil)) {
		return httpNotFound, infoNotFound, nil, errors.New("service route not found")
	}

	target, err := g.selectService(service)
	if err != nil {
		return httpNotFound, infoNotFound, nil, err
	}

	// service limiter
	if route.Limiter != "" {
		second, quantity, err := limiterAnalyze(route.Limiter)
		if err != nil {
			log.Error("limiter", err)
		} else if !g.limiterInspect(breakerKey(target.Addr, service, action), second, quantity) {
			span.SetTag("break", "service limiter")
			return httpTooManyRequests, infoServerLimiter, nil, errors.New("server limiter")
		}
	}

	// service fusing
	if route.Fusing != "" {
		second, quantity, err := g.fusingAnalyze(route.Fusing)
		if err != nil {
			log.Error("fusingAnalyze", err)
		} else if !g.fusingInspect(breakerKey(target.Addr, service, action), second, quantity) {
			span.SetTag("break", "service fusing")
			return httpNotFound, infoServerFusing, nil, errors.New("server fusing")
		}
	}

	// service call retry
	retry, err := retryAnalyze(g.cfg.Service.CallRetry)
	if err != nil {
		log.Error("retryAnalyze", err)
		retry = []int{0}
	}

	code, result, header, err := g.retryGo(service, action, retry, target, span, route, request, args, reply)

	return code, result, header, err
}

func (g *Garden) requestServiceHttp(span opentracing.Span, rawURL string, request *req, timeout int) (int, string, http.Header, error) {
	if request == nil {
		return httpFail, "", nil, errors.New("empty request")
	}
	var body io.Reader
	if len(request.RawBody) > 0 {
		body = bytes.NewReader(request.RawBody)
	}
	r, err := http.NewRequest(request.Method, appendQuery(rawURL, request.UrlParam), body)
	if err != nil {
		return httpFail, "", nil, err
	}

	skip := hopByHop(request.Headers)
	for k, v := range request.Headers {
		name := http.CanonicalHeaderKey(k)
		if _, drop := skip[name]; drop {
			continue
		}
		value, ok := v.(string)
		if !ok {
			continue
		}
		r.Header.Set(name, value)
	}
	if xff := forwardedFor(headerValue(request.Headers, "X-Forwarded-For"), request.RemoteIp); xff != "" {
		r.Header.Set("X-Forwarded-For", xff)
	}
	if r.Header.Get("X-Real-Ip") == "" && request.ClientIp != "" {
		r.Header.Set("X-Real-Ip", request.ClientIp)
	}
	r.Header.Set("Call-Key", g.cfg.Service.CallKey)

	if span != nil {
		opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header))
	}

	client := &http.Client{
		Timeout: time.Millisecond * time.Duration(timeout),
	}
	res, err := client.Do(r)
	if err != nil {
		return httpFail, "", nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != httpOk {
		return res.StatusCode, "", nil, errors.New("http status " + strconv.Itoa(res.StatusCode))
	}
	body2, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return httpFail, "", nil, err
	}
	return httpOk, string(body2), res.Header, nil
}

func appendQuery(rawURL, urlParam string) string {
	if urlParam == "" {
		return rawURL
	}
	q := strings.TrimPrefix(urlParam, "?")
	if q == "" {
		return rawURL
	}
	if strings.Contains(rawURL, "?") {
		return rawURL + "&" + q
	}
	return rawURL + "?" + q
}

func forwardedFor(prior, remote string) string {
	switch {
	case prior != "" && remote != "":
		return prior + ", " + remote
	case prior != "":
		return prior
	default:
		return remote
	}
}

func headerValue(headers MapData, key string) string {
	if headers == nil {
		return ""
	}
	if v, ok := headers[http.CanonicalHeaderKey(key)]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	for k, v := range headers {
		if strings.EqualFold(k, key) {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func hopByHop(headers MapData) map[string]struct{} {
	skip := map[string]struct{}{
		"Connection":          {},
		"Keep-Alive":          {},
		"Proxy-Authenticate":  {},
		"Proxy-Authorization": {},
		"Te":                  {},
		"Trailer":             {},
		"Trailers":            {},
		"Transfer-Encoding":   {},
		"Upgrade":             {},
		"Content-Length":      {},
		"Host":                {},
	}
	for _, part := range strings.Split(headerValue(headers, "Connection"), ",") {
		name := http.CanonicalHeaderKey(strings.TrimSpace(part))
		if name != "" {
			skip[name] = struct{}{}
		}
	}
	return skip
}

// CallRpc call other service rpc method
func (g *Garden) CallRpc(span opentracing.Span, service, action string, args, reply interface{}) error {
	_, _, _, err := g.callService(span, service, action, nil, args, reply)
	return err
}
