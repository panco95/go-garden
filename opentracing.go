package garden

import (
	"encoding/json"
	"github.com/panco95/go-garden/drives/zipkin"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"net/http"
	"reflect"
)

func InitOpenTracing(service, addr, address string) error {
	trace, err := zipkin.Connect(service, addr, address)
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(trace)
	return nil
}

// StartSpanFromHeader Get the opentracing span from the request header
// If no span, in header creates new root span, if any, new child span
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

func RequestTracing(ctx interface{}, span opentracing.Span) {
	t := reflect.TypeOf(ctx)
	switch t.String() {
	case "*gin.Context":
		c := ctx.(*gin.Context)
		requestTracingGin(c, span)
		break
	default:
		break
	}
}

func requestTracingGin(c *gin.Context, span opentracing.Span) {
	request := Request{
		GetMethod(c),
		GetUrl(c),
		GetUrlParam(c),
		GetClientIp(c),
		GetHeaders(c),
		GetBody(c)}
	s, _ := json.Marshal(&request)
	span.SetTag("Request", string(s))

	c.Set("span", span)
	c.Set("request", &request)
}
