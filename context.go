package eudore

// Context定义一个请求上下文

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

/*
Context 定义请求上下文接口，分为请求上下文数据、请求、参数、响应、日志输出五部分。
	context.Context、eudore.ResponseWriter、*http.Request、eudore.Logger对象读写
	中间件机制执行
	基本请求信息
	数据Bind和Validate
	重复读取请求body
	param、query、header、cookie、form读写
	状态码、header、重定向、push、body写入
	数据写入Render
	5级日志格带fields格式化属性
*/
type Context interface {
	// context
	Reset(http.ResponseWriter, *http.Request)
	GetContext() context.Context
	Request() *http.Request
	Response() ResponseWriter
	Value(interface{}) interface{}
	SetContext(context.Context)
	SetRequest(*http.Request)
	SetResponse(ResponseWriter)
	SetValue(interface{}, interface{})
	// handles
	SetHandler(int, HandlerFuncs)
	GetHandler() (int, HandlerFuncs)
	Next()
	End()
	Err() error

	// request info
	Read([]byte) (int, error)
	Host() string
	Method() string
	Path() string
	RealIP() string
	RequestID() string
	ContentType() string
	Istls() bool
	Body() []byte
	Bind(interface{}) error

	// param query header cookie form
	Params() *Params
	GetParam(string) string
	SetParam(string, string)
	Querys() url.Values
	GetQuery(string) string
	GetHeader(string) string
	SetHeader(string, string)
	Cookies() []Cookie
	GetCookie(string) string
	SetCookie(*SetCookie)
	SetCookieValue(string, string, int)
	FormValue(string) string
	FormValues() map[string][]string
	FormFile(string) *multipart.FileHeader
	FormFiles() map[string][]*multipart.FileHeader

	// response
	Write([]byte) (int, error)
	WriteHeader(int)
	Redirect(int, string)
	Push(string, *http.PushOptions) error
	Render(interface{}) error
	WriteString(string) error
	WriteFile(string) error

	// Database interface
	Query(interface{}, DatabaseStmt) error
	Exec(DatabaseStmt) error
	NewRequest(string, string, ...interface{}) error
	// Logger interface
	Debug(...interface{})
	Info(...interface{})
	Warning(...interface{})
	Error(...interface{})
	Fatal(...interface{})
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warningf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
	WithField(string, interface{}) Logger
	WithFields([]string, []interface{}) Logger
}

// contextBase 实现Context接口。
type contextBase struct {
	// context
	index          int
	handler        HandlerFuncs
	httpParams     Params
	config         *contextBaseConfig
	RequestReader  *http.Request
	ResponseWriter ResponseWriter
	context        context.Context
	// data
	contextValues contextBaseValue
	httpResponse  responseWriterHTTP
	cookies       []Cookie
	bodyContent   []byte
}

type contextBaseConfig struct {
	Logger   Logger
	Database Database
	Client   Client
	Bind     func(Context, interface{}) error
	Validate func(Context, interface{}) error
	Filte    func(Context, interface{}) error
	Render   func(Context, interface{}) error
}

type contextBaseValue struct {
	context.Context
	Logger   Logger
	Database Database
	Client   Client
	Error    error
	Values   []interface{}
}

// ResponseWriter 接口用于写入http请求响应体status、header、body。
//
// net/http.response实现了flusher、hijacker、pusher接口。
type ResponseWriter interface {
	// http.ResponseWriter
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)
	// http.Flusher
	Flush()
	// http.Hijacker
	Hijack() (net.Conn, *bufio.ReadWriter, error)
	// http.Pusher
	Push(string, *http.PushOptions) error
	Size() int
	Status() int
}

// responseWriterHTTP 是对net/http.ResponseWriter接口封装
type responseWriterHTTP struct {
	http.ResponseWriter
	code int
	size int
}

// SetCookie 定义响应返回的set-cookie header的数据生成
type SetCookie = http.Cookie

// Cookie 定义请求读取的cookie header的键值对数据存储
type Cookie struct {
	Name  string
	Value string
}

// contextBaseEntry 实现ContextBase使用的Logger对象。
type contextBaseEntry struct {
	Logger
	Context *contextBase
}

// NewContextBasePool 函数从上下文创建一个Context sync.Pool。
//
// 需要从上下文加载ContextKeyApp实现Logger Database Client接口。
//
// ContextBase相关方法文档点击NewContextBase函数跳转到源码查看。
func NewContextBasePool(ctx context.Context) *sync.Pool {
	config := newContextBaseConfig(ctx)
	return &sync.Pool{
		New: func() interface{} {
			return &contextBase{
				config:     config,
				httpParams: Params{ParamRoute, ""},
			}
		},
	}
}

// NewContextBaseFunc 函数使用context.Context创建Context构造函数，用于获取自定义Context。
func NewContextBaseFunc(ctx context.Context) func() Context {
	config := newContextBaseConfig(ctx)
	return func() Context {
		return &contextBase{
			config:     config,
			httpParams: Params{ParamRoute, ""},
		}
	}
}

func newContextBaseConfig(ctx context.Context) *contextBaseConfig {
	bind, _ := ctx.Value(ContextKeyBind).(func(Context, interface{}) error)
	validate, _ := ctx.Value(ContextKeyValidate).(func(Context, interface{}) error)
	filte, _ := ctx.Value(ContextKeyFilte).(func(Context, interface{}) error)
	render, _ := ctx.Value(ContextKeyRender).(func(Context, interface{}) error)
	if bind == nil {
		bind = NewBinds(nil)
	}
	if render == nil {
		render = NewRenders(nil)
	}
	return &contextBaseConfig{
		Logger:   ctx.Value(ContextKeyApp).(Logger).WithField("logger", true),
		Database: ctx.Value(ContextKeyApp).(Database),
		Client:   ctx.Value(ContextKeyApp).(Client),
		Bind:     bind,
		Validate: validate,
		Filte:    filte,
		Render:   render,
	}
}

// Reset Context
func (ctx *contextBase) Reset(w http.ResponseWriter, r *http.Request) {
	ctx.context = &ctx.contextValues
	ctx.ResponseWriter = &ctx.httpResponse
	ctx.RequestReader = r
	ctx.httpParams = ctx.httpParams[0:2]
	ctx.httpParams[1] = ""
	// cookies body
	ctx.contextValues.Reset(r.Context(), ctx.config)
	ctx.httpResponse.Reset(w)
	ctx.cookies = ctx.cookies[0:0]
	ctx.bodyContent = nil
}

// GetContext 获取当前请求的上下文,返回RequestReader的context.Context对象。
//
// 该函数名称如果为Context，会在Context对象组合时出现冲突。
func (ctx *contextBase) GetContext() context.Context {
	return ctx.context
}

// Request 获取请求对象。
// 注意：ctx.Request().Context() 不等于ctx.GetContext()
func (ctx *contextBase) Request() *http.Request {
	return ctx.RequestReader
}

// Response 获得响应对象。
func (ctx *contextBase) Response() ResponseWriter {
	return ctx.ResponseWriter
}

func (ctx *contextBase) Value(key interface{}) interface{} {
	return ctx.contextValues.Value(key)
}

func (ctx *contextBase) SetContext(c context.Context) {
	ctx.context = c
}

// SetRequest 设置请求对象。
func (ctx *contextBase) SetRequest(r *http.Request) {
	ctx.RequestReader = r
}

// SetResponse 设置响应对象。
func (ctx *contextBase) SetResponse(w ResponseWriter) {
	ctx.ResponseWriter = w
}

// SetLogger 方法设置ContextBases输出日志的基础Logger。
//
// 注意确保设置的是Logger，而不是一个Entry。
func (ctx *contextBase) SetValue(key, val interface{}) {
	ctx.contextValues.SetValue(key, val)
}

// SetHandler 方法设置请求上下文的全部请求处理者。
func (ctx *contextBase) SetHandler(index int, hs HandlerFuncs) {
	ctx.index, ctx.handler = index, hs
}

// GetHandler 方法获取请求上下文的当前处理索引和全部请求处理者。
func (ctx *contextBase) GetHandler() (int, HandlerFuncs) {
	return ctx.index, ctx.handler
}

// Next 方法调用请求上下文下一个处理函数。
func (ctx *contextBase) Next() {
	ctx.index++
	for ctx.index < len(ctx.handler) {
		ctx.handler[ctx.index](ctx)
		ctx.index++
	}
}

// End 结束请求上下文的处理。
func (ctx *contextBase) End() {
	ctx.index = DefaultContextMaxHandler
	ctx.httpResponse.writeStatus()
}

// Err 方法返回
func (ctx *contextBase) Err() error {
	return ctx.contextValues.Err()
}

// Read 方法实现io.Reader读取http请求。
func (ctx *contextBase) Read(b []byte) (int, error) {
	return ctx.RequestReader.Body.Read(b)
}

// Host 方法返回请求Host。
func (ctx *contextBase) Host() string {
	return ctx.RequestReader.Host
}

// Method 方法返回请求方法，
func (ctx *contextBase) Method() string {
	return ctx.RequestReader.Method
}

// Path 方法返回请求路径。
func (ctx *contextBase) Path() string {
	return ctx.RequestReader.URL.Path
}

// RealIP 获取用户真实ip，ctx.Request().RemoteAddr()获取远程连接地址。
func (ctx *contextBase) RealIP() string {
	if realip := ctx.RequestReader.Header.Get(HeaderXRealIP); realip != "" {
		return realip
	}
	if xforward := ctx.RequestReader.Header.Get(HeaderXForwardedFor); xforward != "" {
		return strings.SplitN(string(xforward), ",", 2)[0]
	}
	addr := strings.SplitN(ctx.RequestReader.RemoteAddr, ":", 2)[0]
	if addr == "pipe" {
		return strings.SplitN(DefaultClientInternalHost, ":", 2)[0]
	}
	return addr
}

// RequestID 获取响应中的X-Request-Id Header
func (ctx *contextBase) RequestID() string {
	return ctx.GetHeader(HeaderXRequestID)
}

// ContentType 获取请求内容类型，返回Content-Type Header
func (ctx *contextBase) ContentType() string {
	return ctx.GetHeader(HeaderContentType)
}

// Istls 判断是否使用了tls，tls状态使用ctx.Request().TLS()获取。
func (ctx *contextBase) Istls() bool {
	return ctx.RequestReader.TLS != nil
}

var noneSliceByte = make([]byte, 0, 0)

// Body 返回请求的body，并保存到缓存中，可重复调用Body方法,
// 每次调用会重置ctx.Request().Body对象成一个body reader。
//
// ctx.bodyContent 不会随contextBase一起内存复用，正常应该避免调用Body方法；
// 如果使用应该设置middleware.NewBodyLimitFunc。
func (ctx *contextBase) Body() []byte {
	if ctx.bodyContent == nil {
		bts, err := ioutil.ReadAll(ctx.RequestReader.Body)
		if err != nil {
			ctx.bodyContent = noneSliceByte
			ctx.contextValues.Logger.WithField("depth", 1).WithField(ParamCaller, "Context.Body").Error(err)
			return nil
		}
		ctx.bodyContent = bts
	}
	ctx.RequestReader.Body = ioutil.NopCloser(bytes.NewReader(ctx.bodyContent))
	return ctx.bodyContent
}

// Bind 使用Bind解析请求body并绑定数据。
// 如果Validate不为空，则使用Validate校验数据。
func (ctx *contextBase) Bind(i interface{}) error {
	err := ctx.config.Bind(ctx, i)
	if err != nil {
		ctx.contextValues.Logger.WithField("depth", 1).WithField(ParamCaller, "Context.Bind").Error(err)
		return NewErrorStatusCode(err, StatucBindFail, CodeBindFail)
	}
	if ctx.config.Validate != nil {
		err = ctx.config.Validate(ctx, i)
		if err != nil {
			ctx.contextValues.Logger.WithField("depth", 1).WithField(ParamCaller, "Context.Bind(Validate)").Error(err)
			return NewErrorStatusCode(err, StatucValidateFail, CodeValidateFail)
		}
	}
	return nil
}

// Params 获得请求的全部参数。
func (ctx *contextBase) Params() *Params {
	return &ctx.httpParams
}

// GetParam 方法获取一个参数的值。
func (ctx *contextBase) GetParam(key string) string {
	return ctx.httpParams.Get(key)
}

// SetParam 方法设置一个参数。
func (ctx *contextBase) SetParam(key, val string) {
	ctx.httpParams = ctx.httpParams.Set(key, val)
}

// Querys 方法返回http请求的全部uri参数。
func (ctx *contextBase) Querys() url.Values {
	err := ctx.RequestReader.ParseForm()
	if err != nil {
		ctx.contextValues.Logger.WithField("depth", 1).WithField(ParamCaller, "Context.Querys").Error(err)
	}
	return ctx.RequestReader.Form
}

// GetQuery 方法获得一个uri参数的值。
func (ctx *contextBase) GetQuery(key string) string {
	err := ctx.RequestReader.ParseForm()
	if err != nil {
		ctx.contextValues.Logger.WithField("depth", 1).WithField(ParamCaller, "Context.GetQuery").Error(err)
		return ""
	}
	return ctx.RequestReader.Form.Get(key)
}

// GetHeader 方法获取一个请求header，相当于ctx.Request().Header().Get(name)。
func (ctx *contextBase) GetHeader(name string) string {
	return ctx.RequestReader.Header.Get(name)
}

// SetHeader 方法设置一个响应header，相当于ctx.Response().Header().Set(name, val)
func (ctx *contextBase) SetHeader(name string, val string) {
	ctx.ResponseWriter.Header().Set(name, val)
}

// Cookies 方法获取全部请求的cookie,获取的cookie值是首次调用Cookies/GetCookie方法后解析的数据。。
func (ctx *contextBase) Cookies() []Cookie {
	ctx.readCookies()
	return ctx.cookies
}

// GetCookie 获方法得一个请求cookie的值,获取的cookie值是首次调用Cookies/GetCookie方法后解析的数据。。
func (ctx *contextBase) GetCookie(name string) string {
	ctx.readCookies()
	for _, cookie := range ctx.cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

// SetCookie 方法设置一个响应cookie的数据，设置响应 Set-Cookie header。
func (ctx *contextBase) SetCookie(cookie *SetCookie) {
	if v := cookie.String(); v != "" {
		ctx.ResponseWriter.Header().Add(HeaderSetCookie, v)
	}
}

// SetCookieValue 方法设置一个响应cookie，如果maxAge非0则设置Max-Age属性。
func (ctx *contextBase) SetCookieValue(name, value string, maxAge int) {
	if maxAge != 0 {
		ctx.ResponseWriter.Header().Add(HeaderSetCookie, fmt.Sprintf("%s=%s; Max-Age=%d", name, url.QueryEscape(value), maxAge))
	} else {
		ctx.ResponseWriter.Header().Add(HeaderSetCookie, fmt.Sprintf("%s=%s;", name, url.QueryEscape(value)))
	}
}

// FormValue 使用body解析成Form数据，并返回对应key的值
func (ctx *contextBase) FormValue(key string) string {
	if ctx.parseForm() != nil {
		return ""
	}
	val, ok := ctx.RequestReader.MultipartForm.Value[key]
	if ok && len(val) != 0 {
		return val[0]
	}
	return ""
}

// FormValues 使用body解析成Form数据，并返回全部的值
func (ctx *contextBase) FormValues() map[string][]string {
	if ctx.parseForm() != nil {
		return nil
	}
	return ctx.RequestReader.MultipartForm.Value
}

// FormFile 使用body解析成Form数据，并返回对应key的文件
func (ctx *contextBase) FormFile(key string) *multipart.FileHeader {
	if ctx.parseForm() != nil {
		return nil
	}
	val, ok := ctx.RequestReader.MultipartForm.File[key]
	if ok && len(val) != 0 {
		return val[0]
	}
	return nil
}

// FormFiles 使用body解析成Form数据，并返回全部的文件。
func (ctx *contextBase) FormFiles() map[string][]*multipart.FileHeader {
	if ctx.parseForm() != nil {
		return nil
	}
	return ctx.RequestReader.MultipartForm.File
}

// parseForm 解析form数据。
func (ctx *contextBase) parseForm() error {
	err := ctx.RequestReader.ParseMultipartForm(DefaultContextFormMaxMemory)
	if err != nil && err.Error() != "http: multipart handled by MultipartReader" {
		ctx.contextValues.Logger.WithField("depth", 2).WithField(ParamCaller, "Context.Form...").Error(err)
		return err
	}
	return nil
}

// WriteHeader 方法写入响应状态码。
func (ctx *contextBase) WriteHeader(code int) {
	ctx.ResponseWriter.WriteHeader(code)
}

// Redirect implement request redirection.
//
// Redirect 实现请求重定向。
func (ctx *contextBase) Redirect(code int, url string) {
	http.Redirect(ctx.ResponseWriter, ctx.RequestReader, url, code)
}

// Push 实现http2 push
func (ctx *contextBase) Push(target string, opts *http.PushOptions) error {
	if opts == nil {
		opts = &http.PushOptions{
			Header: http.Header{
				HeaderAcceptEncoding: []string{ctx.RequestReader.Header.Get(HeaderAcceptEncoding)},
			},
		}
	}

	err := ctx.ResponseWriter.Push(target, opts)
	if err != nil {
		ctx.contextValues.Logger.WithField("depth", 1).WithField(ParamCaller, "Context.Push").
			Errorf("Failed to push: %v, Resource path: %s.", err, target)
	}
	return err
}

// Render 使用app.Renderer返回数据。
func (ctx *contextBase) Render(i interface{}) error {
	var err error
	if ctx.config.Filte != nil {
		err = ctx.config.Filte(ctx, i)
		if err != nil {
			ctx.contextValues.Logger.WithField("depth", 1).WithField(ParamCaller, "Context.Render(Filte)").Error(err)
			return NewErrorStatusCode(err, StatucFilteFail, CodeFilteFail)
		}
	}

	err = ctx.config.Render(ctx, i)
	if err != nil {
		ctx.contextValues.Logger.WithField("depth", 1).WithField(ParamCaller, "Context.Render").Error(err)
	}
	return NewErrorStatusCode(err, StatucRenderFail, CodeRenderFail)
}

// Write 实现io.Writer，向响应写入数据。
func (ctx *contextBase) Write(data []byte) (n int, err error) {
	n, err = ctx.ResponseWriter.Write(data)
	if err != nil {
		ctx.contextValues.Logger.WithField("depth", 1).WithField(ParamCaller, "Context.Write").Error(err)
	}
	return
}

// WriteString 实现向响应写入一个字符串。
func (ctx *contextBase) WriteString(i string) (err error) {
	header := ctx.ResponseWriter.Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, MimeTextPlainCharsetUtf8)
	}
	_, err = ctx.ResponseWriter.Write([]byte(i))
	if err != nil {
		ctx.contextValues.Logger.WithField("depth", 1).WithField(ParamCaller, "Context.WriteString").Error(err)
	}
	return
}

// WriteFile 使用HandlerFile处理一个静态文件。
func (ctx *contextBase) WriteFile(path string) error {
	http.ServeFile(ctx.ResponseWriter, ctx.RequestReader, path)
	return nil
}

// writeError 方法返回error数据，该方法不应该被直接使用，调用ctx.Fatal方法会自动调用writeError方法。
// 定义次方法用于重写error响应,如果error实现Code() int方法会获取错误响应码。
func (ctx *contextBase) writeError(err error) {
	// 结束Context
	w := ctx.ResponseWriter
	if w.Size() == 0 {
		status := w.Status()
		if status == 200 {
			ctx.WriteHeader(getErrorStatus(err))
		}
		ctx.Render(NewContextMessgae(ctx, err, nil))
	}
	ctx.contextValues.Error = err
	ctx.End()
}

// NewContextMessgae 方法从请求上下文创建一个error或对象响应对象，记录请求上下文相关信息。
func NewContextMessgae(ctx Context, err error, message interface{}) interface{} {
	type contextMessage struct {
		Time       string      `json:"time" protobuf:"1,name=time" xml:"time" yaml:"time"`
		Host       string      `json:"host" protobuf:"2,name=host" xml:"host" yaml:"host"`
		Method     string      `json:"method" protobuf:"3,name=method" xml:"method" yaml:"method"`
		Path       string      `json:"path" protobuf:"4,name=path" xml:"path" yaml:"path"`
		Route      string      `json:"route" protobuf:"5,name=route" xml:"route" yaml:"route"`
		Status     int         `json:"status" protobuf:"6,name=status" xml:"status" yaml:"status"`
		Code       int         `json:"code,omitempty" protobuf:"7,name=code" xml:"code,omitempty" yaml:"code,omitempty"`
		XRequestID string      `json:"x-request-id,omitempty" protobuf:"8,name=x-request-id" xml:"x-request-id,omitempty" yaml:"x-request-id,omitempty"`
		XTraceID   string      `json:"x-trace-id,omitempty" protobuf:"9,name=x-trace-id" xml:"x-trace-id,omitempty" yaml:"x-trace-id,omitempty"`
		Error      string      `json:"error,omitempty" protobuf:"10,name=error" xml:"error,omitempty" yaml:"error,omitempty"`
		Message    interface{} `json:"message,omitempty" protobuf:"11,name=message" xml:"message,omitempty" yaml:"message,omitempty"`
	}
	msg := contextMessage{
		Time:       time.Now().Format(DefaultLoggerTimeFormat),
		Host:       ctx.Host(),
		Method:     ctx.Method(),
		Path:       ctx.Path(),
		Route:      ctx.GetParam(ParamRoute),
		XRequestID: ctx.Response().Header().Get(HeaderXRequestID),
		XTraceID:   ctx.Response().Header().Get(HeaderXTraceID),
		Status:     ctx.Response().Status(),
		Message:    message,
	}
	if err != nil {
		msg.Code = getErrorCode(err)
		msg.Error = err.Error()
	}
	return msg
}

func getErrorStatus(err error) int {
	for {
		stater, ok := err.(interface{ Status() int })
		if ok {
			return stater.Status()
		}
		u, ok := err.(interface {
			Unwrap() error
		})
		if ok {
			err = u.Unwrap()
		} else {
			return 500
		}
	}
}

func getErrorCode(err error) int {
	for {
		coder, ok := err.(interface{ Code() int })
		if ok {
			return coder.Code()
		}
		u, ok := err.(interface {
			Unwrap() error
		})
		if ok {
			err = u.Unwrap()
		} else {
			return 0
		}
	}
}

// Query 方法调用Database.Query查询数据块。
func (ctx *contextBase) Query(data interface{}, stmt DatabaseStmt) error {
	err := ctx.contextValues.Database.Query(ctx.context, data, stmt)
	if err != nil {
		ctx.contextValues.Logger.WithField("depth", 1).WithField(ParamCaller, "Context.Query").Error(err)
	}
	return err
}

// Exec 方法调用Database.Exec执行数据块。
func (ctx *contextBase) Exec(stmt DatabaseStmt) error {
	err := ctx.contextValues.Database.Exec(ctx.context, stmt)
	if err != nil {
		ctx.contextValues.Logger.WithField("depth", 1).WithField(ParamCaller, "Context.Exec").Error(err)
	}
	return err
}

func (ctx *contextBase) NewRequest(method, path string, options ...interface{}) error {
	return ctx.contextValues.Client.NewRequest(ctx.context, method, path, options...)
}

// Debug 方法写入Debug日志。
func (ctx *contextBase) Debug(args ...interface{}) {
	ctx.contextValues.Logger.WithField("depth", 1).Debug(args...)
}

// Info 方法写入Info日志。
func (ctx *contextBase) Info(args ...interface{}) {
	ctx.contextValues.Logger.WithField("depth", 1).Info(args...)
}

// Warning 方法写入Warning日志。
func (ctx *contextBase) Warning(args ...interface{}) {
	ctx.contextValues.Logger.WithField("depth", 1).Warning(args...)
}

// Error 方法写入Error日志。
func (ctx *contextBase) Error(args ...interface{}) {
	ctx.contextValues.Logger.WithField("depth", 1).Error(args...)
}

// Fatal 方法写入Error日志，并结束请求上下文处理。
//
// 注意：如果err中存在敏感信息会被写入到响应中。
func (ctx *contextBase) Fatal(args ...interface{}) {
	err := getMessagError(args)
	ctx.writeError(err)
	ctx.contextValues.Logger.WithField("depth", 1).Error(err.Error())
}

func getMessagError(args []interface{}) error {
	if len(args) == 1 {
		err, ok := args[0].(error)
		if ok {
			return err
		}
	}
	msg := fmt.Sprintln(args...)
	msg = msg[:len(msg)-1]
	return errors.New(msg)
}

// Debugf 方法输出Info日志。
func (ctx *contextBase) Debugf(format string, args ...interface{}) {
	ctx.contextValues.Logger.WithField("depth", 1).Debug(fmt.Sprintf(format, args...))
}

// Infof 方法输出Info日志。
func (ctx *contextBase) Infof(format string, args ...interface{}) {
	ctx.contextValues.Logger.WithField("depth", 1).Info(fmt.Sprintf(format, args...))
}

// Warningf 方法输出Warning日志。
func (ctx *contextBase) Warningf(format string, args ...interface{}) {
	ctx.contextValues.Logger.WithField("depth", 1).Warning(fmt.Sprintf(format, args...))
}

// Errorf 方法输出Error日志。
func (ctx *contextBase) Errorf(format string, args ...interface{}) {
	ctx.contextValues.Logger.WithField("depth", 1).Error(fmt.Sprintf(format, args...))
}

// Fatalf 方法输出Fatal日志，并结束请求上下文处理。
//
// 注意：如果err中存在敏感信息会被写入到响应中。
func (ctx *contextBase) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	ctx.writeError(errors.New(msg))
	ctx.contextValues.Logger.WithField("depth", 1).Errorf(msg)
}

// WithField 方法增加一个日志属性，返回一个新的Logger。
func (ctx *contextBase) WithField(key string, value interface{}) Logger {
	return &contextBaseEntry{
		Logger:  ctx.contextValues.Logger.WithField(key, value),
		Context: ctx,
	}
}

// WithFields 方法增加多个日志属性，返回一个新的Logger。
//
// 如果fields包含file条目属性，则不会添加调用位置信息。
func (ctx *contextBase) WithFields(keys []string, fields []interface{}) Logger {
	return &contextBaseEntry{
		Logger:  ctx.contextValues.Logger.WithFields(keys, fields),
		Context: ctx,
	}
}

// Fatal 方法重写Context的Fatal方法，不执行panic，http返回500和请求id。
func (e *contextBaseEntry) Fatal(args ...interface{}) {
	err := getMessagError(args)
	e.Context.writeError(err)
	e.Logger.WithField("depth", 1).Error(err.Error())
}

// Fatalf 方法重写Context的Fatalf方法，不执行panic，http返回500和请求id。
func (e *contextBaseEntry) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	e.Context.writeError(errors.New(msg))
	e.Logger.WithField("depth", 1).Error(msg)
}

// WithField 方法增加一个日志属性。
func (e *contextBaseEntry) WithField(key string, value interface{}) Logger {
	e.Logger = e.Logger.WithField(key, value)
	return e
}

// WithFields 方法增加多个日志属性。
func (e *contextBaseEntry) WithFields(keys []string, fields []interface{}) Logger {
	e.Logger = e.Logger.WithFields(keys, fields)
	return e
}

// readCookies 方法初始化cookie键值对，form net/http。
func (ctx *contextBase) readCookies() {
	if len(ctx.cookies) > 0 {
		return
	}
	for _, line := range ctx.RequestReader.Header[HeaderCookie] {
		parts := strings.Split(line, "; ")
		// Per-line attributes
		for i := 0; i < len(parts); i++ {
			if len(parts[i]) == 0 {
				continue
			}
			name, val := parts[i], ""
			if j := strings.Index(name, "="); j >= 0 {
				name, val = name[:j], name[j+1:]
			}
			if !isCookieNameValid(name) {
				continue
			}
			val, ok := parseCookieValue(val, true)
			if !ok {
				continue
			}
			ctx.cookies = append(ctx.cookies, Cookie{Name: name, Value: val})
		}
	}
}

func parseCookieValue(raw string, allowDoubleQuote bool) (string, bool) {
	// Strip the quotes, if present.
	if allowDoubleQuote && len(raw) > 1 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	for i := 0; i < len(raw); i++ {
		if !validCookieValueByte(raw[i]) {
			return "", false
		}
	}
	return raw, true
}

func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}

func isCookieNameValid(raw string) bool {
	if raw == "" {
		return false
	}
	return strings.IndexFunc(raw, isNotToken) < 0
}

func isNotToken(r rune) bool {
	i := int(r)
	return !(i < len(isTokenTable) && isTokenTable[i])
}

var isTokenTable = [127]bool{
	'!': true, '#': true, '$': true, '%': true, '&': true, '\'': true, '*': true, '+': true,
	'-': true, '.': true, '0': true, '1': true, '2': true, '3': true, '4': true, '5': true,
	'6': true, '7': true, '8': true, '9': true, 'A': true, 'B': true, 'C': true, 'D': true,
	'E': true, 'F': true, 'G': true, 'H': true, 'I': true, 'J': true, 'K': true, 'L': true,
	'M': true, 'N': true, 'O': true, 'P': true, 'Q': true, 'R': true, 'S': true, 'T': true,
	'U': true, 'W': true, 'V': true, 'X': true, 'Y': true, 'Z': true, '^': true, '_': true,
	'`': true, 'a': true, 'b': true, 'c': true, 'd': true, 'e': true, 'f': true, 'g': true,
	'h': true, 'i': true, 'j': true, 'k': true, 'l': true, 'm': true, 'n': true, 'o': true,
	'p': true, 'q': true, 'r': true, 's': true, 't': true, 'u': true, 'v': true, 'w': true,
	'x': true, 'y': true, 'z': true, '|': true, '~': true,
}

func (ctx *contextBaseValue) Reset(c context.Context, config *contextBaseConfig) {
	ctx.Context = c
	ctx.Logger = config.Logger
	ctx.Database = config.Database
	ctx.Client = config.Client
	ctx.Error = nil
	ctx.Values = ctx.Values[0:0]
}
func (ctx *contextBaseValue) SetValue(key, val interface{}) {
	switch key {
	case ContextKeyLogger:
		ctx.Logger = val.(Logger)
	case ContextKeyDatabase:
		ctx.Database = val.(Database)
	case ContextKeyClient:
		ctx.Client = val.(Client)
	default:
		for i := 0; i < len(ctx.Values); i += 2 {
			if ctx.Values[i] == key {
				ctx.Values[i+1] = val
				return
			}
		}
		ctx.Values = append(ctx.Values, key, val)
	}
}

func (ctx *contextBaseValue) Value(key interface{}) interface{} {
	switch key {
	case ContextKeyLogger:
		return ctx.Logger
	case ContextKeyDatabase:
		return ctx.Database
	case ContextKeyClient:
		return ctx.Client
	}
	for i := 0; i < len(ctx.Values); i += 2 {
		if ctx.Values[i] == key {
			return ctx.Values[i+1]
		}
	}
	return ctx.Context.Value(key)
}

func (ctx *contextBaseValue) Err() error {
	if ctx.Error != nil {
		return ctx.Error
	}
	return ctx.Context.Err()
}

func (ctx *contextBaseValue) String() string {
	var meta []string
	for i := 0; i < len(ctx.Values); i += 2 {
		meta = append(meta, fmt.Sprintf("%v=%v", ctx.Values[i], ctx.Values[i+1]))
	}
	if ctx.Error != nil {
		meta = append(meta, fmt.Sprintf("error=%s", ctx.Error.Error()))
	}
	return fmt.Sprintf("%v.WithEudoreContext(%s)", ctx.Context, strings.Join(meta, ", "))
}

// Reset 方法重置responseWriterHTTP对象。
func (w *responseWriterHTTP) Reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.code = http.StatusOK
	w.size = 0
}

// Write 方法实现io.Writer接口。
func (w *responseWriterHTTP) Write(data []byte) (int, error) {
	w.writeStatus()
	n, err := w.ResponseWriter.Write(data)
	w.size = w.size + n
	return n, err
}

// WriteHeader 方法实现写入http请求状态码。
func (w *responseWriterHTTP) WriteHeader(codeCode int) {
	w.code = codeCode
}

func (w *responseWriterHTTP) writeStatus() {
	if w.code > 0 && w.code != 200 {
		w.ResponseWriter.WriteHeader(w.code)
		w.code *= -1
	}
}

// Flush 方法实现刷新缓冲，将缓冲的请求发送给客户端。
func (w *responseWriterHTTP) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack 方法实现劫持http连接,用于websocket连接。
func (w *responseWriterHTTP) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, ErrResponseWriterHTTPNotHijacker
}

// Push 方法实现http Psuh，如果responseWriterHTTP实现http.Push接口，则Push资源。
func (w *responseWriterHTTP) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return nil
}

// Size 方法获得写入的数据长度。
func (w *responseWriterHTTP) Size() int {
	return w.size
}

// Status 方法获得设置的http状态码。
func (w *responseWriterHTTP) Status() int {
	if w.code < 0 {
		return w.code * -1
	}
	return w.code
}
