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

eudore的Middleware是一个函数，类型是eudore.HandlerFunc,可以在中间件类型使用ctx.Next来调用先后续处理函数，然后再继续执行定义的内容，ctx.End直接忽略后续处理。

ctx.Fatal默认会调用ctx.End，也是和ctx.

在ContextBase实现中，将全部请求处理者使用`SetHandler`方法传递给ctx，然后调用`Next`方法开始循环执行处理，`End`方法将执行索引移到到结束，就结束了全部的处理。

在`Next`方法中调用处理，如果处理中调用了Next方法，那么在处理中就会先将后续处理完，才会继续处理，就巧妙实现Pre和Post的Handler的调用，如果处理中没有调用Next方法，就for循环继续执行处理。

例如：

```golang
func main() {
	app := eudore.NewCore()
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.Println("前执行")
		ctx.Next()
		fmt.Println("后执行")
	})
}
```
