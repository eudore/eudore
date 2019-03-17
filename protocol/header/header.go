package header

type (
	HeaderMap map[string]string
	HeaderArray struct {
		keys	[]string
		vals	[]string
	}
)


func (h HeaderMap) Get(key string) string {
	return h[key]
}

func (h HeaderMap) Set(key string, val string) {
	h[key] = val
}

func (h HeaderMap) Add(key string, val string) {
	h[key] = val
}

func (h HeaderMap) Del(key string) {
	delete(h, key)
}

func (h HeaderMap) Range(fn func(string, string)) {
	for k, v := range h {
		fn(k, v)
	}
}




func (h *HeaderArray) Get(key string) string {
	for i, k := range h.keys {
		if k == key {
			return h.vals[i]
		}
	}
	return ""
}

func (h *HeaderArray) Set(key string, val string) {
	for i, k := range h.keys {
		if k == key {
			h.vals[i] = val
			return
		}
	}
	h.Add(key, val)
}

func (h *HeaderArray) Add(key string, val string) {
	h.keys = append(h.keys, key)
	h.vals = append(h.vals, val)
}

func (h *HeaderArray) Del(key string) {
	for i, k := range h.keys {
		if k == key {
			h.keys[i] = ""
			return
		}
	}
}

func (h *HeaderArray) Range(fn func(string, string)) {
	for i, k := range h.keys {
		fn(k, h.vals[i])
	}
}
