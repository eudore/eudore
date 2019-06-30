package fasthttp

type (
	Header struct {
		header FastHeaer
	}
	FastHeaer interface {
		Set(string, string)
		Add(string, string)
		Del(string)
		Peek(string) []byte
		VisitAll(func([]byte, []byte))
	}
)

func (h *Header) Add(key, val string) {
	h.header.Add(key, val)
}

func (h *Header) Set(key, val string) {
	h.header.Set(key, val)
}

func (h *Header) Del(key string) {
	h.header.Del(key)
}

func (h *Header) Get(key string) string {
	return string(h.header.Peek(key))
}

func (h *Header) Range(fn func(string, string)) {
	h.header.VisitAll(func(key, val []byte) {
		fn(string(key), string(val))
	})
}
