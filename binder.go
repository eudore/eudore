package eudore

/*
定义请求反序列化然后一个对象的数据。
默认根据http请求的Content-Type header指定的请求数据格式来解析数据。
支持设置map和结构体，目前未加入使用uri参数。
*/

import (
	"io"
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
	switch r.Header().Get("Content-Type") {
	case MimeApplicationJson:
		return BinderJSON.Bind(r, i)
	case MimeTextXml, MimeApplicationXml:
		return BinderXML.Bind(r, i)
	case MimeApplicationForm, MimeMultipartForm:
		return BinderForm.Bind(r, i)
	default: //case MIMEPOSTForm, MIMEMultipartPOSTForm:
		return BinderForm.Bind(r, i)
	}
}

func BindFormFunc(r protocol.RequestReader, i interface{}) error {
	d, params, err := mime.ParseMediaType(r.Header().Get("Content-Type"))
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
	return mapFormByTag(i, form.Value, "form")
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
	return mapFormByTag(i, uri, "uri")
}

func BindJsonFunc(r protocol.RequestReader, i interface{}) error {
	return json.NewDecoder(r).Decode(i)
}

func BindXmlFunc(r protocol.RequestReader, i interface{}) error {
	return xml.NewDecoder(r).Decode(i)
}

// source gin
func mapFormByTag(ptr interface{}, form map[string][]string, tag string) error {
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
		inputFieldName := typeField.Tag.Get(tag)
		inputFieldNameList := strings.Split(inputFieldName, ",")
		inputFieldName = inputFieldNameList[0]
		// 从tag获取默认值
		var defaultValue string
		if len(inputFieldNameList) > 1 {
			defaultList := strings.SplitN(inputFieldNameList[1], "=", 2)
			if defaultList[0] == "default" {
				defaultValue = defaultList[1]
			}
		}
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
				err := mapFormByTag(structField.Addr().Interface(), form, tag)
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
			if defaultValue == "" {
				continue
			}
			inputValue = make([]string, 1)
			inputValue[0] = defaultValue
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
/*
func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	switch valueKind {
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	case reflect.Ptr:
		if !structField.Elem().IsValid() {
			structField.Set(reflect.New(structField.Type().Elem()))
		}
		structFieldElem := structField.Elem()
		return setWithProperType(structFieldElem.Kind(), val, structFieldElem)
	default:
		return errors.New("Unknown type")
	}
	return nil
}

func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

func setTimeField(val string, structField reflect.StructField, value reflect.Value) error {
	timeFormat := structField.Tag.Get("time_format")
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	if val == "" {
		value.Set(reflect.ValueOf(time.Time{}))
		return nil
	}

	l := time.Local
	if isUTC, _ := strconv.ParseBool(structField.Tag.Get("time_utc")); isUTC {
		l = time.UTC
	}

	if locTag := structField.Tag.Get("time_location"); locTag != "" {
		loc, err := time.LoadLocation(locTag)
		if err != nil {
			return err
		}
		l = loc
	}

	t, err := time.ParseInLocation(timeFormat, val, l)
	if err != nil {
		return err
	}

	value.Set(reflect.ValueOf(t))
	return nil
}

*/