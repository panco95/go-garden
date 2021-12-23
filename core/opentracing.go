package core

import (
	"context"
	"encoding/json"
	"github.com/opentracing/opentracing-go"
	zkOt "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zkHttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"net/http"
	"time"
)

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

func connZipkin(service, addr, serviceIp string) error {
	reporter := zkHttp.NewReporter(addr)
	endpoint, err := zipkin.NewEndpoint(service, serviceIp)
	if err != nil {
		return err
	}
	trace, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(zkOt.Wrap(trace))
	return nil
}

func connJaeger(service, addr string) error {
	cfg := jaegercfg.Configuration{
		ServiceName: service,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}

	sender, err := jaeger.NewUDPTransport(addr, 0)
	if err != nil {
		return err
	}

	reporter := jaeger.NewRemoteReporter(sender)
	tracer, _, err := cfg.NewTracer(
		jaegercfg.Reporter(reporter),
	)

	opentracing.SetGlobalTracer(tracer)
	return nil
}
