# eudore 使用`*sql.DB`
 
预定义`*sql.Stmt`和处理函数扩展

## 处理函数扩展

```golang
type ContextDB struct {
	eudore.Context
	*sql.DB
}

func main() {
	db,err := sql.Open("", "")
	if err != nil {
		return
	}

	eudore.RegisterHandlerFunc(func(fn func(ContextDB)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			fn(ContextDB{
				Context: ctx,
				DB: db,
			})
		}
	})

	app := eudore.NewCore()
	app.GetFunc("/*", func(ctx ContextDB) {
		ctx.Query("...")
	})
}
```