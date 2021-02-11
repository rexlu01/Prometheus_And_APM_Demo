package ginprometheus

import (
	"strconv"
	"time"

	logger "promandapm/log"
	gutil "promandapm/util"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const defaultMetricPath = "/metrics"

/*----------------------------------------定义结构体--------------------------------------------------*/

type Metric struct {
	MetricCollector prometheus.Collector
	ID              string
	Name            string
	Description     string
	Type            string
	Args            []string
}

type PrometheusPushGateway struct {
	PushIntervalSeconds time.Duration
	PushGatewayURL      string
	MetricsURL          string
	Job                 string
}

type Prometheus struct {
	reqCnt                  *prometheus.CounterVec
	reqDur                  *prometheus.HistogramVec
	reqSz, resSz            prometheus.Summary
	router                  *gin.Engine
	listenAddress           string
	Ppg                     PrometheusPushGateway
	MetricsList             []*Metric
	MetricsPath             string
	ReqCutURLLabelMappingFn func(c *gin.Context) string
	URLLabelFromContext     string
}

var reqCnt = &Metric{
	ID:          "reqCut",
	Name:        "requests_total",
	Description: "How many HTTP requests processed, partitioned by status code and HTTP method.",
	Type:        "counter_vec",
	Args:        []string{"code", "method", "handler", "host", "url"},
}

var reqDur = &Metric{
	ID:          "reqDur",
	Name:        "request_duration_seconds",
	Description: "The HTTP request latencies in seconds.",
	Type:        "histogram_vec",
	Args:        []string{"code", "method", "url"},
}

var resSz = &Metric{
	ID:          "reqSz",
	Name:        "response_size_bytes",
	Description: "The HTTP response sizes in bytes.",
	Type:        "summary",
}

var reqSz = &Metric{
	ID:          "reqSz",
	Name:        "request_size_bytes",
	Description: "The HTTP request sizes in bytes.",
	Type:        "summary",
}

var standardMetrics = []*Metric{
	reqCnt,
	reqDur,
	resSz,
	reqSz,
}

/*----------------------------------------NewPrometheus方法群--------------------------------------------------*/
//使用...[]*Metric的原因是在main函数里不需要传
func NewPrometheus(subsystem string, customMetricsList ...[]*Metric) *Prometheus {

	var metricsList []*Metric

	if len(customMetricsList) > 1 {
		panic("Too many args. NewPrometheus(string,<string,<optional []*Metruc>).")
	} else if len(customMetricsList) == 1 {
		metricsList = customMetricsList[0]
	}

	for _, metric := range standardMetrics {
		metricsList = append(metricsList, metric)
	}

	p := &Prometheus{
		MetricsList: metricsList,
		MetricsPath: defaultMetricPath,
		ReqCutURLLabelMappingFn: func(c *gin.Context) string {
			return c.Request.URL.Path
		},
	}

	p.registerMetrics(subsystem)

	return p
}

func (p *Prometheus) registerMetrics(subsystem string) {
	for _, metricDef := range p.MetricsList {
		metric := NewMetric(metricDef, subsystem)
		if err := prometheus.Register(metric); err != nil {
			logger.Errorf("%s could not be registered in Prometheus", metricDef.Name)
			logger.Fatal(err)
		}
		switch metricDef {
		case reqCnt:
			p.reqCnt = metric.(*(prometheus.CounterVec))
		case reqDur:
			p.reqDur = metric.(*prometheus.HistogramVec)
		case resSz:
			p.resSz = metric.(prometheus.Summary)
		case reqSz:
			p.reqSz = metric.(prometheus.Summary)
		}
		metricDef.MetricCollector = metric
	}
}

func NewMetric(m *Metric, subsystem string) prometheus.Collector {
	var metric prometheus.Collector
	switch m.Type {
	case "counter_vec":
		metric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "counter":
		metric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "gange_vec":
		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "gauge":
		metric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "histogram_vec":
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "histogram":
		metric = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "summary_vec":
		metric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "summary":
		metric = prometheus.NewSummary(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	}
	return metric
}

/*----------------------------------------Use方法相关--------------------------------------------------*/
func (p *Prometheus) Use(e *gin.Engine) {
	e.Use(p.HandlerFunc())
	p.SetMetricsPath(e)
}

func (p *Prometheus) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == p.MetricsPath {
			c.Next()
			return
		}
		start := time.Now()
		reqSz := gutil.ComputeApproximateRequestSize(c.Request)
		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		resSz := float64(c.Writer.Size())

		url := p.ReqCutURLLabelMappingFn(c)

		if len(p.URLLabelFromContext) > 0 {
			u, found := c.Get(p.URLLabelFromContext)
			if !found {
				u = "unknown"
			}
			url = u.(string)
		}

		p.reqDur.WithLabelValues(status, c.Request.Method, url).Observe(elapsed)
		p.reqCnt.WithLabelValues(status, c.Request.Method, c.HandlerName(), c.Request.Host, url).Inc()
		p.reqSz.Observe(float64(reqSz))
		p.resSz.Observe(resSz)
	}

}

func (p *Prometheus) SetMetricsPath(e *gin.Engine) {

	if p.listenAddress != "" {
		p.router.GET(p.MetricsPath, prometheusHandler())
		p.runServer()
	} else {
		e.GET(p.MetricsPath, prometheusHandler())
	}
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func (p *Prometheus) runServer() {
	if p.listenAddress != "" {
		go p.router.Run(p.listenAddress)
	}
}
