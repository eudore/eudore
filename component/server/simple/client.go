package simple

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

// Client 定义http客户端。
type Client struct {
	nc     net.Conn
	rw     *bufio.ReadWriter
	header Header
}

// NewRequest 函数发送一个http请求。
func NewRequest(method, host, url string) error {
	// 建立连接
	nc, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		return err
	}
	// 创建Http Client
	c := &Client{
		nc:     nc,
		header: make(Header),
		rw:     bufio.NewReadWriter(bufio.NewReader(nc), bufio.NewWriter(nc)),
	}
	// 设置net/http.Server唯一必要Header host
	c.header.Add("Host", host)
	// 写入请求行
	fmt.Fprintf(c.rw, "%s %s HTTP/1.1\r\n", method, url)
	// 写入Header
	for k, v := range c.header {
		fmt.Fprintf(c.rw, "%s: %s\r\n", k, v[0])
	}
	// header结束换行
	fmt.Fprintf(c.rw, "\r\n")
	// body结束换行
	fmt.Fprintf(c.rw, "\r\n")

	// 缓冲数据写入，相当于发送
	if err := c.rw.Flush(); err != nil {
		return err
	}
	// 读取返回数据
	for {
		line, err := c.rw.ReadSlice('\n')
		if err != nil {
			return err
		}
		fmt.Print(string(line))
	}
}
