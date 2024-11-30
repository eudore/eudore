package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type App struct {
	*eudore.App
	*Config
	Tracer trace.TracerProvider
}

type Config struct {
	Logger *eudore.LoggerConfig
	Tracer *TracerConfig
}

type TracerConfig struct {
	ServiceName string `json:"servicename" alias:"servicename"`
	Endpoint    string `json:"endpoint" alias:"endpoint"`
	Timeout     time.Duration
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
				// otel-collector or jaeger-collector addr
				Endpoint: "172.19.214.64:4317",
				Timeout:  15 * time.Second,
			},
		},
	}

	app.SetValue(eudore.ContextKeyConfig, eudore.NewConfig(app.Config))
	app.ParseOption(
		app.NewParseLoggerFunc(),
		app.NewParseTracerFunc(),
		app.NewParseClientFunc(),
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
		if config.Endpoint == "" {
			return nil
		}

		exporter, err := otlptracegrpc.New(app,
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(config.Endpoint),
			otlptracegrpc.WithTimeout(config.Timeout),
		)
		if err != nil {
			return err
		}

		app.Infof("init opentelemetry to endpoint: otel://%s", config.Endpoint)
		name, _ := os.Hostname()
		tracer := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(resource.NewSchemaless(
				attribute.String("host.name", name),
				attribute.String("service.name", app.Config.Tracer.ServiceName),
			)),
		)
		otel.SetTracerProvider(tracer)

		app.Tracer = tracer
		app.SetValue(eudore.ContextKeyTrace, tracer)
		app.SetValue(eudore.NewContextKey("otel-shutdown"),
			eudore.Unmounter(func(ctx context.Context) {
				exporter.Shutdown(ctx)
			}),
		)
		return nil
	}
}

func (app *App) NewParseClientFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		if app.Tracer != nil {
			app.SetValue(eudore.ContextKeyClient, app.NewClient(
				NewClientHookTracer(app.Tracer),
			))
		}
		return nil
	}
}
func (app *App) NewParseRouterFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		app.AddMiddleware("global",
			middleware.NewLoggerFunc(app),
		)

		api := app.Group("")
		api.AddMiddleware(
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
		api.GetFunc("/", func(ctx eudore.Context) {
			id := ctx.Response().Header().Get(eudore.HeaderXTraceID)
			ctx.WriteString("request id: " + id)
			ctx.Info("request id:", id)

			wait := sync.WaitGroup{}
			wait.Add(3)
			defer wait.Wait()
			c := context.WithValue(ctx.Context(), eudore.ContextKeyServer, app.Server)
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
	tracer := app.Tracer.Tracer("github.com/eudore/eudore/_example")
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	const spanname = "eudore:app:server"
	return func(ctx eudore.Context) {
		spanCtx := propagator.Extract(ctx.Context(),
			propagation.HeaderCarrier(ctx.Request().Header),
		)
		spanCtx, span := tracer.Start(spanCtx, spanname,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()
		span.SetAttributes(
			attribute.String("http.host", ctx.Host()),
			attribute.String("http.method", ctx.Method()),
			attribute.String("http.target", ctx.Request().RequestURI),
			attribute.String("http.user_agent", ctx.GetHeader(eudore.HeaderUserAgent)),
			attribute.String("http.client_ip", ctx.RealIP()),
		)

		traceid := span.SpanContext().TraceID().String()
		ctx.SetContext(spanCtx)
		ctx.SetValue(eudore.ContextKeyLogger,
			ctx.Value(eudore.ContextKeyLogger).(eudore.Logger).
				WithField("context", spanCtx).
				WithField("x-trace-id", traceid).
				WithField("logger", true),
		)
		ctx.SetHeader(eudore.HeaderXTraceID, traceid)
		ctx.Next()

		status := ctx.Response().Status()
		userid := eudore.GetAny[int](ctx.GetParam(eudore.ParamUserid))
		action := ctx.GetParam(eudore.ParamAction)
		span.SetAttributes(
			attribute.Int("http.status_code", status),
			attribute.String("http.route", ctx.GetParam(eudore.ParamRoute)),
		)
		if userid != 0 {
			span.SetAttributes(attribute.Int("http.user", userid))
		}
		if action != "" {
			span.SetName(action)
			span.SetAttributes(attribute.String("http.action", action))
		}
		if status > 499 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	}
}

type loggerHookTrace struct{}

func (h *loggerHookTrace) HandlerPriority() int {
	return eudore.DefaultLoggerPriorityFormatter - 1
}

func (h *loggerHookTrace) HandlerEntry(e *eudore.LoggerEntry) {
	for i, key := range e.Keys {
		if key == "context" {
			ctx, ok := e.Vals[i].(context.Context)
			if ok {
				e.Keys = e.Keys[:i+copy(e.Keys[i:], e.Keys[i+1:])]
				e.Vals = e.Vals[:i+copy(e.Vals[i:], e.Vals[i+1:])]
				h.writeSpan(e, trace.SpanFromContext(ctx))
			}
			return
		}
	}
}

func (h *loggerHookTrace) writeSpan(e *eudore.LoggerEntry, span trace.Span) {
	if span == nil {
		return
	}
	fields := make([]attribute.KeyValue, 0, len(e.Keys))
	for i := 0; i < len(e.Keys); i++ {
		if e.Keys[i] != "x-trace-id" {
			fields = append(fields, newLoggerAttr(e.Keys[i], e.Vals[i]))
		}
	}
	if e.Message != "" {
		fields = append(fields, attribute.String("message", e.Message))
	}
	span.AddEvent(e.Level.String(),
		trace.WithTimestamp(e.Time),
		trace.WithAttributes(fields...),
	)
	// 可忽略Warning日志警告
	if e.Level == eudore.LoggerError || e.Level == eudore.LoggerWarning {
		span.SetAttributes(attribute.Bool("error", true))
	}
}

func newLoggerAttr(key string, val interface{}) attribute.KeyValue {
	switch i := val.(type) {
	case string:
		return attribute.String(key, i)
	case int:
		return attribute.Int(key, i)
	case bool:
		return attribute.Bool(key, i)
	case int64:
		return attribute.Int64(key, i)
	case float64:
		return attribute.Float64(key, i)
	case []string:
		return attribute.StringSlice(key, i)
	case []int:
		return attribute.IntSlice(key, i)
	case []bool:
		return attribute.BoolSlice(key, i)
	case []int64:
		return attribute.Int64Slice(key, i)
	case []float64:
		return attribute.Float64Slice(key, i)
	case fmt.Stringer:
		return attribute.Stringer(key, i)
	default:
		return attribute.String(key, fmt.Sprint(val))
	}
}

func NewClientHookTracer(tracer trace.TracerProvider) eudore.ClientHook {
	return &httpTrace{
		tracer: tracer.Tracer("github.com/eudore/eudore/_example"),
		propagator: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	}
}

type httpTrace struct {
	next       http.RoundTripper
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

func (*httpTrace) Name() string { return "trace" }

func (hook *httpTrace) Wrap(rt http.RoundTripper) http.RoundTripper {
	return &httpTrace{
		next:       rt,
		tracer:     hook.tracer,
		propagator: hook.propagator,
	}
}

func (hook *httpTrace) RoundTrip(req *http.Request) (*http.Response, error) {
	spanParent := trace.SpanFromContext(req.Context())
	if spanParent == nil {
		return hook.next.RoundTrip(req)
	}

	ctx, span := hook.tracer.Start(req.Context(), "eudore:app:client",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", req.Method),
		attribute.String("http.scheme", req.URL.Scheme),
		attribute.String("http.host", req.URL.Host),
		attribute.String("http.path", req.URL.Path),
	)
	req = req.WithContext(ctx)
	hook.propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := hook.next.RoundTrip(req)
	if resp != nil {
		span.SetAttributes(attribute.Int("http.status", resp.StatusCode))
	}
	if err != nil {
		span.RecordError(err)
	}

	return resp, err
}
