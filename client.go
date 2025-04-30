package eudore

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

// Client defines the http client interface, builds and sends [http.Client]
// requests.
//
// Can be used to send requests or for unit testing.
type Client interface {
	// The NewRequest method creates and sends a request,
	// using options to customize processing.
	NewRequest(method string, path string, options ...any) error
	// The NewClient method uses options to create a new Client,
	// inheriting Options.
	//
	// refer: [NewClient].
	NewClient(options ...any) Client
	GetRequest(path string, options ...any) error
	PostRequest(path string, options ...any) error
	PutRequest(path string, options ...any) error
	DeleteRequest(path string, options ...any) error
	HeadRequest(path string, options ...any) error
	PatchRequest(path string, options ...any) error
}

type ClientHook interface {
	// When Name is not empty, it affects the Client to remove duplicate Hooks.
	Name() string
	// The Wrap method wraps the [http.RoundTripper] and returns a copy.
	Wrap(rt http.RoundTripper) http.RoundTripper
	// The RoundTrip method implements [http.RoundTripper] to send a request.
	//
	// Wrap RoundTrip to implement a custom Hook function.
	RoundTrip(req *http.Request) (*http.Response, error)
}

type MetadataClient struct {
	Health    bool     `json:"health" protobuf:"1,name=health" yaml:"health"`
	Name      string   `json:"name" protobuf:"2,name=name" yaml:"name"`
	Transport string   `json:"transport" protobuf:"3,name=transport" yaml:"transport"`
	Hooks     []string `json:"hooks" protobuf:"4,name=hooks" yaml:"hooks"`
}

type clientStd struct {
	Transport    http.RoundTripper
	RoundTripper http.RoundTripper
	Option       *ClientOption
	Hooks        []ClientHook
	Names        []string
}

// NewClient creates [Client] with custom options.
//
// The options types are: [http.RoundTripper], [ClientHook],
// func(http.RoundTripper), [ClientOption].
//
// By default, [NewClientHookTimeout] and [NewClientHookRedirect] are used.
func NewClient(options ...any) Client {
	client := newClientStd()
	client.Hooks = []ClientHook{
		NewClientHookTimeout(DefaultClientTimeout),
		NewClientHookRedirect(nil),
	}
	client.Names = []string{
		client.Hooks[0].Name(),
		client.Hooks[1].Name(),
	}
	return client.NewClient(options...)
}

// NewClientCustom creates [Client] with empty options.
//
// refer: [NewClient].
func NewClientCustom(options ...any) Client {
	return newClientStd().NewClient(options...)
}

func newClientStd() *clientStd {
	return &clientStd{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           newDialContext(),
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			MaxConnsPerHost:       100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			DisableKeepAlives:     false,
		},
		Option: &ClientOption{
			Context: context.Background(),
		},
	}
}

// newDialContext 函数创建http客户端Dial函数，如果是内部请求Host，从环境上下文获取到Server处理连接。
func newDialContext() func(ctx context.Context, network, addr string,
) (net.Conn, error) {
	fn := (&net.Dialer{
		Timeout:   DefaultClientDialTimeout,
		KeepAlive: DefaultClientDialKeepAlive,
	}).DialContext
	type ServeConn interface{ ServeConn(conn net.Conn) }
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if network == "tcp" &&
			strings.HasPrefix(addr, DefaultClientInternalHost) {
			port := addr[len(DefaultClientInternalHost):]
			if port == ":80" || port == "" {
				server, ok := ctx.Value(ContextKeyServer).(ServeConn)
				if ok {
					serverConn, clientConn := net.Pipe()
					server.ServeConn(serverConn)
					return clientConn, nil
				}
			}
		}

		return fn(ctx, network, addr)
	}
}

// The Mount method saves the [context.Context] as the default context.Context
// for the Client to initiate the request.
func (client *clientStd) Mount(ctx context.Context) {
	if client.Option.Context == context.Background() {
		client.Option.Context = ctx
	}
}

func (client *clientStd) Metadata() any {
	names := make([]string, len(client.Hooks))
	for i := range client.Hooks {
		s, ok := client.Hooks[i].(fmt.Stringer)
		if ok {
			names[i] = s.String()
			continue
		}
		names[i] = client.Hooks[i].Name()
	}
	return MetadataClient{
		Name:      "eudore.clientStd",
		Health:    true,
		Transport: reflect.TypeOf(client.Transport).String(),
		Hooks:     names,
	}
}

func (client *clientStd) NewRequest(method string, path string, options ...any) error {
	option := client.Option.clone().appendOptions(options)
	path = option.parsePath(path)
	ctx := option.Context
	if option.Trace != nil {
		ctx = NewClientTraceWithContext(ctx, option.Trace)
	}

	req, err := http.NewRequestWithContext(ctx, method, path, option.Body)
	if err != nil {
		return err
	}
	option.apply(req)

	resp, err := client.getRoundTripper(option.Hooks).RoundTrip(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}
	return option.release(resp)
}

func (client *clientStd) getRoundTripper(hooks []ClientHook) http.RoundTripper {
	if hooks == nil {
		return client.RoundTripper
	}

	overs := make([]int, len(client.Hooks))
	adds := make([]ClientHook, 0, len(hooks))
	for i, hook := range hooks {
		name := hook.Name()
		pos := sliceIndex(client.Names, name)
		if name != "" && pos != -1 {
			overs[pos] = i + 1
		} else {
			adds = append(adds, hook)
		}
	}

	rt := client.Transport
	if len(hooks) != len(adds) {
		for i, hook := range client.Hooks {
			if overs[i] > 0 {
				hook = hooks[overs[i]-1]
			}
			rt = hook.Wrap(rt)
		}
		hooks = adds
	} else {
		rt = client.RoundTripper
	}

	for _, hook := range hooks {
		rt = hook.Wrap(rt)
	}
	return rt
}

func (client *clientStd) NewClient(options ...any) Client {
	client = &clientStd{
		Transport:    client.Transport,
		RoundTripper: client.RoundTripper,
		Option:       client.Option,
		Hooks:        append([]ClientHook{}, client.Hooks...),
		Names:        append([]string{}, client.Names...),
	}
	var opts []any
	for i := range options {
		switch o := options[i].(type) {
		case ClientHook:
			client.RoundTripper = nil
			name := o.Name()
			pos := sliceIndex(client.Names, name)
			if pos == -1 || name == "" {
				client.Hooks = append(client.Hooks, o)
				client.Names = append(client.Names, name)
			} else {
				client.Hooks[pos] = o
			}
		case http.RoundTripper:
			client.RoundTripper = nil
			client.Transport = o
		case func(http.RoundTripper):
			o(client.Transport)
			options[i] = nil
		default:
			opts = append(opts, o)
		}
	}

	if client.RoundTripper == nil {
		client.RoundTripper = client.Transport
		for _, hook := range client.Hooks {
			client.RoundTripper = hook.Wrap(client.RoundTripper)
		}
	}
	client.Option = client.Option.clone().appendOptions(opts)
	return client
}

func (client *clientStd) GetRequest(path string, options ...any) error {
	return client.NewRequest(MethodGet, path, options...)
}

func (client *clientStd) PostRequest(path string, options ...any) error {
	return client.NewRequest(MethodPost, path, options...)
}

func (client *clientStd) PutRequest(path string, options ...any) error {
	return client.NewRequest(MethodPut, path, options...)
}

func (client *clientStd) DeleteRequest(path string, options ...any) error {
	return client.NewRequest(MethodDelete, path, options...)
}

func (client *clientStd) HeadRequest(path string, options ...any) error {
	return client.NewRequest(MethodHead, path, options...)
}

func (client *clientStd) PatchRequest(path string, options ...any) error {
	return client.NewRequest(MethodPatch, path, options...)
}

// ClientBody defines the client Body.
type ClientBody interface {
	io.ReadCloser
	GetContentType() string
	// The GetBody method returns a shallow copy of the data for
	// request redirection and retry.
	GetBody() (io.ReadCloser, error)
	// The AddValue method sets the data saved by the body.
	AddValue(key string, val any)
	// The AddFile method can add file upload when using MultipartForm.
	//
	// Convert the file type to io.Reader,
	// the file type is []byte as the content;
	// the file type is string to open the os file.
	AddFile(key string, name string, data any)
}

type bodyDecoder struct {
	reader      io.ReadCloser
	values      map[string]any
	data        any
	contentType string
	encoder     func(io.Writer, any)
}

// The NewClientBodyJSON function creates [ClientBody] with [json] encoder.
func NewClientBodyJSON(data any) ClientBody {
	return NewClientBodyDecoder(MimeApplicationJSON, data,
		func(w io.Writer, data any) {
			_ = json.NewEncoder(w).Encode(data)
		},
	)
}

// The NewClientBodyJSON function creates [ClientBody] with [xml] encoder.
func NewClientBodyXML(data any) ClientBody {
	return NewClientBodyDecoder(MimeApplicationXML, data,
		func(w io.Writer, data any) {
			_ = xml.NewEncoder(w).Encode(data)
		},
	)
}

// The NewClientBodyDecoder function creates [ClientBody] encoder,
// which needs to specify [HeaderContentType] and encoder.
func NewClientBodyDecoder(contenttype string, data any,
	encoder func(io.Writer, any),
) ClientBody {
	if data == nil {
		data = make(map[string]any)
	}
	vals, _ := data.(map[string]any)
	body := &bodyDecoder{
		data:        data,
		values:      vals,
		contentType: contenttype,
		encoder:     encoder,
	}
	return body
}

func (body *bodyDecoder) Read(p []byte) (int, error) {
	if body.reader == nil {
		rc, wc := io.Pipe()
		body.reader = rc
		go func() {
			body.encoder(wc, body.data)
			wc.Close()
		}()
	}
	return body.reader.Read(p)
}

func (body *bodyDecoder) Close() error {
	if body.reader != nil {
		return body.reader.Close()
	}
	return nil
}

func (body *bodyDecoder) GetContentType() string {
	return body.contentType
}

func (body *bodyDecoder) GetBody() (io.ReadCloser, error) {
	return &bodyDecoder{
		data:        body.data,
		values:      body.values,
		contentType: body.contentType,
		encoder:     body.encoder,
	}, nil
}

func (body *bodyDecoder) AddValue(key string, val any) {
	if body.values != nil {
		body.values[key] = val
	} else {
		_ = SetAnyByPath(body.data, key, val, nil)
	}
}

func (body *bodyDecoder) AddFile(string, string, any) {}

type bodyFile struct {
	*os.File
	contentType string
}

func NewClientBodyFile(contenttype string, file *os.File) ClientBody {
	if contenttype == "" {
		contenttype = MimeApplicationOctetStream
	}
	return &bodyFile{
		File:        file,
		contentType: contenttype,
	}
}

func (body *bodyFile) GetContentType() string {
	return body.contentType
}

func (body *bodyFile) GetBody() (io.ReadCloser, error) {
	path := body.File.Name()
	return os.OpenFile(path, os.O_RDWR, 0)
}
func (body *bodyFile) AddValue(string, any)        {}
func (body *bodyFile) AddFile(string, string, any) {}

type bodyForm struct {
	reader   io.ReadCloser
	values   url.Values
	files    map[string][]fileContent
	boundary string
	noClone  bool
}

type fileContent struct {
	Name   string
	Body   []byte
	File   string
	Reader io.Reader
}

// The NewClientBodyForm function creates [MimeApplicationForm] or
// [MimeMultipartForm] request body.
//
// If you use the AddFile method to add file,
// the [HeaderContentType] is [MimeMultipartForm].
func NewClientBodyForm(data url.Values) ClientBody {
	return &bodyForm{values: data, boundary: GetStringRandom(30)}
}

func (body *bodyForm) Read(p []byte) (int, error) {
	if body.reader == nil {
		if body.files == nil {
			body.reader = io.NopCloser(strings.NewReader(body.values.Encode()))
		} else {
			rc, wc := io.Pipe()
			body.reader = rc
			body.encode(wc)
		}
	}
	return body.reader.Read(p)
}

func (body *bodyForm) encode(wc io.WriteCloser) {
	w := multipart.NewWriter(wc)
	_ = w.SetBoundary(body.boundary)
	go func() {
		for key, vals := range body.values {
			for _, val := range vals {
				_ = w.WriteField(key, val)
			}
		}
		for key, vals := range body.files {
			for _, val := range vals {
				part, _ := w.CreateFormFile(key, val.Name)
				switch {
				case val.Body != nil:
					_, _ = part.Write(val.Body)
				case val.Reader != nil:
					_, _ = io.Copy(part, val.Reader)
					c, ok := val.Reader.(io.Closer)
					if ok {
						c.Close()
					}
				case val.File != "":
					file, err := os.Open(val.File)
					if err == nil {
						_, _ = io.Copy(part, file)
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
	if body.reader != nil {
		return body.reader.Close()
	}
	return nil
}

func (body *bodyForm) GetContentType() string {
	if body.files == nil {
		return MimeApplicationForm
	}

	return MimeMultipartForm + "; boundary=" + body.boundary
}

func (body *bodyForm) GetBody() (io.ReadCloser, error) {
	if body.noClone {
		return nil, ErrClientBodyNotGetBody
	}
	return &bodyForm{
		values:   body.values,
		files:    body.files,
		boundary: body.boundary,
	}, nil
}

func (body *bodyForm) AddValue(key string, val any) {
	if body.values == nil {
		body.values = make(url.Values)
	}
	body.values.Add(key, GetStringByAny(val))
}

func (body *bodyForm) AddFile(key string, name string, data any) {
	if body.files == nil {
		body.files = make(map[string][]fileContent)
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
		body.noClone = true
		content.Reader = b
	default:
		return
	}
	body.files[key] = append(body.files[key], content)
}

// The NewClientCheckStatus method creates [ClientOption] to check the response
// status code.
func NewClientCheckStatus(status ...int) func(*http.Response) error {
	return func(w *http.Response) error {
		for i := range status {
			if status[i] == w.StatusCode {
				return nil
			}
		}

		return fmt.Errorf(ErrClientCheckStatusError,
			w.Request.Method, w.Request.URL.Path, w.StatusCode, status,
		)
	}
}

// The NewClientParse method creates [ClientOption] to parse body data.
//
// data type is *string or [io.Writer] to write data directly.
//
// If the [HeaderContentType] value is [MimeApplicationJSON]
// [MimeApplicationXML],
// use the corresponding Decoder to parse.
func NewClientParse(data any) func(*http.Response) error {
	return func(w *http.Response) error {
		return clientParseIn(w, 0, 0xffffffff, data)
	}
}

// The NewClientParseIf method creates [ClientOption] and parses the body data
// when the status code is specified.
//
// refer: [NewClientParse].
func NewClientParseIf(status int, data any) func(*http.Response) error {
	return func(w *http.Response) error {
		return clientParseIn(w, status, status, data)
	}
}

// The NewClientParseIn method creates [ClientOption] and parses the body data
// when the status code in range.
//
// refer: [NewClientParse].
func NewClientParseIn(star, end int, data any) func(*http.Response) error {
	return func(w *http.Response) error {
		return clientParseIn(w, star, end, data)
	}
}

// The NewClientParseErr method creates [ClientOption] parsing response error.
//
// When the status code is in [DefaultClientParseErrRange],
// the Error field in the parsed body is returned.
func NewClientParseErr() func(*http.Response) error {
	return func(w *http.Response) error {
		var data struct {
			Status int    `json:"status" protobuf:"6,name=status" yaml:"status"`
			Code   int    `json:"code,omitempty" protobuf:"7,name=code" yaml:"code,omitempty"`
			Error  string `json:"error,omitempty" protobuf:"10,name=error" yaml:"error,omitempty"`
		}
		err := clientParseIn(w,
			DefaultClientParseErrRange[0], DefaultClientParseErrRange[1],
			&data,
		)
		if err != nil {
			return err
		}
		if data.Error != "" {
			if data.Code == 0 {
				return fmt.Errorf("client request status is %d, error: %s",
					data.Status, data.Error,
				)
			}
			return fmt.Errorf("client request status is %d, code is %d, error: %s",
				data.Status, data.Code, data.Error,
			)
		}
		return nil
	}
}

func clientParseIn(w *http.Response, star, end int, data any) error {
	if w.StatusCode < star || w.StatusCode > end || w.Body == nil {
		return nil
	}

	switch body := data.(type) {
	case *string:
		data, err := io.ReadAll(w.Body)
		if err != nil {
			return err
		}
		*body = string(data)
		return nil
	case io.Writer:
		_, err := io.Copy(body, w.Body)
		return err
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
	}
	return fmt.Errorf(ErrClientParseBodyError, mime)
}

// The NewClientCheckBody method creates [ClientOption] to check response body
// contains the specified string.
func NewClientCheckBody(str string) func(*http.Response) error {
	return func(w *http.Response) error {
		body, err := io.ReadAll(w.Body)
		if err != nil {
			return err
		}
		w.Body = io.NopCloser(bytes.NewReader(body))
		if !strings.Contains(string(body), str) {
			if len(body) > DefaultClientCheckBodyLength {
				body = body[:DefaultClientCheckBodyLength]
			}
			return fmt.Errorf("check body '%s' not have string '%s'",
				string(body), str,
			)
		}
		return nil
	}
}
