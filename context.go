package eudore

/*
Context

Context定义一个请求上下文

文件：context.go
*/

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/url"
	"strings"
	"unsafe"

	"github.com/eudore/eudore/protocol"
)

type (
	// Context 定义请求上下文接口。
	Context interface {
		// context
		Reset(context.Context, protocol.ResponseWriter, protocol.RequestReader)
		Context() context.Context
		Request() protocol.RequestReader
		Response() protocol.ResponseWriter
		SetRequest(protocol.RequestReader)
		SetResponse(protocol.ResponseWriter)
		SetHandler(HandlerFuncs)
		Next()
		End()

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

		// param query header cookie session
		Params() Params
		GetParam(string) string
		SetParam(string, string)
		AddParam(string, string)
		Querys() Querys
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
		Push(string, *protocol.PushOptions) error
		Render(interface{}) error
		RenderWith(interface{}, Renderer) error
		// render writer
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
	}

	// ContextBase 实现Context接口。
	ContextBase struct {
		protocol.RequestReader
		protocol.ResponseWriter
		ParamsArray
		QueryURL
		// run handler
		index   int
		handler HandlerFuncs
		// data
		ctx        context.Context
		cookies    []Cookie
		isReadBody bool
		postBody   []byte
		form       *multipart.Form
		// component
		app *App
		log Logger
	}
	// entryContext 实现Context使用的Logout对象。
	entryContext struct {
		Logout
		Context Context
	}
)

// Convert nil to type *ContextBase, detect ContextBase object to implement Context interface
//
// 将nil强制类型转换成*ContextBase，检测ContextBase对象实现Context接口
var _ Context = (*ContextBase)(nil)

// NewContextBase 创建ContextBase对象，实现Context接口。
func NewContextBase(app *App) *ContextBase {
	return &ContextBase{
		app: app,
	}
}

// Reset Context
func (ctx *ContextBase) Reset(pctx context.Context, w protocol.ResponseWriter, r protocol.RequestReader) {
	ctx.ctx = pctx
	ctx.RequestReader = r
	ctx.ResponseWriter = w
	// logger
	ctx.log = ctx.app.Logger

	// query
	err := ctx.QueryURL.readQuery(r.RawQuery())
	if err != nil {
		ctx.log.WithField("caller", "Context.Reset").WithField("check", "request uri raw: "+r.RawQuery()).Error(err)
	}
	// params
	ctx.ParamsArray.Keys = ctx.ParamsArray.Keys[0:0]
	ctx.ParamsArray.Vals = ctx.ParamsArray.Vals[0:0]
	// cookies
	ctx.cookies = ctx.cookies[0:0]
	ctx.readCookies(r.Header().Get(HeaderCookie))
	ctx.form = emptyMultipartForm

	// body
	ctx.isReadBody = false
	ctx.postBody = ctx.postBody[0:0]
}

// Context 获取当前请求的上下文。
func (ctx *ContextBase) Context() context.Context {
	return ctx.ctx
}

// Request 获取请求对象。
func (ctx *ContextBase) Request() protocol.RequestReader {
	return ctx.RequestReader
}

// Response 获得响应对象。
func (ctx *ContextBase) Response() protocol.ResponseWriter {
	return ctx.ResponseWriter
}

// SetRequest 设置请求对象。
func (ctx *ContextBase) SetRequest(r protocol.RequestReader) {
	ctx.RequestReader = r
}

// SetResponse 设置响应对象。
func (ctx *ContextBase) SetResponse(w protocol.ResponseWriter) {
	ctx.ResponseWriter = w
}

// SetHandler 重新设置上下文的全部请求处理者。
func (ctx *ContextBase) SetHandler(fs HandlerFuncs) {
	ctx.index = -1
	ctx.handler = fs
}

// Next 调用请求上下文下一个处理函数。
func (ctx *ContextBase) Next() {
	ctx.index++
	for ctx.index < len(ctx.handler) {
		ctx.handler[ctx.index](ctx)
		ctx.index++
	}
}

// End 结束请求上下文的处理。
func (ctx *ContextBase) End() {
	ctx.index = 0xff
	ctx.form.RemoveAll()
}

// RealIP 获取用户真实ip，ctx.Request().RemoteAddr()获取远程连接地址。
func (ctx *ContextBase) RealIP() string {
	xforward := ctx.RequestReader.Header().Get(HeaderXForwardedFor)
	if "" == xforward {
		return strings.SplitN(ctx.RequestReader.RemoteAddr(), ":", 2)[0]
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
	return ctx.RequestReader.TLS() != nil
}

// Body 返回请求的body，并保存到缓存中，可重复调用Body方法。
func (ctx *ContextBase) Body() []byte {
	if !ctx.isReadBody {
		bts, err := ioutil.ReadAll(ctx.RequestReader)
		if err != nil {
			ctx.logReset(0).WithField("caller", "Context.Body").Error(err)
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
func (ctx *ContextBase) BindWith(i interface{}, r Binder) (err error) {
	err = r(ctx, ctx.getReader(), i)
	if err != nil {
		ctx.logReset(0).WithField("caller", "Context.ReadBind").Error(err)
	}
	return
}

// Bind 使用app.Binder解析请求body并绑定数据。
func (ctx *ContextBase) Bind(i interface{}) (err error) {
	err = ctx.app.Binder(ctx, ctx.getReader(), i)
	if err != nil {
		ctx.logReset(0).WithField("caller", "Context.ReadBind").Error(err)
	}
	return
}

// Params 获得请求的全部参数。
func (ctx *ContextBase) Params() Params {
	return &ctx.ParamsArray
}

// Querys 方法返回http请求的全部uri参数。
func (ctx *ContextBase) Querys() Querys {
	return &ctx.QueryURL
}

// GetQuery 方法获得一个uri参数的值。
func (ctx *ContextBase) GetQuery(key string) string {
	return ctx.QueryURL.Get(key)
}

// GetHeader 获取一个请求header，相当于ctx.Request().Header().Get(name)。
func (ctx *ContextBase) GetHeader(name string) string {
	return ctx.RequestReader.Header().Get(name)
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
		ctx.ResponseWriter.Header().Add("Set-Cookie", v)
	}
}

// SetCookieValue 返回一个cookie。
func (ctx *ContextBase) SetCookieValue(name, value string, maxAge int) {
	ctx.ResponseWriter.Header().Add("Set-Cookie", fmt.Sprintf("%s=%s; Max-Age=%d", name, url.QueryEscape(value), maxAge))
}

/*
// GetSession 获取当前请求上下文的会话数据。
func (ctx *ContextBase) GetSession() SessionData {
	return ctx.app.Session.SessionLoad(ctx)
}

// SetSession 给当前请求上下文设置会话数据。
func (ctx *ContextBase) SetSession(sess SessionData) {
	ctx.app.Session.SessionSave(sess)
}
*/

// 定义空的Form对象。
var emptyMultipartForm = &multipart.Form{
	Value: make(map[string][]string),
	File:  make(map[string][]*multipart.FileHeader),
}

// FormValue 使用body解析成Form数据，并返回对应key的值
func (ctx *ContextBase) FormValue(key string) string {
	ctx.parseForm()
	val, ok := ctx.form.Value[key]
	if ok && len(val) != 0 {
		return val[0]
	}
	return ""
}

// FormValues 使用body解析成Form数据，并返回全部的值
func (ctx *ContextBase) FormValues() map[string][]string {
	ctx.parseForm()
	return ctx.form.Value
}

// FormFile 使用body解析成Form数据，并返回对应key的文件
func (ctx *ContextBase) FormFile(key string) *multipart.FileHeader {
	ctx.parseForm()
	val, ok := ctx.form.File[key]
	if ok && len(val) != 0 {
		return val[0]
	}
	return nil
}

// FormFiles 使用body解析成Form数据，并返回全部的文件。
func (ctx *ContextBase) FormFiles() map[string][]*multipart.FileHeader {
	ctx.parseForm()
	return ctx.form.File
}

// parseForm 解析form数据。
func (ctx *ContextBase) parseForm() error {
	if ctx.form != emptyMultipartForm {
		return nil
	}
	_, params, err := mime.ParseMediaType(ctx.GetHeader(HeaderContentType))
	if err != nil {
		ctx.logReset(0).WithField("caller", "Context.Form...").WithField("check", "request content-type header: "+ctx.ContentType()).Error(err)
		return err
	}

	f, err := multipart.NewReader(ctx, params["boundary"]).ReadForm(defaultMaxMemory)
	if f != nil {
		ctx.form = f
	}
	if err != nil {
		ctx.logReset(0).WithField("caller", "Context.Form...").Error(err)
	}
	return err
}

// Redirect implement request redirection.
//
// Redirect 实现请求重定向。
func (ctx *ContextBase) Redirect(code int, url string) {
	HandlerRedirect(ctx, url, code)
}

// Push 实现http2 push
func (ctx *ContextBase) Push(target string, opts *protocol.PushOptions) error {
	if opts == nil {
		opts = &protocol.PushOptions{
			Header: HeaderMap{
				HeaderAcceptEncoding: []string{ctx.RequestReader.Header().Get(HeaderAcceptEncoding)},
			},
		}
	}

	err := ctx.ResponseWriter.Push(target, opts)
	if err != nil {
		ctx.logReset(0).WithField("caller", "Context.Push").Errorf("Failed to push: %v, Resource path: %s.", err, target)
	}
	return err
}

// Write 实现io.Writer，向响应写入数据。
func (ctx *ContextBase) Write(data []byte) (n int, err error) {
	n, err = ctx.ResponseWriter.Write(data)
	if err != nil {
		ctx.logReset(0).WithField("caller", "Context.Write").Error(err)
	}
	return
}

/*
// WriteView 实现返回一个View渲染的html。
func (ctx *ContextBase) WriteView(path string, i interface{}) error {
	ctx.ResponseWriter.Header().Add(HeaderContentType, MimeTextHTMLCharsetUtf8)
	err := ctx.app.View.ExecuteTemplate(ctx, path, i)
	if err != nil {
		ctx.logReset(0).WithField("caller", "Context.WriteView").Error(err)
	}
	return err
}
*/

// WriteString 实现向响应写入一个字符串。
func (ctx *ContextBase) WriteString(i string) (err error) {
	_, err = ctx.ResponseWriter.Write(*(*[]byte)(unsafe.Pointer(&i)))
	if err != nil {
		ctx.logReset(0).WithField("caller", "Context.WriteString").Error(err)
	}
	return
}

// WriteJSON 使用Json返回数据。
func (ctx *ContextBase) WriteJSON(i interface{}) error {
	return ctx.writeRenderWith(i, RenderJSON)
}

// WriteFile 使用HandlerFile处理一个静态文件。
func (ctx *ContextBase) WriteFile(path string) (err error) {
	err = HandlerFile(ctx, path)
	if err != nil {
		ctx.logReset(0).WithField("caller", "Context.WriteFile").Error(err)
	}
	return
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
		ctx.logReset(1).WithField("caller", "Context.Render Context.Render Context.WriteJSON").Error(err)
	}
	return err
}

// Debug 方法写入Debug日志。
func (ctx *ContextBase) Debug(args ...interface{}) {
	ctx.logReset(0).Debug(fmt.Sprint(args...))
}

// Info 方法写入Info日志。
func (ctx *ContextBase) Info(args ...interface{}) {
	ctx.logReset(0).Info(fmt.Sprint(args...))
}

// Warning 方法写入Warning日志。
func (ctx *ContextBase) Warning(args ...interface{}) {
	ctx.logReset(0).Warning(fmt.Sprint(args...))
}

// Error 方法写入Error日志。
func (ctx *ContextBase) Error(args ...interface{}) {
	// 空错误不处理
	if len(args) == 1 && args[0] == nil {
		return
	}
	ctx.logReset(0).Error(fmt.Sprint(args...))
}

// Fatal 方法写入Fatal日志，并结束请求上下文处理。
func (ctx *ContextBase) Fatal(args ...interface{}) {
	ctx.logReset(0).Error(fmt.Sprint(args...))
	// 结束Context
	if ctx.ResponseWriter.Status() == 200 {
		ctx.WriteHeader(500)
		ctx.Render(map[string]string{
			// "error":        fmt.Sprint(args...),
			"status":       "500",
			"x-request-id": ctx.RequestID(),
		})
	}
	ctx.End()
}

// Debugf 方法输出Info日志。
func (ctx *ContextBase) Debugf(format string, args ...interface{}) {
	ctx.logReset(0).Debug(fmt.Sprintf(format, args...))
}

// Infof 方法输出Info日志。
func (ctx *ContextBase) Infof(format string, args ...interface{}) {
	ctx.logReset(0).Info(fmt.Sprintf(format, args...))
}

// Warningf 方法输出Warning日志。
func (ctx *ContextBase) Warningf(format string, args ...interface{}) {
	ctx.logReset(0).Warning(fmt.Sprintf(format, args...))
}

// Errorf 方法输出Error日志。
func (ctx *ContextBase) Errorf(format string, args ...interface{}) {
	ctx.logReset(0).Error(fmt.Sprintf(format, args...))
}

// Fatalf 方法输出Fatal日志，并结束请求上下文处理。
func (ctx *ContextBase) Fatalf(format string, args ...interface{}) {
	ctx.logReset(0).Error(fmt.Sprintf(format, args...))
	// 结束Context
	if ctx.ResponseWriter.Status() == 200 {
		ctx.WriteHeader(500)
		ctx.Render(map[string]string{
			// "error":        fmt.Sprintf(format, args...),
			"status":       "500",
			"x-request-id": ctx.RequestID(),
		})
	}
	ctx.End()
}

func (ctx *ContextBase) logReset(depth int) Logout {
	fields := make(Fields)
	file, line := logFormatFileLine(depth)
	fields[HeaderXRequestID] = ctx.GetHeader(HeaderXRequestID)
	fields["file"] = file
	fields["line"] = line
	return ctx.log.WithFields(fields)
}

// WithField 方法增加一个日志属性，返回一个新的Logout。
func (ctx *ContextBase) WithField(key string, value interface{}) Logout {
	return &entryContext{
		Logout:  ctx.logReset(0).WithField(key, value),
		Context: ctx,
	}
}

// WithFields 方法增加多个日志属性，返回一个新的Logout。
func (ctx *ContextBase) WithFields(fields Fields) Logout {
	file, line := logFormatFileLine(0)
	fields[HeaderXRequestID] = ctx.GetHeader(HeaderXRequestID)
	fields["file"] = file
	fields["line"] = line
	return &entryContext{
		Logout:  ctx.log.WithFields(fields),
		Context: ctx,
	}
}

// Fatal 方法重写Context的Fatal方法，不执行panic，http返回500和请求id。
func (e *entryContext) Fatal(args ...interface{}) {
	e.Logout.Error(args...)
	contextFatal(e.Context)

}

// Fatalf 方法重写Context的Fatalf方法，不执行panic，http返回500和请求id。
func (e *entryContext) Fatalf(format string, args ...interface{}) {
	e.Logout.Errorf(format, args...)
	contextFatal(e.Context)
}

// WithField 方法增加一个日志属性。
func (e *entryContext) WithField(key string, value interface{}) Logout {
	e.Logout = e.Logout.WithField(key, value)
	return e
}

// WithFields 方法增加多个日志属性。
func (e *entryContext) WithFields(fields Fields) Logout {
	e.Logout = e.Logout.WithFields(fields)
	return e
}

func contextFatal(ctx Context) {
	if ctx.Response().Status() == 200 {
		ctx.WriteHeader(500)
		ctx.Render(map[string]string{
			"status":       "500",
			"x-request-id": ctx.RequestID(),
		})
	}
	ctx.End()
}

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
