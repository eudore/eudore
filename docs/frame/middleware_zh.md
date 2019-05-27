# Handler & Middleware

Handler接口定义了处理Context的方法。

```golang
type (
	// Context handle func
	HandlerFunc func(Context)
	// Handler interface
	Handler interface {
		Handle(Context)
	}
	HandlerFuncs	[]HandlerFunc
)
```

而eudore中间件处理函数与eudore.HanderFunc相同，通过ctx的next实现中间件机制。

```golang
func (ctx *ContextBase) SetHandler(fs HandlerFuncs) {
	ctx.index = -1
	ctx.handler = fs
}

func (ctx *ContextBase) Next() {
	ctx.index++
	for ctx.index < len(ctx.handler) {
		ctx.handler[ctx.index](ctx)
		ctx.index++
	}
}

func (ctx *ContextBase) End() {
	ctx.index = 0xff
}
```

在ContextBase实现中，将全部请求处理者使用`SetHandler`方法传递给ctx，然后调用`Next`方法开始循环执行处理，`End`方法将执行索引移到到结束，就结束了全部的处理。

在`Next`方法中调用处理，如果处理中调用了Next方法，那么在处理中就会先将后续处理完，才会继续处理，就巧妙实现Pre和Post的Handler的调用，如果处理中没有调用Next方法，就for循环继续执行处理。

例如：

```golang
ctx.Println("前执行")
ctx.Next()
fmt.Println("后执行")
```