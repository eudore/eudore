package eudore

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"net/url"
	"strings"
)

/*
Binder

Binder对象用于请求数据反序列化，

默认根据http请求的Content-Type header指定的请求数据格式来解析数据。

支持设置map和结构体，目前未加入使用uri参数。

定义：binder.go

*/

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

// 定义各种默认的Binder对象。
var (
	BinderDefault = BindFunc(BinderDefaultFunc)
	BinderUrl     = BindFunc(BindUrlFunc)
	BindHeader    = BindFunc(BindHeaderFunc)
	BinderForm    = BindFunc(BindFormFunc)
	BinderUrlBody = BindFunc(BindUrlBodyFunc)
	BinderJSON    = BindFunc(BindJsonFunc)
	BinderXML     = BindFunc(BindXmlFunc)
)

type (
	// Binder 定义Binder接口。
	Binder interface {
		Bind(Context, interface{}) error
	}
	// BindFunc 实现Binder接口，将BindFunc转换成Binder对象。
	BindFunc func(Context, interface{}) error
)

// Bind 方法实现Binder接口。
func (fn BindFunc) Bind(ctx Context, i interface{}) error {
	return fn(ctx, i)
}

// BinderDefaultFunc 函数实现默认
func BinderDefaultFunc(ctx Context, i interface{}) error {
	if ctx.Method() == MethodGet || ctx.Method() == MethodHead {
		return BinderUrl.Bind(ctx, i)
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
		err := errors.New("bind not suppert content type: " + ctx.GetHeader(HeaderContentType))
		ctx.Error(err)
		return err
	}
}

// BindUrlFunc 函数使用url参数实现bind。
func BindUrlFunc(ctx Context, i interface{}) error {
	ctx.Querys().Range(func(k, v string) {
		Set(i, k, v)
	})
	return nil
}

// BindFormFunc 函数使用form格式body实现bind。
func BindFormFunc(ctx Context, i interface{}) error {
	ConvertTo(ctx.FormFiles(), i)
	return ConvertTo(ctx.FormValues(), i)
}

// BindUrlBodyFunc 函数使用url格式body实现bind，body读取限制32kb。
func BindUrlBodyFunc(ctx Context, i interface{}) error {
	body, err := ioutil.ReadAll(io.LimitReader(ctx, 32<<10))
	if err != nil {
		return err
	}
	uri, err := url.ParseQuery(string(body))
	if err != nil {
		return err
	}
	return ConvertTo(uri, i)
}

// BindJsonFunc 函数使用json格式body实现bind。
func BindJsonFunc(ctx Context, i interface{}) error {
	return json.NewDecoder(ctx).Decode(i)
}

// BindXmlFunc 函数使用xml格式body实现bind。
func BindXmlFunc(ctx Context, i interface{}) error {
	return xml.NewDecoder(ctx).Decode(i)
}

// BindHeaderFunc 函数实现使用header数据bind。
func BindHeaderFunc(ctx Context, i interface{}) error {
	ctx.Request().Header().Range(func(k, v string) {
		Set(i, k, v)
	})
	return nil
}

// BindWithHeader 实现Binder额外封装bind header。
func BindWithHeader(b Binder) Binder {
	return BindFunc(func(ctx Context, i interface{}) error {
		BindHeader(ctx, i)
		return b.Bind(ctx, i)
	})
}
