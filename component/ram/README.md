# RAM访问控制(Resource Access Management)

RAM 主要为eudore框架而设计,不适合net/http，在nethttp中拿不到action，pbac算不出resource。

RAM实现acl、rbac、pbac三种混合鉴权。

eudor demo

```golang
func main() {
	acl := ram.NewAcl()
	rbac := ram.NewRbac()
	pbac := ram.NewPbac()
	// acl rbac pbac 绑定信息

	app := eudore.NewApp()
	app.AddMiddleware(ram.NewMiddleware(acl, rbac, pbac))

	app.Listen(":80")
	app.Run()
}
```

[数据库相关封装](https://github.com/eudore/website/blob/master/framework/ram.go)
