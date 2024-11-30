package eudore

// Context defines a http request context

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

// Context defines the request context interface,
// Included: Context data, *[http.Request], Request params, [ResponseWriter],
// and log output.
//
// You can modify or add Context methods through
// [App.ContextPool] [HandlerExtender]
// [github.com/eudore/eudore/middleware.NewContextWrapFunc].
type Context interface {
	// context

	Reset(w http.ResponseWriter, r *http.Request)
	// Context Get the context of the current request.
	// ctx.Request().Context() is not equal to ctx.Context().
	Context() context.Context
	Request() *http.Request
	Response() ResponseWriter
	Value(key any) any
	SetContext(c context.Context)
	SetRequest(r *http.Request)
	SetResponse(w ResponseWriter)
	// SetValue sets the Value of the built-in [context.Context],
	// which can be read by calling the [Value] method.
	//
	// String type parameters are prioritized using [SetParam].
	SetValue(key any, val any)
	// handles
	SetHandlers(index int, handlers []HandlerFunc)
	GetHandlers() (int, []HandlerFunc)
	// Next calls the next [HandlerFunc] of the request context.
	//
	// If there are remaining [HandlerFunc], it will be called automatically
	//
	// The code after Next will not be executed because of panic.
	Next()
	// End Ends current handlers of the request context.
	//
	// The Fatal/Fatalf method contains the End method.
	End()
	Err() error

	// request

	// The Read method implements [io.Reader] to read http requests.
	Read(b []byte) (int, error)
	Host() string
	Method() string
	// Path returns the request path, alias ctx.Request().URL.Path.
	Path() string
	// RealIP get the user's real IP, reads
	// [HeaderXRealIP] [HeaderXForwardedFor] and [http.Request.RemoteAddr]
	//
	// If the server does not have a proxy layer,
	// It is necessary to use middleware to filter the request header to
	// prevent forgery of real-ip.
	RealIP() string
	// Body returns the request body and saves it to the cache.
	// The Body method can be called repeatedly.
	// Each call will reset the ctx.Request().Body object to a body reader.
	//
	// ctx.bodyContent will not be reused with memory.
	// Normally, you should avoid calling the Body method;
	// If you use it, you should set [middleware.NewBodyLimitFunc] to
	// avoid large bodies occupying memory.
	//
	// If [NewBodyLimitFunc] is used, error may return [http.MaxBytesError].
	Body() ([]byte, error)
	// Bind uses the [ContextKeyBind] function loaded
	// in [NewContextBaseFunc] to bind data.
	// Use [NewHandlerDataBinds] by default.
	Bind(data any) error

	// param query header cookie form

	// Params returns the [Context] parameter containing the route parameters,
	// [ParamRoute] gets the route.
	Params() *Params
	GetParam(key string) string
	SetParam(key string, val string)
	// The Query method returns the uri parameter get
	// by parsing ctx.Request().URL.RawQuery.
	//
	// The parsed data is saved in Request().Form, and the body is not parsed.
	Querys() (url.Values, error)
	// refer Querys
	GetQuery(key string) string
	// GetHeader gets a request header, alias ctx.Request().Header().Get(name).
	GetHeader(key string) string
	// SetHeader sets a response header,
	// alias ctx.Response().Header().Set(name, val).
	SetHeader(key string, val string)
	// Cookies gets all cookies from [HeaderCookie] and
	// parses the data after the first call to the [Cookies]/[GetCookie] method.
	Cookies() []Cookie
	// refer [Cookies]
	GetCookie(key string) string
	// SetCookie sets [HeaderSetCookie], allowing the cookie option to be set.
	SetCookie(cookie *CookieSet)
	// SetCookieValue sets [HeaderSetCookie] key and value.
	SetCookieValue(name string, value string, age int)
	// The FormValue method returns the parsed body data in
	// MimeApplicationForm/MimeMultipartForm/MimeMultipartMixed format.
	//
	// The uri parameter is used only when the body is empty.
	//
	// The parsed data is saved in Request().PostForm or MultipartForm.
	FormValue(key string) string
	// refer FormValue
	FormValues() (map[string][]string, error)
	// refer FormValue
	FormFile(key string) *multipart.FileHeader
	// refer FormValue
	FormFiles() map[string][]*multipart.FileHeader

	// response

	Write(b []byte) (int, error)
	WriteString(s string) (int, error)
	// WriteStatus sets the status code but does not write.
	//
	// Automatically write code at the first Write or WriteString.
	WriteStatus(code int)
	// WriteHeader method writing status code and [http.Header],
	// [http.Header] cannot be set after calling.
	WriteHeader(code int)
	// WriteFile opens the file and responds using [http.ServeContent].
	WriteFile(path string) error
	// The Redirect method uses [http.Redirect] to redirect url.
	Redirect(code int, url string) error
	// Render uses the [ContextKeyRender] function loaded
	// in [NewContextBaseFunc] to Render data.
	// Use [NewHandlerDataRenders] by default.
	//
	// Use the [WriteStatus] method to set the status code;
	// if the [WriteHeader] method is used,
	// Render cannot write [HeaderContentType].
	Render(data any) error

	// Logger interface

	Debug(args ...any)
	Info(args ...any)
	Warning(args ...any)
	Error(args ...any)
	// The Fatal method outputs the [LoggerError] logger, returns a message,
	// and ends request processing.
	//
	// Use the [WriteStatus] to write the status code of the render Message.
	//
	// If Response.Size=0, the response will be written.
	Fatal(args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warningf(format string, args ...any)
	Errorf(format string, args ...any)
	// refer Fatal
	Fatalf(format string, args ...any)
	// The WithField method uses [Logger.WithField] to return the [Logger]
	// associated with the [Context].
	// It is not allowed to be used after the HandlerFunc ends.
	//
	// Use ctx.Value(eudore.ContextKeyLogger).(eudore.Logger) to
	// get a non-associated Logger.
	WithField(key string, val any) Logger
	// refer WithField
	WithFields(keys []string, vals []any) Logger
}

// contextBase implements the Context interface.
type contextBase struct {
	// context
	index          int
	handlers       []HandlerFunc
	params         Params
	config         *contextBaseConfig
	RequestReader  *http.Request
	ResponseWriter ResponseWriter
	context        context.Context
	// data
	contextValues contextBaseValue
	httpResponse  responseWriterHTTP
	wantStatus    int
	cookies       []Cookie
	bodyContent   []byte
}

type contextBaseConfig struct {
	Logger                 Logger
	Bind                   func(Context, any) error
	Render                 func(Context, any) error
	MaxApplicationFormSize int64
	MaxMultipartFormMemory int64
}

// The ResponseWriter interface writes the http response body status, header,
// and body.
//
// Combines [http.ResponseWriter], [http.Flusher], [http.Hijacker],
// and [http.Pusher] interfaces.
type ResponseWriter interface {
	// The Write method implements the [io.Writer] interface.
	Write(b []byte) (int, error)
	// The WriteString method implements the [io.StringWriter] interface.
	WriteString(s string) (int, error)
	// The WriteHeader method implements writing status code and [http.Header],
	// which will be written only at the first Write.
	WriteHeader(code int)
	// The WriteStatus method records the status code set by [Context] and
	// will not be written automatically.
	WriteStatus(code int)

	Header() http.Header
	// The Flush method implements the [http.Flusher] interface,
	// refreshes and sends the buffer.
	Flush()
	// The Hijack method implements the [http.Hijacker] interface,
	// used for websocket.
	Hijack() (net.Conn, *bufio.ReadWriter, error) // Only http1
	// The Push method implements the [http.Pusher] interface.
	//
	// support of HTTP/2 Server Push will be disabled by default in
	// Chrome 106 and other Chromium-based browsers.
	Push(path string, opts *http.PushOptions) error // Only http2

	Size() int
	Status() int
}

// CookieSet defines the generation of the [HeaderSetCookie] value, which is
// an alias for [http.Cookie].
type CookieSet = http.Cookie

// Cookie defines the data of the [HeaderCookie] read by the request.
type Cookie struct {
	Name  string
	Value string
}

// The NewContextBasePool function creates a [Context] [sync.Pool] from the
// [context.Context].
//
// refer: [NewContextBaseFunc].
func NewContextBasePool(ctx context.Context) *sync.Pool {
	config := newContextBaseConfig(ctx)
	return &sync.Pool{
		New: func() any {
			return &contextBase{
				config: config,
				params: Params{ParamRoute, ""},
			}
		},
	}
}

// The NewContextBaseFunc function uses [context.Context] to create a [Context]
// constructor.
//
// Load [ContextKeyApp] implement the [Logger] interface from the
// [context.Context].
//
// Load [ContextKeyBind] [ContextKeyRender] is [HandlerDataFunc] from the
// [context.Context].
//
// If the [App] updates this data,
// you need to reset the [ContextKeyContextPool].
func NewContextBaseFunc(ctx context.Context) func() Context {
	config := newContextBaseConfig(ctx)
	return func() Context {
		return &contextBase{
			config: config,
			params: Params{ParamRoute, ""},
		}
	}
}

func newContextBaseConfig(ctx context.Context) *contextBaseConfig {
	bind, _ := ctx.Value(ContextKeyBind).(func(Context, any) error)
	render, _ := ctx.Value(ContextKeyRender).(func(Context, any) error)
	if bind == nil {
		bind = NewHandlerDataBinds(nil)
	}
	if render == nil {
		render = NewHandlerDataRenders(nil)
	}
	return &contextBaseConfig{
		Logger:                 NewLoggerWithContext(ctx),
		Bind:                   bind,
		Render:                 render,
		MaxApplicationFormSize: DefaultContextMaxApplicationFormSize,
		MaxMultipartFormMemory: DefaultContextMaxMultipartFormMemory,
	}
}

// The Reset function resets the Context data.
func (ctx *contextBase) Reset(w http.ResponseWriter, r *http.Request) {
	ctx.context = &ctx.contextValues
	ctx.ResponseWriter = &ctx.httpResponse
	ctx.RequestReader = r
	ctx.params = ctx.params[0:2]
	ctx.params[1] = ""
	ctx.contextValues.Reset(r.Context(), ctx.config)
	ctx.httpResponse.Reset(w)
	ctx.wantStatus = StatusOK
	ctx.cookies = ctx.cookies[:0]
	ctx.bodyContent = nil
}

func (ctx *contextBase) Context() context.Context {
	base, ok := ctx.context.(*contextBaseValue)
	if ok && base == &ctx.contextValues {
		ctx.context = base.Clone()
	}
	return ctx.context
}

func (ctx *contextBase) Request() *http.Request {
	return ctx.RequestReader
}

func (ctx *contextBase) Response() ResponseWriter {
	return ctx.ResponseWriter
}

func (ctx *contextBase) Value(key any) any {
	return ctx.context.Value(key)
}

func (ctx *contextBase) SetContext(c context.Context) {
	ctx.context = c
}

func (ctx *contextBase) SetRequest(r *http.Request) {
	ctx.RequestReader = r
}

func (ctx *contextBase) SetResponse(w ResponseWriter) {
	ctx.ResponseWriter = w
}

func (ctx *contextBase) SetValue(key, val any) {
	base, ok := ctx.context.(interface{ SetValue(key any, val any) })
	if ok {
		base.SetValue(key, val)
		return
	}
	ctx.context = context.WithValue(ctx.context, key, val)
}

func (ctx *contextBase) SetHandlers(index int, handlers []HandlerFunc) {
	ctx.index, ctx.handlers = index, handlers
}

func (ctx *contextBase) GetHandlers() (int, []HandlerFunc) {
	return ctx.index, ctx.handlers
}

func (ctx *contextBase) Next() {
	ctx.index++
	for ctx.index < len(ctx.handlers) {
		ctx.handlers[ctx.index](ctx)
		ctx.index++
	}
}

func (ctx *contextBase) End() {
	ctx.index = DefaultContextMaxHandler
}

func (ctx *contextBase) Err() error {
	return ctx.context.Err()
}

func (ctx *contextBase) Read(b []byte) (int, error) {
	return ctx.RequestReader.Body.Read(b)
}

func (ctx *contextBase) Host() string {
	return ctx.RequestReader.Host
}

func (ctx *contextBase) Method() string {
	return ctx.RequestReader.Method
}

func (ctx *contextBase) Path() string {
	return ctx.RequestReader.URL.Path
}

func (ctx *contextBase) RealIP() string {
	if val := ctx.RequestReader.Header.Get(HeaderXRealIP); val != "" {
		return val
	}
	if val := ctx.RequestReader.Header.Get(HeaderXForwardedFor); val != "" {
		return strings.SplitN(val, ",", 2)[0]
	}

	addr := ctx.RequestReader.RemoteAddr
	if addr == "pipe" {
		return "127.0.0.1"
	}
	pos := strings.LastIndexByte(addr, ':')
	if pos != -1 {
		addr = addr[:pos]
		// ipv6
		if len(addr) > 1 && addr[0] == '[' {
			addr = addr[1 : len(addr)-1]
		}
	}
	return addr
}

var noneSliceByte = make([]byte, 0)

// Body returns the request body and saves it to the cache.
// The Body method can be called repeatedly.
// Each call will reset the ctx.Request().Body object to a body reader.
//
// ctx.bodyContent will not be reused with memory.
// If reuse is required, modify app.ContextPool to use the [Context] that
// implements Body and Reset method.
//
// Normally, you should avoid calling the Body method;
// If you use it, you should set [middleware.NewBodyLimitFunc] to
// avoid large bodies occupying memory.
func (ctx *contextBase) Body() ([]byte, error) {
	r := ctx.RequestReader
	if r.ContentLength == 0 {
		return nil, nil
	}
	if ctx.bodyContent == nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			ctx.bodyContent = noneSliceByte
			ctx.loggerDebug("Context.Body", err)
			return nil, err
		}
		if r.ContentLength == -1 {
			r.ContentLength = int64(len(body))
		}
		ctx.bodyContent = body
	}
	r.Body = io.NopCloser(bytes.NewReader(ctx.bodyContent))
	return ctx.bodyContent, nil
}

func (ctx *contextBase) Bind(i any) error {
	err := ctx.config.Bind(ctx, i)
	if err != nil {
		ctx.loggerDebug("Context.Bind", err)
		return err
	}
	return nil
}

func (ctx *contextBase) Params() *Params {
	return &ctx.params
}

func (ctx *contextBase) GetParam(key string) string {
	return ctx.params.Get(key)
}

func (ctx *contextBase) SetParam(key, val string) {
	ctx.params = ctx.params.Set(key, val)
}

func (ctx *contextBase) Querys() (url.Values, error) {
	r := ctx.RequestReader
	if r.Form == nil {
		var err error
		r.Form, err = url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			ctx.loggerDebug("Context.Querys", err)
			return nil, err
		}
	}
	return r.Form, nil
}

func (ctx *contextBase) GetQuery(key string) string {
	r := ctx.RequestReader
	if r.Form == nil {
		var err error
		r.Form, err = url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			ctx.loggerDebug("Context.Querys", err)
			return ""
		}
	}
	return r.Form.Get(key)
}

func (ctx *contextBase) GetHeader(key string) string {
	return ctx.RequestReader.Header.Get(key)
}

func (ctx *contextBase) SetHeader(key string, val string) {
	ctx.ResponseWriter.Header().Set(key, val)
}

// Cookies gets all cookies from [HeaderCookie] and
// parses the data after the first call to the [Cookies]/[GetCookie] method.
func (ctx *contextBase) Cookies() []Cookie {
	ctx.parseCookies()
	return ctx.cookies
}

func (ctx *contextBase) GetCookie(key string) string {
	ctx.parseCookies()
	for _, cookie := range ctx.cookies {
		if cookie.Name == key {
			return cookie.Value
		}
	}
	return ""
}

// SetCookie sets [HeaderSetCookie], allowing the cookie option to be set.
func (ctx *contextBase) SetCookie(cookie *CookieSet) {
	if v := cookie.String(); v != "" {
		ctx.ResponseWriter.Header().Add(HeaderSetCookie, v)
	}
}

// The SetCookieValue method sets [HeaderSetCookie] and
// sets the Max-Age attribute if age is non-zero.
func (ctx *contextBase) SetCookieValue(name, value string, age int) {
	ctx.SetCookie(&CookieSet{
		Name:   name,
		Value:  value,
		MaxAge: age,
	})
}

// FormValue uses body to parse Form data and returns the value of the key.
func (ctx *contextBase) FormValue(key string) string {
	r := ctx.RequestReader
	if r.PostForm == nil {
		err := ctx.parseForm(r)
		if err != nil {
			r.PostForm = make(url.Values)
			ctx.loggerDebug("Context.FormValue", err)
			return ""
		}
	}

	val, ok := r.PostForm[key]
	if ok && len(val) != 0 {
		return val[0]
	}
	return ""
}

// FormValues uses body to parse Form data and returns all values.
func (ctx *contextBase) FormValues() (map[string][]string, error) {
	r := ctx.RequestReader
	if r.PostForm == nil {
		err := ctx.parseForm(r)
		if err != nil {
			r.PostForm = make(url.Values)
			ctx.loggerDebug("Context.FormValues", err)
			return nil, err
		}
	}
	return r.PostForm, nil
}

// FormFile uses body to parse Form data and returns the file corresponding to
// the key.
func (ctx *contextBase) FormFile(key string) *multipart.FileHeader {
	r := ctx.RequestReader
	if r.PostForm == nil {
		err := ctx.parseForm(r)
		if err != nil {
			r.PostForm = make(url.Values)
			ctx.loggerDebug("Context.FormFile", err)
			return nil
		}
	}

	if r.MultipartForm != nil {
		val, ok := r.MultipartForm.File[key]
		if ok && len(val) != 0 {
			return val[0]
		}
	}
	return nil
}

// FormFiles uses body to parse into Form data and returns all files.
func (ctx *contextBase) FormFiles() map[string][]*multipart.FileHeader {
	r := ctx.RequestReader
	if r.PostForm == nil {
		err := ctx.parseForm(r)
		if err != nil {
			r.PostForm = make(url.Values)
			ctx.loggerDebug("Context.FormFiles", err)
			return nil
		}
	}

	if r.MultipartForm != nil {
		return r.MultipartForm.File
	}
	return nil
}

// Write implements [io.Writer] and writes data to the response.
func (ctx *contextBase) Write(b []byte) (n int, err error) {
	ctx.writeStatus()
	n, err = ctx.ResponseWriter.Write(b)
	if err != nil {
		ctx.internalError("Context.Write", err)
	}
	return
}

// WriteString implements [io.StringWriter] and writes a string to the response.
func (ctx *contextBase) WriteString(s string) (n int, err error) {
	ctx.writeStatus()
	n, err = ctx.ResponseWriter.WriteString(s)
	if err != nil {
		ctx.internalError("Context.WriteString", err)
	}
	return
}

func (ctx *contextBase) writeStatus() {
	if ctx.wantStatus > 0 {
		ctx.ResponseWriter.WriteHeader(ctx.wantStatus)
		ctx.wantStatus = -ctx.wantStatus
	}
}

// The WriteStatus method set the response status code.
func (ctx *contextBase) WriteStatus(code int) {
	if ctx.wantStatus > 0 {
		ctx.wantStatus = code
		ctx.ResponseWriter.WriteStatus(code)
	}
}

func (ctx *contextBase) WriteHeader(code int) {
	ctx.ResponseWriter.WriteHeader(code)
}

// WriteFile 使用HandlerFile处理一个静态文件。
func (ctx *contextBase) WriteFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, _ := file.Stat()
	http.ServeContent(ctx.ResponseWriter, ctx.RequestReader,
		stat.Name(), stat.ModTime(), file,
	)
	return nil
}

// Redirect implements request redirection.
// The status code needs to be 30x or 201.
func (ctx *contextBase) Redirect(code int, u string) error {
	_, err := url.Parse(u)
	if err != nil {
		return err
	}
	if (code < http.StatusMultipleChoices ||
		code > http.StatusPermanentRedirect) && code != StatusCreated {
		err = fmt.Errorf(ErrContextRedirectInvalid, code)
		ctx.internalError("Context.Redirect", err)
		return err
	}
	http.Redirect(ctx.ResponseWriter, ctx.RequestReader, u, code)
	return nil
}

// Render uses Render to return data.
func (ctx *contextBase) Render(data any) error {
	err := ctx.config.Render(ctx, data)
	if err != nil {
		ctx.internalError("Context.Render", err)
		return err
	}
	return nil
}

// The writeFatal method returns error data.
//
// If the response is not written, return error; and end request processing.
//
// This method should not be used directly.
// Calling the ctx.Fatal method will automatically call the writeFatal method.
func (ctx *contextBase) writeFatal(err error) {
	w := ctx.ResponseWriter
	if w.Size() == 0 {
		msg := NewContextMessgae(ctx, err, nil)
		status := w.Status()
		if status == StatusOK {
			ctx.WriteStatus(getErrorStatus(err))
		}
		_ = ctx.Render(msg)
	}
	base, ok := ctx.context.Value(&baseCtxKey).(*contextBaseValue)
	if ok {
		base.SetValue(ContextKeyError, err)
	}
	// stop Context
	ctx.End()
}

func (ctx *contextBase) logger() Logger {
	log, ok := ctx.context.Value(ContextKeyLogger).(Logger)
	if ok {
		return log
	}
	return ctx.config.Logger
}

// The loggerDebug method outputs the error caused by the client data.
func (ctx *contextBase) loggerDebug(call string, err error) {
	log := ctx.logger()
	if log.GetLevel() > LoggerDebug {
		return
	}
	log.WithField(ParamDepth, 2).
		WithField(ParamCaller, call).
		Error(err)
}

// The internalError method outputs the error caused by the server data.
func (ctx *contextBase) internalError(call string, err error) {
	ctx.logger().WithField(ParamDepth, 2).
		WithField(ParamCaller, call).
		Error(err)
}
