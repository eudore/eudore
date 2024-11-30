package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"time"
)

func main() {
	fmt.Println(NewRequest("GET", "http://localhost:8088/api?name=eudore"))
}

// Client 定义http客户端。
type Client struct {
	nc     net.Conn
	rw     *bufio.ReadWriter
	header Header
}

type Header map[string][]string

// NewRequest 函数发送一个http请求。
func NewRequest(method, path string) error {
	u, err := url.Parse(path)
	if err != nil {
		return err
	}
	// 建立连接
	nc, err := net.DialTimeout("tcp", u.Host, 2*time.Second)
	if err != nil {
		return err
	}
	defer nc.Close()

	// 创建Http Client
	c := &Client{
		nc:     nc,
		header: make(Header),
		rw:     bufio.NewReadWriter(bufio.NewReader(nc), bufio.NewWriter(nc)),
	}
	// HTTP/1.1 唯一必要Header host
	c.header["Host"] = []string{u.Host}
	// 不使用长连接 未处理
	c.header["Connection"] = []string{"close"}

	path = u.Path
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}
	// ---------- 请求行 ----------
	fmt.Fprintf(c.rw, "%s %s HTTP/1.1\r\n", method, path)
	// ---------- Header ----------
	for k, v := range c.header {
		fmt.Fprintf(c.rw, "%s: %s\r\n", textproto.CanonicalMIMEHeaderKey(k), v[0])
	}
	fmt.Fprintf(c.rw, "\r\n")
	// ---------- NoBody ----------
	fmt.Fprintf(c.rw, "\r\n")

	// 缓冲数据写入，相当于发送
	if err := c.rw.Flush(); err != nil {
		return err
	}
	// 读取返回数据
	for {
		line, err := c.rw.ReadSlice('\n')
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		fmt.Print(string(line))
	}
}
