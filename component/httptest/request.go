package httptest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type (
	// RequestReaderTest 实现protocol.RequestReader接口，用于执行测试请求。
	RequestReaderTest struct {
		//
		Client *Client
		File   string
		Line   int
		// data
		*http.Request
		json      interface{}
		formValue map[string][]string
		formFile  map[string][]string
	}
)

// NewRequestReaderTest 函数创建一个测试http请求。
func NewRequestReaderTest(client *Client, method, path string) *RequestReaderTest {
	r := &RequestReaderTest{
		Client: client,
		Request: &http.Request{
			Method:     method,
			RequestURI: path,
			Header:     make(http.Header),
			Proto:      "HTTP/1.0",
			ProtoMajor: 1,
			Host:       "eudore-httptest",
			RemoteAddr: "192.0.2.1:1234",
		},
	}
	var err error
	r.Request.URL, err = url.ParseRequestURI(path)
	if err != nil {
		r.Error(err)
	}
	r.Form, err = url.ParseQuery(r.Request.URL.RawQuery)
	if err != nil {
		r.Error(err)
	}
	return r
}

func (r *RequestReaderTest) Error(err error) {
	r.Errorf("%s", err.Error())
}

// Errorf 方法输出错误信息。
func (r *RequestReaderTest) Errorf(format string, args ...interface{}) {
	err := fmt.Errorf(format, args...)
	err = fmt.Errorf("httptest request %s %s of file location %s:%d, error: %v", r.Method, r.RequestURI, r.File, r.Line, err)
	r.Client.Errs = append(r.Client.Errs, err)
}

// WithAddQuery 方法给请求添加一个url参数。
func (r *RequestReaderTest) WithAddQuery(key, val string) *RequestReaderTest {
	r.Form.Add(key, val)
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

// WithBody 方法设置请求的body。
func (r *RequestReaderTest) WithBody(reader interface{}) *RequestReaderTest {
	body, err := transbody(reader)
	if err != nil {
		r.Errorf("%v", err)
	} else if body != nil {
		r.Request.Body = ioutil.NopCloser(body)
	}
	return r
}

func transbody(body interface{}) (io.Reader, error) {
	if body == nil {
		return nil, nil
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

// WithBodyByte 方法设置请的字节body。
func (r *RequestReaderTest) WithBodyByte(b []byte) *RequestReaderTest {
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
func (r *RequestReaderTest) WithBodyJSONValue(key string, val interface{}, args ...interface{}) *RequestReaderTest {
	data, ok := r.json.(map[string]interface{})
	if !ok {
		if r.json == nil {
			data = make(map[string]interface{})
			r.json = data
		} else {
			return r
		}
	}
	data[key] = val
	args = initSlice(args)
	for i := 0; i < len(args); i += 2 {
		data[fmt.Sprint(args[i])] = args[i+1]
	}
	return r
}

// WithBodyFromValue 方法使用Form表单，添加一条键值数据。
func (r *RequestReaderTest) WithBodyFromValue(key, val string, args ...string) *RequestReaderTest {
	if r.formValue == nil {
		r.formValue = make(map[string][]string)
	}
	r.formValue[key] = append(r.formValue[key], val)

	args = initSliceSrting(args)
	for i := 0; i < len(args); i += 2 {
		r.formValue[args[i]] = append(r.formValue[args[i]], args[i+1])
	}
	return r
}

// WithBodyFromValues 方法使用Form表单，添加多条键值数据。
func (r *RequestReaderTest) WithBodyFromValues(data map[string][]string) *RequestReaderTest {
	if r.formValue == nil {
		r.formValue = make(map[string][]string)
	}
	for key, vals := range data {
		r.formValue[key] = vals
	}
	return r
}

// WithBodyFromFile 方法设置请求body Form的文件，值为实际文件路径
func (r *RequestReaderTest) WithBodyFromFile(key, val string, args ...string) *RequestReaderTest {
	if r.formFile == nil {
		r.formFile = make(map[string][]string)
	}
	r.formFile[key] = append(r.formFile[key], val)

	args = initSliceSrting(args)
	for i := 0; i < len(args); i += 2 {
		r.formFile[args[i]] = append(r.formFile[args[i]], args[i+1])
	}
	return r
}

// Do 方法发送这个请求，使用客户端处理这个请求返回响应。
func (r *RequestReaderTest) Do() *ResponseWriterTest {
	// 附加客户端公共参数
	for key, vals := range r.Client.Args {
		for _, val := range vals {
			r.Request.Form.Add(key, val)
		}
	}
	r.Request.URL.RawQuery = r.Form.Encode()
	r.Form = nil

	for key, vals := range r.Client.Headers {
		for _, val := range vals {
			r.Request.Header.Add(key, val)
		}
	}

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
					file, err := os.Open(val)
					if err != nil {
						r.Error(err)
					} else {
						part, _ := w.CreateFormFile(key, file.Name())
						io.Copy(part, file)
						file.Close()
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
	// defer r.Body.Close()

	// 创建响应并处理
	resp := NewResponseWriterTest(r.Client, r)
	r.Client.Handler.ServeHTTP(resp, r.Request)
	return resp
}

func initSlice(args []interface{}) []interface{} {
	if len(args)%2 == 0 {
		return args
	}
	return args[:len(args)-1]
}

func initSliceSrting(args []string) []string {
	if len(args)%2 == 0 {
		return args
	}
	return args[:len(args)-1]
}
