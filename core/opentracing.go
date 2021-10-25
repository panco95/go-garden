package core

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	zkOt "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zkHttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/smallnest/rpcx/share"
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

// StartSpanFromHeader Get the opentracing span from the request header
// If no span, will create new root span, if any, new child span
func StartSpanFromHeader(header http.Header, operateName string) opentracing.Span {
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

// StartSpanFromTextMap Get the opentracing span from textMap
// If no span, will create new root span, if any, new child span
func StartSpanFromTextMap(textMap opentracing.TextMapCarrier, operateName string) opentracing.Span {
	var span opentracing.Span
	wireContext, _ := opentracing.GlobalTracer().Extract(
		opentracing.TextMap,
		textMap)
	span = opentracing.StartSpan(
		operateName,
		opentracing.ChildOf(wireContext),
	)
	return span
}

// StartSpanFormRpc start and get opentracing span fro rpc
func StartSpanFormRpc(ctx context.Context, operateName string) opentracing.Span {
	reqMeta := ctx.Value(share.ReqMetaDataKey).(map[string]string)
	span := StartSpanFromTextMap(reqMeta, operateName)
	span.SetTag("CallType", "Rpc")
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
