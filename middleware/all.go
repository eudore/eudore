package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/eudore/eudore"
)

// HandlerAdmin 函数返回Admin UI界面。
func HandlerAdmin(ctx eudore.Context) {
	ctx.SetHeader(eudore.HeaderXEudoreAdmin, "ui")
	ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextHTMLCharsetUtf8)
	http.ServeContent(
		ctx.Response(), ctx.Request(), "admin.html",
		now, strings.NewReader(AdminStatic),
	)
}

// NewBasicAuthFunc 创建一个Basic auth认证中间件。
//
// names为保存用户密码的map。
//
// 注意: BasicAuth需要放置在CORS之后。
func NewBasicAuthFunc(names map[string]string) eudore.HandlerFunc {
	checks := make(map[string]string, len(names))
	for name, pass := range names {
		checks[base64.StdEncoding.EncodeToString([]byte(name+":"+pass))] = name
	}
	return func(ctx eudore.Context) {
		auth := ctx.GetHeader(eudore.HeaderAuthorization)
		if len(auth) > 5 && auth[:6] == "Basic " {
			name, ok := checks[auth[6:]]
			if ok {
				ctx.SetParam(eudore.ParamBasicAuth, name)
				return
			}
		}
		ctx.SetHeader(eudore.HeaderWWWAuthenticate, "Basic")
		ctx.WriteHeader(eudore.StatusUnauthorized)
		ctx.End()
	}
}

// NewBodyLimitFunc 函数创建显示请求body长度的处理中间件。
func NewBodyLimitFunc(size int64) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		req := ctx.Request()
		switch {
		case req.Body == http.NoBody:
		case req.ContentLength > size:
			ctx.SetHeader(eudore.HeaderConnection, "close")
			ctx.WriteHeader(http.StatusRequestEntityTooLarge)
			_ = ctx.Render(eudore.NewContextMessgae(ctx, nil, &http.MaxBytesError{Limit: size}))
			ctx.End()
		default:
			var w http.ResponseWriter = ctx.Response()
			for {
				unwraper, ok := w.(interface{ Unwrap() http.ResponseWriter })
				if !ok {
					break
				}
				w = unwraper.Unwrap()
			}
			req.Body = http.MaxBytesReader(w, req.Body, size)
		}
	}
}

// NewContextWarpFunc 函数中间件使之后的处理函数使用的eudore.Context对象为新的Context。
//
// 装饰器下可以直接对Context进行包装，
// 而责任链下无法修改Context主体故设计该中间件作为中间件执行机制补充。
func NewContextWarpFunc(fn func(eudore.Context) eudore.Context) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		index, handler := ctx.GetHandler()
		wctx := &contextWarp{
			Context: fn(ctx),
			index:   index,
			handler: handler,
		}
		wctx.Next()
		ctx.SetHandler(wctx.index, wctx.handler)
	}
}

type contextWarp struct {
	eudore.Context
	index   int
	handler eudore.HandlerFuncs
}

// SetHandler 方法设置请求上下文的全部请求处理者。
func (ctx *contextWarp) SetHandler(index int, hs eudore.HandlerFuncs) {
	ctx.index, ctx.handler = index, hs
}

// GetHandler 方法获取请求上下文的当前处理索引和全部请求处理者。
func (ctx *contextWarp) GetHandler() (int, eudore.HandlerFuncs) {
	return ctx.index, ctx.handler
}

// Next 方法调用请求上下文下一个处理函数。
func (ctx *contextWarp) Next() {
	ctx.index++
	for ctx.index < len(ctx.handler) {
		ctx.handler[ctx.index](ctx)
		ctx.index++
	}
}

// End 结束请求上下文的处理。
func (ctx *contextWarp) End() {
	ctx.index = 0xff
	ctx.Context.End()
}

// NewHeaderFunc 函数创建响应header写入中间件。
func NewHeaderFunc(h http.Header) eudore.HandlerFunc {
	if len(h) == 0 {
		return nil
	}
	return func(ctx eudore.Context) {
		header := ctx.Response().Header()
		for k, v := range h {
			header[k] = append(header[k], v...)
		}
	}
}

// NewHeaderWithSecureFunc 函数创建响应header写入中间件，并额外附加基本安全header。
func NewHeaderWithSecureFunc(h http.Header) eudore.HandlerFunc {
	header := http.Header{
		eudore.HeaderXXSSProtection:      []string{"1; mode=block"},
		eudore.HeaderXFrameOptions:       []string{"SAMEORIGIN"},
		eudore.HeaderXContentTypeOptions: []string{"nosniff"},
	}
	for k, v := range h {
		header[k] = append(header[k], v...)
	}
	return NewHeaderFunc(header)
}

// NewHeaderFilteFunc 函数创建请求header过滤中间件，对来源于外部ip请求，过滤指定header。
func NewHeaderFilteFunc(iplist, names []string) eudore.HandlerFunc {
	if iplist == nil {
		iplist = []string{
			"10.0.0.0/8", "172.16.0.0/12", "192.0.0.0/24",
			"127.0.0.1", "127.0.0.10",
		}
	}
	if names == nil {
		names = []string{
			eudore.HeaderXRealIP, eudore.HeaderXForwardedFor,
			eudore.HeaderXForwardedHost, eudore.HeaderXForwardedProto,
			eudore.HeaderXRequestID, eudore.HeaderXTraceID,
		}
	}
	var list BlackNode
	for _, ip := range iplist {
		list.Insert(ip)
	}
	return func(ctx eudore.Context) {
		addr := ctx.Request().RemoteAddr
		pos := strings.IndexByte(addr, ':')
		if pos != -1 {
			addr = addr[:pos]
		}
		if list.Look(ip2int(addr)) {
			return
		}
		h := ctx.Request().Header
		for _, name := range names {
			h.Del(name)
		}
	}
}

// NewLoggerFunc 函数创建一个请求日志记录中间件。
//
// log参数设置用于输出eudore.Logger，
// params获取Context.Params如果不为空则添加到输出日志条目中
//
// 状态码如果为50x输出日志级别为Error。
func NewLoggerFunc(log eudore.Logger, params ...string) eudore.HandlerFunc {
	log = log.WithField("depth", "disable").WithField("logger", true)
	keys := [...]string{
		"method", "path", "realip", "proto", "host", "status", "request-time", "size",
	}
	headerkeys := [...]string{
		eudore.HeaderXRequestID,
		eudore.HeaderXTraceID,
		eudore.HeaderLocation,
	}
	headernames := [...]string{"x-request-id", "x-trace-id", "location"}
	return func(ctx eudore.Context) {
		now := time.Now()
		ctx.Next()
		status := ctx.Response().Status()
		// 连续WithField保证field顺序
		out := log.WithFields(keys[:], []any{
			ctx.Method(), ctx.Path(), ctx.RealIP(), ctx.Request().Proto,
			ctx.Host(), status, time.Since(now).String(), ctx.Response().Size(),
		})

		for _, param := range params {
			val := ctx.GetParam(param)
			if val != "" {
				out = out.WithField(strings.ToLower(param), val)
			}
		}

		if xforward := ctx.GetHeader(eudore.HeaderXForwardedFor); len(xforward) > 0 {
			out = out.WithField("x-forward-for", xforward)
		}
		headers := ctx.Response().Header()
		for i, key := range headerkeys {
			val := headers.Get(key)
			if val != "" {
				out = out.WithField(headernames[i], val)
			}
		}

		if status < 500 {
			out.Info()
		} else {
			if err := ctx.Err(); err != nil {
				out = out.WithField("error", err.Error())
			}
			out.Error()
		}
	}
}

// NewLoggerLevelFunc 函数创建一个设置一次请求日志级别的中间件。
//
// 通过一个函数处理请求，返回一个0-4,代表日志级别Debug-Fatal,默认处理函数使用debug参数转换成日志级别数字。
func NewLoggerLevelFunc(fn func(ctx eudore.Context) int) eudore.HandlerFunc {
	if fn == nil {
		fn = func(ctx eudore.Context) int {
			level := ctx.GetQuery("eudore_debug")
			if level != "" {
				return eudore.GetAnyByString[int](level)
			}
			return -1
		}
	}
	return func(ctx eudore.Context) {
		l := fn(ctx)
		if -1 < l && l < 5 {
			log := ctx.Value(eudore.ContextKeyLogger).(eudore.Logger).WithField("logger", true)
			log.SetLevel(eudore.LoggerLevel(l))
			ctx.SetValue(eudore.ContextKeyLogger, log)
		}
	}
}

// NewRecoverFunc 函数创建一个错误捕捉中间件，并返回500。
func NewRecoverFunc() eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		defer func() {
			r := recover()
			if r == nil {
				return
			}

			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			stack := eudore.GetCallerStacks(3)
			ctx.WithField("stack", stack).Error(err)
			if ctx.Response().Size() == 0 {
				ctx.WriteHeader(eudore.StatusInternalServerError)
				_ = ctx.Render(eudore.NewContextMessgae(ctx, err, stack))
			}
		}()
		ctx.Next()
	}
}

// NewRequestIDFunc 函数创建一个请求ID注入处理函数，不给定请求ID创建函数，
// 默认使用时间戳和随机数,会将request-id写入协议和附加到日志field。
func NewRequestIDFunc(fn func(eudore.Context) string) eudore.HandlerFunc {
	if fn == nil {
		fn = func(eudore.Context) string {
			randkey := make([]byte, 3)
			_, _ = io.ReadFull(rand.Reader, randkey)
			return fmt.Sprintf("%d-%x", time.Now().UnixNano(), randkey)
		}
	}
	return func(ctx eudore.Context) {
		requestID := ctx.GetHeader(eudore.HeaderXRequestID)
		if requestID == "" {
			requestID = fn(ctx)
		}
		ctx.SetHeader(eudore.HeaderXRequestID, requestID)

		log := ctx.Value(eudore.ContextKeyLogger).(eudore.Logger)
		log = log.WithField("x-request-id", requestID).WithField("logger", true)
		ctx.SetValue(eudore.ContextKeyLogger, log)
	}
}
