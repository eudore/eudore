package eudore

import (
	"io"
	"fmt"
	"time"
	"unsafe"
	"strings"
	"context"
	"net/http"
	"io/ioutil"
	"crypto/tls"
	"golang.org/x/net/http2"
)

const sniffLen = 512
const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

type (
	Context interface {
		// context
		Reset(context.Context, ResponseWriter, RequestReader)
		Request() RequestReader
		Response() ResponseWriter
		SetRequest(RequestReader)
		SetResponse(ResponseWriter)
		SetHandler(Middleware)
		Next()
		End()
		NewRequest(string, string, io.Reader) (ResponseReader, error)
		// context
		Deadline() (time.Time, bool)
		Done() <-chan struct{}
		Err() error
		Value(key interface{}) interface{}
		SetValue(interface{}, interface{})

		// request info
		Read([]byte) (int, error)
		Host() string
		Method() string
		Path() string
		RemoteAddr() string
		RequestID() string
		Referer() string
		ContentType() string
		Istls() bool
		Body() []byte

		// param header cookie
		Params() Params
		GetParam(string) string
		SetParam(string, string)
		AddParam(string, string)
		GetHeader(name string) string
		SetHeader(string, string)
		Cookies() []*CookieRead
		GetCookie(name string) string
		SetCookie(cookie *CookieWrite)
		SetCookieValue(string, string, int)


		// response
		Write([]byte) (int, error)
		WriteHeader(int)
		Redirect(int, string)
		Push(string, *PushOptions) error
		// render writer 
		WriteString(string) error
		WriteView(string, interface{}) error
		WriteJson(interface{}) error
		WriteFile(string) (int, error)
		// binder and renderer
		ReadBind(interface{}) error
		WriteRender(interface{}) error
		// log LogOut interface
		Debug(...interface{})
		Info(...interface{})
		Warning(...interface{})
		Error(...interface{})
		Fatal(...interface{})
		WithField(key string, value interface{}) LogOut
		WithFields(fields Fields) LogOut
		// app
		App() *App
	}

	/* 实现Context接口 */
	ContextHttp struct {
		context.Context
		RequestReader
		ResponseWriter
		Middleware
		// data
		keys		map[interface{}]interface{}
		path 		string
		rawQuery	string
		pkeys		[]string
		pvals		[]string
		cookies 	[]*CookieRead
		isReadBody	bool
		postBody	[]byte
		isrun		bool
		handler		Handler
		// component
		app			*App
		log			Logger
	}

)

// check interface
var _ Context			=	&ContextHttp{}

// context
func (ctx *ContextHttp) Reset(pctx context.Context, w ResponseWriter, r RequestReader) {
	ctx.Context = pctx
	ctx.RequestReader = r
	ctx.ResponseWriter = w
	ctx.keys = nil
	// path and raw
	uri := r.RequestURI()
	pos := strings.IndexByte(uri, '?')
	if pos == -1 {
		ctx.path = uri
		ctx.rawQuery = ""
	}else {
		ctx.path = uri[:pos]
		ctx.rawQuery = uri[pos + 1:]
	}

	ctx.isrun = true
	ctx.pkeys = ctx.pkeys[0:0]
	ctx.pvals = ctx.pvals[0:0]
	// ctx.params = make(Params)
	// ctx.AddParam(ParamRoutePath, ctx.path)
	// ctx.AddParam(ParamRouteMethod, ctx.Method())
	ctx.cookies = ReadCookies(r.Header()[HeaderCookie])
	ctx.isReadBody = false
	ctx.postBody = ctx.postBody[0:0]
	ctx.log = ctx.app.Logger
	readQuery(ctx.rawQuery, ctx)
}

func (ctx *ContextHttp) Request() RequestReader {
	return ctx.RequestReader
}

func (ctx *ContextHttp) Response() ResponseWriter {
	return ctx.ResponseWriter
}

func (ctx *ContextHttp) SetRequest(r RequestReader) {
	ctx.RequestReader = r
}

func (ctx *ContextHttp) SetResponse(w ResponseWriter) {
	ctx.ResponseWriter = w
}

func (ctx *ContextHttp) SetHandler(m Middleware) {
	ctx.Middleware = m
}

func (ctx *ContextHttp) Next() {
	for ctx.Middleware != nil && ctx.isrun {
		ctx.handler = ctx.Middleware
		ctx.Middleware = ctx.Middleware.GetNext()
		ctx.handler.Handle(ctx)
	}
}

func (ctx *ContextHttp) End() {
	ctx.isrun = false
}

func (ctx *ContextHttp) NewRequest(method, url string, body io.Reader) (ResponseReader, error) {
	tr := &http2.Transport{
		AllowHTTP: true, //充许非加密的链接
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	httpClient := http.Client{Transport: tr}

	cctx, cancel := context.WithCancel(ctx)
	time.AfterFunc(5*time.Second, func() {
		cancel()
	})

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add(HeaderXRequestID, ctx.RequestID())
	req = req.WithContext(cctx)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	// check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("resp StatusCode: %d", resp.StatusCode)
	}
	return NewResponseReaderHttp(resp), err
}

func (ctx *ContextHttp) Value(key interface{}) interface{} {
	v, ok := ctx.keys[key]
	if ok {
		return v
	}
	return ctx.Context.Value(key)
}

func (ctx *ContextHttp) SetValue(key interface{}, val interface{}) {
	if ctx.keys == nil {
		ctx.keys = make(map[interface{}]interface{})
	}
	ctx.keys[key] = val
}













func (ctx *ContextHttp) Path() string {
	return ctx.path
}

func (ctx *ContextHttp) RemoteAddr() string {
	xforward := ctx.RequestReader.Header().Get(HeaderXForwardedFor)
	if "" == xforward {
		return strings.SplitN(ctx.RequestReader.RemoteAddr(), ":", 2)[0]
	}
	return strings.SplitN(string(xforward), ",", 2)[0]
}

func (ctx *ContextHttp) RequestID() string {
	return ctx.GetHeader(HeaderXRequestID)
}

func (ctx *ContextHttp) Referer() string {
	return ctx.GetHeader(HeaderReferer)
}

func (ctx *ContextHttp) ContentType() string {
	return ctx.GetHeader(HeaderContentType)
}

func (ctx *ContextHttp) Istls() bool {
	return ctx.RequestReader.TLS() != nil
}

func (ctx *ContextHttp) Body() []byte {
	if !ctx.isReadBody {
		bts, err := ioutil.ReadAll(ctx.RequestReader)
		if err != nil {
			return []byte{}
		} else {
			ctx.isReadBody = true
			ctx.postBody = bts
		}
	}
	return ctx.postBody
}



func (ctx *ContextHttp) Params() Params {
	return ctx
}

func (ctx *ContextHttp) GetParam(key string) string {
	for i, str := range ctx.pkeys {
		if str == key {
			return ctx.pvals[i]
		}
	}
	return ""
}

func (ctx *ContextHttp) AddParam(key string, val string) {
	ctx.pkeys = append(ctx.pkeys, key)
	ctx.pvals = append(ctx.pvals, val)
}

func (ctx *ContextHttp) SetParam(key string, val string) {
	for i, str := range ctx.pkeys {
		if str == key {
			ctx.pvals[i] = val
			return
		}
	}
	ctx.AddParam(key, val)
}
/*

func (ctx *ContextHttp) Params() Params {
	return ctx.params
}

func (ctx *ContextHttp) GetParam(key string) string {
	return ctx.params.GetPa(key)
}

func (ctx *ContextHttp) SetParam(key string, val string) {
	ctx.params.Set(key, val)
}

func (ctx *ContextHttp) AddParam(key string, val string) {
	ctx.params.Add(key, val)
}*/


func (ctx *ContextHttp) GetHeader(name string) string {
	return ctx.RequestReader.Header().Get(name)
}

func (ctx *ContextHttp) SetHeader(name string, val string) {
	ctx.ResponseWriter.Header().Set(name, val)
}

func (ctx *ContextHttp) Cookies() []*CookieRead {
	return ctx.cookies
}

func (ctx *ContextHttp) GetCookie(name string) string {
	for _, ctx := range ctx.cookies {
		if ctx.Name == name {
			return ctx.Value	
		}
	}
	return ""
}

func (ctx *ContextHttp) SetCookie(cookie *CookieWrite) {
	// ctx.RequestReader.AddCookie(cookie)	
	if v := cookie.String(); v != "" {
		ctx.ResponseWriter.Header().Add("Set-Cookie", v)
	}
}
func (ctx *ContextHttp) SetCookieValue(name, value string, maxAge int) {
	ctx.ResponseWriter.Header().Add("Set-Cookie", fmt.Sprintf("%s=%s; Max-Age=%d", name, value ,maxAge))
//	ctx.SetCookie(&http.Cookie{Name: name, Value: url.QueryEscape(value), Path: "/", MaxAge: maxAge})
}





// response
//
// func (ctx *ContextHttp) Write([]byte) (int, error) from ResponseWriter
//
// func (ctx *ContextHttp) WriteHeader(int) from ResponseWriter

// Implement request redirection, divided into internal redirection and external redirection.
//
// No host information is internally redirected, rerouted and processed.
//
// There is host information for external redirects, and the request returns redirect information.
//
// 实现请求重定向，分内部重定向和外部重定向。
//
// 无主机信息为内部重定向，重新路由并处理。
//
// 有主机信息为外部重定向，请求返回重定向信息。
func (ctx *ContextHttp) Redirect(code int, url string) {
	Redirect(ctx, url, code)
}

func (ctx *ContextHttp) Push(target string, opts *PushOptions) error {
	if opts == nil {
		opts = &PushOptions{
			Header: http.Header{
				HeaderAcceptEncoding: ctx.RequestReader.Header()[HeaderAcceptEncoding],
			},
		}
	}
	err := ctx.ResponseWriter.Push(target, opts)
	if err != nil {
		ctx.Debug(fmt.Sprintf("Failed to push: %v, Resource path: %s.", err, target))
	}
	return err
}

func (ctx *ContextHttp) WriteView(path string,i interface{}) error {	
	if i == nil {
		i = ctx.keys
	}
	ctx.ResponseWriter.Header().Add(HeaderContentType, MimeTextHTMLCharsetUtf8)
	return ctx.app.View.ExecuteTemplate(ctx.ResponseWriter, path, i)
}

func (ctx *ContextHttp) WriteString(i string) (err error) {
	_, err = ctx.Write(*(*[]byte)(unsafe.Pointer(&i)))
	return 
}

func (ctx *ContextHttp) WriteJson(i interface{}) error {
	return ctx.WriteRenderWith(i, RendererJson)
}

func (ctx *ContextHttp) WriteXml(i interface{}) error {
	return ctx.WriteRenderWith(i, RendererXml)
}

func (ctx *ContextHttp) WriteFile(path string) (int, error) {
	n, err := ServeFile(ctx, path)
	if err != nil {
		ctx.Fatal(err)
	}
	return n, err
}



func (ctx *ContextHttp) ReadBind(i interface{}) error {
	if i == nil {
		if ctx.keys == nil {
			ctx.keys = make(map[interface{}]interface{})
		}
		i = ctx.keys
	}
	return ctx.app.Binder.Bind(ctx.Request(), i)
}


func (ctx *ContextHttp) WriteRender(i interface{}) error {
	var r Renderer
	for _, accept := range strings.Split(ctx.GetHeader(HeaderAccept), ",") {
		if accept[0] == ' ' {
			accept = accept[1:]
		}
		switch accept {
		case MimeApplicationJson:
			r = RendererJson
		case MimeApplicationXml, MimeTextXml:
			r = RendererXml
		case MimeTextPlain, MimeText:
			r = RendererText
		case MimeTextHTML:
			temp := ctx.GetParam(ParamTemplate)
			if len(temp) > 0 {
				return ctx.WriteView(temp, i)
			}
		default:
			// return fmt.Errorf("undinf accept: %v", c.GetHeader(HeaderAccept))
		}
		if r != nil {
			break
		}
	}
	if r == nil {
		r = RendererText

	}
	return ctx.WriteRenderWith(i, r)
}

func (ctx *ContextHttp) WriteRenderWith(i interface{}, r Renderer) error {
	if i == nil {
		i = ctx.keys
	}
	header := ctx.ResponseWriter.Header()
	if val := header[HeaderContentType]; len(val) == 0 {
		header.Add(HeaderContentType, r.ContentType())
	}
	err := r.Render(ctx.ResponseWriter, i)
	if err != nil {
		ctx.Fatal(err)
	}
	return err
}

// logger
func (ctx *ContextHttp) Debug(args ...interface{}) {
	ctx.logReset().Debug(fmt.Sprint(args...))
}
func (ctx *ContextHttp) Info(args ...interface{}) {
	ctx.logReset().Info(fmt.Sprint(args...))
}
func (ctx *ContextHttp) Warning(args ...interface{}) {
	ctx.logReset().Warning(fmt.Sprint(args...))
}
func (ctx *ContextHttp) Error(args ...interface{}) {
	ctx.logReset().Error(fmt.Sprint(args...))
}

func (ctx *ContextHttp) Fatal(args ...interface{}) {
	NewEntryContext(ctx, ctx.app.Logger).Fatal(fmt.Sprint(args...))
}

func (ctx *ContextHttp) logReset() LogOut {
	file, line := LogFormatFileLine(0)
	f := Fields{
		HeaderXRequestID:	ctx.GetHeader(HeaderXRequestID),
		"file":				file,
		"line":				line,
	}
	return ctx.log.WithFields(f)
}

func (ctx *ContextHttp) WithField(key string, value interface{}) LogOut {
	return NewEntryContext(ctx, ctx.app.Logger).WithField(key, value)
}

func (ctx *ContextHttp) WithFields(fields Fields) LogOut {
	return NewEntryContext(ctx, ctx.app.Logger).WithFields(fields)
}



func (ctx *ContextHttp) App() *App {
	return ctx.app
}
















func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
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
	'!':  true,
	'#':  true,
	'$':  true,
	'%':  true,
	'&':  true,
	'\'': true,
	'*':  true,
	'+':  true,
	'-':  true,
	'.':  true,
	'0':  true,
	'1':  true,
	'2':  true,
	'3':  true,
	'4':  true,
	'5':  true,
	'6':  true,
	'7':  true,
	'8':  true,
	'9':  true,
	'A':  true,
	'B':  true,
	'C':  true,
	'D':  true,
	'E':  true,
	'F':  true,
	'G':  true,
	'H':  true,
	'I':  true,
	'J':  true,
	'K':  true,
	'L':  true,
	'M':  true,
	'N':  true,
	'O':  true,
	'P':  true,
	'Q':  true,
	'R':  true,
	'S':  true,
	'T':  true,
	'U':  true,
	'W':  true,
	'V':  true,
	'X':  true,
	'Y':  true,
	'Z':  true,
	'^':  true,
	'_':  true,
	'`':  true,
	'a':  true,
	'b':  true,
	'c':  true,
	'd':  true,
	'e':  true,
	'f':  true,
	'g':  true,
	'h':  true,
	'i':  true,
	'j':  true,
	'k':  true,
	'l':  true,
	'm':  true,
	'n':  true,
	'o':  true,
	'p':  true,
	'q':  true,
	'r':  true,
	's':  true,
	't':  true,
	'u':  true,
	'v':  true,
	'w':  true,
	'x':  true,
	'y':  true,
	'z':  true,
	'|':  true,
	'~':  true,
}
