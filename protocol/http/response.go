package http

import (
	"io"
	"fmt"
	"net"
	"time"
	"bufio"
	"github.com/eudore/eudore/protocol"
	"github.com/eudore/eudore/protocol/header"
)

const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

var Status = map[int]string{
	200:    "OK",
}

type Response struct {
	request		*Request
	conn		net.Conn
	writer		*bufio.Writer
	header		protocol.Header
	status		int
	iswrite		bool
	chunked		bool
	// buffer 
	buf 		[]byte
	n			int
	err			error
}


func (w *Response) Reset(conn net.Conn) {
	w.conn = conn
	w.writer.Reset(conn)
	w.header = make(header.HeaderMap)
	w.status = 200
	w.iswrite = false
	w.chunked = false
	w.err = nil
	w.n = 0
}

func (w *Response) Header() protocol.Header {
	return w.header
}

func (w *Response) WriteHeader(codeCode int) {
	w.status = codeCode
}

func (w *Response) Write(p []byte) (nn int, err error) {
	// 数据大于缓冲，发送数据
	for len(p) > len(w.buf) - w.n && w.err == nil {
		// 数据大于缓冲，使用分块传输
		w.chunked = true
		// 写入数据
		var n int
		if w.n == 0 {
			// Large write, empty buffer.
			// Write directly from p to avoid copy.
			w.writerResponseLine()	
			fmt.Fprintf(w.writer, "%x\r\n", len(p))	
			n, w.err = w.writer.Write(p)
			w.writer.Write([]byte{13, 10})
		} else {
			n = copy(w.buf[w.n:], p)
			w.n += n
			w.flush()
		}
		nn += n
		p = p[n:]
	}
	if w.err != nil {
		return nn, w.err
	}
	// 数据小于缓存，保存
	n := copy(w.buf[w.n:], p)
	w.n += n
	nn += n
	return nn, nil
}


func (w *Response) writerResponseLine() {
	if !w.iswrite {
		// 设置写入标志为true。
		w.iswrite = true
		// Write response line
		// 写入响应行
		fmt.Fprintf(w.writer, "%s %d %s\r\n", w.request.Proto(), w.status, Status[w.status])
		// Write headers
		// 写入headers
		w.header.Range(func(k, v string){
			fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
		})
		fmt.Fprintf(w.writer, "Date: %s\r\nServer: eudore\r\n", time.Now().Format(TimeFormat))
		if w.chunked {
			fmt.Fprintf(w.writer, "Transfer-Encoding: chunked\r\n")
		}else{
			fmt.Fprintf(w.writer, "Content-Length: %d\r\n", w.n)
		}
		// Write header separator
		// 写入header后分割符
		w.writer.Write([]byte("\r\n"))
	}
}

func (w *Response) Flush() {
	w.chunked = true	
	w.flush()
	w.writer.Flush()
}

func (w *Response) flushend() error {
	w.flush()
	if w.chunked {
		w.writer.Write([]byte{0x30, 13, 10, 13, 10})
	}
	return w.writer.Flush()
}

func (w *Response) flush() error {
	w.writerResponseLine()	
	if w.err != nil {
		return w.err
	}
	if w.n == 0 {
		return nil
	}
	// 写入数据，如果分块加入块长度和分割符
	if w.chunked {
		fmt.Fprintf(w.writer, "%x\r\n", w.n)	
	}
	n, err := w.writer.Write(w.buf[0:w.n])
	if w.chunked {
		w.writer.Write([]byte{13, 10})	
	}
	
	if n < w.n && err == nil {
		err = io.ErrShortWrite
	}
	if err != nil {
		if n > 0 && n < w.n {
			copy(w.buf[0:w.n-n], w.buf[n:w.n])
		}
		w.n -= n
		w.err = err
		return err
	}
	w.n = 0
	return nil
}

func (w *Response) Hijack() (net.Conn, error) {
	return w.request.conn, nil
}

func (*Response) Push(string, *protocol.PushOptions) error {
	return nil
}

func (w *Response) Status() int {
	return w.status
}

func (w *Response) Size() int {
	return 0
}
