package eudore

import (
	"io"
	"net/http"
	"github.com/eudore/eudore/protocol"
)

type (
	Client interface {
		NewRequest(string, string, io.Reader) Client
		Header() protocol.Header
		Do() (protocol.ResponseReader, error)
	}
	ClientHttp struct {
		*http.Client
		req		*http.Request
		header	protocol.Header
	}
)
var (
	DefultClientHttp = NewClientHttp()
)

func NewRequest(method string, url string, body io.Reader) Client {
	return DefultClientHttp.NewRequest(method, url, body)
}

func NewClientHttp() Client {
	return &ClientHttp{
		Client:	&http.Client{},
	}
}

func (clt *ClientHttp) NewRequest(method string, url string, body io.Reader) Client {
	clt.req, _ = http.NewRequest(method ,url ,body)
	clt.header = HeaderHttp(clt.req.Header)
	return clt
}

func (clt *ClientHttp) Header() protocol.Header {
	return clt.header
}

func (clt *ClientHttp) Do() (protocol.ResponseReader, error) {
	resp, err := clt.Client.Do(clt.req)
	if err != nil {
		return nil, err
	}
	return NewResponseReaderHttp(resp), nil
}