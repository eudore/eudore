package simple

import (
	"fmt"
	"net"
	"time"
	"bufio"
)

type Client struct {
	nc		net.Conn
	rw		*bufio.ReadWriter
	header	Header
}

func NewRequest(method, host , url string) error {
	nc, err := net.DialTimeout("tcp", host, 2 * time.Second)
	if err != nil {
		return err
	}
	c := &Client{
		nc:	nc,
		header:	Header{},
		rw:   bufio.NewReadWriter(bufio.NewReader(nc), bufio.NewWriter(nc)),
	}
	c.header.Add("Host", host)
	fmt.Fprintf(c.rw, "%s %s HTTP/1.1\r\n", method, url)
	for k, v := range c.header {
		fmt.Fprintf(c.rw, "%s: %s\r\n", k, v)
	}
	fmt.Fprintf(c.rw, "\r\n")
	fmt.Fprintf(c.rw, "\r\n")

	if err := c.rw.Flush(); err != nil {
		return err
	}
	for {
		line, err := c.rw.ReadSlice('\n')
		if err != nil {
			return err
		}
		fmt.Print(string(line))
	}
}