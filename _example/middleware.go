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

		// rewrite, must be before the router
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

	// disable admin group logger
	admin := app.Group("/eudore/debug loggerkind=~all")
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

	static := app.Group("")
	static.AddMiddleware(
		middleware.NewHeaderSecureFunc(http.Header{
			eudore.HeaderCacheControl: []string{"no-cache"},
		}),
		middleware.NewRefererCheckFunc(map[string]bool{
			"":            true,
			"origin":      true,
			"*":           false,
			"*.eudore.cn": true,
		}),
		// 4MB/s
		middleware.NewRateSpeedFunc(4<<20, 128<<20,
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
				"127.0.0.1":      true,
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
		middleware.NewBearerAuthFunc("secret"),
		middleware.NewRateRequestFunc(3, 30,
			middleware.NewOptionRateCleanup(app, time.Second*600, 100),
			middleware.NewOptionKeyFunc(func(ctx eudore.Context) string {
				user := ctx.GetParam(eudore.ParamUserid)
				if user != "" {
					return user
				}
				return ctx.RealIP()
			}),
		),
		middleware.NewCircuitBreakerFunc(middleware.NewOptionRouter(admin)),
		middleware.NewSecurityPolicysFunc([]string{
			`{"user": "<Guest User>", "policy": ["Guest"]}`,
			`{"user": "10000", "policy": ["Local","Administrator"], "data":["Local"]}`,
			`{"name":"Administrator", "statement": [{"effect": true, "action": ["*"]}]}`,
			`{"name":"Guest", "statement": [{"effect": true, "action": ["*:*:Get*"]}]}`,
			`{"name":"Local", "statement": [{"effect": true, "data": {"menu":["Get*"]}, "action": ["*"], "conditions": {
				"method": ["GET", "POST", "OPTIONS","XX"],
				"sourceip": ["127.0.0.1", "192.168.0.0/24", "43.227.0.0/16"],
				"date": {"after": "2025-07-31", "before": "2025-08-31"},
				"params": {
					"Userid": ["1", "2", "10000"],
					"group_id": ["1",""]
				}
			}}]}`}, middleware.NewOptionRouter(admin),
		),
		middleware.NewDumpFunc(admin),
		middleware.NewBodySizeFunc(),
		middleware.NewBodyLimitFunc(4<<20), // 4MB
		middleware.NewServerTimingFunc(),
		middleware.NewLoggerLevelFunc(nil),
		middleware.NewCacheFunc(time.Second*10),
		middleware.NewRecoveryFunc(),
		middleware.NewTimeoutFunc(app.ContextPool, time.Second*10),
	)
	api.AddHandler(eudore.MethodOptions, "/", eudore.HandlerRouter403)

	eudore.DefaultControllerParam = "Action={{Package}}:{{Name}}:{{Method}}"
	apiv3 := api.Group("/v3 loggerkind=~handler")
	apiv3.AddController(
		NewUsersController(),
	)

	app.GetRequest("/api/v3/users/eudore")
	app.GetRequest("/api/v3/users/zzz")
	app.GetRequest("/api/v3/users/zzz/xi")

	app.GetFunc("/custom",
		NewPreFunc(eudore.HandlerEmpty),
		NewPostFunc(eudore.HandlerEmpty),
		NewPrePostFunc(eudore.HandlerEmpty, eudore.HandlerEmpty),
	)

	app.Listen(":8088")
	app.Run()
}

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

func (ctl *UsersController) Get() {
}

func (ctl *UsersController) Any(eudore.Context) {
}

func (ctl *UsersController) GetZZZ(eudore.Context) {
}

func (ctl *UsersController) GetZZZXi(eudore.Context) {
}
