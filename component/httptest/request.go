package httptest

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// RequestReaderTest 实现protocol.RequestReader接口，用于执行测试请求。
type RequestReaderTest struct {
	//
	Client *Client
	File   string
	Line   int
	err    error
	// data
	*http.Request
	websocketHandle func(net.Conn)
	json            interface{}
	formValue       map[string][]string
	formFile        map[string][]fileContent
}
type fileContent struct {
	Name string
	io.Reader
}

// NewRequestReaderTest 函数创建一个测试http请求。
func NewRequestReaderTest(client *Client, method, path string) *RequestReaderTest {
	r := &RequestReaderTest{
		Client: client,
		Request: &http.Request{
			Method:     method,
			RequestURI: path,
		},
	}
	r.File, r.Line = logFormatFileLine(3)
	u, err := url.ParseRequestURI(path)
	if err != nil {
		r.Error(err)
		u = new(url.URL)
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" {
		u.Host = HTTPTestHost
	}
	r.Request = &http.Request{
		Method:     method,
		Host:       u.Host,
		RequestURI: u.RequestURI(),
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}
	r.Form, err = url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		r.Error(err)
	}
	return r
}

// Errorf 方法输出错误信息。
func (r *RequestReaderTest) Error(err error) {
	r.err = err
	r.Client.Print(fmt.Errorf("httptest request %s %s of file location %s:%d, error: %v", r.Method, r.RequestURI, r.File, r.Line, err))
}

// WithTLS 方法在模拟请求时设置tls状态。
func (r *RequestReaderTest) WithTLS() *RequestReaderTest {
	r.TLS = &tls.ConnectionState{
		Version:           tls.VersionTLS12,
		HandshakeComplete: true,
		ServerName:        r.Host,
	}
	return r
}

// WithAddQuery 方法给请求添加一个url参数。
func (r *RequestReaderTest) WithAddQuery(key, val string) *RequestReaderTest {
	r.Form.Add(key, val)
	return r
}

// WithRemoteAddr 方法设置请求的客户端ip地址端口。
func (r *RequestReaderTest) WithRemoteAddr(addr string) *RequestReaderTest {
	r.RemoteAddr = addr
	return r
}

// WithHeaders 方法给请求添加多个header。
func (r *RequestReaderTest) WithHeaders(headers http.Header) *RequestReaderTest {
	for key, vals := range headers {
		for _, val := range vals {
			r.Request.Header.Add(key, val)
		}
	}
	return r
}

// WithHeaderValue 方法给请求添加一个header的值。
func (r *RequestReaderTest) WithHeaderValue(key, val string) *RequestReaderTest {
	r.Request.Header.Add(key, val)
	return r
}

// WithBody 方法设置请求的body,允许使用string、、[]byte、io.ReadCloser、io.Reader类型。
func (r *RequestReaderTest) WithBody(reader interface{}) *RequestReaderTest {
	body, err := getIOReader(reader)
	if err != nil {
		r.Error(err)
	} else if body != nil {
		r.Request.Body = ioutil.NopCloser(body)
	}
	return r
}

func getIOReader(body interface{}) (io.Reader, error) {
	if body == nil {
		return nil, fmt.Errorf("getIOReader body is nil")
	}
	switch t := body.(type) {
	case string:
		return strings.NewReader(t), nil
	case []byte:
		return bytes.NewReader(t), nil
	case io.Reader:
		return t, nil
	default:
		return nil, fmt.Errorf("unknown type used for body: %+v", body)
	}
}

// WithBodyString 方法设置请求的字符串body。
func (r *RequestReaderTest) WithBodyString(s string) *RequestReaderTest {
	r.Body = ioutil.NopCloser(strings.NewReader(s))
	r.ContentLength = int64(len(s))
	return r
}

// WithBodyBytes 方法设置请的字节body。
func (r *RequestReaderTest) WithBodyBytes(b []byte) *RequestReaderTest {
	r.Body = ioutil.NopCloser(bytes.NewReader(b))
	r.ContentLength = int64(len(b))
	return r
}

// WithBodyJSON 方法设置body为一个对象的json字符串。
func (r *RequestReaderTest) WithBodyJSON(data interface{}) *RequestReaderTest {
	r.json = data
	return r
}

// WithBodyJSONValue 方法设置一条json数据，使用map[string]interface{}保存json数据。
func (r *RequestReaderTest) WithBodyJSONValue(key string, val interface{}) *RequestReaderTest {
	if r.json == nil {
		r.json = make(map[string]interface{})
	}
	data, ok := r.json.(map[string]interface{})
	if !ok {
		return r
	}
	data[key] = val
	return r
}

// WithBodyFormValue 方法使用Form表单，添加一条键值数据。
func (r *RequestReaderTest) WithBodyFormValue(key, val string) *RequestReaderTest {
	if r.formValue == nil {
		r.formValue = make(map[string][]string)
	}
	r.formValue[key] = append(r.formValue[key], val)
	return r
}

// WithBodyFormValues 方法使用Form表单，添加多条键值数据。
func (r *RequestReaderTest) WithBodyFormValues(data map[string][]string) *RequestReaderTest {
	if r.formValue == nil {
		r.formValue = make(map[string][]string)
	}
	for key, vals := range data {
		r.formValue[key] = vals
	}
	return r
}

// WithBodyFormFile 方法使用Form表单，添加一个文件名称和内容。
func (r *RequestReaderTest) WithBodyFormFile(key, name string, val interface{}) *RequestReaderTest {
	if r.formFile == nil {
		r.formFile = make(map[string][]fileContent)
	}

	body, err := getIOReader(val)
	if err != nil {
		r.Error(err)
		return r
	}

	r.formFile[key] = append(r.formFile[key], fileContent{name, body})
	return r
}

// WithBodyFormLocalFile 方法设置请求body Form的文件，值为实际文件路径
func (r *RequestReaderTest) WithBodyFormLocalFile(key, name, path string) *RequestReaderTest {
	if r.formFile == nil {
		r.formFile = make(map[string][]fileContent)
	}

	file, err := os.Open(path)
	if err != nil {
		r.Error(err)
		return r
	}

	r.formFile[key] = append(r.formFile[key], fileContent{name, file})
	return r
}

// WithWebsocket 方法定义websock处理函数。
func (r *RequestReaderTest) WithWebsocket(fn func(net.Conn)) *RequestReaderTest {
	r.websocketHandle = fn
	r.Request.Method = "GET"
	r.Request.Header.Set("Host", r.Request.Host)
	r.Request.Header.Add("Upgrade", "websocket")
	r.Request.Header.Add("Connection", "Upgrade")
	r.Request.Header.Add("Sec-WebSocket-Key", "x3JJHMbDL1EzLkh9GBhXDw==")
	r.Request.Header.Add("Sec-WebSocket-Version", "13")
	r.Request.Header.Add("Origin", "http://"+r.Request.Host)
	return r
}

// Do 方法发送这个请求，使用客户端处理这个请求返回响应。
func (r *RequestReaderTest) Do() *ResponseWriterTest {
	if r.err != nil {
		resp := NewResponseWriterTest(r.Client, r)
		resp.Code = 500
		resp.Body = bytes.NewBufferString(r.err.Error())
		return resp
	}
	r.initArgs()
	r.initBody()
	ctx, cancel := context.WithCancel(r.Request.Context())
	defer cancel()
	r.Request = r.Request.WithContext(ctx)

	// 创建响应并处理
	resp := NewResponseWriterTest(r.Client, r)
	if r.URL.Host == HTTPTestHost {
		if r.RemoteAddr == "" {
			r.RemoteAddr = r.Client.RemoteAddr
		}
		r.Client.ServeHTTP(resp, r.Request)
		// 等待websocket客户端连接处理启动
		resp.Wait()
		r.handleCookie(&http.Response{Header: resp.Header()})
	} else {
		r.RequestURI = ""
		httpResp, err := r.DoHTTP()
		if err == nil {
			resp.HandleRespone(httpResp)
			r.handleCookie(httpResp)
		} else {
			r.Error(err)
			resp.Code = 500
			resp.Body = bytes.NewBufferString(err.Error())
		}

	}
	return resp
}

func (r *RequestReaderTest) initArgs() {
	// 附加客户端公共参数
	for key, vals := range r.Client.Querys {
		for _, val := range vals {
			r.Request.Form.Add(key, val)
		}
	}
	r.Request.URL.RawQuery = r.Form.Encode()
	r.Request.RequestURI = r.Request.URL.RequestURI()
	r.Form = nil

	for key, vals := range r.Client.Headers {
		for _, val := range vals {
			r.Request.Header.Add(key, val)
		}
	}
	// set host
	host := r.Header.Get("Host")
	if host != "" {
		r.Request.Host = host
		r.Header.Del("Host")
	}
	// set cookie header
	for _, cookie := range r.Client.CookieJar.Cookies(r.URL) {
		r.Request.Header.Add("Cookie", cookie.String())
	}
}

func (r *RequestReaderTest) initBody() {
	switch {
	case r.json != nil:
		r.Request.Header.Add("Content-Type", "application/json")
		reader, writer := io.Pipe()
		r.Request.Body = reader
		go func() {
			json.NewEncoder(writer).Encode(r.json)
			writer.Close()
		}()
	case r.formValue != nil || r.formFile != nil:
		reader, writer := io.Pipe()
		r.Request.Body = reader
		w := multipart.NewWriter(writer)
		r.Request.Header.Add("Content-Type", w.FormDataContentType())
		go func() {
			for key, vals := range r.formValue {
				for _, val := range vals {
					w.WriteField(key, val)
				}
			}
			for key, vals := range r.formFile {
				for _, val := range vals {
					part, _ := w.CreateFormFile(key, val.Name)
					io.Copy(part, val)
					cr, ok := val.Reader.(io.Closer)
					if ok {
						cr.Close()
					}
				}
			}
			w.Close()
			writer.Close()
		}()
	case r.Request.Body == nil:
		r.Request.Body = http.NoBody
		r.Request.ContentLength = -1
	}
}

// DoHTTP 方法发送这个请求。
func (r *RequestReaderTest) DoHTTP() (*http.Response, error) {
	if r.websocketHandle == nil {
		return r.Client.Do(r.Request)
	}

	conn, err := r.dialConn()
	if err != nil {
		return nil, err
	}
	err = r.Request.Write(conn)
	if err != nil {
		return nil, err
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), r.Request)
	if err == nil {
		go r.websocketHandle(conn)
	}
	return resp, err
}

func (r *RequestReaderTest) handleCookie(resp *http.Response) {
	r.Client.CookieJar.SetCookies(r.URL, resp.Cookies())
}

var zeroDialer net.Dialer

func (r *RequestReaderTest) dialConn() (net.Conn, error) {
	ts := new(http.Transport)
	if r.Client.Transport != nil {
		ts = r.Client.Transport.(*http.Transport)
	}

	if r.URL.Scheme == "http" {
		if ts.DialContext != nil {
			return ts.DialContext(r.Request.Context(), "tcp", r.Request.URL.Host)
		}
		if ts.Dial != nil {
			return ts.Dial("tcp", r.Request.URL.Host)
		}
		return zeroDialer.DialContext(r.Request.Context(), "tcp", r.Request.URL.Host)
	}
	// by go1.14
	// if ts.DialTLSContext  != nil {
	// 	return ts.DialTLSContext(r.Request.Context(), "tcp", r.Request.Host)
	// }
	if ts.DialTLS != nil {
		return ts.DialTLS("tcp", r.Request.URL.Host)
	}
	return tls.Dial("tcp", r.Request.URL.Host, &tls.Config{InsecureSkipVerify: true})
}
