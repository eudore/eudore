package eudore

import (
	"net/http"
	"net/http/httptest"
)

type (
	// Context handle func
	HandlerFunc func(Context)
	// Handler interface
	Handler interface {
		Handle(Context)
	}
	HandlerFuncs	[]HandlerFunc
/*	// Middleware interface
	Middleware interface {
		Handler
		GetNext() Middleware
		SetNext(Middleware)
	}

	MiddlewareBase struct {
		Handler
		Next Middleware
	}*/
)
/*
// Convert the HandlerFunc function to a Handler interface.
//
// 转换HandlerFunc函数成Handler接口。
func (f HandlerFunc) Handle(ctx Context) {
	f(ctx)
}

func NewMiddleware(h Handler) Middleware {
	m, ok := h.(Middleware)
	if ok {
		return m
	}
	return NewMiddlewareBase(h)
}

// 创建一个基础Middleware，组合一个Handler。
func NewMiddlewareBase(h Handler) Middleware {
	return &MiddlewareBase{
		Handler:	h,
		Next:		nil,
	}
}

func (m *MiddlewareBase) Handle(ctx Context) {
	m.Handler.Handle(ctx)
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
}*/

func HandlerEmpty(Context) {
	// Do nothing because empty handler does not process entries.
}


func CombineHandlers(hs1, hs2 HandlerFuncs) HandlerFuncs {
	// if nil
	if len(hs1) == 0 {
		return hs2
	}
	if len(hs2) == 0 {
		return hs1
	}
	// combine
	const abortIndex int8 = 63
	finalSize := len(hs1) + len(hs2)
	if finalSize >= int(abortIndex) {
		panic("too many handlers")
	}
	hs := make(HandlerFuncs, finalSize)
	copy(hs, hs1)
	copy(hs[len(hs1):], hs2)
	return hs
}

func TestHttpHandler(h http.Handler, method, path string) {
	r := httptest.NewRequest(method, path, nil)	
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
}
