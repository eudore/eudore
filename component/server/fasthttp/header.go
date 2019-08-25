package fasthttp

type (
	// Header 定义适配header，实现protocol.Header接口。
	Header struct {
		header FastHeaer
	}
	// FastHeaer 将header方法提取成接口，可以使用公用header
	FastHeaer interface {
		Set(string, string)
		Add(string, string)
		Del(string)
		Peek(string) []byte
		VisitAll(func([]byte, []byte))
	}
)

// Add 方法添加一个Header值。
func (h *Header) Add(key, val string) {
	h.header.Add(key, val)
}

// Set 方法设置一个Header值。
func (h *Header) Set(key, val string) {
	h.header.Set(key, val)
}

// Del 方法删除一个Header值。
func (h *Header) Del(key string) {
	h.header.Del(key)
}

// Get 方法获得一个Header值。
func (h *Header) Get(key string) string {
	return string(h.header.Peek(key))
}

// Range 方法遍历Header全部键值。
func (h *Header) Range(fn func(string, string)) {
	h.header.VisitAll(func(key, val []byte) {
		fn(string(key), string(val))
	})
}
