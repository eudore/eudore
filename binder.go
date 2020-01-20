package eudore

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
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

type (
	// Binder 定义Bind函数处理请求。
	Binder func(Context, io.Reader, interface{}) error
)

// BindDefault 函数实现默认Binder。
func BindDefault(ctx Context, r io.Reader, i interface{}) error {
	if ctx.Method() == MethodGet || ctx.Method() == MethodHead {
		return BindURL(ctx, r, i)
	}
	switch strings.SplitN(ctx.GetHeader(HeaderContentType), ";", 2)[0] {
	case MimeApplicationJSON, MimeApplicationJSONUtf8:
		return BindJSON(ctx, r, i)
	case MimeTextXML, MimeTextXMLCharsetUtf8, MimeApplicationXML, MimeApplicationxmlCharsetUtf8:
		return BindXML(ctx, r, i)
	case MimeMultipartForm, MimeApplicationFormCharsetUtf8:
		return BindForm(ctx, r, i)
	case MimeApplicationForm:
		return BindURLBody(ctx, r, i)
	default:
		err := fmt.Errorf(ErrFormatBindDefaultNotSupportContentType, ctx.GetHeader(HeaderContentType))
		ctx.Error(err)
		return err
	}
}

// BindURL 函数使用url参数实现bind。
func BindURL(ctx Context, _ io.Reader, i interface{}) error {
	for key, vals := range ctx.Querys() {
		Set(i, key, vals[0])
	}
	return nil
}

// BindForm 函数使用form格式body实现bind。
func BindForm(ctx Context, _ io.Reader, i interface{}) error {
	ConvertTo(ctx.FormFiles(), i)
	return ConvertTo(ctx.FormValues(), i)
}

// BindURLBody 函数使用url格式body实现bind，body读取限制32kb。
func BindURLBody(_ Context, r io.Reader, i interface{}) error {
	body, err := ioutil.ReadAll(io.LimitReader(r, 32<<10))
	if err != nil {
		return err
	}
	uri, err := url.ParseQuery(string(body))
	if err != nil {
		return err
	}
	return ConvertTo(uri, i)
}

// BindJSON 函数使用json格式body实现bind。
func BindJSON(_ Context, r io.Reader, i interface{}) error {
	return json.NewDecoder(r).Decode(i)
}

// BindXML 函数使用xml格式body实现bind。
func BindXML(_ Context, r io.Reader, i interface{}) error {
	return xml.NewDecoder(r).Decode(i)
}

// BindHeader 函数实现使用header数据bind。
func BindHeader(ctx Context, _ io.Reader, i interface{}) error {
	for k, v := range ctx.Request().Header {
		Set(i, k, v)
	}
	return nil
}

// NewHeaderBinder 实现Binder额外封装bind header。
func NewHeaderBinder(fn Binder) Binder {
	return func(ctx Context, r io.Reader, i interface{}) error {
		BindHeader(ctx, r, i)
		return fn(ctx, r, i)
	}
}

// NewURLBinder 实现Binder在非get和head方法下实现BindURL。
func NewURLBinder(fn Binder) Binder {
	return func(ctx Context, r io.Reader, i interface{}) error {
		if ctx.Method() != MethodGet && ctx.Method() != MethodHead {
			BindURL(ctx, r, i)
		}
		return fn(ctx, r, i)
	}
}

// NewValidateBinder 实际Binder后调用ValidateStruct检测结构体对象数据是否有效。
func NewValidateBinder(fn Binder) Binder {
	return func(ctx Context, r io.Reader, i interface{}) error {
		err := fn(ctx, r, i)
		if err == nil {
			err = Validate(i)
			if err != nil {
				ctx.Error(err)
			}
		}
		return err
	}
}
