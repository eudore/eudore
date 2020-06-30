package eudore

// Context定义一个请求上下文

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// Context 定义请求上下文接口。
type Context interface {
	// context
	Reset(context.Context, http.ResponseWriter, *http.Request)
	Request() *http.Request
	Response() ResponseWriter
	WithContext(context.Context)
	SetRequest(*http.Request)
	SetResponse(ResponseWriter)
	SetHandler(int, HandlerFuncs)
	GetContext() context.Context
	GetHandler() (int, HandlerFuncs)
	Next()
	End()
	Done() <-chan struct{}
	Err() error

	// request info
	Read([]byte) (int, error)
	Host() string
	Method() string
	Path() string
	RealIP() string
	RequestID() string
	Referer() string
	ContentType() string
	Istls() bool
	Body() []byte
	Bind(interface{}) error
	BindWith(interface{}, Binder) error
	Validate(interface{}) error

	// param query header cookie session
	Params() *Params
	GetParam(string) string
	SetParam(string, string)
	AddParam(string, string)
	Querys() url.Values
	GetQuery(string) string
	GetHeader(name string) string
	SetHeader(string, string)
	Cookies() []Cookie
	GetCookie(name string) string
	SetCookie(cookie *SetCookie)
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
	RenderWith(interface{}, Renderer) error
	WriteString(string) error
	WriteJSON(interface{}) error
	WriteFile(string) error

	// log Logout interface
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
	WithField(key string, value interface{}) Logout
	WithFields(fields Fields) Logout
	Logger() Logout
}

// ContextBase 实现Context接口。
type ContextBase struct {
	RequestReader  *http.Request
	ResponseWriter ResponseWriter
	httpResponse   ResponseWriterHTTP
	httpParams     Params
	index          int
	depth          int
	handler        HandlerFuncs
	ctx            context.Context
	err            string
	querys         url.Values
	cookies        []Cookie
	isReadBody     bool
	postBody       []byte
	// component
	app  *App
	pool *sync.Pool
}

// entryContextBase 实现ContextBase使用的Logout对象。
type entryContextBase struct {
	Logout
	Context *ContextBase
}

// NewContextBaseFunc 函数创建一个NewContext函数用于获取Context对象。
func NewContextBaseFunc(app *App) func() Context {
	pool := &sync.Pool{}
	return func() Context {
		return NewContextBase(app, pool)
	}
}

// NewContextBase 函数创建ContextBase对象，实现Context接口。
// 依赖app.Logger、app.Binder、app.Validater、app.Render
func NewContextBase(app *App, pool *sync.Pool) *ContextBase {
	return &ContextBase{
		app:  app,
		pool: pool,
	}
}

// Reset Context
func (ctx *ContextBase) Reset(pctx context.Context, w http.ResponseWriter, r *http.Request) {
	ctx.ctx = pctx
	ctx.RequestReader = r
	ctx.httpResponse.Reset(w)
	ctx.ResponseWriter = &ctx.httpResponse
	ctx.err = ""

	// data
	ctx.depth = 0
	ctx.querys = nil
	ctx.httpParams.Keys = ctx.httpParams.Keys[0:0]
	ctx.httpParams.Vals = ctx.httpParams.Vals[0:0]
	// cookies
	ctx.cookies = ctx.cookies[0:0]
	ctx.readCookies(r.Header.Get(HeaderCookie))

	// body
	ctx.isReadBody = false
	ctx.postBody = ctx.postBody[0:0]
}

// GetContext 获取当前请求的上下文,Context的context.Context对象由更高层传递下来，禁止SetContext方法修改。
func (ctx *ContextBase) GetContext() context.Context {
	return ctx.ctx
}

// Request 获取请求对象。
func (ctx *ContextBase) Request() *http.Request {
	return ctx.RequestReader
}

// Response 获得响应对象。
func (ctx *ContextBase) Response() ResponseWriter {
	return ctx.ResponseWriter
}

// WithContext 设置当前请求上下文的ctx，必须是请求上下文的衍生上下文。
//
// ctx.WithContext(context.WithValue("key", ctx.Context()))
func (ctx *ContextBase) WithContext(cctx context.Context) {
	ctx.ctx = cctx
}

// SetRequest 设置请求对象。
func (ctx *ContextBase) SetRequest(r *http.Request) {
	ctx.RequestReader = r
}

// SetResponse 设置响应对象。
func (ctx *ContextBase) SetResponse(w ResponseWriter) {
	ctx.ResponseWriter = w
}

// SetHandler 方法设置请求上下文的全部请求处理者。
func (ctx *ContextBase) SetHandler(index int, hs HandlerFuncs) {
	ctx.index, ctx.handler = index, hs
}

// GetHandler 方法获取请求上下文的当前处理索引和全部请求处理者。
func (ctx *ContextBase) GetHandler() (int, HandlerFuncs) {
	return ctx.index, ctx.handler
}

// Next 方法调用请求上下文下一个处理函数。
func (ctx *ContextBase) Next() {
	ctx.index++
	ctx.depth++
	for ctx.index < len(ctx.handler) {
		ctx.handler[ctx.index](ctx)
		ctx.index++
	}
	ctx.depth--
	if ctx.depth == 0 {
		ctx.pool.Put(ctx)
	}
}

// End 结束请求上下文的处理。
func (ctx *ContextBase) End() {
	ctx.index = 0xff
}

// Done 方法返回判断Context是否完成，在调用End方法时会cancel。
func (ctx *ContextBase) Done() <-chan struct{} {
	return ctx.ctx.Done()
}

// Err 方法返回
func (ctx *ContextBase) Err() error {
	if ctx.err != "" {
		return errors.New(ctx.err)
	}
	return ctx.ctx.Err()
}

// Read 方法实现io.Reader读取http请求。
func (ctx *ContextBase) Read(b []byte) (int, error) {
	return ctx.RequestReader.Body.Read(b)
}

// Host 方法返回请求Host。
func (ctx *ContextBase) Host() string {
	return ctx.RequestReader.Host
}

// Method 方法返回请求方法，
func (ctx *ContextBase) Method() string {
	return ctx.RequestReader.Method
}

// Path 方法返回请求路径。
func (ctx *ContextBase) Path() string {
	return ctx.RequestReader.URL.Path
}

// RealIP 获取用户真实ip，ctx.Request().RemoteAddr()获取远程连接地址。
func (ctx *ContextBase) RealIP() string {
	xforward := ctx.RequestReader.Header.Get(HeaderXForwardedFor)
	if "" == xforward {
		return strings.SplitN(ctx.RequestReader.RemoteAddr, ":", 2)[0]
	}
	return strings.SplitN(string(xforward), ",", 2)[0]
}

// RequestID 获取X-Request-Id Header
func (ctx *ContextBase) RequestID() string {
	return ctx.GetHeader(HeaderXRequestID)
}

// Referer 获取Referer Header
func (ctx *ContextBase) Referer() string {
	return ctx.GetHeader(HeaderReferer)
}

// ContentType 获取请求内容类型，返回Content-Type Header
func (ctx *ContextBase) ContentType() string {
	return ctx.GetHeader(HeaderContentType)
}

// Istls 判断是否使用了tls，tls状态使用ctx.Request().TLS()获取。
func (ctx *ContextBase) Istls() bool {
	return ctx.RequestReader.TLS != nil
}

// Body 返回请求的body，并保存到缓存中，可重复调用Body方法。
func (ctx *ContextBase) Body() []byte {
	if !ctx.isReadBody {
		bts, err := ioutil.ReadAll(ctx.RequestReader.Body)
		if err != nil {
			ctx.logReset(1).WithField(ParamCaller, "Context.Body").Error(err)
			return []byte{}
		}
		ctx.isReadBody = true
		ctx.postBody = bts
	}
	return ctx.postBody
}

// getReader 如果调用过Body方法，返回Body封装的io.Reader可重复获得。
func (ctx *ContextBase) getReader() io.Reader {
	if ctx.isReadBody {
		return bytes.NewReader(ctx.postBody)
	}
	return ctx
}

// BindWith 使用指定Binder解析请求body并绑定数据。
func (ctx *ContextBase) BindWith(i interface{}, r Binder) error {
	return ctx.bind(i, r)
}

// Bind 使用app.Binder解析请求body并绑定数据。
func (ctx *ContextBase) Bind(i interface{}) error {
	return ctx.bind(i, ctx.app.Binder)
}

func (ctx *ContextBase) bind(i interface{}, r Binder) error {
	err := r(ctx, ctx.getReader(), i)
	if err == nil && ctx.GetParam("valid") != "" {
		err = ctx.app.Validater.Validate(i)
	}
	if err != nil {
		ctx.logReset(2).WithField(ParamCaller, "Context.ReadBind").Error(err)
	}
	return err
}

// Validate 方法调用app.Validater校验结构体对象。
func (ctx *ContextBase) Validate(i interface{}) error {
	return ctx.app.Validater.Validate(i)
}

// Params 获得请求的全部参数。
func (ctx *ContextBase) Params() *Params {
	return &ctx.httpParams
}

// GetParam 方法获取一个参数的值。
func (ctx *ContextBase) GetParam(key string) string {
	return ctx.httpParams.Get(key)
}

// SetParam 方法设置一个参数。
func (ctx *ContextBase) SetParam(key, val string) {
	ctx.httpParams.Set(key, val)
}

// AddParam 方法给参数添加一个新参数。
func (ctx *ContextBase) AddParam(key, val string) {
	ctx.httpParams.Add(key, val)
}

// Querys 方法返回http请求的全部uri参数。
func (ctx *ContextBase) Querys() url.Values {
	if ctx.querys == nil && ctx.RequestReader.URL != nil {
		newValues, err := url.ParseQuery(ctx.RequestReader.URL.RawQuery)
		if err != nil {
			ctx.Error(err)
			ctx.querys = make(url.Values)
		} else {
			ctx.querys = newValues
		}
	}
	return ctx.querys
}

// GetQuery 方法获得一个uri参数的值。
func (ctx *ContextBase) GetQuery(key string) string {
	return ctx.Querys().Get(key)
}

// GetHeader 获取一个请求header，相当于ctx.Request().Header().Get(name)。
func (ctx *ContextBase) GetHeader(name string) string {
	return ctx.RequestReader.Header.Get(name)
}

// SetHeader 设置一个响应header，相当于ctx.Response().Header().Set(name, val)
func (ctx *ContextBase) SetHeader(name string, val string) {
	ctx.ResponseWriter.Header().Set(name, val)
}

// Cookies 获取全部请求的cookie。
func (ctx *ContextBase) Cookies() []Cookie {
	return ctx.cookies
}

// GetCookie 获得一个请求cookie的值。
func (ctx *ContextBase) GetCookie(name string) string {
	for _, ctx := range ctx.cookies {
		if ctx.Name == name {
			return ctx.Value
		}
	}
	return ""
}

// SetCookie 设置一个Set-Cookie header，返回设置的cookie。
func (ctx *ContextBase) SetCookie(cookie *SetCookie) {
	if v := cookie.String(); v != "" {
		ctx.ResponseWriter.Header().Add(HeaderSetCookie, v)
	}
}

// SetCookieValue 返回一个cookie。
func (ctx *ContextBase) SetCookieValue(name, value string, maxAge int) {
	ctx.ResponseWriter.Header().Add(HeaderSetCookie, fmt.Sprintf("%s=%s; Max-Age=%d", name, url.QueryEscape(value), maxAge))
}

// FormValue 使用body解析成Form数据，并返回对应key的值
func (ctx *ContextBase) FormValue(key string) string {
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
func (ctx *ContextBase) FormValues() map[string][]string {
	if ctx.parseForm() != nil {
		return nil
	}
	return ctx.RequestReader.MultipartForm.Value
}

// FormFile 使用body解析成Form数据，并返回对应key的文件
func (ctx *ContextBase) FormFile(key string) *multipart.FileHeader {
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
func (ctx *ContextBase) FormFiles() map[string][]*multipart.FileHeader {
	if ctx.parseForm() != nil {
		return nil
	}
	return ctx.RequestReader.MultipartForm.File
}

// parseForm 解析form数据。
func (ctx *ContextBase) parseForm() error {
	if ctx.RequestReader.MultipartForm != nil {
		return nil
	}
	_, params, err := mime.ParseMediaType(ctx.GetHeader(HeaderContentType))
	if params == nil || params["boundary"] == "" {
		err = errors.New("content-type Header parse boundary is empty")
	}
	if err != nil {
		ctx.logReset(2).WithField(ParamCaller, "Context.Form...").WithField("check", "request content-type header: "+ctx.ContentType()).Error(err)
		return err
	}

	f, err := multipart.NewReader(ctx, params["boundary"]).ReadForm(DefaultBodyMaxMemory)
	if f != nil {
		ctx.RequestReader.MultipartForm = f
	}
	if err != nil {
		ctx.logReset(2).WithField(ParamCaller, "Context.Form...").Error(err)
	}
	return err
}

// WriteHeader 方法写入响应状态码。
func (ctx *ContextBase) WriteHeader(code int) {
	ctx.ResponseWriter.WriteHeader(code)
}

// Redirect implement request redirection.
//
// Redirect 实现请求重定向。
func (ctx *ContextBase) Redirect(code int, url string) {
	http.Redirect(ctx.ResponseWriter, ctx.RequestReader, url, code)
}

// Push 实现http2 push
func (ctx *ContextBase) Push(target string, opts *http.PushOptions) error {
	if opts == nil {
		opts = &http.PushOptions{
			Header: http.Header{
				HeaderAcceptEncoding: []string{ctx.RequestReader.Header.Get(HeaderAcceptEncoding)},
			},
		}
	}

	err := ctx.ResponseWriter.Push(target, opts)
	if err != nil {
		ctx.logReset(1).WithField(ParamCaller, "Context.Push").Errorf("Failed to push: %v, Resource path: %s.", err, target)
	}
	return err
}

// Write 实现io.Writer，向响应写入数据。
func (ctx *ContextBase) Write(data []byte) (n int, err error) {
	n, err = ctx.ResponseWriter.Write(data)
	if err != nil {
		ctx.logReset(1).WithField(ParamCaller, "Context.Write").Error(err)
	}
	return
}

// WriteString 实现向响应写入一个字符串。
func (ctx *ContextBase) WriteString(i string) (err error) {
	_, err = ctx.ResponseWriter.Write([]byte(i))
	if err != nil {
		ctx.logReset(1).WithField(ParamCaller, "Context.WriteString").Error(err)
	}
	return
}

// WriteJSON 使用Json返回数据。
func (ctx *ContextBase) WriteJSON(i interface{}) error {
	return ctx.writeRenderWith(i, RenderJSON)
}

// WriteFile 使用HandlerFile处理一个静态文件。
func (ctx *ContextBase) WriteFile(path string) (err error) {
	http.ServeFile(ctx.ResponseWriter, ctx.RequestReader, path)
	return nil
}

// Render 使用app.Renderer返回数据。
func (ctx *ContextBase) Render(i interface{}) error {
	return ctx.writeRenderWith(i, ctx.app.Renderer)
}

// RenderWith 使用指定的Render返回数据。
func (ctx *ContextBase) RenderWith(i interface{}, r Renderer) error {
	return ctx.writeRenderWith(i, r)
}

func (ctx *ContextBase) writeRenderWith(i interface{}, r Renderer) error {
	err := r(ctx, i)
	if err != nil {
		ctx.logReset(2).WithField(ParamCaller, "Context.Render Context.Render Context.WriteJSON").Error(err)
	}
	return err
}

// Debug 方法写入Debug日志。
func (ctx *ContextBase) Debug(args ...interface{}) {
	ctx.logReset(1).Debug(args...)
}

// Info 方法写入Info日志。
func (ctx *ContextBase) Info(args ...interface{}) {
	ctx.logReset(1).Info(args...)
}

// Warning 方法写入Warning日志。
func (ctx *ContextBase) Warning(args ...interface{}) {
	ctx.logReset(1).Warning(args...)
}

// Error 方法写入Error日志。
func (ctx *ContextBase) Error(args ...interface{}) {
	// 空错误不处理
	if len(args) == 1 && args[0] == nil {
		return
	}
	ctx.logReset(1).Error(args...)
}

// Fatal 方法写入Fatal日志，并结束请求上下文处理。
//
// 注意：如果err中存在敏感信息会被写入到响应中。
func (ctx *ContextBase) Fatal(args ...interface{}) {
	if len(args) == 1 && args[0] == nil {
		return
	}
	msg := fmt.Sprintln(args...)
	ctx.err = msg[:len(msg)-1]
	ctx.logReset(1).Error(ctx.err)
	ctx.logFatal()
}

// Debugf 方法输出Info日志。
func (ctx *ContextBase) Debugf(format string, args ...interface{}) {
	ctx.logReset(1).Debug(fmt.Sprintf(format, args...))
}

// Infof 方法输出Info日志。
func (ctx *ContextBase) Infof(format string, args ...interface{}) {
	ctx.logReset(1).Info(fmt.Sprintf(format, args...))
}

// Warningf 方法输出Warning日志。
func (ctx *ContextBase) Warningf(format string, args ...interface{}) {
	ctx.logReset(1).Warning(fmt.Sprintf(format, args...))
}

// Errorf 方法输出Error日志。
func (ctx *ContextBase) Errorf(format string, args ...interface{}) {
	ctx.logReset(1).Error(fmt.Sprintf(format, args...))
}

// Fatalf 方法输出Fatal日志，并结束请求上下文处理。
//
// 注意：如果err中存在敏感信息会被写入到响应中。
func (ctx *ContextBase) Fatalf(format string, args ...interface{}) {
	ctx.err = fmt.Sprintf(format, args...)
	ctx.logReset(1).Errorf(ctx.err)
	ctx.logFatal()
}

// logReset 方法添加Context基础信息。
func (ctx *ContextBase) logReset(depth int) Logout {
	fields := make(Fields)
	if depth != 0 {
		fields["depth"] = depth
	}
	requestid := ctx.GetHeader(HeaderXRequestID)
	if requestid != "" {
		fields[HeaderXRequestID] = requestid
	}
	return ctx.app.Logger.WithFields(fields)
}

// logFatal 方法执行Fatal方法的返回信息。
func (ctx *ContextBase) logFatal() {
	// 结束Context
	status := ctx.ResponseWriter.Status()
	if ctx.ResponseWriter.Size() == 0 {
		status = 500
		ctx.WriteHeader(500)
	}
	if status > 399 {
		ctx.Render(map[string]interface{}{
			"error":        ctx.err,
			"status":       status,
			"x-request-id": ctx.RequestID(),
		})
	}
	ctx.End()
}

// WithField 方法增加一个日志属性，返回一个新的Logout。
func (ctx *ContextBase) WithField(key string, value interface{}) Logout {
	return &entryContextBase{
		Logout:  ctx.logReset(0).WithField(key, value),
		Context: ctx,
	}
}

// WithFields 方法增加多个日志属性，返回一个新的Logout。
//
// 如果fields包含file条目属性，则不会添加调用位置信息。
func (ctx *ContextBase) WithFields(fields Fields) Logout {
	if fields != nil {
		fields[HeaderXRequestID] = ctx.GetHeader(HeaderXRequestID)
	}
	return &entryContextBase{
		Logout:  ctx.logReset(0).WithFields(fields),
		Context: ctx,
	}
}

// Logger 直接返回app的Logger对象，通常用于Hijack并释放Context后使用Logout。
func (ctx *ContextBase) Logger() Logout {
	return ctx.logReset(0).WithField("logout", true)
}

// Fatal 方法重写Context的Fatal方法，不执行panic，http返回500和请求id。
func (e *entryContextBase) Fatal(args ...interface{}) {
	msg := fmt.Sprintln(args...)
	e.Context.err = msg[:len(msg)-1]
	e.Logout.WithField("depth", 1).Error(msg)
	e.Context.logFatal()

}

// Fatalf 方法重写Context的Fatalf方法，不执行panic，http返回500和请求id。
func (e *entryContextBase) Fatalf(format string, args ...interface{}) {
	e.Context.err = fmt.Sprintf(format, args...)
	e.Logout.WithField("depth", 1).Error(e.Context.err)
	e.Context.logFatal()
}

// WithField 方法增加一个日志属性。
func (e *entryContextBase) WithField(key string, value interface{}) Logout {
	e.Logout = e.Logout.WithField(key, value)
	return e
}

// WithFields 方法增加多个日志属性。
func (e *entryContextBase) WithFields(fields Fields) Logout {
	e.Logout = e.Logout.WithFields(fields)
	return e
}

// readCookies 方法初始化cookie键值对，form net/http。
func (ctx *ContextBase) readCookies(line string) {
	if len(line) == 0 {
		return
	}
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

// ContextData 扩展Context对象，加入获取数据类型转换。
type ContextData struct {
	Context
}

// NewExtendContextData 转换ContextData处理函数为Context处理函数。
func NewExtendContextData(fn func(ContextData)) HandlerFunc {
	return func(ctx Context) {
		fn(ContextData{Context: ctx})
	}
}

// GetParamBool 获取参数转换成bool类型。
func (ctx ContextData) GetParamBool(key string) bool {
	return GetStringDefaultBool(ctx.GetParam(key), false)
}

// GetParamDefaultBool 获取参数转换成bool类型，转换失败返回默认值。
func (ctx ContextData) GetParamDefaultBool(key string, b bool) bool {
	return GetStringDefaultBool(ctx.GetParam(key), b)
}

// GetParamInt 获取参数转换成int类型。
func (ctx ContextData) GetParamInt(key string) int {
	return GetStringDefaultInt(ctx.GetParam(key), 0)
}

// GetParamDefaultInt 获取参数转换成int类型，转换失败返回默认值。
func (ctx ContextData) GetParamDefaultInt(key string, i int) int {
	return GetStringDefaultInt(ctx.GetParam(key), i)
}

// GetParamInt64 获取参数转换成int64类型。
func (ctx ContextData) GetParamInt64(key string) int64 {
	return GetStringDefaultInt64(ctx.GetParam(key), 0)
}

// GetParamDefaultInt64 获取参数转换成int64类型，转换失败返回默认值。
func (ctx ContextData) GetParamDefaultInt64(key string, i int64) int64 {
	return GetStringDefaultInt64(ctx.GetParam(key), i)
}

// GetParamFloat32 获取参数转换成int32类型。
func (ctx ContextData) GetParamFloat32(key string) float32 {
	return GetStringDefaultFloat32(ctx.GetParam(key), 0)
}

// GetParamDefaultFloat32 获取参数转换成int32类型，转换失败返回默认值。
func (ctx ContextData) GetParamDefaultFloat32(key string, f float32) float32 {
	return GetStringDefaultFloat32(ctx.GetParam(key), f)
}

// GetParamFloat64 获取参数转换成float64类型。
func (ctx ContextData) GetParamFloat64(key string) float64 {
	return GetStringDefaultFloat64(ctx.GetParam(key), 0)
}

// GetParamDefaultFloat64 获取参数转换成float64类型，转换失败返回默认值。
func (ctx ContextData) GetParamDefaultFloat64(key string, f float64) float64 {
	return GetStringDefaultFloat64(ctx.GetParam(key), f)
}

// GetParamDefaultString 获取一个参数，如果为空字符串返回默认值。
func (ctx ContextData) GetParamDefaultString(key, str string) string {
	return GetStringDefault(ctx.GetParam(key), str)
}

// GetHeaderBool 获取header转换成bool类型。
func (ctx ContextData) GetHeaderBool(key string) bool {
	return GetStringDefaultBool(ctx.GetHeader(key), false)
}

// GetHeaderDefaultBool 获取header转换成bool类型，转换失败返回默认值。
func (ctx ContextData) GetHeaderDefaultBool(key string, b bool) bool {
	return GetStringDefaultBool(ctx.GetHeader(key), b)
}

// GetHeaderInt 获取header转换成int类型。
func (ctx ContextData) GetHeaderInt(key string) int {
	return GetStringDefaultInt(ctx.GetHeader(key), 0)
}

// GetHeaderDefaultInt 获取header转换成int类型，转换失败返回默认值。
func (ctx ContextData) GetHeaderDefaultInt(key string, i int) int {
	return GetStringDefaultInt(ctx.GetHeader(key), i)
}

// GetHeaderInt64 获取header转换成int64类型。
func (ctx ContextData) GetHeaderInt64(key string) int64 {
	return GetStringDefaultInt64(ctx.GetHeader(key), 0)
}

// GetHeaderDefaultInt64 获取header转换成int64类型，转换失败返回默认值。
func (ctx ContextData) GetHeaderDefaultInt64(key string, i int64) int64 {
	return GetStringDefaultInt64(ctx.GetHeader(key), i)
}

// GetHeaderFloat32 获取header转换成float32类型。
func (ctx ContextData) GetHeaderFloat32(key string) float32 {
	return GetStringDefaultFloat32(ctx.GetHeader(key), 0)
}

// GetHeaderDefaultFloat32 获取header转换成float32类型，转换失败返回默认值。
func (ctx ContextData) GetHeaderDefaultFloat32(key string, f float32) float32 {
	return GetStringDefaultFloat32(ctx.GetHeader(key), f)
}

// GetHeaderFloat64 获取header转换成float64类型。
func (ctx ContextData) GetHeaderFloat64(key string) float64 {
	return GetStringDefaultFloat64(ctx.GetHeader(key), 0)
}

// GetHeaderDefaultFloat64 获取header转换成float64类型，转换失败返回默认值。
func (ctx ContextData) GetHeaderDefaultFloat64(key string, f float64) float64 {
	return GetStringDefaultFloat64(ctx.GetHeader(key), f)
}

// GetHeaderDefaultString 获取header，如果为空字符串返回默认值。
func (ctx ContextData) GetHeaderDefaultString(key, str string) string {
	return GetStringDefault(ctx.GetHeader(key), str)
}

// GetQueryBool 获取uri参数值转换成bool类型。
func (ctx ContextData) GetQueryBool(key string) bool {
	return GetStringDefaultBool(ctx.GetQuery(key), false)
}

// GetQueryDefaultBool 获取uri参数值转换成bool类型，转换失败返回默认值。
func (ctx ContextData) GetQueryDefaultBool(key string, b bool) bool {
	return GetStringDefaultBool(ctx.GetQuery(key), b)
}

// GetQueryInt 获取uri参数值转换成int类型。
func (ctx ContextData) GetQueryInt(key string) int {
	return GetStringDefaultInt(ctx.GetQuery(key), 0)
}

// GetQueryDefaultInt 获取uri参数值转换成int类型，转换失败返回默认值。
func (ctx ContextData) GetQueryDefaultInt(key string, i int) int {
	return GetStringDefaultInt(ctx.GetQuery(key), i)
}

// GetQueryInt64 获取uri参数值转换成int64类型。
func (ctx ContextData) GetQueryInt64(key string) int64 {
	return GetStringDefaultInt64(ctx.GetQuery(key), 0)
}

// GetQueryDefaultInt64 获取uri参数值转换成int64类型，转换失败返回默认值。
func (ctx ContextData) GetQueryDefaultInt64(key string, i int64) int64 {
	return GetStringDefaultInt64(ctx.GetQuery(key), i)
}

// GetQueryFloat32 获取url参数值转换成float32类型。
func (ctx ContextData) GetQueryFloat32(key string) float32 {
	return GetStringDefaultFloat32(ctx.GetQuery(key), 0)
}

// GetQueryDefaultFloat32 获取url参数值转换成float32类型，转换失败返回默认值。
func (ctx ContextData) GetQueryDefaultFloat32(key string, f float32) float32 {
	return GetStringDefaultFloat32(ctx.GetQuery(key), f)
}

// GetQueryFloat64 获取url参数值转换成float64类型。
func (ctx ContextData) GetQueryFloat64(key string) float64 {
	return GetStringDefaultFloat64(ctx.GetQuery(key), 0)
}

// GetQueryDefaultFloat64 获取url参数值转换成float64类型，转换失败返回默认值。
func (ctx ContextData) GetQueryDefaultFloat64(key string, f float64) float64 {
	return GetStringDefaultFloat64(ctx.GetQuery(key), f)
}

// GetQueryDefaultString 获取一个uri参数的值，如果为空字符串返回默认值。
func (ctx ContextData) GetQueryDefaultString(key, str string) string {
	return GetStringDefault(ctx.GetQuery(key), str)
}

// GetCookieBool 获取一个cookie的转换成bool类型。
func (ctx ContextData) GetCookieBool(key string) bool {
	return GetStringDefaultBool(ctx.GetCookie(key), false)
}

// GetCookieDefaultBool 获取一个cookie的转换成bool类型，转换失败返回默认值
func (ctx ContextData) GetCookieDefaultBool(key string, b bool) bool {
	return GetStringDefaultBool(ctx.GetCookie(key), b)
}

// GetCookieInt 获取一个cookie的转换成int类型。
func (ctx ContextData) GetCookieInt(key string) int {
	return GetStringDefaultInt(ctx.GetCookie(key), 0)
}

// GetCookieDefaultInt 获取一个cookie的转换成int类型，转换失败返回默认值
func (ctx ContextData) GetCookieDefaultInt(key string, i int) int {
	return GetStringDefaultInt(ctx.GetCookie(key), i)
}

// GetCookieInt64 获取一个cookie的转换成int64类型。
func (ctx ContextData) GetCookieInt64(key string) int64 {
	return GetStringDefaultInt64(ctx.GetCookie(key), 0)
}

// GetCookieDefaultInt64 获取一个cookie的转换成int64类型，转换失败返回默认值
func (ctx ContextData) GetCookieDefaultInt64(key string, i int64) int64 {
	return GetStringDefaultInt64(ctx.GetCookie(key), i)
}

// GetCookieFloat32 获取一个cookie的转换成float32类型。
func (ctx ContextData) GetCookieFloat32(key string) float32 {
	return GetStringDefaultFloat32(ctx.GetCookie(key), 0)
}

// GetCookieDefaultFloat32 获取一个cookie的转换成float32类型，转换失败返回默认值
func (ctx ContextData) GetCookieDefaultFloat32(key string, f float32) float32 {
	return GetStringDefaultFloat32(ctx.GetCookie(key), f)
}

// GetCookieFloat64 获取一个cookie的转换成float64类型。
func (ctx ContextData) GetCookieFloat64(key string) float64 {
	return GetStringDefaultFloat64(ctx.GetCookie(key), 0)
}

// GetCookieDefaultFloat64 获取一个cookie的转换成float64类型，转换失败返回默认值
func (ctx ContextData) GetCookieDefaultFloat64(key string, f float64) float64 {
	return GetStringDefaultFloat64(ctx.GetCookie(key), f)
}

// GetCookieDefaultString 获取一个cookie的值，如果为空字符串返回默认值。
func (ctx ContextData) GetCookieDefaultString(key, str string) string {
	return GetStringDefault(ctx.GetCookie(key), str)
}
