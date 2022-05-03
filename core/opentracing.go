package core

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	zkOt "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zkHttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/log/zap"
)

func (g *Garden) bootOpenTracing() {
	var err error
	switch g.cfg.Service.TracerDrive {
	case "jaeger":
		err = connJaeger(g.cfg.Service.ServiceName, g.cfg.Service.JaegerAddress)
		break
	case "zipkin":
		err = connZipkin(g.cfg.Service.ServiceName, g.cfg.Service.ZipkinAddress, g.GetServiceIp())
		break
	default:
		break
	}
	if err != nil {
		g.Log(FatalLevel, "openTracing", err)
	}
	l, err := g.GetLog()
	if err != nil {
		g.Log(FatalLevel, "openTracing GetLog", err)
	}
	zap.NewLoggingTracer(l.Desugar(), opentracing.GlobalTracer())
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
	span.SetTag("ServiceIp", g.GetServiceIp())
	span.SetTag("ServiceId", g.GetServiceId())
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
	cfg := jaegerCfg.Configuration{
		ServiceName: service,
		Sampler: &jaegerCfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegerCfg.ReporterConfig{
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
		jaegerCfg.Reporter(reporter),
	)

	opentracing.SetGlobalTracer(tracer)
	return nil
}
