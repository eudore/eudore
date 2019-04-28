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
		Request() protocol.RequestReader
		Response() protocol.ResponseWriter
		SetRequest(protocol.RequestReader)
		SetResponse(protocol.ResponseWriter)
		SetHandler(HandlerFuncs)
		Next()
		End()
		NewRequest(string, string, io.Reader) (protocol.ResponseReader, error)
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
		GetQuery(string) string
		GetHeader(name string) string
		SetHeader(string, string)
		Cookies() []*Cookie
		GetCookie(name string) string
		SetCookie(cookie *SetCookie)
		SetCookieValue(string, string, int)


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
		WithField(key string, value interface{}) LogOut
		WithFields(fields Fields) LogOut
		// app
		App() *App
	}

	/* 实现Context接口 */
	ContextHttp struct {
		context.Context
		protocol.RequestReader
		protocol.ResponseWriter
		ParamsArray
		QueryUrl
		// run handler
		index		int
		handler		HandlerFuncs
		// data
		keys		map[interface{}]interface{}
		path 		string
		rawQuery	string
		cookies 	[]*Cookie
		isReadBody	bool
		postBody	[]byte
		// component
		app			*App
		log			Logger
		fields		Fields
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

// Convert nil to type *ContextHttp, detect ContextHttp object to implement Context interface
//
// 将nil强制类型转换成*ContextHttp，检测ContextHttp对象实现Context接口
var _ Context			=	(*ContextHttp)(nil)

// context
func (ctx *ContextHttp) Reset(pctx context.Context, w protocol.ResponseWriter, r protocol.RequestReader) {
	ctx.Context = pctx
	ctx.RequestReader = r
	ctx.ResponseWriter = w
	ctx.keys = nil
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

func (ctx *ContextHttp) Request() protocol.RequestReader {
	return ctx.RequestReader
}

func (ctx *ContextHttp) Response() protocol.ResponseWriter {
	return ctx.ResponseWriter
}

func (ctx *ContextHttp) SetRequest(r protocol.RequestReader) {
	ctx.RequestReader = r
}

func (ctx *ContextHttp) SetResponse(w protocol.ResponseWriter) {
	ctx.ResponseWriter = w
}

func (ctx *ContextHttp) SetHandler(fs HandlerFuncs) {
	ctx.index = -1
	ctx.handler = fs
}

func (ctx *ContextHttp) Next() {
	ctx.index++
	for ctx.index < len(ctx.handler) {
		ctx.handler[ctx.index](ctx)
		ctx.index++
	}
}

func (ctx *ContextHttp) End() {
	ctx.index = 0xff
}

func (ctx *ContextHttp) NewRequest(method, url string, body io.Reader) (protocol.ResponseReader, error) {
	// tr := &http2.Transport{
	// 	AllowHTTP: true, //充许非加密的链接
	// 	TLSClientConfig: &tls.Config{
	// 		InsecureSkipVerify: true,
	// 	},
	// }
	// httpClient := http.Client{Transport: tr}

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
	return &ctx.ParamsArray
}

func (ctx *ContextHttp) GetHeader(name string) string {
	return ctx.RequestReader.Header().Get(name)
}

func (ctx *ContextHttp) SetHeader(name string, val string) {
	ctx.ResponseWriter.Header().Set(name, val)
}

func (ctx *ContextHttp) Cookies() []*Cookie {
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

func (ctx *ContextHttp) SetCookie(cookie *SetCookie) {
	if v := cookie.String(); v != "" {
		ctx.ResponseWriter.Header().Add("Set-Cookie", v)
	}
}

func (ctx *ContextHttp) SetCookieValue(name, value string, maxAge int) {
	ctx.ResponseWriter.Header().Add("Set-Cookie", fmt.Sprintf("%s=%s; Max-Age=%d", name, value ,maxAge))
//	ctx.SetCookie(&http.Cookie{Name: name, Value: url.QueryEscape(value), Path: "/", MaxAge: maxAge})
}



// Implement request redirection.
//
// 实现请求重定向。
func (ctx *ContextHttp) Redirect(code int, url string) {
	HandlerRedirectExternal(ctx, url, code)
}

func (ctx *ContextHttp) Push(target string, opts *protocol.PushOptions) error {
	if opts == nil {
		opts = &protocol.PushOptions{
			Header: HeaderHttp{
				HeaderAcceptEncoding: []string{ctx.RequestReader.Header().Get(HeaderAcceptEncoding)},
			},
		}
	}
	// TODO: add opts
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

func (ctx *ContextHttp) WriteFile(path string) (err error) {
	err = HandlerFile(ctx, path)
	if err != nil {
		ctx.Fatal(err)
	}
	return
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
	// 空错误不处理
	if len(args) == 1 && args[0] == nil {
		return
	}
	ctx.logReset().Error(fmt.Sprint(args...))
}

func (ctx *ContextHttp) Fatal(args ...interface{}) {
	ctx.logReset().Error(fmt.Sprint(args...))
	// 结束Context
	ctx.WriteHeader(500)
	ctx.WriteRender(map[string]string{
		"status":	"500",
		"x-request-id":	ctx.RequestID(),
	})
	ctx.End()
}

func (ctx *ContextHttp) logReset() LogOut {
	file, line := LogFormatFileLine(0)
	ctx.fields[HeaderXRequestID] = ctx.GetHeader(HeaderXRequestID)
	ctx.fields["file"] = file
	ctx.fields["line"] = line
	return ctx.app.Logger.WithFields(ctx.fields)
}

func (ctx *ContextHttp) WithField(key string, value interface{}) LogOut {
	if ctx.fields == nil {
		ctx.fields = make(Fields)
	}
	ctx.fields[key] = value
	return ctx
}

func (ctx *ContextHttp) WithFields(fields Fields) LogOut {
	ctx.fields = fields
	return ctx
}


func (ctx *ContextHttp) App() *App {
	return ctx.app
}


func (ctx *ContextHttp) readCookies(line string) {
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
