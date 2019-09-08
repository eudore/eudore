package http

import (
	"net/textproto"
)

// Header 定义实现protocol.Header
type Header struct {
	Keys []string
	Vals []string
}

// Reset 方法实现重置
func (h *Header) Reset() {
	h.Keys = h.Keys[0:0]
	h.Vals = h.Vals[0:0]
}

// Get 方法获得一个值
func (h *Header) Get(key string) string {
	key = textproto.CanonicalMIMEHeaderKey(key)
	for i, k := range h.Keys {
		if k == key {
			return h.Vals[i]
		}
	}
	return ""
}

// Set 方法设置一个值
func (h *Header) Set(key string, val string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	for i, k := range h.Keys {
		if k == key {
			h.Vals[i] = val
			return
		}
	}
	h.Add(key, val)
}

// Add 方法添加一个值
func (h *Header) Add(key string, val string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	h.Keys = append(h.Keys, key)
	h.Vals = append(h.Vals, val)
}

// Del 方法删除一个值
func (h *Header) Del(key string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	for i, k := range h.Keys {
		if k == key {
			h.Keys[i] = ""
			return
		}
	}
}

// Range 方法实现遍历全部header
func (h *Header) Range(fn func(string, string)) {
	for i, k := range h.Keys {
		if k != "" {
			fn(k, h.Vals[i])
		}
	}
}
