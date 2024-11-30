package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/eudore/eudore"
)

// Middleware is [eudore.HandlerFunc] alias, show handler sort in godoc.
type Middleware = eudore.HandlerFunc

// The NewAdminFunc function returns the Admin UI interface.
//
//go:noinline
func NewAdminFunc() Middleware {
	return func(ctx eudore.Context) {
		ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextHTMLCharsetUtf8)
		http.ServeContent(
			ctx.Response(), ctx.Request(), "",
			eudore.DefaultHandlerEmbedTime,
			strings.NewReader(DefaultPageAdmin),
		)
	}
}

// NewBasicAuthFunc function creates middleware to implement
// Basic auth authentication.
//
// names is a map that stores user passwords.
//
// Note: BasicAuth needs to be placed after [NewCORSFunc].
func NewBasicAuthFunc(names map[string]string) Middleware {
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
		writePage(ctx, eudore.StatusUnauthorized, DefaultPageBasicAuth, "")
		ctx.End()
	}
}

// The NewBodyLimitFunc function creates middleware to implement
// that limits the request body length.
//
// If the body length exceeds the limit,
// [eudore.StatusRequestEntityTooLarge] is returned.
//
// http/1.x cannot merge Reader Body with [NewBodySizeFunc].
//
// refer: [http.MaxBytesReader].
//
//go:noinline
func NewBodyLimitFunc(size int64) Middleware {
	return func(ctx eudore.Context) {
		r := ctx.Request()
		switch {
		case r.ContentLength == -1:
			var w http.ResponseWriter = ctx.Response()
			if ctx.Request().ProtoMajor == 1 {
				for {
					unwraper, ok := w.(interface{ Unwrap() http.ResponseWriter })
					if !ok {
						break
					}
					w = unwraper.Unwrap()
				}
			}
			r.Body = http.MaxBytesReader(w, r.Body, size)
		case r.ContentLength > size:
			ctx.SetHeader(eudore.HeaderConnection, "close")
			writePage(ctx, eudore.StatusRequestEntityTooLarge,
				DefaultPageBodyLimit, strconv.FormatInt(size, 10),
			)
			ctx.End()
		}
	}
}

// The NewBodySizeFunc function creates the middleware implement
// update the request Body Size.
//
// If ctx.Request().ContentLength == -1, update ContentLength on Close.
//
//go:noinline
func NewBodySizeFunc() Middleware {
	return func(ctx eudore.Context) {
		r := ctx.Request()
		if r.ContentLength == -1 {
			reader := &readerLength{r.Body, 0}
			r.Body = reader
			defer reader.release(ctx)
			ctx.Next()
		}
	}
}

type readerLength struct {
	io.ReadCloser
	Length int64
}

func (r *readerLength) release(ctx eudore.Context) {
	r.Close()
	ctx.Request().ContentLength = r.Length
}

func (r *readerLength) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	r.Length += int64(n)
	return n, err
}

func (r *readerLength) Close() error {
	n, _ := io.Copy(io.Discard, r.ReadCloser)
	r.Length += n
	return r.ReadCloser.Close()
}

type csrf struct {
	GetKeyFunc func(eudore.Context) string
	Cookie     http.Cookie
}

// The NewCSRFFunc function creates middleware to implement CSRF double
// verification.
//
// key specifies the GetQuery parameter for get the CSRF value.
//
// options: [NewOptionKeyFunc] [NewOptionCSRFCookie].
func NewCSRFFunc(key string, options ...Option) Middleware {
	c := &csrf{
		GetKeyFunc: func(ctx eudore.Context) string {
			return ctx.GetQuery(key)
		},
		Cookie: http.Cookie{
			Name: "_csrf",
		},
	}
	applyOption(c, options)
	return func(ctx eudore.Context) {
		key := ctx.GetCookie(c.Cookie.Name)
		if key == "" {
			key = eudore.GetStringRandom(32)
			newcookie := c.Cookie
			newcookie.Value = key
			ctx.SetCookie(&newcookie)
		}
		switch ctx.Method() {
		case eudore.MethodGet, eudore.MethodHead, eudore.MethodOptions, eudore.MethodTrace:
			return
		}

		have := c.GetKeyFunc(ctx)
		if have != key {
			writePage(ctx, eudore.StatusBadRequest, DefaultPageCSRF, have)
			ctx.End()
		}
	}
}

// The NewContextWrapperFunc function creates middleware to implement
// modify the [eudore.Context] used by Next [eudore.HandlerFunc].
//
// [eudore.Context] will be reset in [NewTimeoutFunc].
//
// This middleware is a supplement to the middleware execution mechanism.
//
//go:noinline
func NewContextWrapperFunc(fn func(eudore.Context) eudore.Context) Middleware {
	return func(ctx eudore.Context) {
		wrap := &contextWraper{contextBase: fn(ctx)}
		wrap.index, wrap.handlers = ctx.GetHandlers()
		defer wrap.release()
		wrap.Next()
	}
}

type contextWraper struct {
	contextBase
	index    int
	handlers []eudore.HandlerFunc
}

type contextBase = eudore.Context

func (ctx *contextWraper) release() {
	ctx.contextBase.SetHandlers(ctx.index, ctx.handlers)
}

// The SetHandler method sets all [HandlerFunc] for the request context.
func (ctx *contextWraper) SetHandlers(index int, hs []eudore.HandlerFunc) {
	ctx.index, ctx.handlers = index, hs
}

// The GetHandler method gets the current processing index
// and all [HandlerFunc] of the request context.
func (ctx *contextWraper) GetHandlers() (int, []eudore.HandlerFunc) {
	return ctx.index, ctx.handlers
}

// The Next method modifies the [eudore.Context] used by [HandlerFunc].
func (ctx *contextWraper) Next() {
	ctx.index++
	for ctx.index < len(ctx.handlers) {
		ctx.handlers[ctx.index](ctx)
		ctx.index++
	}
}

// End Ends processing of the request context.
func (ctx *contextWraper) End() {
	ctx.index = eudore.DefaultContextMaxHandler
}

// The NewHealthCheckFunc function creates [eudore.HandlerFunc] to check
// metadata health.
//
// Get metadata for [eudore.ContextKeyAppValues] from [context.Context],
// and returns only [eudore.StatusServiceUnavailable] if Health=false exists.
//
//go:noinline
func NewHealthCheckFunc(app context.Context) Middleware {
	return func(ctx eudore.Context) {
		vals := app.Value(eudore.ContextKeyAppValues).([]any)
		for i := 0; i < len(vals); i += 2 {
			meta := anyMetadata(vals[i+1])
			if meta != nil && !handlerHealthValue(meta) {
				writePage(ctx, eudore.StatusServiceUnavailable,
					DefaultPageHealth, fmt.Sprint(vals[i]),
				)
				return
			}
		}
		_, _ = ctx.WriteString("OK")
	}
}

// The NewMetadataFunc function creates [eudore.HandlerFunc] to gets all
// metadata.
//
// Get metadata for [eudore.ContextKeyAppValues] from [context.Context].
// The NewHandlerMetadata function gets the metadata of
// [eudore.ContextKeyAppValues].
//
// If Health=false exists, only [eudore.StatusServiceUnavailable] is returned.
//
// All metadata will be returned; contentKey can be specified
// using the route params 'name'.
//
//go:noinline
func NewMetadataFunc(app context.Context) Middleware {
	return func(ctx eudore.Context) {
		name := ctx.GetParam("name")
		if name != "" {
			meta := anyMetadata(app.Value(eudore.NewContextKey(name)))
			if meta != nil {
				_ = ctx.Render(meta)
			} else {
				eudore.HandlerRouter404(ctx)
			}
			return
		}

		vals := app.Value(eudore.ContextKeyAppValues).([]any)
		healthy := true
		metas := make(map[string]any, len(vals)/2)
		for i := 0; i < len(vals); i += 2 {
			meta := anyMetadata(vals[i+1])
			if meta != nil {
				if healthy {
					healthy = handlerHealthValue(meta)
				}
				metas[fmt.Sprint(vals[i])] = meta
			}
		}
		if !healthy {
			ctx.WriteStatus(eudore.StatusServiceUnavailable)
		}
		_ = ctx.Render(metas)
	}
}

func anyMetadata(i any) any {
	metaer, ok := i.(interface{ Metadata() any })
	if ok {
		return metaer.Metadata()
	}
	return nil
}

func handlerHealthValue(i any) bool {
	iValue := reflect.ValueOf(i)
	for {
		switch iValue.Kind() {
		case reflect.Ptr, reflect.Interface:
			iValue = iValue.Elem()
		case reflect.Struct:
			iValue = iValue.FieldByName("Health")
			if iValue.Kind() == reflect.Bool {
				return iValue.Bool()
			}
			return true
		default:
			return true
		}
	}
}

// The NewHeaderAddFunc function creates middleware to implement
// add response [http.Header].
//
//go:noinline
func NewHeaderAddFunc(h http.Header) Middleware {
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

// The NewHeaderAddSecureFunc function creates middleware to implement
// add a response [http.Header],
// and additionally appends a basic security [http.Header].
//
// Append security headers: [eudore.HeaderXXSSProtection]
// [eudore.HeaderXFrameOptions] [eudore.HeaderXContentTypeOptions].
//
//go:noinline
func NewHeaderAddSecureFunc(h http.Header) Middleware {
	header := http.Header{
		eudore.HeaderXContentTypeOptions: []string{"nosniff"},
		eudore.HeaderXFrameOptions:       []string{"SAMEORIGIN"},
		eudore.HeaderXXSSProtection:      []string{"1; mode=block"},
	}
	headerCopy(header, h)
	return NewHeaderAddFunc(header)
}

// The NewHeaderDeleteFunc function creates middleware to implement
// delete request [http.Header].
// If the IP is not in the sets, it delete the specified header to
// prevent forgery of internal headers.
//
// Delete headers by default:
// [eudore.HeaderXRealIP] [eudore.HeaderXForwardedFor]
// [eudore.HeaderXForwardedHost] [eudore.HeaderXForwardedProto]
// [eudore.HeaderXRequestID] [eudore.HeaderXTraceID].
func NewHeaderDeleteFunc(iplist, names []string) Middleware {
	if iplist == nil {
		iplist = []string{
			"127.0.0.1", "10.0.0.0/8", "172.16.0.0/12", "192.0.0.0/24",
		}
	}
	if names == nil {
		names = []string{
			eudore.HeaderXRealIP, eudore.HeaderXForwardedFor,
			eudore.HeaderXForwardedHost, eudore.HeaderXForwardedProto,
			eudore.HeaderXRequestID, eudore.HeaderXTraceID,
		}
	}

	list := &subnetListMixin{
		V4: &subnetListV4{},
		V6: &subnetListV6{},
	}
	for _, ip := range iplist {
		list.Insert(ip)
	}
	return func(ctx eudore.Context) {
		addr := ctx.Request().RemoteAddr
		if addr == "pipe" {
			return
		}
		addr = addr[:strings.LastIndexByte(addr, ':')]
		// ipv6
		if len(addr) > 1 && addr[0] == '[' {
			addr = addr[1 : len(addr)-1]
		}

		if list.Look(addr) {
			return
		}
		h := ctx.Request().Header
		for _, name := range names {
			h.Del(name)
		}
	}
}

// The NewRecoveryFunc function creates middleware to implement recover errors
// and return 500 and a detailed message.
//
//go:noinline
func NewRecoveryFunc() Middleware {
	type m interface {
		Unwrap() error
		Stack() []string
	}
	release := func(ctx eudore.Context) {
		r := recover()
		if r == nil {
			return
		}

		var err error
		stack := eudore.GetCallerStacks(3)
		switch v := r.(type) {
		case error:
			err = v
		case m:
			err = v.Unwrap()
			stack = append(v.Stack(), stack[1:]...)
		default:
			err = fmt.Errorf("%v", r)
		}
		if ctx.Response().Size() == 0 {
			ctx.WriteStatus(eudore.StatusInternalServerError)
			_ = ctx.Render(eudore.NewContextMessgae(ctx, err, stack))
		}
		ctx.WithField("stack", stack).Error(err)
		ctx.End()
	}
	return func(ctx eudore.Context) {
		defer release(ctx)
		ctx.Next()
	}
}

// The NewRequestIDFunc function creates middleware to implement
// setting [eudore.HeaderXRequestID]
// and appends x-request-id to the log field.
//
// Timestamp and random number are used by default.
//
//go:noinline
func NewRequestIDFunc(fn func(eudore.Context) string) Middleware {
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

// The NewRoutesFunc function creates middleware to implement
// uses Routes to create [NewRouterFunc] middleware.
func NewRoutesFunc(routes map[string]any) Middleware {
	router := eudore.NewRouterCoreMux()
	router.HandleFunc("404", "", []eudore.HandlerFunc{})
	router.HandleFunc("405", "", []eudore.HandlerFunc{})
	for k, v := range routes {
		h := eudore.DefaultHandlerExtender.CreateHandlers(k, v)
		pos := strings.IndexByte(k, '/')
		if pos > 1 {
			router.HandleFunc(strings.TrimSpace(k[:pos]), k[pos:], h)
		} else {
			router.HandleFunc(eudore.MethodAny, k, h)
		}
	}
	return NewRouterFunc(router)
}

// The NewRouterFunc function creates middleware to implement execution Router.
//
// It can be used as a front [eudore.Router] of the [App.Router],
// or as a sub [eudore.Router] of the [App.Router].
//
// [eudore.Params] matched by this [eudore.Router] will also be added to the
// [eudore.Context]. May pollute the [eudore.Params] added by [app.Router].
//
// This router uses the End method to stop external [eudore.HandlerFunc].
//
//go:noinline
func NewRouterFunc(router eudore.RouterCore) Middleware {
	release := func(ctx eudore.Context, index int, handlers []Middleware) {
		i, _ := ctx.GetHandlers()
		if i > eudore.DefaultContextMaxHandler {
			index = i
		}
		ctx.SetHandlers(index, handlers)
	}
	return func(ctx eudore.Context) {
		route := ctx.GetParam(eudore.ParamRoute)
		// reset ParamRoute
		if route != "" {
			defer ctx.SetParam(eudore.ParamRoute, route)
		}

		h := router.Match(ctx.Method(), ctx.Path(), ctx.Params())
		switch len(h) {
		case 0:
		case 1:
			h[0](ctx)
		default:
			// reset handlers
			index, handlers := ctx.GetHandlers()
			defer release(ctx, index, handlers)
			ctx.SetHandlers(-1, h)
			ctx.Next()
		}
	}
}

// The NewServerTimingFunc function creates middleware to implement writing
// [eudore.HeaderServerTiming].
//
// Record the time from the start to the first message written.
//
//go:noinline
func NewServerTimingFunc() Middleware {
	return func(ctx eudore.Context) {
		ctx.SetResponse(&responseWriterTiming{ctx.Response(), time.Now(), true})
	}
}

type responseWriterTiming struct {
	eudore.ResponseWriter
	Now    time.Time
	timing bool
}

func (w *responseWriterTiming) Write(p []byte) (int, error) {
	w.writeTiming()
	return w.ResponseWriter.Write(p)
}

func (w *responseWriterTiming) WriteString(p string) (int, error) {
	w.writeTiming()
	return w.ResponseWriter.WriteString(p)
}

func (w *responseWriterTiming) WriteHeader(code int) {
	w.writeTiming()
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterTiming) Flush() {
	w.writeTiming()
	w.ResponseWriter.Flush()
}

func (w *responseWriterTiming) writeTiming() {
	if w.timing {
		w.timing = false
		dura := float64(time.Since(w.Now)) / float64(time.Millisecond)
		tims := w.Header()[eudore.HeaderServerTiming]
		tims = append(tims, "total;dur="+strconv.FormatFloat(dura, 'f', 2, 64))
		w.Header().Set(eudore.HeaderServerTiming, strings.Join(tims, ", "))
	}
}
