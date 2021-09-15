package zipkin

import (
	"github.com/opentracing/opentracing-go"
	zkOt "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zkHttp "github.com/openzipkin/zipkin-go/reporter/http"
)

func Connect(service, addr, address string) (opentracing.Tracer, error) {
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
