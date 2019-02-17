package eudore

import (
	"io"
	"fmt"
	"net"
	"bufio"
	"bytes"
	"net/http"
	"crypto/tls"
)

type (
	// ResponseWriter接口用于写入http请求响应体status、header、body。
	//
	// net/http.response实现了flusher、hijacker、pusher接口。
	ResponseWriter interface {
		// http.ResponseWriter
		Header() http.Header
		Write([]byte) (int, error)
		WriteHeader(codeCode int)
		// http.Flusher 
		Flush()
		// http.Hijacker
		Hijack() (net.Conn, *bufio.ReadWriter, error)
		// http.Pusher
		Push(string, *PushOptions) error
		Size() int
		Status() int
	}

	// ResponseReader is used to read the http protocol response message information.
	//
	// ResponseReader用于读取http协议响应报文信息。
	ResponseReader interface {
		Proto() string
		Statue() int
		Code() string
		Header() Header
		Read([]byte) (int, error)
		TLS() *tls.ConnectionState
		Close() error
	}

	// Encapsulate the net/http.Response response message and convert it to the ResponseReader interface.
	//
	// 封装net/http.Response响应报文，转换成ResponseReader接口
	ResponseReaderHttp struct {
		io.ReadCloser
		Data 	*http.Response
	}
	// net/http.ResponseWriter接口封装
	ResponseWriterHttp struct {
		http.ResponseWriter
		code		int
		size		int
	}
	// 带缓存的ResponseWriter，需要调用Flush然后写入数据。
	ResponseWriterBuffer struct {
		ResponseWriter
		Buf 	*bytes.Buffer
	}
)

var _ ResponseWriter	=	&ResponseWriterHttp{}
var _ ResponseWriter	=	&ResponseWriterBuffer{}


func NewResponseWriterHttp(w http.ResponseWriter) ResponseWriter{
	return &ResponseWriterHttp{ResponseWriter: w}
}

func ResetResponseWriterHttp(hw *ResponseWriterHttp, w http.ResponseWriter) ResponseWriter {
	hw.ResponseWriter = w
	hw.code = http.StatusOK
	hw.size = 0
	return hw
}

func (w *ResponseWriterHttp) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *ResponseWriterHttp) Write(data []byte) (int, error) {
	n, err :=  w.ResponseWriter.Write(data)
	w.size = w.size + n
	return n, err
}

func (w *ResponseWriterHttp) WriteHeader(codeCode int) {
	w.code = codeCode
	w.ResponseWriter.WriteHeader(w.code)
}

func (w *ResponseWriterHttp) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *ResponseWriterHttp) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("http.Hijacker interface is not supported")
}

// 如果ResponseWriterHttp实现http.Push接口，则Push资源。
func (w *ResponseWriterHttp) Push(target string, opts *PushOptions) error {	
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)	
	}	
	return nil
}

func (w *ResponseWriterHttp) Size() int {
	return w.size
}

func (w *ResponseWriterHttp) Status() int {
	return w.code
}



func NewResponseWriterBuffer(w ResponseWriter) ResponseWriter {
	return &ResponseWriterBuffer{
		ResponseWriter:		w,
		Buf:				new(bytes.Buffer),
	}
}


func (w *ResponseWriterBuffer) Write(p []byte) (int, error) {
	return w.Buf.Write(p)
}

func (w *ResponseWriterBuffer) Flush() {
	io.Copy(w.ResponseWriter, w.Buf)
	w.Buf.Reset()
}



func NewResponseReaderHttp(resp *http.Response) ResponseReader {
	return &ResponseReaderHttp{
		ReadCloser:	resp.Body,
		Data:		resp,
	}
}

func (r *ResponseReaderHttp) Proto() string {
	return r.Data.Proto
}

func (r *ResponseReaderHttp) Statue() int {
	return r.Data.StatusCode
}

func (r *ResponseReaderHttp) Code() string {
	return r.Data.Status
}

func (r *ResponseReaderHttp) Header() Header {
	return Header(r.Data.Header)
}

func (r *ResponseReaderHttp) TLS() *tls.ConnectionState {
	return r.Data.TLS
}

