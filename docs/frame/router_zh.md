# Router


## Interface

### Define

接口定义：

```golang
type (
	// Router method
	RouterMethod interface {
		AllRouterMethod() []string
		Any(string, Handler) Handler
		AnyFunc(string, HandlerFunc) Handler
		Delete(string, Handler) Handler
		DeleteFunc(string, HandlerFunc) Handler
		Get(string, Handler) Handler
		GetFunc(string, HandlerFunc) Handler
		Head(string, Handler) Handler
		HeadFunc(string, HandlerFunc) Handler
		Options(string, Handler) Handler
		OptionsFunc(string, HandlerFunc) Handler
		Patch(string, Handler) Handler
		PatchFunc(string, HandlerFunc) Handler
		Post(string, Handler) Handler
		PostFunc(string, HandlerFunc) Handler
		Put(string, Handler) Handler
		PutFunc(string, HandlerFunc) Handler
	}
	// Router Register
	RouterRegister interface {
		RegisterFunc(method string, path string, handle HandlerFunc) Handler
		RegisterHandler(method string, path string, handler Handler) Handler
		RegisterSubRoute(path string, router Router) Handler
		RegisterHandlers(...Handler) []Handler
	}
	// router
	Router interface {
		Component
		Handler
		RouterMethod
		RouterRegister
		// method path
		Match(string, string, Params) ([]Handler, string)
		GetSubRouter(string) Router
		NotFoundFunc(Handler)
	}
)
```


### RouterMethod

RouterMethod接口主要有下面三个方法

```golang
AllRouterMethod() []string
Any(string, Handler) Handler
AnyFunc(string, HandlerFunc) Handler
```

AllRouterMethod方法返回，默认Any全部请求方法

Any调用RegisterHandler函数，注册any方法下路径的处理者，本接口其他方法函数相同。

AnyFunc调用RegisterFunc函数，注册any方法下路径的处理函数，本接口其他方法函数相同。

### RouterRegister

### Router

## Component

### RouterRadix

基于radix tree算法实现。

支持4种路径，'\'结尾的常量、常量、':name'形式的变量、'\*'结尾的通配符;第一个路径空格后可以增加额外的匹配命中参数。

```
/
/index
/api/v1/* version:v1
/:name
/*
```

## Features

- 通配符匹配
- 捕捉路由参数
- 默认路由参数
- 子路由
- 获取路由规则
