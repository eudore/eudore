package eudore

import (
	"io"
	"net/http"
	"github.com/eudore/eudore/protocol"
)

type (
	Client interface {
		NewRequest(string, string, io.Reader) protocol.RequestWriter
	}
	ClientHttp struct {
		*http.Client
	}
)
var (
	DefultClientHttp = NewClientHttp()
)

func NewRequest(method string, url string, body io.Reader) protocol.RequestWriter {
	return DefultClientHttp.NewRequest(method, url, body)
}

func NewClientHttp() Client {
	return &ClientHttp{
		Client:	&http.Client{},
	}
}

func (clt *ClientHttp) NewRequest(method string, url string, body io.Reader) protocol.RequestWriter {
	req, err := http.NewRequest(method ,url ,body)
	return &RequestWriterHttp{
		Client:		clt.Client,
		Request:	req,
		err:		err,
	}
}
