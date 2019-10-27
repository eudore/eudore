package eudore

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// TimeFormat 定义响应header写入Date的时间格式。
const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

// Status 定义http状态码响应值
var Status = map[int]string{
	100: "Continue",
	101: "Switching Protocols",
	102: "Processing",

	200: "OK",
	201: "Created",
	202: "Accepted",
	203: "Non-Authoritative Information",
	204: "No Content",
	205: "Reset Content",
	206: "Partial Content",
	207: "Multi-Status",
	208: "Already Reported",
	226: "IM Used",

	300: "Multiple Choices",
	301: "Moved Permanently",
	302: "Found",
	303: "See Other",
	304: "Not Modified",
	305: "Use Proxy",
	307: "Temporary Redirect",
	308: "Permanent Redirect",

	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Request Entity Too Large",
	414: "Request URI Too Long",
	415: "Unsupported Media Type",
	416: "Requested Range Not Satisfiable",
	417: "Expectation Failed",
	418: "I'm a teapot",
	421: "Misdirected Request",
	422: "Unprocessable Entity",
	423: "Locked",
	424: "Failed Dependency",
	425: "Too Early",
	426: "Upgrade Required",
	428: "Precondition Required",
	429: "Too Many Requests",
	431: "Request Header Fields Too Large",
	451: "Unavailable For Legal Reasons",

	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
	506: "Variant Also Negotiates",
	507: "Insufficient Storage",
	508: "Loop Detected",
	510: "Not Extended",
	511: "Network Authentication Required",
}

// Response 定义http响应对象
type Response struct {
	request Request
	writer  *bufio.Writer
	header  http.Header
	status  int
	size    int
	iswrite bool
	chunked bool
	ishjack bool
	// buffer
	buf []byte
	n   int
	err error
	//
	cancel context.CancelFunc
}

// cancelConn 定义Conn在Close时执行Context cancel
type cancelConn struct {
	net.Conn
	cancel context.CancelFunc
}

// Reset 方法重置http响应状态
func (w *Response) Reset(conn net.Conn) {
	w.writer.Reset(conn)
	w.header = make(http.Header)
	w.status = 200
	w.size = 0
	w.iswrite = false
	w.chunked = false
	w.ishjack = false
	w.err = nil
	w.n = 0
}

// Header 方法获得http响应header对象。
func (w *Response) Header() http.Header {
	return w.header
}

// WriteHeader 方法写入状态码
func (w *Response) WriteHeader(codeCode int) {
	w.status = codeCode
}

// Write 方法写入数据，如果写入数据长度小于缓冲，不会立刻返回，也不会写入状态行。
func (w *Response) Write(p []byte) (int, error) {
	// 数据大于缓冲，发送数据
	if w.n+len(p) > len(w.buf) {
		// 写入数据
		n, _ := w.writeDate(p, len(p))
		// 更新数据长度
		w.size += n
		return n, w.err
	}
	// 数据小于缓存，保存
	n := copy(w.buf[w.n:], p)
	w.n += n
	// 更新数据长度
	w.size += n
	return n, nil
}

// writeDate 方法写入数据并返回。
//
// 会先写入缓冲数据，然后将当前数据写入
//
// 提升分块效率，会将大小两块合并发送。
func (w *Response) writeDate(p []byte, length int) (n int, err error) {
	// 写入状态行
	w.writerResponseLine()
	// 如果有写入错误，或者数据长度为0则返回。
	if w.err != nil || (length+w.n) == 0 {
		return 0, w.err
	}
	// 数据写入
	if w.chunked {
		// 分块写入
		fmt.Fprintf(w.writer, "%x\r\n", length+w.n)
		// 写入缓冲数据和当前数据
		w.writer.Write(w.buf[0:w.n])
		n, err = w.writer.Write(p)
		// 分块结束
		w.writer.Write([]byte{13, 10})
	} else {
		w.writer.Write(w.buf[0:w.n])
		n, err = w.writer.Write(p)
	}
	w.n = 0
	// 检测写入的长度
	if n < length {
		err = io.ErrShortWrite
	}
	w.err = err
	return
}

// writerResponseLine 方法写入状态行
func (w *Response) writerResponseLine() {
	// 已经写入则返回
	if w.iswrite {
		return
	}
	// 设置写入标志为true。
	w.iswrite = true
	// Write response line
	// 写入响应行
	fmt.Fprintf(w.writer, "%s %d %s\r\n", w.request.Proto, w.status, Status[w.status])
	// Write headers
	// 写入headers
	for key, vals := range w.header {
		for _, val := range vals {
			fmt.Fprintf(w.writer, "%s: %s\r\n", key, val)
		}
	}
	// 写入时间和Server
	fmt.Fprintf(w.writer, "Date: %s\r\nServer: eudore\r\n", time.Now().UTC().Format(TimeFormat))
	// 检测是否有写入长度，没有则进行分块传输。
	// 未检测Content-Length值是否合法
	w.chunked = len(w.header.Get("Content-Length")) == 0 && w.header.Get("Upgrade") == ""
	if w.chunked {
		fmt.Fprintf(w.writer, "Transfer-Encoding: chunked\r\n")
	}
	// Write header separator
	// 写入header后分割符
	w.writer.Write([]byte("\r\n"))
}

// Flush 方法数据写入
func (w *Response) Flush() {
	// 将缓冲数据写入
	w.writeDate(nil, 0)
	w.n = 0
	// 发送writer的全部数据
	w.writer.Flush()
}

// finalFlush 方法请求结束时flush写入数据。
func (w *Response) finalFlush() (err error) {
	// 如果没有写入状态行，并且没有指定内容长度。
	// 设置内容长度为当前缓冲数据。
	if !w.iswrite && len(w.header.Get("Content-Length")) == 0 {
		w.header.Set("Content-Length", fmt.Sprint(w.n))
	}
	// 将缓冲数据写入
	w.writeDate(nil, 0)
	// 处理分段传输
	if w.chunked {
		// 处理Trailer header
		tr := w.header.Get("Trailer")
		if len(tr) == 0 {
			// 没有Trailer,直接写入结束
			w.writer.Write([]byte{0x30, 0x0d, 0x0a, 0x0d, 0x0a})
		} else {
			// 写入结尾
			w.writer.Write([]byte{0x30, 0x0d, 0x0a})
			// 写入Trailer的值
			for _, k := range strings.Split(tr, ",") {
				fmt.Fprintf(w.writer, "%s: %s\r\n", k, w.header.Get(k))
			}
			w.writer.Write([]byte{0x0d, 0x0a})
		}
	}
	// 发送数据
	err = w.writer.Flush()
	w.cancel()
	return
}

// Hijack 方法劫持http连接。
func (w *Response) Hijack() (net.Conn, error) {
	w.ishjack = true
	// w.request.conn.SetDeadline(time.Time{})
	// return &cancelConn{w.request.conn, w.cancel}, nil
	return nil, nil
}

// Push 方法http协议不支持push方法。
func (*Response) Push(string, *http.PushOptions) error {
	return nil
}

// Status 方法返回当前状态码。
func (w *Response) Status() int {
	return w.status
}

// Size 方法返回写入的数据长度。
func (w *Response) Size() int {
	return w.size
}

// Close 方法在net.Conn关闭时，执行context cancel。
func (c *cancelConn) Close() (err error) {
	err = c.Conn.Close()
	c.cancel()
	return
}
