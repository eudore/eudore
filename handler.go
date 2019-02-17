package eudore

type (
	// Context handle func
	HandlerFunc func(Context)
	// Handler interface
	Handler interface {
		Handle(Context)
	}

	// Middleware interface
	Middleware interface {
		Handler
		GetNext() Middleware
		SetNext(Middleware)
	}

	MiddlewareBase struct {
		Handler
		Next Middleware
	}
)

// Convert the HandlerFunc function to a Handler interface.
//
// 转换HandlerFunc函数成Handler接口。
func (f HandlerFunc) Handle(ctx Context) {
	f(ctx)
}


// 创建一个基础Middleware，组合一个Handler。
func NewMiddlewareBase(h Handler) Middleware {
	return &MiddlewareBase{
		Handler:	h,
		Next:		nil,
	}
}

func (m *MiddlewareBase) GetNext() Middleware {
	return m.Next
}

func (m *MiddlewareBase) SetNext(nm Middleware) {
	m.Next = nm
}


// 将多个Handler转换成一个Middleware。
// func NewMutilHandler(hs ...Handler) Middleware {
func NewMiddlewareLink(hs ...Handler) Middleware {
	var head, link Middleware
	for _, h := range hs {
		m, ok := h.(Middleware)
		if !ok {
			m = NewMiddlewareBase(h)
		}
		if link == nil {
			head, link = m, m
		}else {
			link = GetMiddlewareEnd(link)
			link.SetNext(m)
		}
	}
	return head
}


// 读取Middleware的最后一个元素。
func GetMiddlewareEnd(m Middleware) Middleware {
	link := m
	next := link.GetNext()
	for next != nil {
		link = next
		next = link.GetNext()
	}
	return link
}
