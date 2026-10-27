package core

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var metricName = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)

type gardenCollector struct {
	g *Garden
}

func (c gardenCollector) Describe(chan<- *prometheus.Desc) {}

func (c gardenCollector) Collect(ch chan<- prometheus.Metric) {
	c.constMetric(ch, "garden_requests_in_flight", "Requests currently being handled.", prometheus.GaugeValue, float64(atomic.LoadInt64(&c.g.requestProcess)))
	c.constMetric(ch, "garden_requests_finished_total", "Requests finished by this process.", prometheus.CounterValue, float64(atomic.LoadInt64(&c.g.requestFinish)))
	c.g.metrics.Range(func(key, value interface{}) bool {
		name, ok := key.(string)
		if !ok || !metricName.MatchString(name) {
			return true
		}
		number, ok := metricValue(value)
		if !ok {
			return true
		}
		c.constMetric(ch, name, "Custom metric.", prometheus.GaugeValue, number)
		return true
	})
}

func (c gardenCollector) constMetric(ch chan<- prometheus.Metric, name, help string, valueType prometheus.ValueType, value float64) {
	desc := prometheus.NewDesc(name, help, nil, nil)
	metric, err := prometheus.NewConstMetric(desc, valueType, value)
	if err != nil {
		return
	}
	ch <- metric
}

func metricValue(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint64:
		return float64(n), true
	case uint32:
		return float64(n), true
	case float64:
		return n, true
	case float32:
		return float64(n), true
	default:
		return 0, false
	}
}

func (g *Garden) prometheus() gin.HandlerFunc {
	reg := prometheus.NewRegistry()
	reg.MustRegister(gardenCollector{g: g})
	h := promhttp.HandlerFor(prometheus.Gatherers{
		reg,
		prometheus.DefaultGatherer,
	}, promhttp.HandlerOpts{})
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// PushGateway upload metric to pushGateway
func (g *Garden) PushGateway(job string, data MapData) (string, error) {
	client := &http.Client{
		Timeout: time.Millisecond * time.Duration(5000),
	}

	instance := url.PathEscape(g.GetServiceId())
	url := fmt.Sprintf("http://%s/metrics/job/%s/instance/%s", g.cfg.Service.PushGatewayAddress, url.PathEscape(job), instance)
	r, err := http.NewRequest("POST", url, strings.NewReader(metricFormat(data)))
	if err != nil {
		return "", err
	}

	res, err := client.Do(r)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body2, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body2), nil
}

func metricFormat(data MapData) string {
	body := ""
	for k, v := range data {
		body += fmt.Sprintf("%s %v\n", k, v)
	}
	return body
}

// SetMetric stores a numeric value exported by /metrics.
func (g *Garden) SetMetric(key string, val interface{}) {
	g.metrics.Store(key, val)
}
