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
	"time"
	"errors"
	"reflect"
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
	ErrNotMultipart = 	errors.New("request Content-Type isn't multipart/form-data")
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
	default: //case MIMEPOSTForm, MIMEMultipartPOSTForm:
		fmt.Println("default bind", r.Header().Get(HeaderContentType))
		return BinderForm.Bind(r, i)
	}
}

func BindFormFunc(r protocol.RequestReader, i interface{}) error {
	d, params, err := mime.ParseMediaType(r.Header().Get(HeaderContentType))
	if err != nil {
		return nil
	}
	if d != "multipart/form-data" && d != "multipart/mixed" {
		return ErrNotMultipart
	}
	form, err := multipart.NewReader(r, params["boundary"]).ReadForm(defaultMaxMemory)
	if err != nil {
		return nil
	}
	return mapFormByTag(i, form.Value)
}

// body读取限制32kb.
func BindUrlFunc(r protocol.RequestReader, i interface{}) error {
	body, err := ioutil.ReadAll(io.LimitReader(r, 32 << 10))
	if err != nil {
		return nil
	}
	uri, err := url.ParseQuery(string(body))
	if err != nil {
		return nil
	}
	return mapFormByTag(i, uri)
}

func BindJsonFunc(r protocol.RequestReader, i interface{}) error {
	return json.NewDecoder(r).Decode(i)
}

func BindXmlFunc(r protocol.RequestReader, i interface{}) error {
	return xml.NewDecoder(r).Decode(i)
}

// source gin
func mapFormByTag(ptr interface{}, form map[string][]string) error {
	typ := reflect.TypeOf(ptr).Elem()
	val := reflect.ValueOf(ptr).Elem()
	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if !structField.CanSet() {
			continue
		}

		structFieldKind := structField.Kind()
		// 从tag获取属性名称
		inputFieldName := typeField.Tag.Get("bind")

		if inputFieldName == "" {
			// 设置名称为结构体属性名
			inputFieldName = typeField.Name

			// if "form" tag is nil, we inspect if the field is a struct or struct pointer.
			// this would not make sense for JSON parsing but it does for a form
			// since data is flatten
			if structFieldKind == reflect.Ptr {
				if !structField.Elem().IsValid() {
					structField.Set(reflect.New(structField.Type().Elem()))
				}
				structField = structField.Elem()
				structFieldKind = structField.Kind()
			}
			// 如果对象属性是结构体，则设置该对象属性
			if structFieldKind == reflect.Struct {
				err := mapFormByTag(structField.Addr().Interface(), form)
				if err != nil {
					return err
				}
				continue
			}
		}
		// 从输入数据读取值
		inputValue, exists := form[inputFieldName]

		// 如果值不存在，有默认值则初始化对象并使用默认值
		if !exists {
			inputValue = strings.Split(typeField.Tag.Get("default"), ",")
		}
		if len(inputValue) == 0 {
			continue
		}

		numElems := len(inputValue)
		// 处理数组
		if structFieldKind == reflect.Slice && numElems > 0 {
			sliceOf := structField.Type().Elem().Kind()
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for i := 0; i < numElems; i++ {
				if err := setWithString(sliceOf, slice.Index(i), inputValue[i]); err != nil {
					return err
				}
			}
			val.Field(i).Set(slice)
			continue
		}
		// 处理时间类型
		if _, isTime := structField.Interface().(time.Time); isTime {
			if err := setTimeField(inputValue[0], typeField, structField); err != nil {
				return err
			}
			continue
		}
		// 处理其他类型
		if err := setWithString(typeField.Type.Kind(), structField, inputValue[0]); err != nil {
			return err
		}
	}
	return nil
}