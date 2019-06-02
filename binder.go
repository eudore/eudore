/*
Binder

Binder对象用于请求数据反序列化，

默认根据http请求的Content-Type header指定的请求数据格式来解析数据。

支持设置map和结构体，目前未加入使用uri参数。

定义：binder.go

*/
package eudore

import (
	"io"
	"fmt"
	"mime"
	"errors"
	"strings"
	// "strconv"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"encoding/xml"
	"mime/multipart"
	"github.com/eudore/eudore/protocol"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)


var (
	BinderDefault	=	BindFunc(BinderDefaultFunc)
	BinderForm		=	BindFunc(BindFormFunc)
	BinderUrl		=	BindFunc(BindUrlFunc)
	BinderJSON		=	BindFunc(BindJsonFunc)
	BinderXML		=	BindFunc(BindXmlFunc)
)


type (
	Binder interface {
		Bind(protocol.RequestReader, interface{}) error
	}
	BindFunc func(protocol.RequestReader, interface{}) error
)

func (fn BindFunc) Bind(r protocol.RequestReader, i interface{}) error {
	return fn(r, i)
}

func BinderDefaultFunc(r protocol.RequestReader, i interface{}) error {
	switch strings.SplitN(r.Header().Get(HeaderContentType), ";", 2)[0] {
	case MimeApplicationJson:
		return BinderJSON.Bind(r, i)
	case MimeTextXml, MimeApplicationXml:
		return BinderXML.Bind(r, i)
	case MimeMultipartForm:
		return BinderForm.Bind(r, i)
	case MimeApplicationForm:
		return BinderUrl.Bind(r, i)
	default:
		fmt.Println(errors.New("bind not suppert content type " + r.Header().Get(HeaderContentType)))
		return errors.New("bind not suppert content type " + r.Header().Get(HeaderContentType))
	}
}

func BindFormFunc(r protocol.RequestReader, i interface{}) error {
	_, params, err := mime.ParseMediaType(r.Header().Get(HeaderContentType))
	if err != nil {
		return err
	}

	form, err := multipart.NewReader(r, params["boundary"]).ReadForm(defaultMaxMemory)
	if err != nil {
		return err
	}
	ConvertTo(form.File, i)
	return ConvertTo(form.Value, i)
}

// body读取限制32kb.
func BindUrlFunc(r protocol.RequestReader, i interface{}) error {
	body, err := ioutil.ReadAll(io.LimitReader(r, 32 << 10))
	if err != nil {
		return err
	}
	uri, err := url.ParseQuery(string(body))
	if err != nil {
		return err
	}
	return ConvertTo(uri, i)
}

func BindJsonFunc(r protocol.RequestReader, i interface{}) error {
	return json.NewDecoder(r).Decode(i)
}

func BindXmlFunc(r protocol.RequestReader, i interface{}) error {
	return xml.NewDecoder(r).Decode(i)
}