# Handler

标准处理函数定义：

```golang
type (
	// HandlerFunc 是处理一个Context的函数
	HandlerFunc func(Context)
	// HandlerFuncs 是HandlerFunc的集合，表示多个请求处理函数。
	HandlerFuncs []HandlerFunc
)
```

其他框架处理函数都是一个固定的函数，而eudore几乎支持任意处理函数，只需要额外注册一下转换函数，将任意函数转换成eudore.HanderFunc对象即可，在基于路由匹配返回字符串的benckmark测试中，处理函数转换造成的性能丢失低于1%。

基于处理函数扩展机制，可以给处理请求的请求上下文添加任意函数。

可以使用eudore.ListExtendHandlerFun()函数查看内置支持的任意函数类型，如果是注册了不支持的处理函数类型会触发panic。

内置扩展处理函数类型：

```godoc
func(eudore.Context) error
func(eudore.Context) (interface {}, error)
func(eudore.ContextData)
func(eudore.Context, map[string]interface {}) (map[string]interface {}, error)
```

```golang

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})
	app.GetFunc("/check", func(ctx eudore.Context) error {
		if len(ctx.GetQuery("value")) > 3 {
			return fmt.Errorf("value is %s len great 3", ctx.GetQuery("value"))
		}
		return nil
	})
	app.GetFunc("/data", func(ctx eudore.Context) (interface{}, error) {
		return map[string]string{
			"a": "1",
			"b": "2",
		}, nil
	})
	app.Listen(":8080")
	app.Run()
}
```

### 实现一个扩展函数

MyContext额外实现了一个Hello方法，然后使用eudore.RegisterHandlerFunc注册一个转换函数，转换函数要求参数是一个函数，返回参数是一个eudore.HandlerFunc。

闭包一个`func(fn func(MyContext)) eudore.HandlerFunc`转换函数，就将MyContext类型处理函数转换成了eudore.HandlerFunc，然后就可以使用路由注册自己定义的处理函数。

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
