package middleware

import (
	"github.com/eudore/eudore"
)

// NewContextWarpFunc 函数中间件使之后的处理函数使用的eudore.Context对象为新的Context。
//
// 装饰器下可以直接对Context进行包装，而责任链下无法修改Context主体故设计该中间件作为中间件执行机制补充。
func NewContextWarpFunc(fn func(eudore.Context) eudore.Context) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		index, handler := ctx.GetHandler()
		wctx := &contextWarp{
			Context: fn(ctx),
			index:   index,
			handler: handler,
		}
		wctx.Next()
		ctx.SetHandler(wctx.index, wctx.handler)
	}
}

type contextWarp struct {
	eudore.Context
	index   int
	handler eudore.HandlerFuncs
}

// SetHandler 方法设置请求上下文的全部请求处理者。
func (ctx *contextWarp) SetHandler(index int, hs eudore.HandlerFuncs) {
	ctx.index, ctx.handler = index, hs
}

// GetHandler 方法获取请求上下文的当前处理索引和全部请求处理者。
func (ctx *contextWarp) GetHandler() (int, eudore.HandlerFuncs) {
	return ctx.index, ctx.handler
}

// Next 方法调用请求上下文下一个处理函数。
func (ctx *contextWarp) Next() {
	ctx.index++
	for ctx.index < len(ctx.handler) {
		ctx.handler[ctx.index](ctx)
		ctx.index++
	}
}

// End 结束请求上下文的处理。
func (ctx *contextWarp) End() {
	ctx.index = 0xff
}
