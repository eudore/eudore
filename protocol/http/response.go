package http

import (
	"fmt"
	"net"
	"time"
	"bufio"
	"github.com/eudore/eudore/protocol"
)

type Response struct {
		request		*Request
		iswrite		bool
		status		int
		header		protocol.Header
		writer		*bufio.ReadWriter

	}


func (w *Response) Header() protocol.Header {
	return w.header
}

func (w *Response) Write(b []byte) (int, error) {
	// If it is the first time to write to the body, write the response line and headers before this.
	// 如果是第一次写入body，在此之前写入响应行和headers。
	if !w.iswrite {
		// Set default headers
		// 设置默认headers
		w.Header().Add("Date", time.Now().Format(TimeFormat))
		// Write response line
		// 写入响应行
		fmt.Fprintf(w.writer, "%s %d %s\r\n", w.request.Proto(), w.status, Status[w.status])
		// Write headers
		// 写入headers
		for k, v := range w.header {
			fmt.Fprintf(w.writer, "%s: %s\r\n", k, v[0])
		}
		// Write header separator
		// 写入header后分割符
		w.writer.Write([]byte("\r\n"))
		// Set the write standard to true.
		// 设置写入标准为true。
		w.iswrite = true
	}
	return w.writer.Write(b)
}

func (w *Response) WriteHeader(codeCode int) {
	w.status = codeCode
}

func (w *Response) Flush() {
	w.writer.Flush()
}

func (w *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.request.conn, w.writer, nil
}

func (w *Response) Status() int {
	return w.status
}

func (w *Response) Size() int {
	return 0
}
