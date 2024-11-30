package main

import (
	"embed"
	"net/http"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
	"github.com/google/uuid"
	"github.com/klauspost/compress/zstd"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

/*
NewCompressionMixinsFunc
NewCompressionFunc
NewGzipFunc

HeaderCacheControl
NewHeaderAddFunc

NewRouterFunc
NewRoutesFunc

NewTimeoutFunc
NewTimeoutSkipFunc
*/

//go:embed *.go
var RootFS embed.FS

func init() {
	// load br zstd compress
	middleware.DefaultCompressionEncoder[middleware.CompressionNameBrotli] = func() any {
		return brotli.NewWriter(nil)
	}
	middleware.DefaultCompressionEncoder[middleware.CompressionNameZstandard] = func() any {
		w, _ := zstd.NewWriter(nil)
		return w
	}
}

func main() {
	app := eudore.NewApp()
	app.AddMiddleware("global",
		middleware.NewLoggerFunc(app,
			"remote-addr",
			"param:"+eudore.ParamBasicAuth,
		),
		middleware.NewHeaderDeleteFunc(nil, nil),
		middleware.NewRequestIDFunc(func(eudore.Context) string {
			return uuid.New().String()
		}),
		middleware.NewCompressionMixinsFunc(nil),

		// rewrite
		middleware.NewRewriteFunc(map[string]string{
			"/api/v2/*": "/api/3/$0",
		}),
		middleware.NewRoutesFunc(map[string]any{
			"/api/v1/*": func(ctx eudore.Context) {
				ctx.Warningf("rewrite api v1 path: %s", ctx.Path())
				path := "/api/v3/" + ctx.GetParam("*")
				req := ctx.Request()
				req.URL.Path = path
				req.RequestURI = req.URL.String()
			},
		}),
	)
	app.GetFunc("/metrics", promhttp.Handler())
	app.GetFunc("/health", middleware.NewHealthCheckFunc(app))
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)

	admin := app.Group("/eudore/debug")
	admin.AddMiddleware(
		middleware.NewBasicAuthFunc(map[string]string{
			"username": "password",
			"eudore":   "12345678",
		}),
	)
	admin.GetFunc("/admin/ui", middleware.NewAdminFunc())
	admin.GetFunc("/pprof/*", middleware.NewPProfFunc())
	admin.GetFunc("/look/*", middleware.NewLookFunc(app))
	admin.GetFunc("/metadata/*", middleware.NewMetadataFunc(app))
	// disable admin group logger
	admin = admin.Group(" loggerkind=~all")

	static := app.Group("")
	static.AddMiddleware(
		middleware.NewHeaderAddSecureFunc(http.Header{
			eudore.HeaderCacheControl: []string{"no-cache"},
		}),
		middleware.NewRefererCheckFunc(map[string]bool{
			"":            true,
			"origin":      true,
			"*":           false,
			"*.eudore.cn": true,
		}),
		// 4MB/s
		middleware.NewRateSpeedFunc(4<<20, 32<<32,
			middleware.NewOptionRateCleanup(app, time.Second*1800, 64),
		),
	)
	static.GetFunc("/static/* autoindex=true", eudore.NewHandlerFileSystems(RootFS, "."))

	api := app.Group("/api")
	api.AddMiddleware(
		"/",
		middleware.NewBlackListFunc(
			map[string]bool{
				"0.0.0.0/0":      false,
				"10.0.0.0/8":     true,
				"172.16.0.0/12":  true,
				"192.168.0.0/16": true,
				"43.227.0.0/16":  true, //  me
				"::0/0":          false,
				"::1":            true,
			},
			middleware.NewOptionRouter(admin),
		),
		middleware.NewCORSFunc([]string{"*.eudore.cn", "127.0.0.1:*"}, map[string]string{
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, HEAD",
			"Access-Control-Allow-Headers":     "Content-Type,X-Request-Id,X-CustomHeader",
			"Access-Control-Expose-Headers":    "X-Request-Id",
			"Access-Control-Max-Age":           "1000",
		}),
		middleware.NewCSRFFunc("_csrf"),
		middleware.NewRateRequestFunc(3, 30,
			middleware.NewOptionRateCleanup(app, time.Second*600, 100),
		),
		middleware.NewCircuitBreakerFunc(
			middleware.NewOptionRouter(admin),
		),

		middleware.NewDumpFunc(admin),
		middleware.NewBodySizeFunc(),
		middleware.NewBodyLimitFunc(4<<20), // 4MB
		middleware.NewServerTimingFunc(),
		middleware.NewLoggerLevelFunc(nil),
		middleware.NewCacheFunc(time.Second*10),
		middleware.NewRecoveryFunc(),
		middleware.NewTimeoutFunc(app.ContextPool, time.Second*10),
		middleware.NewContextWrapperFunc(func(ctx eudore.Context) eudore.Context {
			return contextWraper{ctx}
		}),
	)
	apiv3 := api.Group("/v3 loggerkind=~handler")
	apiv3.AddController(
		NewUsersController(),
	)

	app.GetFunc("/custom",
		NewPreFunc(eudore.HandlerEmpty),
		NewPostFunc(eudore.HandlerEmpty),
		NewPrePostFunc(eudore.HandlerEmpty, eudore.HandlerEmpty),
	)

	app.Listen(":8088")
	app.Run()
}

type contextWraper struct {
	contextBase
}

type contextBase = eudore.Context

func NewPreFunc(pre eudore.HandlerFunc) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		pre(ctx)
	}
}

func NewPostFunc(post eudore.HandlerFunc) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		defer post(ctx)
		ctx.Next()
	}
}

func NewPrePostFunc(pre, post eudore.HandlerFunc) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		pre(ctx)
		defer post(ctx)
		ctx.Next()
	}
}

type UsersController struct {
	eudore.ControllerAutoRoute
}

func NewUsersController() eudore.Controller {
	return &UsersController{}
}

func (ctl *UsersController) Get(eudore.Context) {
}

func (ctl *UsersController) Any(eudore.Context) {
}

func (ctl *UsersController) GetZZZ(eudore.Context) {
}

func (ctl *UsersController) GetZZZXi(eudore.Context) {
}
