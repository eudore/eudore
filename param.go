package eudore

import (
	"net/url"
)

type (
	// Params is used to manage the k-v key-value pair parameters of the string type.
	//
	// Params用于管理字符串类型的k-v键值对参数。
	Params interface {
		Add(string, string)
		Del(string)
		Get(string) string
		Set(string, string)
	}
	// Values等于url.Values
	//
	// Values等于url.Values
	ParamsValues = url.Values
	// ParamsMap is a map storage parameter.
	//
	// MapParams是一个map存储参数。
	ParamsMap map[string]string
)

func NewParamsValue() Params {
	return ParamsValues{}
}


func NewParamsMap() Params {
	return ParamsMap{}
}

func (p ParamsMap) Add(key, value string) {
	p[key] = value
}

func (p ParamsMap) Del(key string) {
	delete(p, key)
}

func (p ParamsMap) Get(key string) string {
	if p == nil {
		return ""
	}
	v, ok := p[key]
	if ok {
		return v
	}
	return ""
}

func (p ParamsMap) Set(key, value string) {
	p[key] = value
}