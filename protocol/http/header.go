package http

import (
	"net/textproto"
)

type Header struct {
	Keys []string
	Vals []string
}

func (h *Header) Reset() {
	h.Keys = h.Keys[0:0]
	h.Vals = h.Vals[0:0]
}

func (h *Header) Get(key string) string {
	key = textproto.CanonicalMIMEHeaderKey(key)
	for i, k := range h.Keys {
		if k == key {
			return h.Vals[i]
		}
	}
	return ""
}

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

func (h *Header) Add(key string, val string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	h.Keys = append(h.Keys, key)
	h.Vals = append(h.Vals, val)
}

func (h *Header) Del(key string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	for i, k := range h.Keys {
		if k == key {
			h.Keys[i] = ""
			return
		}
	}
}

func (h *Header) Range(fn func(string, string)) {
	for i, k := range h.Keys {
		if k != "" {
			fn(k, h.Vals[i])
		}
	}
}
