package eudore

import (
	// "strings"
	"net/http"
	"net/textproto"
)

type (
	Params interface {
		GetParam(string) string
		AddParam(string, string)
		SetParam(string, string)
	}
	HeaderHttp map[string][]string

	// 用于响应返回的set-cookie header的数据生成
	SetCookie = http.Cookie
	// 用于请求读取的cookie header的键值对数据存储
	Cookie struct {
		Name  string
		Value string
	}
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
