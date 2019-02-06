package eudore

import (
	// "net/http"
)

type (
	// Context handle func
	HandlerFunc func(Context)
	// Handler interface
	Handler interface {
		Handle(Context)
	}
	ComposeHandler interface {
		ComposeHandle([]Handler)
	}
	// ArgsHandler struct {
	// 	key		[]string
	// 	val		[]string
	// }
	MutilHandler []Handler
)
/*
func NewHttpHandlerFunc(h http.HandlerFunc) Handler {
	return HandlerFunc(func(ctx Context) {
		// h(ctx.Response(), ctx.Request().(*http.Request))
	})
}

func NewHttpHandler(h http.Handler) Handler {	
	return HandlerFunc(func(ctx Context) {
		// h.ServeHTTP(ctx.Response(), ctx.Request())
	})
}
*/
func NewMutilHandler(hs ...Handler) Handler {
	return MutilHandler(hs)
}

func (hs MutilHandler) Handle(ctx Context) {
	for _, h := range hs {
		h.Handle(ctx)	
	}
}




func (f HandlerFunc) Handle(ctx Context) {
	f(ctx)
}

// func NewArgsHandler(ks, vs []string) Handler {
// 	return &ArgsHandler{
// 		key:	ks,
// 		val:	vs,
// 	}
// }

// func (h *ArgsHandler) Handle(ctx Context) {
// 	for i, k := range h.key {
// 		ctx.SetParam(k, h.val[i])
// 	}	
// }

