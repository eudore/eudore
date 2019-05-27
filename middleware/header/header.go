package header

import (

	"github.com/eudore/eudore"
)

type Header struct {
	keys 	[]string
	vals	[]string
}

func (h *Header) AddHeader(key, val string) {
	h.keys = append(h.keys, key)
	h.vals = append(h.vals, val)
}

func (h *Header) Handler(ctx eudore.Context) {
	for i, key := range h.keys {
		ctx.SetHeader(key, h.vals[i])
	}
}