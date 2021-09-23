package core

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	zkOt "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zkHttp "github.com/openzipkin/zipkin-go/reporter/http"
	"net/http"
)

func (g *Garden) initOpenTracing(service, addr, address string) error {
	trace, err := connZipkin(service, addr, address)
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(trace)
	return nil
}

// startSpanFromHeader Get the opentracing span from the request header
// If no span, in header creates new root span, if any, new child span
func startSpanFromHeader(header http.Header, operateName string) opentracing.Span {
	var span opentracing.Span
	wireContext, _ := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(header))
	span = opentracing.StartSpan(
		operateName,
		//ext.RPCServerOption(wireContext),
		opentracing.ChildOf(wireContext),
	)
	return span
}

func requestTracing(c *gin.Context, span opentracing.Span) {
	request := Request{
		getMethod(c),
		getUrl(c),
		getUrlParam(c),
		getClientIp(c),
		getHeaders(c),
		getBody(c)}
	s, _ := json.Marshal(&request)
	span.SetTag("Request", string(s))

	c.Set("span", span)
	c.Set("request", &request)
}

func connZipkin(service, addr, address string) (opentracing.Tracer, error) {
	reporter := zkHttp.NewReporter(addr)
	endpoint, err := zipkin.NewEndpoint(service, address)
	if err != nil {
		return nil, err
	}
	trace, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		return nil, err
	}
	return zkOt.Wrap(trace), nil
}
