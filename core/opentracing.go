package core

import (
	"context"
	"encoding/json"
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

// StartRpcTrace rpc method use this method first
func (g *Garden) StartRpcTrace(ctx context.Context, args interface{}, method string) opentracing.Span {
	span := StartSpanFormRpc(ctx, method)
	span.SetTag("CallType", "Rpc")
	span.SetTag("ServiceIp", g.ServiceIp)
	span.SetTag("ServiceId", g.ServiceId)
	span.SetTag("Status", "unfinished")
	s, _ := json.Marshal(&args)
	span.SetTag("Args", string(s))
	return span
}

// FinishRpcTrace rpc method use this method last
func (g *Garden) FinishRpcTrace(span opentracing.Span) {
	span.SetTag("Status", "finished")
	span.Finish()
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
