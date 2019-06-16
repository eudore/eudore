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
	"mime"
	"errors"
	"strings"
	"context"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"encoding/xml"
	"mime/multipart"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)


var (
	BinderDefault	=	BindFunc(BinderDefaultFunc)
	BinderUrl		=	BindFunc(BindUrlFunc)
	BinderForm		=	BindFunc(BindFormFunc)
	BinderUrlBody	=	BindFunc(BindUrlBodyFunc)
	BinderJSON		=	BindFunc(BindJsonFunc)
	BinderXML		=	BindFunc(BindXmlFunc)
)


type (
	Binder interface {
		Bind(Context, interface{}) error
	}
	BindFunc func(Context, interface{}) error
)

func (fn BindFunc) Bind(ctx Context, i interface{}) error {
	return fn(ctx, i)
}

func BinderDefaultFunc(ctx Context, i interface{}) error {
	if ctx.Method() == MethodGet || ctx.Method() == MethodHead {
		return nil
	}
	switch strings.SplitN(ctx.GetHeader(HeaderContentType), ";", 2)[0] {
	case MimeApplicationJson:
		return BinderJSON.Bind(ctx, i)
	case MimeTextXml, MimeApplicationXml:
		return BinderXML.Bind(ctx, i)
	case MimeMultipartForm:
		return BinderForm.Bind(ctx, i)
	case MimeApplicationForm:
		return BinderUrlBody.Bind(ctx, i)
	default:
		err := errors.New("bind not suppert content type " + ctx.GetHeader(HeaderContentType))
		ctx.Error(err)
		return err
	}
}

func BindUrlFunc(ctx Context, i interface{}) error {
	return nil
}

func BindFormFunc(ctx Context, i interface{}) error {
	_, params, err := mime.ParseMediaType(ctx.GetHeader(HeaderContentType))
	if err != nil {
		return err
	}

	form, err := multipart.NewReader(ctx, params["boundary"]).ReadForm(defaultMaxMemory)
	if err != nil {
		return err
	}
	go func(ctx context.Context) {
		for {
			select {
			case <- ctx.Done():
				form.RemoveAll()
				return
			}
		}

	}(ctx.Context())
	ConvertTo(form.File, i)
	return ConvertTo(form.Value, i)
}

// body读取限制32kb.
func BindUrlBodyFunc(ctx Context, i interface{}) error {
	body, err := ioutil.ReadAll(io.LimitReader(ctx, 32 << 10))
	if err != nil {
		return err
	}
	uri, err := url.ParseQuery(string(body))
	if err != nil {
		return err
	}
	return ConvertTo(uri, i)
}

func BindJsonFunc(ctx Context, i interface{}) error {
	return json.NewDecoder(ctx).Decode(i)
}

func BindXmlFunc(ctx Context, i interface{}) error {
	return xml.NewDecoder(ctx).Decode(i)
}