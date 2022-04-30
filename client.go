package eudore

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/*
Client 定义http协议客户端。
	使用http.Handler处理请求
	构建json form
	httptrace
*/
type Client interface {
	Do(*http.Request) (*http.Response, error)
}

// NewClientStd 函数创建一个net/http.Client。
//
// 如果DialContext的host为DefaultClientHost，上下文获取ContextKeyServer实现ServeConn方法会使用net.Pipe连接。
func NewClientStd(options ...interface{}) Client {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           newDialContext(),
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	// fix less 1.13
	// Set(client, "Transport.ForceAttemptHTTP2", true)
	for _, option := range options {
		switch val := option.(type) {
		case http.RoundTripper:
			ConvertTo(val, client.Transport)
		case func(http.RoundTripper) http.RoundTripper:
			tp := val(client.Transport)
			if tp != nil {
				client.Transport = tp
			}
		case http.CookieJar:
			client.Jar = val
		}
	}
	return client
}

func newDialContext() func(ctx context.Context, network, addr string) (net.Conn, error) {
	fn := (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if network == "tcp" && addr == DefaultClientHost {
			server, ok := ctx.Value(ContextKeyServer).(interface{ ServeConn(conn net.Conn) })
			if ok {
				serverConn, clientConn := net.Pipe()
				server.ServeConn(serverConn)
				return clientConn, nil
			}
		}

		return fn(ctx, network, addr)
	}
}

// ClientWarp 定义http客户端包装方法
type ClientWarp struct {
	Client  `alias:"client" json:"client"`
	Context context.Context `alias:"context" json:"context"`
	Querys  url.Values      `alias:"querys" json:"querys"`
	Headers http.Header     `alias:"headers" json:"headers"`
	Cookies http.CookieJar  `alias:"cookies" json:"cookies"`
}

// NewClientWarp 函数创建ClientWarp对象，提供http请求创建方法。
func NewClientWarp(options ...interface{}) *ClientWarp {
	client := &ClientWarp{
		Context: context.Background(),
		Querys:  make(url.Values),
		Headers: make(http.Header),
	}
	client.Cookies, _ = cookiejar.New(nil)
	client.Client = NewClientStd(append([]interface{}{client.Cookies}, options...)...)
	return client
}

// Mount 方法保存context.Context作为Client默认发起请求的context.Context
func (client *ClientWarp) Mount(ctx context.Context) {
	client.Context = ctx
}

func headerCopy(dst, src map[string][]string) map[string][]string {
	for key, vals := range src {
		dst[key] = append(dst[key], vals...)
	}
	return dst
}

// AddQuery 方法给客户端增加请求参数，客户端发起每个请求都会附加参数。
func (client *ClientWarp) AddQuery(key, val string) {
	client.Querys.Add(key, val)

}

// AddQuerys 方法给客户端增加请求参数，客户端发起每个请求都会附加参数。
func (client *ClientWarp) AddQuerys(querys url.Values) {
	client.Querys = headerCopy(client.Querys, querys)
}

// AddHeader 方法给客户端增加Header，客户端发起每个请求都会附加Header。
func (client *ClientWarp) AddHeader(key, val string) {
	client.Headers.Add(key, val)
}

// AddHeaders 方法给客户端增加Header，客户端发起每个请求都会附加Header。
func (client *ClientWarp) AddHeaders(headers http.Header) {
	headerCopy(client.Headers, headers)
}

// AddBasicAuth 方法给客户端添加Basic Auth信息。
func (client *ClientWarp) AddBasicAuth(name, pass string) {
	client.Headers.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(name+":"+pass)))
}

// AddCookie 方法给指定host下添加cookie。
func (client *ClientWarp) AddCookie(host, key, val string) {
	u, err := url.Parse(host)
	if err != nil {
		return
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" {
		u.Host = DefaultClientHost
	}
	if u.Path == "" {
		u.Path = "/"
	}
	client.Cookies.SetCookies(u, []*http.Cookie{{Name: key, Value: val}})
}

// GetCookie 方法读取指定host的cookie。
func (client *ClientWarp) GetCookie(host, key string) string {
	u, err := url.Parse(host)
	if err != nil {
		return ""
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" {
		u.Host = DefaultClientHost
	}
	if u.Path == "" {
		u.Path = "/"
	}
	for _, cookie := range client.Cookies.Cookies(u) {
		if cookie.Name == key {
			return cookie.Value
		}
	}
	return ""
}

// NewRequest 方法从客户端创建http请求。
//
// 如果schme为空默认为http，如果host为空默认为eudore.DefaultClientHost(go1.9 运行将异常阻塞)
func (client *ClientWarp) NewRequest(method string, path string) *RequestWriterWarp {
	u, err := url.Parse(path)
	if err == nil {
		if u.Scheme == "" {
			u.Scheme = "http"
		}
		if u.Host == "" {
			u.Host = DefaultClientHost
		}
		path = u.String()
	}
	req, _ := http.NewRequest(method, path, nil)
	req = req.WithContext(client.Context)
	headerCopy(req.Header, client.Headers)
	req.URL.RawQuery = url.Values(headerCopy(req.URL.Query(), client.Querys)).Encode()
	return &RequestWriterWarp{
		Request: req,
		Client:  client,
	}
}

// RequestWriterWarp 定义http请求对象。
type RequestWriterWarp struct {
	*http.Request
	Client *ClientWarp
}

// Do 方法发生请求。
func (req *RequestWriterWarp) Do() ResponseReader {
	resp, err := req.Client.Do(req.Request)
	return &responseReaderHttp{
		Request:  req.Request,
		Response: resp,
		Error:    err,
	}
}

// AddQuery 方法给请求添加url参数。
func (req *RequestWriterWarp) AddQuery(key, val string) *RequestWriterWarp {
	querys := req.URL.Query()
	querys.Add(key, val)
	req.URL.RawQuery = querys.Encode()
	return req
}

// AddHeaders 方法给请求添加headers。
func (req *RequestWriterWarp) AddHeaders(headers http.Header) *RequestWriterWarp {
	headerCopy(req.Header, headers)
	return req
}

// AddHeader 方法给请求添加header。
func (req *RequestWriterWarp) AddHeader(key, val string) *RequestWriterWarp {
	if strings.ToLower(key) == "host" {
		req.Host = val
	}
	req.Header.Add(key, val)
	return req
}

// Body 方法格局对象类型创建http请求body。
func (req *RequestWriterWarp) Body(i interface{}) *RequestWriterWarp {
	switch body := i.(type) {
	case string:
		req.BodyString(body)
	case []byte:
		req.BodyBytes(body)
	case io.ReadCloser:
		req.Request.Body = body
		req.initGetBody()
	case io.Reader:
		req.Request.Body = ioutil.NopCloser(body)
		req.initGetBody()
	default:
		req.BodyJSON(body)
	}
	return req
}

func (req *RequestWriterWarp) initGetBody() {
	req.GetBody = func() (io.ReadCloser, error) {
		body, ok := req.Request.Body.(interface{ Clone() io.ReadCloser })
		if ok {
			return body.Clone(), nil
		}
		return http.NoBody, nil
	}
}

// BodyBytes 方法使用[]byte创建body。
func (req *RequestWriterWarp) BodyBytes(b []byte) *RequestWriterWarp {
	if req.Request.Body == nil {
		req.Request.Body = &bodyBuffer{}
		req.initGetBody()
	}
	body, ok := req.Request.Body.(*bodyBuffer)
	if ok {
		body.Write(b)
		req.ContentLength = int64(body.Len())
	}
	return req
}

// BodyString 方法使用string创建body。
func (req *RequestWriterWarp) BodyString(s string) *RequestWriterWarp {
	if req.Request.Body == nil {
		req.Request.Body = &bodyBuffer{}
		req.initGetBody()
	}
	body, ok := req.Request.Body.(*bodyBuffer)
	if ok {
		body.WriteString(s)
		req.ContentLength = int64(body.Len())
	}
	return req
}

type bodyBuffer struct {
	bytes.Buffer
}

func (body *bodyBuffer) Clone() io.ReadCloser {
	return &bodyBuffer{*bytes.NewBuffer(body.Bytes())}
}

func (body *bodyBuffer) Close() error {
	return nil
}

// BodyJSON 方法使用任意类型创建json请求body。
func (req *RequestWriterWarp) BodyJSON(data interface{}) *RequestWriterWarp {
	if req.Request.Body == nil && req.Header.Get("Content-Type") == "" {
		req.ContentLength = -1
		req.Header.Add("Content-Type", "application/json")
		body, ok := data.(map[string]interface{})
		if ok {
			req.Request.Body = &bodyJSON{values: body}
		} else {
			req.Request.Body = &bodyJSON{data: data}
		}
		req.initGetBody()
	}
	return req
}

// BodyJSONValue 方法使用json键值创建json请求body。
func (req *RequestWriterWarp) BodyJSONValue(key string, val interface{}) *RequestWriterWarp {
	if req.Request.Body == nil && req.Header.Get("Content-Type") == "" {
		req.ContentLength = -1
		req.Header.Add("Content-Type", "application/json")
		req.Request.Body = &bodyJSON{values: make(map[string]interface{})}
		req.initGetBody()
	}
	body, ok := req.Request.Body.(*bodyJSON)
	if ok {
		body.values[key] = val
	}
	return req
}

type bodyJSON struct {
	reader *io.PipeReader
	writer *io.PipeWriter
	data   interface{}
	values map[string]interface{}
}

func (body *bodyJSON) Clone() io.ReadCloser {
	return &bodyJSON{
		data:   body.data,
		values: body.values,
	}
}

func (body *bodyJSON) Read(p []byte) (n int, err error) {
	if body.reader == nil {
		body.reader, body.writer = io.Pipe()
		go func() {
			if body.data != nil {
				json.NewEncoder(body.writer).Encode(body.data)
			} else {
				json.NewEncoder(body.writer).Encode(body.values)
			}
			body.writer.Close()
		}()
	}
	return body.reader.Read(p)
}

func (body *bodyJSON) Close() error {
	return body.reader.Close()
}

func (req *RequestWriterWarp) initbodyForm() {
	if req.Request.Body == nil && req.Header.Get("Content-Type") == "" {
		var buf [30]byte
		io.ReadFull(rand.Reader, buf[:])
		boundary := fmt.Sprintf("%x", buf[:])
		req.ContentLength = -1
		req.Header.Add("Content-Type", "multipart/form-data; boundary="+boundary)
		req.Request.Body = &bodyForm{
			Boundary: boundary,
			Values:   make(map[string][]string),
			Files:    make(map[string][]fileContent),
		}
		req.initGetBody()
	}
}

// BodyFormValue 方法给form添加键值数据。
func (req *RequestWriterWarp) BodyFormValue(key, val string) *RequestWriterWarp {
	req.initbodyForm()
	body, ok := req.Request.Body.(*bodyForm)
	if ok {
		body.Values[key] = append(body.Values[key], val)
	}
	return req
}

// BodyFormValues 方法给form添加键值数据。
func (req *RequestWriterWarp) BodyFormValues(data map[string][]string) *RequestWriterWarp {
	req.initbodyForm()
	body, ok := req.Request.Body.(*bodyForm)
	if ok {
		for key, vals := range data {
			body.Values[key] = append(body.Values[key], vals...)
		}
	}
	return req
}

// BodyFormFile 方法给form添加文件内容。
func (req *RequestWriterWarp) BodyFormFile(key, name string, val interface{}) *RequestWriterWarp {
	req.initbodyForm()
	body, ok := req.Request.Body.(*bodyForm)
	if ok {
		var content fileContent
		switch body := val.(type) {
		case []byte:
			content.Body = body
		case string:
			content.Body = []byte(body)
		case io.ReadCloser:
			content.Reader = body
		case io.Reader:
			content.Reader = ioutil.NopCloser(body)
		default:
			return req
		}
		content.Name = name
		body.Files[key] = append(body.Files[key], content)
	}
	return req
}

// BodyFormLocalFile 方法给form添加本地文件内容。
func (req *RequestWriterWarp) BodyFormLocalFile(key, name, path string) *RequestWriterWarp {
	req.initbodyForm()
	body, ok := req.Request.Body.(*bodyForm)
	if ok {
		if name == "" {
			name = filepath.Base(path)
		}
		body.Files[key] = append(body.Files[key], fileContent{
			Name: name,
			File: path,
		})
	}
	return req
}

type bodyForm struct {
	reader   *io.PipeReader
	writer   *io.PipeWriter
	Boundary string
	Values   map[string][]string
	Files    map[string][]fileContent
}

type fileContent struct {
	Name   string
	Body   []byte
	File   string
	Reader io.ReadCloser
}

func (body *bodyForm) Clone() io.ReadCloser {
	return &bodyForm{
		Boundary: body.Boundary,
		Values:   body.Values,
		Files:    body.Files,
	}
}

func (body *bodyForm) Read(p []byte) (n int, err error) {
	if body.reader == nil {
		body.reader, body.writer = io.Pipe()
		w := multipart.NewWriter(body.writer)
		w.SetBoundary(body.Boundary)
		go func() {
			for key, vals := range body.Values {
				for _, val := range vals {
					w.WriteField(key, val)
				}
			}
			for key, vals := range body.Files {
				for _, val := range vals {
					part, _ := w.CreateFormFile(key, val.Name)
					switch {
					case val.Body != nil:
						part.Write(val.Body)
					case val.Reader != nil:
						io.Copy(part, val.Reader)
						val.Reader.Close()
					case val.File != "":
						file, err := os.Open(val.File)
						if err == nil {
							io.Copy(part, file)
							file.Close()
						}
					}
				}
			}
			w.Close()
			body.writer.Close()
		}()
	}
	return body.reader.Read(p)
}

func (body *bodyForm) Close() error {
	return body.reader.Close()
}

// ResponseReader 定义响应读取方法。
type ResponseReader interface {
	Proto() string
	Status() int
	Reason() string
	Header() http.Header
	Cookies() []*http.Cookie
	Read([]byte) (int, error)
	Body() []byte
	Err() error
	Callback(...ResponseReaderCheck)
}

// ResponseReaderCheck 定义响应检查方法。
type ResponseReaderCheck = func(ResponseReader, *http.Request, Logger) error

type responseReaderHttp struct {
	Request   *http.Request
	Response  *http.Response
	Error     error
	BodyBytes []byte
}

// Proto 方法获取响应协议版本。
func (resp *responseReaderHttp) Proto() string {
	return resp.Response.Proto
}

// Status 方法获取响应状态码。
func (resp *responseReaderHttp) Status() int {
	return resp.Response.StatusCode
}

// Reason 方法获取响应描述文本。
func (resp *responseReaderHttp) Reason() string {
	return resp.Response.Status
}

// Header 方法获取响应header。
func (resp *responseReaderHttp) Header() http.Header {
	return resp.Response.Header
}

// Cookies 方法获取响应cookies。
func (resp *responseReaderHttp) Cookies() []*http.Cookie {
	return resp.Response.Cookies()
}

// Read 方法读取响应body。
func (resp *responseReaderHttp) Read(b []byte) (int, error) {
	return resp.Response.Body.Read(b)
}

// Body 方法获取响应body内容。
func (resp *responseReaderHttp) Body() []byte {
	if resp.BodyBytes == nil {
		// TODO: isgzip
		bts, err := ioutil.ReadAll(resp.Response.Body)
		if err != nil {
			resp.BodyBytes = make([]byte, 0)
			resp.Error = err
			return nil
		}
		resp.BodyBytes = bts
		resp.Response.Body.Close()
	}
	resp.Response.Body = ioutil.NopCloser(bytes.NewReader(resp.BodyBytes))
	return resp.BodyBytes
}

// Err 方法返回响应error。
func (resp *responseReaderHttp) Err() error {
	return resp.Error
}

// Callback 方法执行响应检查函数。
func (resp *responseReaderHttp) Callback(calls ...ResponseReaderCheck) {
	log := clientLogger(resp.Request)
	if resp.Error == nil {
		for _, call := range calls {
			err := call(resp, resp.Request, log)
			if err != nil {
				resp.Error = err
				log.WithFields([]string{"method", "path"}, []interface{}{resp.Request.Method, resp.Request.URL.Path}).Error(resp.Error.Error())
				return
			}
		}
	}
}

func clientLogger(req *http.Request) Logger {
	log, ok := req.Context().Value(ContextKeyApp).(Logger)
	if ok {
		// TODO: x-id
		return log
	}
	return DefaultLoggerNull
}

// NewResponseReaderCheckStatus 函数检查响应状态码
func NewResponseReaderCheckStatus(status int) ResponseReaderCheck {
	return func(resp ResponseReader, req *http.Request, log Logger) error {
		if status != resp.Status() {
			return fmt.Errorf("ResponseReader check status %d error: status is %d", status, resp.Status())
		}
		return nil
	}
}

// NewResponseReaderCheckBody 函数检查响应body是否包含状态码。
func NewResponseReaderCheckBody(str string) ResponseReaderCheck {
	return func(resp ResponseReader, req *http.Request, log Logger) error {
		body := resp.Body()
		if !strings.Contains(string(body), str) {
			return fmt.Errorf("ResponseReader check body key %s not found,body prefix %s", str, body[0:64])
		}
		return resp.Err()
	}
}

// NewResponseReaderOutBody 函数输出响应body内容。
func NewResponseReaderOutBody() ResponseReaderCheck {
	return func(resp ResponseReader, req *http.Request, log Logger) error {
		log.Infof("%s", resp.Body())
		return resp.Err()
	}
}

// NewResponseReaderOutHead 函数输出响应状态行。
func NewResponseReaderOutHead() ResponseReaderCheck {
	return func(resp ResponseReader, req *http.Request, log Logger) error {
		log.WithField("header", resp.Header()).Infof("%s %d %s", resp.Proto(), resp.Status(), resp.Reason())
		return resp.Err()
	}
}
