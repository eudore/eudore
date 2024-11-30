package main

/*
2024年5月opentracing弃用，改为使用OpenTelemetry。
*/
import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/uber/jaeger-client-go"
	jaegerconfig "github.com/uber/jaeger-client-go/config"
	jaegerprometheus "github.com/uber/jaeger-lib/metrics/prometheus"
)

type App struct {
	*eudore.App
	*Config
	Tracer     opentracing.Tracer
	Prometheus prometheus.Registerer
}

type Config struct {
	Logger *eudore.LoggerConfig
	Tracer *TracerConfig
	// 其他配置
}

type TracerConfig struct {
	ServiceName       string `json:"servicename" alias:"servicename"`
	LocalAgent        string `json:"localagent" alias:"localagent"`
	CollectorEndpoint string `json:"collectorendpoint" alias:"collectorendpoint"`
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
			},
			Tracer: &TracerConfig{
				ServiceName: "eudore-example",
				// "127.0.0.1:6831"
				CollectorEndpoint: "http://172.19.214.64:14268/api/traces",
			},
		},
	}

	app.SetValue(eudore.ContextKeyConfig, eudore.NewConfig(app.Config))
	app.ParseOption(
		app.NewParseLoggerFunc(),
		app.NewParseTracerFunc(),
		app.NewParseRouterFunc(),
	)
	return app
}

// NewParseLoggerFunc 方法创建一个日志配置解析函数。
func (app *App) NewParseLoggerFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		conf := app.Config.Logger
		conf.Handlers = append(conf.Handlers, &loggerHookTrace{})
		app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(conf))
		app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))
		return nil
	}
}

// NewParseTracerFunc 方法创建Traing配置解析函数。
func (app *App) NewParseTracerFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		config := app.Config.Tracer
		if config.LocalAgent == "" && config.CollectorEndpoint == "" {
			return nil
		}

		cfg := jaegerconfig.Configuration{
			ServiceName: config.ServiceName,
			Sampler: &jaegerconfig.SamplerConfig{
				Type:  jaeger.SamplerTypeConst,
				Param: 1,
			},
			Reporter: &jaegerconfig.ReporterConfig{
				// agent collector二选一
				LocalAgentHostPort: config.LocalAgent,
				CollectorEndpoint:  config.CollectorEndpoint,
				// 不要将span打印到标准输出
				LogSpans: false,
			},
			Headers: &jaeger.HeadersConfig{
				TraceContextHeaderName:   "uber-trace-id",
				TraceBaggageHeaderPrefix: "uber-context-",
				JaegerBaggageHeader:      "uber-baggage",
				JaegerDebugHeader:        "uber-debug-id",
			},
		}

		tracer, closer, err := cfg.NewTracer(
			jaegerconfig.Logger(&jaegerLogger{app.App.Logger}),
			jaegerconfig.Metrics(jaegerprometheus.New(
				jaegerprometheus.WithRegisterer(app.Prometheus),
			)),
		)
		if err != nil {
			return err
		}

		app.Tracer = tracer
		opentracing.SetGlobalTracer(tracer)
		app.SetValue(eudore.NewContextKey("tracer-closer"),
			eudore.Unmounter(func(ctx context.Context) {
				closer.Close()
			}),
		)
		return nil
	}
}

func (app *App) NewParseRouterFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		app.AddMiddleware("global",
			middleware.NewLoggerFunc(app),
			app.NewTracerHandler(),
		)

		api := app.Group("")
		app.AddMiddleware(
			app.NewTracerHandler(),
			middleware.NewRecoveryFunc(),
			func(eudore.Context) {
				// slow
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(20)+10))
			},
		)

		api.GetFunc("/call action=myservice:exmple:call", func(ctx eudore.Context) {
			id := ctx.Response().Header().Get(eudore.HeaderXTraceID)
			ctx.WriteString("request id: " + id)
		})
		api.GetFunc("/err action=myservice:exmple:err", func(ctx eudore.Context) {
			id := ctx.Response().Header().Get(eudore.HeaderXTraceID)
			ctx.WriteString("request id: " + id)
			ctx.Error("request id:", id)
		})
		api.GetFunc("/sleep action=myservice:exmple:sleep", func(ctx eudore.Context) {
			id := ctx.Response().Header().Get(eudore.HeaderXTraceID)
			ctx.WriteString("sleep request id: " + id)
			time.Sleep(time.Millisecond * 100)
		})
		api.GetFunc("/*", func(ctx eudore.Context) {
			id := ctx.Response().Header().Get(eudore.HeaderXTraceID)
			ctx.WriteString("request id: " + id)
			ctx.Info("request id:", id)

			wait := sync.WaitGroup{}
			wait.Add(3)
			defer wait.Wait()
			c := context.WithValue(ctx.Context(), eudore.ContextKeyServer, app.Server)
			// app.Client 没有实现传递Span。
			for _, p := range []string{"/call", "/err", "/sleep"} {
				go func(p string) {
					app.NewRequest("GET", p, c)
					wait.Done()
				}(p)
			}
		})

		return app.Listen(":8088")
	}
}

// NewTracerHandler 方法创建Tracing处理中间件函数。
func (app *App) NewTracerHandler() eudore.HandlerFunc {
	if app.Tracer == nil {
		return nil
	}
	const spanname = "eudore:app:server"
	return func(ctx eudore.Context) {
		spanCtx, _ := app.Tracer.Extract(opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(ctx.Request().Header),
		)
		span := app.Tracer.StartSpan(
			// TODO
			eudore.GetAnyDefault(ctx.GetParam(eudore.ParamAction), spanname),
			opentracing.ChildOf(spanCtx),
			opentracing.StartTime(time.Now()),
			opentracing.Tags{
				"span.kind":       "server",
				"http.host":       ctx.Host(),
				"http.method":     ctx.Method(),
				"http.target":     ctx.Request().RequestURI,
				"http.user_agent": ctx.GetHeader(eudore.HeaderUserAgent),
				"http.client_ip":  ctx.RealIP(),
			},
		)
		defer span.Finish()

		traceid := span.Context().(jaeger.SpanContext).TraceID().String()
		ctx.SetContext(opentracing.ContextWithSpan(ctx.Context(), span))
		ctx.SetValue(eudore.ContextKeyLogger,
			ctx.Value(eudore.ContextKeyLogger).(eudore.Logger).
				WithField("context", ctx.Context()).
				WithField("x-trace-id", traceid).
				WithField("logger", true),
		)
		ctx.SetHeader(eudore.HeaderXTraceID, traceid)
		ctx.Next()

		status := ctx.Response().Status()
		userid := eudore.GetAny[int](ctx.GetParam(eudore.ParamUserid))
		action := ctx.GetParam(eudore.ParamAction)
		span.SetTag("http.status_code", status)
		span.SetTag("http.route", ctx.GetParam(eudore.ParamRoute))

		if userid != 0 {
			span.SetTag("http.user", userid)
		}
		if action != "" {
			span.SetTag("http.action", action)
		}
		if status > 499 {
			span.SetTag("error", "true")
		}
	}
}

type jaegerLogger struct {
	eudore.Logger
}

func (log jaegerLogger) Error(msg string) {
	log.Logger.Error(msg)
}

func (log jaegerLogger) Infof(msg string, args ...interface{}) {
	log.Logger.Infof(msg, args...)
}

type loggerHookTrace struct{}

func (h *loggerHookTrace) HandlerPriority() int {
	return 25
}

func (h *loggerHookTrace) HandlerEntry(e *eudore.LoggerEntry) {
	for i, key := range e.Keys {
		if key == "context" {
			ctx, ok := e.Vals[i].(context.Context)
			if ok {
				e.Keys = e.Keys[:i+copy(e.Keys[i:], e.Keys[i+1:])]
				e.Vals = e.Vals[:i+copy(e.Vals[i:], e.Vals[i+1:])]
				h.writeSpan(e, opentracing.SpanFromContext(ctx))
			}
			return
		}
	}
}

func (h *loggerHookTrace) writeSpan(e *eudore.LoggerEntry,
	span opentracing.Span,
) {
	if span == nil {
		return
	}
	fields := make([]log.Field, 0, len(e.Keys)+1)
	fields = append(fields, log.String("event", e.Level.String()))
	for i := 0; i < len(e.Keys); i++ {
		if e.Keys[i] != "x-trace-id" {
			fields = append(fields, log.Object(e.Keys[i], e.Vals[i]))
		}
	}
	if e.Message != "" {
		fields = append(fields, log.String("message", e.Message))
	}
	span.LogFields(fields...)
	// 可忽略Warning日志警告
	if e.Level == eudore.LoggerError || e.Level == eudore.LoggerWarning {
		span.SetTag("error", true)
	}
}
