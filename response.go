package eudore

import (
	"io"
	"fmt"
	"net"
	"bytes"
	"net/http"
	"crypto/tls"
	"github.com/eudore/eudore/protocol"
)

type (
/*	ResponseWriter interface {
		// http.ResponseWriter
		Header() Header
		Write([]byte) (int, error)
		WriteHeader(int)
		// http.Flusher 
		Flush()
		// http.Hijacker
		Hijack() (net.Conn, *bufio.ReadWriter, error)
		// http.Pusher
		Push(string, *PushOptions) error
		Size() int
		Status() int
	}*/



	// Encapsulate the net/http.Response response message and convert it to the ResponseReader interface.
	//
	// 封装net/http.Response响应报文，转换成ResponseReader接口
	ResponseReaderHttp struct {
		io.ReadCloser
		Data 	*http.Response
		header	protocol.Header
	}
	// net/http.ResponseWriter接口封装
	ResponseWriterHttp struct {
		http.ResponseWriter
		header	protocol.Header
		code		int
		size		int
	}
	// 带缓存的ResponseWriter，需要调用Flush然后写入数据。
	ResponseWriterBuffer struct {
		protocol.ResponseWriter
		Buf 	*bytes.Buffer
	}
)

var _ protocol.ResponseWriter	=	&ResponseWriterHttp{}
var _ protocol.ResponseWriter	=	&ResponseWriterBuffer{}


func NewResponseWriterHttp(w http.ResponseWriter) protocol.ResponseWriter {
	return &ResponseWriterHttp{
		ResponseWriter: w,
		header:			HeaderHttp(w.Header()),
	}
}

func ResetResponseWriterHttp(hw *ResponseWriterHttp, w http.ResponseWriter) protocol.ResponseWriter {
	hw.ResponseWriter = w
	hw.header = HeaderHttp(w.Header())
	hw.code = http.StatusOK
	hw.size = 0
	return hw
}

func (w *ResponseWriterHttp) Header() protocol.Header {
	return w.header
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

func (w *ResponseWriterHttp) Hijack() (conn net.Conn, err error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		conn, _, err =  hj.Hijack()
		return 
	}
	err = fmt.Errorf("http.Hijacker interface is not supported")
	return
}

// 如果ResponseWriterHttp实现http.Push接口，则Push资源。
func (w *ResponseWriterHttp) Push(target string, opts *protocol.PushOptions) error {	
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		// TODO: add con
		return pusher.Push(target, &http.PushOptions{

		})	
	}	
	return nil
}

func (w *ResponseWriterHttp) Size() int {
	return w.size
}

func (w *ResponseWriterHttp) Status() int {
	return w.code
}



func NewResponseWriterBuffer(w protocol.ResponseWriter) protocol.ResponseWriter {
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



func NewResponseReaderHttp(resp *http.Response) protocol.ResponseReader {
	return &ResponseReaderHttp{
		ReadCloser:	resp.Body,
		Data:		resp,
		header:		HeaderHttp(resp.Header),
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

func (r *ResponseReaderHttp) Header() protocol.Header {
	return r.header
}

func (r *ResponseReaderHttp) TLS() *tls.ConnectionState {
	return r.Data.TLS
}

