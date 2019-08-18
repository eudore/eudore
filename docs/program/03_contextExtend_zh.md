# eudore Context扩展

eudore Context扩展主要是针对eudore.Context增加额外的方法，如果对eudore.Context接口方法修改可以使用扩展或者接口重写实现。

Context扩展最好附加处理方法和添加统一处理过程。

## Context 扩展

eudore.RouterMethod注册路由的参数是`...interface{}`,默认RouterMethod允许参数是多个函数，会调用eudore.NewHandlerFuncs()函数将其他处理函数转换成eudore.HandlerFunc.

由于注册处理函数类型多种多样，需要预先注册转换函数例如`func(fn func(MyContext)) eudore.HandlerFunc`,将func(MyContext)转换成eudore.HandlerFunc。

扩展添加方法:

```golang
type MyContext eudore.Context

func (ctx MyContext) Hello() {
	ctx.WriteString("hellp")
}

func main() {
	eudore.RegisterHandlerFunc(func(fn func(MyContext)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			fn(MyContext(ctx))
		}
	}) 

	app := eudore.NewCore()
	app.GetFunc("/*", func(ctx MyContext) {
		ctx.Hello()
	})
}
```

扩展统一处理过程：

```golang
// NewContextRenderErrorHanderFunc 函数处理func(Context) (interface{}, error)返回数据渲染和error处理。
func NewContextRenderErrorHanderFunc(fn func(Context) (interface{}, error)) HandlerFunc {
	return func(ctx Context) {
		data, err := fn(ctx)
		if err != nil {
			ctx.Fatal(err)
		}
		err = ctx.WriteRender(data)
		if err != nil {
			ctx.Fatal(err)
		}
	}
}
```