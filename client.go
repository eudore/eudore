package eudore

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Client 定义http客户端接口，构建并发送http请求。
type Client interface {
	NewRequest(context.Context, string, string, ...any) error
	WithClient(...any) Client
	GetClient() *http.Client
}

// clientStd 定义http客户端默认实现。
type clientStd struct {
	Client *http.Client  `alias:"client"`
	Option *ClientOption `alias:"option"`
}

/*
ClientBody defines the client Body.

The GetBody method returns a shallow copy of the data for request redirection and retry.

The AddValue method sets the data saved by the body.

The AddFile method can add file upload when using MultipartForm.

ClientBody 定义客户端Body。

GetBody方法返回数据浅复制用于请求重定向和重试。

AddValue方法设置body保存的数据。

AddFile方法在MultipartForm时可以添加文件上传。
*/
type ClientBody interface {
	io.ReadCloser
	GetContentType() string
	GetBody() (io.ReadCloser, error)
	AddValue(string, any)
	AddFile(string, string, any)
}

// NewClient 函数创建默认http客户端实现，参数为默认选项。
func NewClient(options ...any) Client {
	return &clientStd{
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
			Timeout: DefaultClientTimeout,
		},
		Option: NewClientOption(context.Background(), options),
	}
}

// newDialContext 函数创建http客户端Dial函数，如果是内部请求Host，从环境上下文获取到Server处理连接。
func newDialContext() func(ctx context.Context, network, addr string) (net.Conn, error) {
	fn := (&net.Dialer{
		Timeout:   DefaultClientDialTimeout,
		KeepAlive: DefaultClientDialKeepAlive,
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

// Mount 方法保存context.Context作为Client默认发起请求的context.Context。
func (client *clientStd) Mount(ctx context.Context) {
	client.Option.Context = ctx
}

// WithClient 方法给客户端追加新的选项，返回客户端深拷贝。
/*
	Timeout
	http.CookieJar
	*http.Transport
*/
func (client *clientStd) WithClient(options ...any) Client {
	c := client.Client
	if canCopyClient(options) {
		c = &http.Client{}
		*c = *client.Client
	}

	for i := range options {
		switch o := options[i].(type) {
		case *http.Client:
			c = o
		case *http.Transport:
			tp, ok := c.Transport.(*http.Transport)
			if ok {
				SetAnyDefault(tp, o)
			} else {
				c.Transport = o
			}
		case http.RoundTripper:
			c.Transport = o
		case http.CookieJar:
			c.Jar = o
		case time.Duration:
			c.Timeout = o
		}
	}

	return &clientStd{
		Client: c,
		Option: client.Option.clone().appendOptions(client.Option.Context, options),
	}
}

func canCopyClient(options []any) bool {
	for i := range options {
		switch options[i].(type) {
		case *http.Client, *http.Transport, http.RoundTripper, http.CookieJar, time.Duration:
			return true
		}
	}
	return false
}

// GetClient 方法返回*http.Client对象，用于修改属性。
func (client *clientStd) GetClient() *http.Client {
	return client.Client
}

// NewRequest 方法发送http请求。
func (client *clientStd) NewRequest(ctx context.Context, method string, path string, options ...any) error {
	option := client.Option.clone().appendOptions(ctx, options)
	path = initRequestPath(option.Context, path)
	if option.Trace != nil {
		option.Context = NewClientTraceWithContext(option.Context, option.Trace)
	}

	ctx = option.Context
	if option.Retrys == nil && option.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(option.Context, option.Timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, method, path, option.Body)
	if err != nil {
		return err
	}
	option.apply(req)

	resp, err := client.dotry(req, option)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	return option.release(req, resp, err)
}

func (client *clientStd) dotry(req *http.Request, option *ClientOption) (*http.Response, error) {
	if option.Retrys == nil {
		return client.Client.Do(req)
	}

	attempts := make([]int, len(option.Retrys))
	for {
		r := req
		// retry set timeout
		if option.Timeout > 0 {
			ctx, cancel := context.WithTimeout(option.Context, option.Timeout)
			defer cancel()
			r = req.WithContext(ctx)
		}

		resp, err := client.Client.Do(r)
		if err == nil && resp.StatusCode < StatusTooManyRequests && resp.StatusCode != StatusUnauthorized {
			return resp, err
		}

		// If body has been sent
		if resp != nil && req.Body != nil {
			if req.GetBody == nil {
				return resp, err
			}
			body, err2 := req.GetBody()
			if err2 != nil {
				return resp, err
			}
			req.Body = body
		}

		notry := true
		for i, retry := range option.Retrys {
			if attempts[i] < retry.Max && retry.Condition(attempts[i], resp, err) {
				attempts[i]++
				notry = false
				break
			}
		}
		if notry {
			return resp, err
		}
	}
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

type bodyDecoder struct {
	Reader  io.ReadCloser
	Values  map[string]any
	Data    any
	Type    string
	Encoder func(io.Writer, any)
}

// NewClientBodyJSON 函数创建一个json编码器。
func NewClientBodyJSON(data any) ClientBody {
	return NewClientBodyDecoder(MimeApplicationJSON, data, func(w io.Writer, data any) {
		json.NewEncoder(w).Encode(data)
	})
}

// NewClientBodyXML 函数创建一个xml编码器。
func NewClientBodyXML(data any) ClientBody {
	return NewClientBodyDecoder(MimeApplicationXML, data, func(w io.Writer, data any) {
		xml.NewEncoder(w).Encode(data)
	})
}

// NewClientBodyProtobuf 函数创建一个protobuf编码器。
func NewClientBodyProtobuf(data any) ClientBody {
	return NewClientBodyDecoder(MimeApplicationProtobuf, data, func(w io.Writer, data any) {
		NewProtobufEncoder(w).Encode(data)
	})
}

// The NewClientBodyDecoder function creates a ClientBody encoder,
// which needs to specify contenttype and encoder.
//
// NewClientBodyDecoder 函数创建一个ClientBody编码器，需要指定contenttype和encoder。
func NewClientBodyDecoder(contenttype string, data any, encoder func(io.Writer, any)) ClientBody {
	if data == nil {
		data = make(map[string]any)
	}
	vals, _ := data.(map[string]any)
	body := &bodyDecoder{
		Data:    data,
		Values:  vals,
		Type:    contenttype,
		Encoder: encoder,
	}
	return body
}

func (body *bodyDecoder) Read(p []byte) (int, error) {
	if body.Reader == nil {
		rc, wc := io.Pipe()
		body.Reader = rc
		go func() {
			body.Encoder(wc, body.Data)
			wc.Close()
		}()
	}
	return body.Reader.Read(p)
}

func (body *bodyDecoder) Close() error {
	if body.Reader != nil {
		return body.Reader.Close()
	}
	return nil
}

func (body *bodyDecoder) GetContentType() string {
	return body.Type
}

func (body *bodyDecoder) GetBody() (io.ReadCloser, error) {
	return &bodyDecoder{
		Data:    body.Data,
		Values:  body.Values,
		Type:    body.Type,
		Encoder: body.Encoder,
	}, nil
}

func (body *bodyDecoder) AddValue(key string, val any) {
	if body.Values != nil {
		body.Values[key] = val
	} else {
		SetAnyByPath(body.Data, key, val)
	}
}

func (body *bodyDecoder) AddFile(string, string, any) {}

type bodyForm struct {
	Reader   io.ReadCloser
	Values   url.Values
	Files    map[string][]fileContent
	Boundary string
	NoClone  bool
}

type fileContent struct {
	Name   string
	Body   []byte
	File   string
	Reader io.Reader
}

// NewClientBodyForm 函数创建ApplicationForm或MultipartForm请求body。
//
// AddFile方法允许data类型为[]byte io.Reader；如果类型为string则加载这个本地文件。
//
// 如果使用AddFile方法添加文件ContentType为MultipartForm。
func NewClientBodyForm(data url.Values) ClientBody {
	return &bodyForm{Values: data, Boundary: GetStringRandom(30)}
}

func (body *bodyForm) Read(p []byte) (n int, err error) {
	if body.Reader == nil {
		if body.Files == nil {
			body.Reader = io.NopCloser(strings.NewReader(body.Values.Encode()))
		} else {
			rc, wc := io.Pipe()
			body.Reader = rc
			body.encode(wc)
		}
	}
	return body.Reader.Read(p)
}

func (body *bodyForm) encode(wc io.WriteCloser) {
	w := multipart.NewWriter(wc)
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
					c, ok := val.Reader.(io.Closer)
					if ok {
						c.Close()
					}
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
		wc.Close()
	}()
}

func (body *bodyForm) Close() error {
	if body.Reader != nil {
		return body.Reader.Close()
	}
	return nil
}

func (body *bodyForm) GetContentType() string {
	if body.Files == nil {
		return MimeApplicationForm
	}

	return "multipart/form-data; boundary=" + body.Boundary
}

func (body *bodyForm) GetBody() (io.ReadCloser, error) {
	if body.NoClone {
		return nil, ErrClientBodyFormNotGetBody
	}
	return &bodyForm{
		Values:   body.Values,
		Files:    body.Files,
		Boundary: body.Boundary,
	}, nil
}

func (body *bodyForm) AddValue(key string, val any) {
	if body.Values == nil {
		body.Values = make(url.Values)
	}
	body.Values.Add(key, GetStringByAny(val))
}

func (body *bodyForm) AddFile(key string, name string, data any) {
	if body.Files == nil {
		body.Files = make(map[string][]fileContent)
	}

	content := fileContent{Name: name}
	switch b := data.(type) {
	case []byte:
		content.Body = b
	case string:
		if name == "" {
			content.Name = filepath.Base(b)
		}
		content.File = b
	case io.Reader:
		body.NoClone = true
		content.Reader = b
	default:
		return
	}
	body.Files[key] = append(body.Files[key], content)
}

// NewClientCheckStatus 方法创建响应选项检查响应状态码。
func NewClientCheckStatus(status ...int) func(*http.Response) error {
	return func(w *http.Response) error {
		for i := range status {
			if status[i] == w.StatusCode {
				return nil
			}
		}

		return fmt.Errorf(ErrFormatClintCheckStatusError, w.StatusCode, status)
	}
}

// NewClienProxyWriter 函数将客户端响应写入另外Writer，
//
// 如果Writer实现http.ResponseWriter接口会写入状态码和Header。
func NewClienProxyWriter(writer io.Writer) func(*http.Response) error {
	return func(w *http.Response) error {
		wr, ok := writer.(http.ResponseWriter)
		if ok {
			wr.WriteHeader(w.StatusCode)
			h := w.Header.Clone()
			for _, key := range DefaultClinetHopHeaders {
				h.Del(key)
			}
			for key, vals := range h {
				for _, val := range vals {
					wr.Header().Add(key, val)
				}
			}
		}

		_, err := io.Copy(writer, w.Body)
		return err
	}
}

// NewClientParse 方法创建响应选项解析body数据。
func NewClientParse(data any) func(*http.Response) error {
	return func(w *http.Response) error {
		return clientParseIn(w, 0, 0xffffffff, data)
	}
}

// NewClientParseIf 方法创建响应选项，在指定状态码时解析body数据。
func NewClientParseIf(status int, data any) func(*http.Response) error {
	return func(w *http.Response) error {
		return clientParseIn(w, status, status, data)
	}
}

// NewClientParseIn 方法创建响应选项，在指定状态码范围时解析body数据。
func NewClientParseIn(star, end int, data any) func(*http.Response) error {
	return func(w *http.Response) error {
		return clientParseIn(w, star, end, data)
	}
}

// NewClientParseErr 方法创建响应选项，在默认范围时解析body中的Error字段返回。
func NewClientParseErr() func(*http.Response) error {
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
			return errors.New(data.Error)
		}
		return nil
	}
}

func clientParseIn(w *http.Response, star, end int, data any) error {
	if w.StatusCode < star || w.StatusCode > end || w.Body == nil {
		return nil
	}
	body, ok := data.(*string)
	if ok {
		data, err := io.ReadAll(w.Body)
		if err != nil {
			return err
		}
		*body = string(data)
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
	return fmt.Errorf(ErrFormatClintParseBodyError, mime)
}

// NewClientCheckBody 方法创建响应选项检查响应body是否包含指定字符串。
func NewClientCheckBody(str string) func(*http.Response) error {
	return func(w *http.Response) error {
		body, err := io.ReadAll(w.Body)
		if err != nil {
			return err
		}
		w.Body = io.NopCloser(bytes.NewReader(body))
		if !strings.Contains(string(body), str) {
			return fmt.Errorf("check body not have string '%s'"+string(body[:20]), str)
		}
		return nil
	}
}
