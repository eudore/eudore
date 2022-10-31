package eudore

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

// Client 定义http客户端接口，构建并发送http请求。
type Client interface {
	NewRequest(context.Context, string, string, ...interface{}) error
	WithClient(...interface{}) Client
	GetClient() *http.Client
}

// ClientRequestOption 定义http请求选项。
type ClientRequestOption func(*http.Request)

// ClientResponseOption 定义http响应选项。
type ClientResponseOption func(*http.Response) error

// clientStd 定义http客户端默认实现。
type clientStd struct {
	Context context.Context
	Client  *http.Client
	Options []interface{}
}

// NewClientStd 函数创建默认http客户端实现，参数为默认选项。
func NewClientStd(options ...interface{}) Client {
	return &clientStd{
		Context: context.Background(),
		Client: &http.Client{
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				DialContext:           newDialContext(),
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		Options: options,
	}
}

// newDialContext 函数创建http客户端Dial函数，如果是内部请求Host，从环境上下文获取到Server处理连接。
func newDialContext() func(ctx context.Context, network, addr string) (net.Conn, error) {
	fn := (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if network == "tcp" && addr == DefaultClientInternalHost {
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

// Mount 方法保存context.Context作为Client默认发起请求的context.Context
func (client *clientStd) Mount(ctx context.Context) {
	client.Context = ctx
}

// NewRequest 方法发送http请求。
func (client *clientStd) NewRequest(ctx context.Context, method string, path string, options ...interface{}) error {
	if ctx == nil {
		ctx = client.Context
	}

	req, err := http.NewRequestWithContext(ctx, method, initRequestPath(ctx, path), nil)
	if err != nil {
		return err
	}
	ro, wo := initRequestOptions(req, append(client.Options, options...))
	for i := range ro {
		ro[i](req)
	}

	resp, err := client.Client.Do(req)
	if err != nil {
		return err
	}
	for i := range wo {
		err = wo[i](resp)
		if err != nil {
			return err
		}
	}
	return nil
}

// initRequestPath 函数初始化请求url，如果Host为空设置默认或内部Host，如果请求协议为空设置为http。
func initRequestPath(ctx context.Context, path string) string {
	u, err := url.ParseRequestURI(path)
	if err != nil {
		return path
	}

	if u.Host == "" {
		_, ok := ctx.Value(ContextKeyServer).(interface{ ServeConn(conn net.Conn) })
		if ok {
			u.Host = DefaultClientInternalHost
		} else {
			u.Host = DefaultClientHost
		}
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	return u.String()
}

// initRequestOptions 函数初始化请求选项，并返回全部ClientRequestOption和ClientResponseOption。
func initRequestOptions(r *http.Request, options []interface{}) (ro []ClientRequestOption, wo []ClientResponseOption) {
	for i := range options {
		switch option := options[i].(type) {
		case io.ReadCloser:
			r.Body = option
			r.GetBody = initGetBody(r.Body)
		case io.Reader:
			r.Body = ioutil.NopCloser(option)
			r.GetBody = initGetBody(r.Body)
		case string:
			ro = append(ro, NewClientBodyString(option))
		case []byte:
			ro = append(ro, NewClientBodyString(string(option)))
		case http.Header:
			headerCopy(r.Header, option)
		case *http.Cookie:
			r.AddCookie(option)
		case url.Values:
			v, err := url.ParseQuery(r.URL.RawQuery)
			if err == nil {
				headerCopy(v, option)
				r.URL.RawQuery = v.Encode()
			}
		case ClientRequestOption:
			ro = append(ro, option)
		case func(*http.Request):
			ro = append(ro, option)
		case ClientResponseOption:
			wo = append(wo, option)
		case func(*http.Response) error:
			wo = append(wo, option)
		default:
			ro = append(ro, NewClientBody(option))
		}
	}
	return
}

// WithClient 方法给客户端追加新的选项，返回客户端深拷贝。
func (client *clientStd) WithClient(options ...interface{}) Client {
	options = append(client.Options, options...)
	return &clientStd{
		Context: client.Context,
		Client:  client.Client,
		Options: options,
	}
}

// GetClient 方法返回*http.Client对象，用于修改属性。
func (client *clientStd) GetClient() *http.Client {
	return client.Client
}

// NewClientHost 数创建请求选项修改Host。
func NewClientHost(host string) ClientRequestOption {
	return func(r *http.Request) {
		r.Host = host
		r.Header.Set("Host", host)
	}
}

// NewClientQuery 函数创建请求选项追加请求参数。
func NewClientQuery(key, val string) ClientRequestOption {
	return func(r *http.Request) {
		v, err := url.ParseQuery(r.URL.RawQuery)
		if err == nil {
			v.Add(key, val)
			r.URL.RawQuery = v.Encode()
		}
	}
}

// NewClientQuerys 函数创建请求选项加请求参数。
func NewClientQuerys(querys url.Values) ClientRequestOption {
	return func(r *http.Request) {
		v, err := url.ParseQuery(r.URL.RawQuery)
		if err == nil {
			headerCopy(v, querys)
			r.URL.RawQuery = v.Encode()
		}
	}
}

// NewClientHeader 函数创建请求选项追加Header。
func NewClientHeader(key, val string) ClientRequestOption {
	return func(r *http.Request) {
		r.Header.Add(key, val)
	}
}

// NewClientHeaders 函数创建请求选项追加Header。
func NewClientHeaders(headers http.Header) ClientRequestOption {
	return func(r *http.Request) {
		headerCopy(r.Header, headers)
	}
}

// NewClientCookie 函数创建请求选项追加请求Cookie
func NewClientCookie(key, val string) ClientRequestOption {
	return func(r *http.Request) {
		r.AddCookie(&http.Cookie{Name: key, Value: val})
	}
}

// NewClientBasicAuth 函数创建请求选项追加BasicAuth用户信息。
func NewClientBasicAuth(username, password string) ClientRequestOption {
	return func(r *http.Request) {
		r.SetBasicAuth(username, password)
	}
}

// NewClientBody 方法追加请求Body字符串。
func NewClientBody(data interface{}) ClientRequestOption {
	return func(r *http.Request) {
		var contenttype string
		switch reflect.Indirect(reflect.ValueOf(data)).Kind() {
		case reflect.Struct:
			contenttype = r.Header.Get(HeaderContentType)
			if contenttype == "" {
				contenttype = DefaultClientBodyContextType
			}
		case reflect.Slice, reflect.Map:
			contenttype = MimeApplicationJSON
		default:
			return
		}
		switch contenttype {
		case MimeApplicationJSON, MimeApplicationJSONCharsetUtf8:
			r.Body = &bodyEncoder{data: data, contenttype: MimeApplicationJSON}
		case MimeApplicationXML, MimeApplicationXMLCharsetUtf8:
			r.Body = &bodyEncoder{data: data, contenttype: MimeApplicationXML}
		case MimeApplicationProtobuf:
			r.Body = &bodyEncoder{data: data, contenttype: MimeApplicationProtobuf}
		default:
			return
		}
		r.GetBody = initGetBody(r.Body)
	}
}

// NewClientBodyString 方法追加请求Body字符串。
func NewClientBodyString(str string) ClientRequestOption {
	return func(r *http.Request) {
		if r.Body == nil {
			r.Body = &bodyBuffer{}
			r.GetBody = initGetBody(r.Body)
		}
		body, ok := r.Body.(*bodyBuffer)
		if ok {
			body.WriteString(str)
			r.ContentLength = int64(body.Len())
		}
	}
}

// NewClientBodyJSON 方法追加请求json值或json对象。
func NewClientBodyJSON(data interface{}) ClientRequestOption {
	return func(r *http.Request) {
		r.ContentLength = -1
		r.Header.Add(HeaderContentType, "application/json")
		body, ok := data.(map[string]interface{})
		if ok {
			r.Body = &bodyJSON{values: body}
		} else {
			r.Body = &bodyJSON{data: data}
		}
		r.GetBody = initGetBody(r.Body)
	}
}

// NewClientBodyJSONValue 方法追加请求json值。
func NewClientBodyJSONValue(key string, val interface{}) ClientRequestOption {
	return func(r *http.Request) {
		if r.Body == nil && r.Header.Get(HeaderContentType) == "" {
			r.ContentLength = -1
			r.Header.Add(HeaderContentType, "application/json")
			r.Body = &bodyJSON{values: make(map[string]interface{})}
			r.GetBody = initGetBody(r.Body)
		}
		body, ok := r.Body.(*bodyJSON)
		if ok {
			body.values[key] = val
		}
	}
}

// NewClientBodyFormValue 方法追加请求Form值。
func NewClientBodyFormValue(key string, val string) ClientRequestOption {
	return func(r *http.Request) {
		initBodyForm(r)
		body, ok := r.Body.(*bodyForm)
		if ok {
			body.Values[key] = append(body.Values[key], val)
		}
	}
}

// NewClientBodyFormValues 方法追加请求Form值。
func NewClientBodyFormValues(data map[string]string) ClientRequestOption {
	return func(r *http.Request) {
		initBodyForm(r)
		body, ok := r.Body.(*bodyForm)
		if ok {
			for key, val := range data {
				body.Values[key] = append(body.Values[key], val)
			}
		}
	}
}

// NewClientBodyFormFile 方法给form添加文件内容，文件类型可以为[]byte string io.ReadCloser io.Reader。
func NewClientBodyFormFile(key, name string, val interface{}) ClientRequestOption {
	return func(r *http.Request) {
		initBodyForm(r)
		body, ok := r.Body.(*bodyForm)
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
				return
			}
			content.Name = name
			body.Files[key] = append(body.Files[key], content)
		}
	}
}

// NewClientBodyFormLocalFile 方法给form添加本地文件内容。
func NewClientBodyFormLocalFile(key, name, path string) ClientRequestOption {
	return func(r *http.Request) {
		initBodyForm(r)
		body, ok := r.Body.(*bodyForm)
		if ok {
			if name == "" {
				name = filepath.Base(path)
			}
			body.Files[key] = append(body.Files[key], fileContent{
				Name: name,
				File: path,
			})
		}
	}
}

// initBodyForm
func initBodyForm(r *http.Request) {
	if r.Body == nil && r.Header.Get(HeaderContentType) == "" {
		var buf [30]byte
		io.ReadFull(rand.Reader, buf[:])
		boundary := fmt.Sprintf("%x", buf[:])
		r.ContentLength = -1
		r.Header.Add(HeaderContentType, "multipart/form-data; boundary="+boundary)
		r.Body = &bodyForm{
			Boundary: boundary,
			Values:   make(map[string][]string),
			Files:    make(map[string][]fileContent),
		}
		r.GetBody = initGetBody(r.Body)
	}
}

// initGetBody 函数创建http.GetBody函数，如果body实现Clone() io.ReadCloser方法，在调用GetBody时使用。
func initGetBody(data interface{}) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		body, ok := data.(interface{ Clone() io.ReadCloser })
		if ok {
			return body.Clone(), nil
		}
		return http.NoBody, nil
	}
}

// NewClientTimeout 函数创建请求选项设置请求超时时间。
func NewClientTimeout(timeout time.Duration) ClientRequestOption {
	return func(r *http.Request) {
		ctx, _ := context.WithTimeout(r.Context(), timeout)
		*r = *r.WithContext(ctx)
	}
}

// NewClientTrace 函数创建请求选项在请求上下文保存ClientTrace对象和httptrace.ClientTrace，实现http客户端追踪。
func NewClientTrace() ClientRequestOption {
	return func(r *http.Request) {
		trace := &ClientTrace{
			HTTPStart:    time.Now(),
			WroteHeaders: make(http.Header),
		}
		ctx := context.WithValue(r.Context(), ContextKeyClientTrace, trace)
		ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
			DNSStart: func(info httptrace.DNSStartInfo) { trace.DNSStart = time.Now(); trace.DNSHost = info.Host },
			DNSDone:  func(info httptrace.DNSDoneInfo) { trace.DNSDone = time.Now(); trace.DNSAddrs = info.Addrs },
			ConnectStart: func(network, addr string) {
				trace.ConnectStart = time.Now()
				trace.ConnectNetwork = network
				trace.ConnectAddress = addr
			},
			ConnectDone:          func(string, string, error) { trace.ConnectDone = time.Now() },
			GetConn:              func(hostPort string) { trace.GetConn = time.Now(); trace.GetConnHostPort = hostPort },
			GotConn:              func(httptrace.GotConnInfo) { trace.GotConn = time.Now() },
			GotFirstResponseByte: func() { trace.GotFirstResponseByte = time.Now() },
			TLSHandshakeStart:    func() { trace.TLSHandshakeStart = time.Now() },
			TLSHandshakeDone: func(state tls.ConnectionState, _ error) {
				trace.TLSHandshakeDone = time.Now()
				trace.TLSHandshakeState = &state
			},
			WroteHeaderField: func(key string, value []string) { trace.WroteHeaders[key] = value },
		})
		*r = *r.WithContext(ctx)
	}
}

// ClientTrace 定义http客户端请求追踪记录的数据
type ClientTrace struct {
	HTTPStart             time.Time            `json:"http-start" xml:"http-start"`
	HTTPDone              time.Time            `json:"http-done" xml:"http-done"`
	HTTPDuration          time.Duration        `json:"http-duration" xml:"http-duration"`
	DNSStart              time.Time            `json:"dns-start,omitempty" xml:"dns-start,omitempty"`
	DNSDone               time.Time            `json:"dns-done,omitempty" xml:"dns-done,omitempty"`
	DNSDuration           time.Duration        `json:"dns-duration,omitempty" xml:"dns-duration,omitempty"`
	DNSHost               string               `json:"dns-host,omitempty" xml:"dns-host,omitempty"`
	DNSAddrs              []net.IPAddr         `json:"dns-addrs,omitempty" xml:"dns-addrs,omitempty"`
	ConnectStart          time.Time            `json:"connect-start,omitempty" xml:"connect-start,omitempty"`
	ConnectDone           time.Time            `json:"connect-done,omitempty" xml:"connect-done,omitempty"`
	ConnectDuration       time.Duration        `json:"connect-duration,omitempty" xml:"connect-duration,omitempty"`
	ConnectNetwork        string               `json:"connect-network,omitempty" xml:"connect-network,omitempty"`
	ConnectAddress        string               `json:"connect-address,omitempty" xml:"connect-address,omitempty"`
	GetConn               time.Time            `json:"get-conn" xml:"get-conn"`
	GetConnHostPort       string               `json:"get-conn-host-port" xml:"get-conn-host-port"`
	GotConn               time.Time            `json:"got-conn" xml:"got-conn"`
	GotFirstResponseByte  time.Time            `json:"got-first-response-byte" xml:"got-first-response-byte"`
	TLSHandshakeStart     time.Time            `json:"tls-handshake-start,omitempty" xml:"tls-handshake-start,omitempty"`
	TLSHandshakeDone      time.Time            `json:"tls-handshake-done,omitempty" xml:"tls-handshake-done,omitempty"`
	TLSHandshakeDuration  time.Duration        `json:"tls-handshake-duration,omitempty" xml:"tls-handshake-duration,omitempty"`
	TLSHandshakeState     *tls.ConnectionState `json:"tls-handshake-state,omitempty" xml:"tls-handshake-state,omitempty"`
	TLSHandshakeIssuer    string               `json:"tls-handshake-issuer,omitempty" xml:"tls-handshake-issuer,omitempty"`
	TLSHandshakeSubject   string               `json:"tls-handshake-subject,omitempty" xml:"tls-handshake-subject,omitempty"`
	TLSHandshakeNotBefore time.Time            `json:"tls-handshake-not-before,omitempty" xml:"tls-handshake-not-before,omitempty"`
	TLSHandshakeNotAfter  time.Time            `json:"tls-handshake-not-after,omitempty" xml:"tls-handshake-not-after,omitempty"`
	TLSHandshakeDigest    string               `json:"tls-handshake-digest,omitempty" xml:"tls-handshake-digest,omitempty"`
	WroteHeaders          http.Header          `json:"-" xml:"-" description:"http write header"`
}

// NewClientParse 方法创建响应选项解析body数据。
func NewClientParse(data interface{}) ClientResponseOption {
	return func(w *http.Response) error {
		return clientParseIn(w, 0, 0xffffffff, data)
	}
}

// NewClientParseIf 方法创建响应选项，在指定状态码时解析body数据。
func NewClientParseIf(status int, data interface{}) ClientResponseOption {
	return func(w *http.Response) error {
		return clientParseIn(w, status, status, data)
	}
}

// NewClientParseIn 方法创建响应选项，在指定状态码范围时解析body数据。
func NewClientParseIn(star, end int, data interface{}) ClientResponseOption {
	return func(w *http.Response) error {
		return clientParseIn(w, star, end, data)
	}
}

// NewClientParseErr 方法创建响应选项，在默认范围时解析body中的Error字段返回。
func NewClientParseErr() ClientResponseOption {
	return func(w *http.Response) error {
		var data struct {
			Status int    `json:"status" protobuf:"6,name=status" xml:"status" yaml:"status"`
			Code   int    `json:"code,omitempty" protobuf:"7,name=code" xml:"code,omitempty" yaml:"code,omitempty"`
			Error  string `json:"error,omitempty" protobuf:"10,name=error" xml:"error,omitempty" yaml:"error,omitempty"`
		}
		err := clientParseIn(w, DefaultClientParseErrStar, DefaultClientParseErrEnd, &data)
		if err != nil {
			return err
		}
		if data.Error != "" {
			return fmt.Errorf(data.Error)
		}
		return nil
	}
}

func clientParseIn(w *http.Response, star, end int, data interface{}) error {
	if w.StatusCode < star || w.StatusCode > end {
		return nil
	}
	mime := w.Header.Get(HeaderContentType)
	pos := strings.IndexByte(mime, ';')
	if pos != -1 {
		mime = mime[:pos]
	}
	switch mime {
	case MimeApplicationJSON:
		return json.NewDecoder(w.Body).Decode(data)
	case MimeApplicationXML:
		return xml.NewDecoder(w.Body).Decode(data)
	case MimeApplicationProtobuf:
		return NewProtobufDecoder(w.Body).Decode(data)
	}
	return fmt.Errorf("eudore client parse not suppert Content-Type: %s", mime)
}

// NewClientCheckStatus 方法创建响应选项检查响应状态码。
func NewClientCheckStatus(status ...int) ClientResponseOption {
	return func(w *http.Response) error {
		for i := range status {
			if status[i] == w.StatusCode {
				return nil
			}
		}

		err := fmt.Errorf("check status is %d not in %v", w.StatusCode, status)
		r := w.Request
		NewLoggerWithContext(w.Request.Context()).WithFields(
			[]string{"method", "host", "path", "query", "status-code", "status"},
			[]interface{}{r.Method, r.Host, r.URL.Path, r.URL.RawQuery, w.StatusCode, w.Status},
		).Error(err)
		return err
	}
}

// NewClientCheckBody 方法创建响应选项检查响应body是否包含指定字符串。
func NewClientCheckBody(str string) ClientResponseOption {
	return func(w *http.Response) error {
		body, err := ioutil.ReadAll(w.Body)
		if err != nil {
			return err
		}
		w.Body = ioutil.NopCloser(bytes.NewReader(body))
		if strings.Index(string(body), str) == -1 {
			err := fmt.Errorf("check body not have string '%s'", str)
			r := w.Request
			NewLoggerWithContext(w.Request.Context()).WithFields(
				[]string{"method", "host", "path", "query", "status-code", "status"},
				[]interface{}{r.Method, r.Host, r.URL.Path, r.URL.RawQuery, w.StatusCode, w.Status},
			).Error(err)
			return err
		}
		return nil
	}
}

// NewClientDumpHead 方法创建响应选项从环境上下文获取Logger输出请求基本信息和Trace信息。
func NewClientDumpHead() ClientResponseOption {
	return newClientDumpWithBody(false)
}

// NewClientDumpBody 方法创建响应选项从环境上下文获取Logger输出响应head和body内容。
func NewClientDumpBody() ClientResponseOption {
	return newClientDumpWithBody(true)
}

func newClientDumpWithBody(hasbody bool) ClientResponseOption {
	return func(w *http.Response) error {
		log := NewLoggerWithContext(w.Request.Context())
		if log == DefaultLoggerNull || log.GetLevel() > LoggerDebug {
			return nil
		}
		r := w.Request
		log = log.WithFields(
			[]string{"proto", "method", "host", "path", "query", "status-code", "status", "request-headers", "response-headers"},
			[]interface{}{w.Proto, r.Method, r.Host, r.URL.Path, r.URL.RawQuery, w.StatusCode, w.Status, r.Header, w.Header},
		)

		for _, name := range []string{HeaderXRequestID, HeaderXTraceID} {
			id := w.Header.Get(name)
			if id != "" {
				log.WithField(strings.ToLower(name), id)
				break
			}
		}

		trace, ok := w.Request.Context().Value(ContextKeyClientTrace).(*ClientTrace)
		if ok {
			trace.HTTPDone = time.Now()
			trace.HTTPDuration = trace.HTTPDone.Sub(trace.HTTPStart)
			trace.DNSDuration = trace.DNSDone.Sub(trace.DNSStart)
			trace.ConnectDuration = trace.ConnectDone.Sub(trace.ConnectStart)
			trace.TLSHandshakeDuration = trace.TLSHandshakeDone.Sub(trace.TLSHandshakeStart)
			if trace.TLSHandshakeState != nil {
				cert := trace.TLSHandshakeState.PeerCertificates[0]
				h := sha256.New()
				h.Write(cert.Raw)
				trace.TLSHandshakeIssuer = cert.Issuer.String()
				trace.TLSHandshakeSubject = cert.Subject.String()
				trace.TLSHandshakeNotBefore = cert.NotBefore
				trace.TLSHandshakeNotAfter = cert.NotAfter
				trace.TLSHandshakeDigest = hex.EncodeToString(h.Sum(nil))
				trace.TLSHandshakeState = nil
			}
			log = log.WithFields([]string{"wrote-headers", "trace"}, []interface{}{trace.WroteHeaders, trace})
		}

		if hasbody {
			body, err := ioutil.ReadAll(w.Body)
			if err != nil {
				return err
			}
			w.Body = ioutil.NopCloser(bytes.NewReader(body))
			log.WithField("body", string(body))
		}
		log.Debug()
		return nil
	}
}

func headerCopy(dst, src map[string][]string) map[string][]string {
	for key, vals := range src {
		dst[key] = append(dst[key], vals...)
	}
	return dst
}

type bodyBuffer struct {
	bytes.Buffer
}

func (body *bodyBuffer) Clone() io.ReadCloser {
	buf := &bodyBuffer{}
	buf.Write(body.Bytes())
	return buf
}

func (body *bodyBuffer) Close() error {
	return nil
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

type bodyEncoder struct {
	reader      *io.PipeReader
	writer      *io.PipeWriter
	contenttype string
	data        interface{}
}

func (body *bodyEncoder) Clone() io.ReadCloser {
	return &bodyEncoder{
		contenttype: body.contenttype,
		data:        body.data,
	}
}

func (body *bodyEncoder) Read(p []byte) (n int, err error) {
	if body.reader == nil {
		body.reader, body.writer = io.Pipe()
		go func() {
			switch body.contenttype {
			case MimeApplicationJSON:
				json.NewEncoder(body.writer).Encode(body.data)
			case MimeApplicationXML:
				json.NewEncoder(body.writer).Encode(body.data)
			case MimeApplicationProtobuf:
				NewProtobufEncoder(body.writer).Encode(body.data)
			}
			body.writer.Close()
		}()
	}
	return body.reader.Read(p)
}

func (body *bodyEncoder) Close() error {
	return body.reader.Close()
}
