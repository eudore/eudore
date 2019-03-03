package http

import (
	"net"
	"fmt"
	"bufio"
	"context"
	"strings"
	"net/textproto"
	"github.com/eudore/eudore/protocol"
)



const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"
var Status = map[int]string{
	200:    "OK",
}

type HttpHandler struct {

}

// Handling http connections
//
// 处理http连接
func (hh *HttpHandler) EudoreConn(ctx context.Context, c net.Conn, h protocol.Handler) {
	var ok bool
	// Create the currently connected io buffer object.
	// 创建当前连接的io缓冲对象。
	rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
	// Create a text protocol parsing object.
	// 创建一个文本协议解析对象。
	reader := textproto.NewReader(rw.Reader)
	fmt.Println("conn serve:", c.RemoteAddr().String())
	for {
		// Initialize the request object.
		// 初始化请求对象。
		req := &Request{
			header:	make(protocol.Header),
			reader: rw,
			conn:	c,
		}
		resp := &Response{
			request:req,
			status: 200,
			header:	make(protocol.Header),
			writer:	rw,
		}
		// Read the http request line.
		// 读取http请求行。
		line, err := reader.ReadLine()
		if err != nil {
			return
		}
		fmt.Println("read line:", line)
		// Split the http request line.
		// 拆分http请求行。
		req.method, req.requestURI, req.proto, ok = parseRequestLine(line)
		if !ok {
			break
		}
		// read http headers
		// 读取http headers
		for {
			// Read a line of content.
			// 读取一行内容。
			line, err := reader.ReadLine()
			if err != nil || len(line) == 0 {
				break
			}
			// fmt.Println("read header:", line)
			// Split into headers and store them in the request.
			// 分割成header存储到请求中。
			req.header.Add(split2(line, ": "))
		}
		fmt.Println("handler start")
		// Call the handle object to handle the request.
		// 调用handle对象处理这个请求。
		h.EudoreHTTP(ctx, resp, req)
		// Write the cached data and send it back to the client.
		// 将缓存数据写入，发送返回给客户端。
		resp.Flush()
		// // Close the connection and do not implement connection multiplexing.
		// 关闭连接，未实现连接复用。
		c.Close()
		fmt.Println("handler end")
	}
}


// parseRequestLine parses "GET /foo HTTP/1.1" into its three parts.
func parseRequestLine(line string) (method, requestURI, proto string, ok bool) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], line[s2+1:], true
}

func split2(str string, s string) (string, string) {
	ss := strings.SplitN(str, s, 2)
	if len(ss) == 2 {
		return ss[0], ss[1]
	}
	return "", ""
}

