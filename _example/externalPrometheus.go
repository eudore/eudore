package main

import (
	"context"
	"strconv"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/common/expfmt"
)

const ServiceVersion = "undefined"

type App struct {
	*eudore.App
	*Config
	Prometheus prometheus.Registerer
	Collector  *eudoreCollector
}
type Config struct {
	Logger *eudore.LoggerConfig
	// 其他配置
}

type ConfigPrometheus struct {
	MetricPrefix  string
	EnabeleLogger bool
	EnableServer  bool
	EnableClient  bool
}

func main() {
	app := NewApp()
	app.Parse()
	app.Run()
}

func NewApp() *App {
	app := &App{
		App: eudore.NewApp(),
		Config: &Config{
			Logger: &eudore.LoggerConfig{
				Stdout:   true,
				StdColor: true,
				HookMeta: true,
			},
		},
	}

	app.SetValue(eudore.ContextKeyConfig, eudore.NewConfig(app.Config))
	app.ParseOption(
		app.NewParseLoggerFunc(),
		app.NewParseMeterFunc(),
		app.NewParseRouterFunc(),
	)
	return app
}

// NewParseLoggerFunc 方法创建一个日志配置解析函数。
func (app *App) NewParseLoggerFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(app.Config.Logger))
		app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))
		return nil
	}
}

// NewParseMeterFunc 方法创建Traing配置解析函数。
func (app *App) NewParseMeterFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		app.Collector = NewEudoreCollector(app)
		app.Prometheus = prometheus.NewRegistry()
		app.Prometheus.MustRegister(
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
			app.Collector,
		)

		app.SetValue(eudore.NewContextKey("prometheus"), app.Prometheus)
		app.GetFunc("/metrics", NewMetricsFunc(app.Prometheus.(prometheus.Gatherer)))
		return nil
	}
}

func (app *App) NewParseRouterFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		app.GetFunc("/health", middleware.NewHealthCheckFunc(app))
		app.AddMiddleware("global",
			middleware.NewLoggerFunc(app),
			middleware.NewRequestIDFunc(nil),
			middleware.NewRecoveryFunc(),
		)
		app.AddMiddleware(app.Collector.NewHandlerFunc())
		app.AddHandler("404", "", eudore.HandlerRouter404)
		app.AddHandler("405", "", eudore.HandlerRouter405)
		app.GetFunc("/*", func(ctx eudore.Context) {
			ctx.WriteString("hello")
		})
		return app.Listen(":8088")
	}
}

// NewMeterMetrics 方法创建metrics处理函数。
func NewMetricsFunc(registry prometheus.Gatherer) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		mfs, err := registry.Gather()
		if err != nil {
			ctx.Error("error gathering metrics:", err)
			return
		}

		contentType := expfmt.Negotiate(ctx.Request().Header)
		ctx.SetHeader(eudore.HeaderContentType, string(contentType))
		enc := expfmt.NewEncoder(ctx, contentType)

		for _, mf := range mfs {
			err := enc.Encode(mf)
			if err != nil {
				ctx.Error("error encoding and sending metric family:", err)
				return
			}
		}

		if closer, ok := enc.(expfmt.Closer); ok {
			err := closer.Close()
			if err != nil {
				ctx.Error("error encoding and sending metric family:", err)
				return
			}
		}
	}
}

type eudoreCollector struct {
	prometheus.Collector
	loggerMetadata                eudoreMetadata
	appInfo                       *prometheus.Desc
	loggerCount                   *prometheus.Desc
	loggerSize                    *prometheus.Desc
	serverRequestsInflight        *prometheus.GaugeVec
	serverRequestsCount           *prometheus.CounterVec
	serverRequestsDurationSeconds *prometheus.HistogramVec
	serverResponseSize            *prometheus.HistogramVec
}

type eudoreMetadata interface {
	Metadata() any
}

func NewEudoreCollector(ctx context.Context) *eudoreCollector {
	loggerMetadata, ok := ctx.Value(eudore.ContextKeyLogger).(eudoreMetadata)
	if ok {
		if loggerMetadata.Metadata() == nil {
			loggerMetadata = nil
		}
	}

	labels := []string{"method", "handler", "code"}
	return &eudoreCollector{
		loggerMetadata: loggerMetadata,
		appInfo: prometheus.NewDesc(
			"eudore_app_info",
			"eudore app name and version",
			nil,
			prometheus.Labels{
				"service_language": "go",
				"service_version":  ServiceVersion,
			},
		),
		loggerCount: prometheus.NewDesc(
			"eudore_logger_entries_total",
			"eudore logger level count",
			[]string{"level"}, nil,
		),
		loggerSize: prometheus.NewDesc(
			"eudore_logger_size_bytes",
			"eudore logger written size",
			nil, nil,
		),
		serverRequestsInflight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "eudore_server_requests_in_flight",
				Help: "Current number of scrapes being served.",
			},
			labels[:2],
		),
		serverRequestsCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "eudore_server_requests_total",
				Help: "Total number of scrapes by HTTP status code.",
			},
			labels,
		),
		serverRequestsDurationSeconds: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "eudore_server_request_duration_seconds",
				Help: "Histogram of latencies for HTTP requests.",
			},
			labels,
		),
		serverResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "eudore_server_response_size_bytes",
				Help: "Histogram of response size for HTTP requests.",
			},
			labels,
		),
	}
}

func (c *eudoreCollector) NewHandlerFunc() eudore.HandlerFunc {
	release := func(labels prometheus.Labels, start time.Time, resp eudore.ResponseWriter) {
		labels["code"] = strconv.Itoa(resp.Status())
		c.serverRequestsCount.With(labels).Inc()
		c.serverRequestsDurationSeconds.With(labels).Observe(time.Since(start).Seconds())
		c.serverResponseSize.With(labels).Observe(float64(resp.Size()))
	}
	return func(ctx eudore.Context) {
		now := time.Now()
		handler := ctx.GetParam(eudore.ParamAction)
		if handler == "" {
			handler = ctx.GetParam(eudore.ParamRoute)
		}
		labels := prometheus.Labels{
			"method":  ctx.Method(),
			"handler": handler,
		}
		inflight := c.serverRequestsInflight.With(labels)
		inflight.Inc()

		defer inflight.Dec()
		defer release(labels, now, ctx.Response())
		ctx.Next()
	}
}

func (c *eudoreCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.appInfo
	c.serverRequestsInflight.Describe(ch)
	c.serverRequestsCount.Describe(ch)
	c.serverRequestsDurationSeconds.Describe(ch)
	c.serverResponseSize.Describe(ch)
	if c.loggerMetadata != nil {
		ch <- c.loggerCount
		ch <- c.loggerSize
	}
}

// Collect implements Collector.
func (c *eudoreCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.appInfo, prometheus.CounterValue, 1)
	c.serverRequestsInflight.Collect(ch)
	c.serverRequestsCount.Collect(ch)
	c.serverRequestsDurationSeconds.Collect(ch)
	c.serverResponseSize.Collect(ch)
	if c.loggerMetadata != nil {
		meta := c.loggerMetadata.Metadata().(eudore.MetadataLogger)
		ch <- prometheus.MustNewConstMetric(c.loggerSize,
			prometheus.CounterValue, float64(meta.Size),
		)
		for i := range meta.Count {
			ch <- prometheus.MustNewConstMetric(
				c.loggerCount,
				prometheus.CounterValue,
				float64(meta.Count[i]),
				eudore.DefaultLoggerLevelStrings[i],
			)
		}
	}
}
