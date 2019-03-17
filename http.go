package eudore

import (
	// "strings"
	"net/http"
	"net/textproto"
)

type (
	// Header = textproto.MIMEHeader
/*	Header interface {
		Get(string) string
		Set(string, string)
		Add(string, string)
		Range(func(string, string))
	}*/
	Params interface {
		GetParam(string) string
		AddParam(string, string)
		SetParam(string, string)
	}
	HeaderHttp map[string][]string
/*	Params3 struct {
		Data	[]Param2
	}*/
	// From net/http.Cookie
	CookieWrite = http.Cookie
	CookieRead struct {
		Name  string
		Value string
	}
	// source net/http
	//
	// 来源net/http
	PushOptions = http.PushOptions
)



func (h HeaderHttp) Get(key string) string {
	return textproto.MIMEHeader(h).Get(key)
}

func (h HeaderHttp) Set(key , value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

func (h HeaderHttp) Add(key , value string) {
	textproto.MIMEHeader(h).Add(key, value)
}

func (h HeaderHttp) Del(key string) {
	textproto.MIMEHeader(h).Del(key)
}


func (h HeaderHttp) Range(fn func(string, string)) {
	for k, v := range h {
		for _, vv := range v {
			fn(k, vv)
		}
	}
}
