package http

type Header struct {
	keys	[]string
	vals	[]string
}

func (h *Header) Reset() {
	h.keys = h.keys[0:0]
	h.vals = h.vals[0:0]
}

func (h *Header) Get(key string) string {
	for i, k := range h.keys {
		if k == key {
			return h.vals[i]
		}
	}
	return ""
}

func (h *Header) Set(key string, val string) {
	for i, k := range h.keys {
		if k == key {
			h.vals[i] = val
			return
		}
	}
	h.Add(key, val)
}

func (h *Header) Add(key string, val string) {
	h.keys = append(h.keys, key)
	h.vals = append(h.vals, val)
}

func (h *Header) Del(key string) {
	for i, k := range h.keys {
		if k == key {
			h.keys[i] = ""
			return
		}
	}
}

func (h *Header) Range(fn func(string, string)) {
	for i, k := range h.keys {
		fn(k, h.vals[i])
	}
}
