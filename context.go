/*
Context

Context定义一个请求上下文

文件：context.go
*/
package eudore

import (
	"io"
	"fmt"
	"time"
	"unsafe"
	"strings"
	"context"
	"net/http"
	"net/url"
	"io/ioutil"
	// "crypto/tls"
	// "golang.org/x/net/http2"
	"github.com/eudore/eudore/protocol"
)

const sniffLen = 512
const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

type (
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
		NewRequest(string, string, io.Reader) (protocol.ResponseReader, error)

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

		// param header cookie session
		Params() Params
		GetParam(string) string
		SetParam(string, string)
		AddParam(string, string)
		GetQuery(string) string
		GetHeader(name string) string
		SetHeader(string, string)
		Cookies() []*Cookie
		GetCookie(name string) string
		SetCookie(cookie *SetCookie)
		SetCookieValue(string, string, int)
		GetSession() SessionData
		SetSession(SessionData)


		// response
		Write([]byte) (int, error)
		WriteHeader(int)
		Redirect(int, string)
		Push(string, *protocol.PushOptions) error
		// render writer 
		WriteString(string) error
		WriteView(string, interface{}) error
		WriteJson(interface{}) error
		WriteFile(string) error
		// binder and renderer
		ReadBind(interface{}) error
		WriteRender(interface{}) error

		// log LogOut interface
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
		WithField(key string, value interface{}) LogOut
		WithFields(fields Fields) LogOut
		// app
		App() *App
	}

	/* 实现Context接口 */
	ContextBase struct {
		protocol.RequestReader
		protocol.ResponseWriter
		ParamsArray
		QueryUrl
		// run handler
		index		int
		handler		HandlerFuncs
		// data
		ctx			context.Context
		path 		string
		rawQuery	string
		cookies 	[]*Cookie
		isReadBody	bool
		postBody	[]byte
		// component
		app			*App
		log			Logger
	}
	ParamsArray struct {
		Keys		[]string
		Vals		[]string
	}
	QueryUrl struct {
		keys		[]string
		vals		[]string
	}
)

// Convert nil to type *ContextBase, detect ContextBase object to implement Context interface
//
// 将nil强制类型转换成*ContextBase，检测ContextBase对象实现Context接口
var _ Context			=	(*ContextBase)(nil)


func NewContextBase(app *App) *ContextBase {
	return &ContextBase{
		app:	app,
	}
}

// context
func (ctx *ContextBase) Reset(pctx context.Context, w protocol.ResponseWriter, r protocol.RequestReader) {
	ctx.ctx = pctx
	ctx.RequestReader = r
	ctx.ResponseWriter = w
	// logger
	ctx.log = ctx.app.Logger

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
	
	// query
	err := ctx.QueryUrl.readQuery(ctx.rawQuery)
	if err != nil {
		ctx.Error(err)
	}
	// url parsm
	ctx.path, err = url.QueryUnescape(ctx.path)
	if err != nil {
		ctx.Error(err)
	}
	
	// body
	ctx.isReadBody = false
	ctx.postBody = ctx.postBody[0:0]

	// params
	ctx.ParamsArray.Keys = ctx.ParamsArray.Keys[0:0]
	ctx.ParamsArray.Vals = ctx.ParamsArray.Vals[0:0]
	// cookies
	ctx.cookies = ctx.cookies[0:0]
	ctx.readCookies(r.Header().Get(HeaderCookie))
}


func (ctx *ContextBase) Context() context.Context {
	return ctx.ctx
}
func (ctx *ContextBase) Request() protocol.RequestReader {
	return ctx.RequestReader
}

func (ctx *ContextBase) Response() protocol.ResponseWriter {
	return ctx.ResponseWriter
}

func (ctx *ContextBase) SetRequest(r protocol.RequestReader) {
	ctx.RequestReader = r
}

func (ctx *ContextBase) SetResponse(w protocol.ResponseWriter) {
	ctx.ResponseWriter = w
}

func (ctx *ContextBase) SetHandler(fs HandlerFuncs) {
	ctx.index = -1
	ctx.handler = fs
}

func (ctx *ContextBase) Next() {
	ctx.index++
	for ctx.index < len(ctx.handler) {
		ctx.handler[ctx.index](ctx)
		ctx.index++
	}
}

func (ctx *ContextBase) End() {
	ctx.index = 0xff
}

func (ctx *ContextBase) NewRequest(method, url string, body io.Reader) (protocol.ResponseReader, error) {
	// tr := &http2.Transport{
	// 	AllowHTTP: true, //充许非加密的链接
	// 	TLSClientConfig: &tls.Config{
	// 		InsecureSkipVerify: true,
	// 	},
	// }
	// httpClient := http.Client{Transport: tr}

	cctx, cancel := context.WithCancel(ctx.ctx)
	time.AfterFunc(5*time.Second, func() {
		cancel()
	})

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add(HeaderXRequestID, ctx.RequestID())
	req = req.WithContext(cctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	// check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("resp StatusCode: %d", resp.StatusCode)
	}
	return NewResponseReaderHttp(resp), err
}


func (ctx *ContextBase) Path() string {
	return ctx.path
}

func (ctx *ContextBase) RemoteAddr() string {
	xforward := ctx.RequestReader.Header().Get(HeaderXForwardedFor)
	if "" == xforward {
		return strings.SplitN(ctx.RequestReader.RemoteAddr(), ":", 2)[0]
	}
	return strings.SplitN(string(xforward), ",", 2)[0]
}

func (ctx *ContextBase) RequestID() string {
	return ctx.GetHeader(HeaderXRequestID)
}

func (ctx *ContextBase) Referer() string {
	return ctx.GetHeader(HeaderReferer)
}

func (ctx *ContextBase) ContentType() string {
	return ctx.GetHeader(HeaderContentType)
}

func (ctx *ContextBase) Istls() bool {
	return ctx.RequestReader.TLS() != nil
}

func (ctx *ContextBase) Body() []byte {
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



func (ctx *ContextBase) Params() Params {
	return &ctx.ParamsArray
}

func (ctx *ContextBase) GetHeader(name string) string {
	return ctx.RequestReader.Header().Get(name)
}

func (ctx *ContextBase) SetHeader(name string, val string) {
	ctx.ResponseWriter.Header().Set(name, val)
}

func (ctx *ContextBase) Cookies() []*Cookie {
	return ctx.cookies
}

func (ctx *ContextBase) GetCookie(name string) string {
	for _, ctx := range ctx.cookies {
		if ctx.Name == name {
			return ctx.Value	
		}
	}
	return ""
}

func (ctx *ContextBase) SetCookie(cookie *SetCookie) {
	if v := cookie.String(); v != "" {
		ctx.ResponseWriter.Header().Add("Set-Cookie", v)
	}
}

func (ctx *ContextBase) SetCookieValue(name, value string, maxAge int) {
	ctx.ResponseWriter.Header().Add("Set-Cookie", fmt.Sprintf("%s=%s; Max-Age=%d", name, value ,maxAge))
//	ctx.SetCookie(&http.Cookie{Name: name, Value: url.QueryEscape(value), Path: "/", MaxAge: maxAge})
}

func (ctx *ContextBase) GetSession() SessionData {
	return ctx.app.Session.SessionLoad(ctx)
} 
func (ctx *ContextBase) SetSession(sess SessionData) {
	ctx.app.Session.SessionSave(sess)
}



// Implement request redirection.
//
// 实现请求重定向。
func (ctx *ContextBase) Redirect(code int, url string) {
	HandlerRedirect(ctx, url, code)
}

// 实现http2 push
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
		ctx.Errorf("Failed to push: %v, Resource path: %s.", err, target)
	}
	return err
}

func (ctx *ContextBase) WriteView(path string,i interface{}) error {	
	if i == nil {
		i = ctx.keys
	}
	ctx.ResponseWriter.Header().Add(HeaderContentType, MimeTextHTMLCharsetUtf8)
	return ctx.app.View.ExecuteTemplate(ctx.ResponseWriter, path, i)
}

func (ctx *ContextBase) WriteString(i string) (err error) {
	_, err = ctx.Write(*(*[]byte)(unsafe.Pointer(&i)))
	return 
}

func (ctx *ContextBase) WriteJson(i interface{}) error {
	return ctx.WriteRenderWith(i, RendererJson)
}

func (ctx *ContextBase) WriteXml(i interface{}) error {
	return ctx.WriteRenderWith(i, RendererXml)
}

func (ctx *ContextBase) WriteFile(path string) (err error) {
	err = HandlerFile(ctx, path)
	if err != nil {
		ctx.Fatal(err)
	}
	return
}



func (ctx *ContextBase) ReadBind(i interface{}) error {
	return ctx.app.Binder.Bind(ctx, i)
}


func (ctx *ContextBase) WriteRender(i interface{}) error {
	var r Renderer
	for _, accept := range strings.Split(ctx.GetHeader(HeaderAccept), ",") {
		if accept != "" && accept[0] == ' ' {
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
			r = RendererText
		default:
			return fmt.Errorf("undinf accept: %v", ctx.GetHeader(HeaderAccept))
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

func (ctx *ContextBase) WriteRenderWith(i interface{}, r Renderer) error {
	if i == nil {
		i = ctx.keys
	}
	header := ctx.ResponseWriter.Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, r.ContentType())
	}
	err := r.Render(ctx.ResponseWriter, i)
	if err != nil {
		ctx.Fatal(err)
	}
	return err
}

// logger
func (ctx *ContextBase) Debug(args ...interface{}) {
	ctx.logReset().Debug(fmt.Sprint(args...))
}
func (ctx *ContextBase) Info(args ...interface{}) {
	ctx.logReset().Info(fmt.Sprint(args...))
}
func (ctx *ContextBase) Warning(args ...interface{}) {
	ctx.logReset().Warning(fmt.Sprint(args...))
}
func (ctx *ContextBase) Error(args ...interface{}) {
	// 空错误不处理
	if len(args) == 1 && args[0] == nil {
		return
	}
	ctx.logReset().Error(fmt.Sprint(args...))
}

func (ctx *ContextBase) Fatal(args ...interface{}) {
	ctx.logReset().Error(fmt.Sprint(args...))
	// 结束Context
	if ctx.ResponseWriter.Status() == 200 {
		ctx.WriteHeader(500)
		ctx.WriteRender(map[string]string{
			// "error":	fmt.Sprint(args...),
			"status":	"500",
			"x-request-id":	ctx.RequestID(),
		})
	}
	ctx.End()
}


func (ctx *ContextBase) Debugf(format string, args ...interface{}) {
	ctx.logReset().Debug(fmt.Sprintf(format, args...))
}
func (ctx *ContextBase) Infof(format string, args ...interface{}) {
	ctx.logReset().Info(fmt.Sprintf(format, args...))
}
func (ctx *ContextBase) Warningf(format string, args ...interface{}) {
	ctx.logReset().Warning(fmt.Sprintf(format, args...))
}

func (ctx *ContextBase) Errorf(format string, args ...interface{}) {
	ctx.logReset().Error(fmt.Sprintf(format, args...))
}

func (ctx *ContextBase) Fatalf(format string, args ...interface{}) {
	ctx.logReset().Error(fmt.Sprintf(format, args...))
	// 结束Context
	ctx.WriteHeader(500)
	ctx.WriteRender(map[string]string{
		"status":	"500",
		"x-request-id":	ctx.RequestID(),
	})
	ctx.End()
}

func (ctx *ContextBase) logReset() LogOut {
	fields := make(Fields)
	file, line := LogFormatFileLine(0)
	fields[HeaderXRequestID] = ctx.GetHeader(HeaderXRequestID)
	fields["file"] = file
	fields["line"] = line
	return ctx.log.WithFields(fields)
}

func (ctx *ContextBase) WithField(key string, value interface{}) LogOut {
	return ctx.logReset().WithField(key, value)
}

func (ctx *ContextBase) WithFields(fields Fields) LogOut {
	return ctx.log.WithFields(fields)
}


func (ctx *ContextBase) App() *App {
	return ctx.app
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
		ctx.cookies = append(ctx.cookies, &Cookie{Name: name, Value: val})
	}
}



func (p *ParamsArray) GetParam(key string) string {
	for i, str := range p.Keys {
		if str == key {
			return p.Vals[i]
		}
	}
	return ""
}

func (p *ParamsArray) AddParam(key string, val string) {
	p.Keys = append(p.Keys, key)
	p.Vals = append(p.Vals, val)
}

func (p *ParamsArray) SetParam(key string, val string) {
	for i, str := range p.Keys {
		if str == key {
			p.Vals[i] = val
			return
		}
	}
	p.AddParam(key, val)
}

func (q *QueryUrl) GetQuery(key string) string {
	for i, str := range q.keys {
		if str == key {
			return q.vals[i]
		}
	}
	return ""
}

func (q *QueryUrl) readQuery(query string) (err error) {
	q.keys = q.keys[0:0]
	q.vals = q.vals[0:0]
	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&;"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		q.keys = append(q.keys, key)
		q.vals = append(q.vals, value)
	}
	return err
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
