# Router


## Interface

### Define

接口定义：

```golang
type (
	// RouterMethod the route is directly registered by default. Other methods can be directly registered using the RouterRegister interface.
	//
	// RouterMethod 路由默认直接注册的方法，其他方法可以使用RouterRegister接口直接注册。
	RouterMethod interface {
		Group(string) RouterMethod
		GetParam(string) string
		SetParam(string, string) RouterMethod
		AddHandler(string, string, ...interface{}) RouterMethod
		AddMiddleware(...HandlerFunc) RouterMethod
		AddController(...Controller) RouterMethod
		AnyFunc(string, ...interface{})
		GetFunc(string, ...interface{})
		PostFunc(string, ...interface{})
		PutFunc(string, ...interface{})
		DeleteFunc(string, ...interface{})
		HeadFunc(string, ...interface{})
		PatchFunc(string, ...interface{})
		OptionsFunc(string, ...interface{})
	}
	// RouterCore interface, performs routing, middleware registration, and matches a request and returns to the handler.
	//
	// RouterCore接口，执行路由、中间件的注册和匹配一个请求并返回处理者。
	RouterCore interface {
		RegisterMiddleware(string, HandlerFuncs)
		RegisterHandler(string, string, HandlerFuncs)
		Match(string, string, Params) HandlerFuncs
	}
	// Router 接口，需要实现路由器方法、路由器核心两个接口。
	Router interface {
		RouterCore
		RouterMethod
	}
)
```

### Router

路由器接口组合路由器核心接口、路由器方法接口。

### RouterCore

路由器核心实现路由的注册和匹配，以及添加请求处理中间件。

RegisterHandler就是注册一个方法路径下的多个请求处理者。

Match就是对应的匹配实现，根据方法路径匹配出对应的多个请求处理者，同时添加对应的Params。

RegisterMiddleware用于增加请求处理中间件。

### RouterMethod

RouterMethod接口主要有下面方法

```golang
type RouterMethod {
	Group(string) RouterMethod
	AddHandler(string, string, ...interface{}) RouterMethod
	AddMiddleware(...HandlerFunc) RouterMethod
	AddController(...Controller) RouterMethod
	AnyFunc(string, ...interface{})
	...
}
```

路由器方法主要有三类函数。

第一类是Group，用于实现组路由。

第二类是Add...的方法，实现路由器方法注册对象，路由器核心的注册方法都是Register前缀，以作区分。

第三类是Any、AnyFunc等的辅助路由注册方法。

## 对象实现

### RouterRadix

RouterRadix基于radix tree算法实现，是eudore当前使用的默认路由器。

- 通配符匹配
- 捕捉路由参数
- 默认路由参数
- 子路由
- 获取路由规则

支持4种路径，'\'结尾的常量、常量、':name'形式的变量、'\*'结尾的通配符;第一个路径空格后可以增加额外的匹配命中参数。

一二是常量，三是通配符和附加参数、四是参数、五是通配符。

```
/
/index
/api/v1/* version:v1
/:name
/*
```

#### 定义

对象定义：

```golang
type (
	// Basic function router based on radix tree implementation.
	//
	// There are three basic functions: path parameter, wildcard parameter, default parameter, and parameter verification.
	// 基于基数树实现的基本功能路由器。
	//
	// 具有路径参数、通配符参数、默认参数三项基本功能。
	RouterRadix struct {
		RouterMethod
		// save middleware
		// 保存注册的中间件信息
		Print		func(...interface{})	`set:"print"`
		middtree	*middNode
		// exception handling method
		// 异常处理方法
		node404		radixNode
		nodefunc404	HandlerFuncs
		node405		radixNode
		nodefunc405	HandlerFuncs
		// various methods routing tree
		// 各种方法路由树
		root		radixNode
		get			radixNode
		post		radixNode
		put			radixNode
		delete		radixNode
		options		radixNode
		head		radixNode
		patch		radixNode
	}
	// radix节点的定义
	radixNode struct {
		// 基本信息
		kind		uint8
		path		string
		name		string
		// 每次类型子节点
		Cchildren	[]*radixNode
		Pchildren	[]*radixNode
		Wchildren	*radixNode
		// 当前节点的数据
		tags		[]string
		vals		[]string
		handlers		HandlerFuncs
	}
)
```

RouterRadix组合了RouterMethod接口对象，只需要实现路由器核心方法即可。

每个方法都是一个radixNode组成的基数树，由基数树实现了方法的注册和匹配。

radixNode的kind、path、name保存着结点类型、当前匹配路径、匹配的变量名称。

CPW就是三类结点的集合，tags和vals是默认参数、handlers是多个请求处理者。

#### RegisterMiddleware

注册中间件主要给middtree添加路径下的中间件。

middtree也是一棵基数树实现，按照路径前缀匹配来匹配，如果前缀相同就将按树从上往下全部处理者组合出来，成为一个路径下需要使用的中间件。

**注册时路径结尾为‘/’才会前缀匹配，否则只是常量路径注册，只要常量完全相同才会匹配**

```golang
// 注册中间件到中间件树中，如果存在则追加处理者。
// 如果方法非空，路径为空，修改路径为'/'。
func (r *RouterRadix) RegisterMiddleware(method ,path string, hs HandlerFuncs) {
	if len(method) != 0 && len(path) == 0 {
		path = "/"
	}
	r.Print("RegisterMiddleware:", method, path, GetHandlerNames(hs))
	if method == MethodAny {
		if path == "/" {
			r.middtree.Insert("", hs)
			r.node404.handlers = append(r.middtree.val, r.nodefunc404...)
			r.node405.Wchildren.handlers = append(r.middtree.val, r.nodefunc405...)
			return
		}
		for _, method = range RouterAllMethod {
			r.middtree.Insert(method + path, hs)
		}
	}else {
		r.middtree.Insert(method + path, hs)
	}
}
```

#### RegisterHandler

RegisterHandler注册Any方法会给全部方法树都注册一遍。

调用insertRoute方法来实现添加，同时将中间件树匹配到的请求中间件和处理者合并，从全部处理中间件和多个请求处理者合并成一个完整请求处理链。

```golang
func (r *RouterRadix) RegisterHandler(method string, path string, handler HandlerFuncs) {
	r.Print("RegisterHandler:", method, path, GetHandlerNames(handler))
	if method == MethodAny{
		for _, method := range RouterAllMethod {
			r.insertRoute(method, path, CombineHandlerFuncs(r.middtree.Lookup(method + path), handler))
		}
	}else {
		r.insertRoute(method, path, CombineHandlerFuncs(r.middtree.Lookup(method + path), handler))
	}
}
```

insertRoute方法先拿到对应的基数树，如果方法是不允许的就拿到的405树，就结束了添加。

然后将路由注册路径按结点类型切割成多段，按照常量、变量、字符串三类将路径切割出来，每类结点就是一种结点的类型，当前切割实现对应特殊的正则规则支持不够。

getSpiltPath函数将字符串路径切割例子：

```
/				[/]
/api/note/		[/api/note/]
//api/*			[/api/ *]
//api/*name		[/api/ *name]
/api/get/		[/api/get/]
/api/get		[/api/get]
/api/:get		[/api/ :get]
/api/:get/*			[/api/ :get / *]
/api/:name/info/*	[/api/ :name /info/ *]
```

然后给基数树添加结点，如果结点是存在的就返回的存在的，所以要更新当前的根节点，然后依次向下添加。

最后一个结点就是树底了，然后给树底的新结点设置多个请求处理者和默认参数。

```golang
// routerRadix.go
func (t *RouterRadix) insertRoute(method, key string, val HandlerFuncs) {
	var currentNode *radixNode = t.getTree(method)
	if currentNode == &t.node405 {
		return
	}

	// 创建节点
	args := strings.Split(key, " ")
	for _, path := range getSpiltPath(args[0]) {
		currentNode = currentNode.InsertNode(path, newRadixNode(path))
	}

	currentNode.handlers = val
	currentNode.SetTags(args)
}
```

newRadixNode函数就是简单的根据首字符来设置结点的类型和名称。

InsertNode方法就根据结点类型来添加对对应的结点集合中，常量结点需要进行分叉操作，将相同的前缀提取出来。

```golang
// 创建一个Radix树Node，会根据当前路由设置不同的节点类型和名称。
// '*'前缀为通配符节点，':'前缀为参数节点，其他未常量节点。
func newRadixNode(path string) *radixNode {
	newNode := &radixNode{path: path}
	switch path[0] {
	case '*':
		newNode.kind = radixNodeKindWildcard
		if len(path) == 1 {
			newNode.name = "*"
		}else{
			newNode.name = path[1:]
		}
	case ':':
		newNode.kind = radixNodeKindParam
		newNode.name = path[1:]
	default:
		newNode.kind = radixNodeKindConst
	}
	return newNode
}

// 给当前节点路径下添加一个子节点。
//
// 如果新节点类型是常量节点，寻找是否存在相同前缀路径的结点，
// 如果存在路径为公共前缀的结点，直接添加新结点为匹配前缀结点的子节点；
// 如果只是两结点只是拥有公共前缀，则先分叉然后添加子节点。
//
// 如果新节点类型是参数结点，会检测当前参数是否存在，存在返回已处在的节点。
//
// 如果新节点类型是通配符结点，直接设置为当前节点的通配符处理节点。
func (r *radixNode) InsertNode(path string, nextNode *radixNode) *radixNode {
	if len(path) == 0 {
		return r
	}
	nextNode.path = path
	switch nextNode.kind {
	case radixNodeKindConst:
		for i, _ := range r.Cchildren {
			subStr, find := getSubsetPrefix(path, r.Cchildren[i].path)
			if find {
				if subStr == r.Cchildren[i].path {
					nextTargetKey := strings.TrimPrefix(path, r.Cchildren[i].path)
					return r.Cchildren[i].InsertNode(nextTargetKey, nextNode)	
				}else {
					newNode := r.SplitNode(subStr, r.Cchildren[i].path)
					if newNode == nil {
						panic("Unexpect error on split node")
					}
					return newNode.InsertNode(strings.TrimPrefix(path, subStr), nextNode)
				}
			}
		}
		r.Cchildren = append(r.Cchildren, nextNode)
	case radixNodeKindParam:
		for _, i := range r.Pchildren {
			if i.path == path {
				return i
			}
		}
		r.Pchildren = append(r.Pchildren, nextNode)
	case radixNodeKindWildcard:
		r.Wchildren = nextNode
	default:
		panic("Undefined radix node type")
	}
	return nextNode
}
```


#### Match

匹配是先获得对应的方法树，然后递归查找，没有就返回404

```golang
// 匹配一个请求，如果方法不不允许直接返回node405，未匹配返回node404。
// 注意：404不支持额外参数，未实现。
func (t *RouterRadix) Match(method, path string, params Params) HandlerFuncs {
	if n := t.getTree(method).recursiveLoopup(path, params); n != nil {
		return n
	}

	// 处理404
	t.node404.AddTagsToParams(params)
	return t.node404.handlers
}
```

递归查找主要分为四步。

第一步检测当前节点是否匹配，匹配则添加参数然后返回。

第二步检测当前节点的常量子节点是否和匹配路径前缀，有前缀表示有可能匹配到了，然后截取路径递归匹配，然后不为空就表示匹配命中，返回对象

第三步检测当前节点的变量子节点是否匹配，直接截取两个‘/’间路径为当前的变量匹配内容，然后检测进一步匹配。

第四步检测当前节点是否拥有通配符结点，如果有直接执行通配符匹配。

最后如果前面四步没有匹配名字，返回nil。

```golang
// 按照顺序匹配一个路径。
// 依次检查常量节点、参数节点、通配符节点，如果有一个匹配就直接返回。
func (searchNode *radixNode) recursiveLoopup(searchKey string, params Params) HandlerFuncs {
	// 如果路径为空，当前节点就是需要匹配的节点，直接返回。
	if len(searchKey) == 0 && searchNode.handlers != nil {
		searchNode.AddTagsToParams(params)
		return searchNode.handlers
	}

	// 遍历常量Node匹配，寻找具有相同前缀的那个节点
	for _, edgeObj := range searchNode.Cchildren {
		if contrainPrefix(searchKey, edgeObj.path) {
			nextSearchKey := searchKey[len(edgeObj.path):]
			if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
				return n
			}
			break
		}
	}

	if len(searchNode.Pchildren) > 0 && len(searchKey) > 0 {
		pos := strings.IndexByte(searchKey, '/')
		if pos == -1 {
			pos = len(searchKey) 
		}
		nextSearchKey := searchKey[pos:]

		// Whether the variable Node matches in sequence is satisfied
		// 遍历参数节点是否后续匹配
		for _, edgeObj := range searchNode.Pchildren {
			if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
				params.AddParam(edgeObj.name, searchKey[:pos])
				return n
			}
		}
	}
	
	// If the current Node has a wildcard processing method that directly matches, the result is returned.
	// 若当前节点有通配符处理方法直接匹配，返回结果。
	if searchNode.Wchildren != nil {
		searchNode.Wchildren.AddTagsToParams(params)
		params.AddParam(searchNode.Wchildren.name, searchKey)
		return searchNode.Wchildren.handlers
	}

	// can't match, return nil
	// 无法匹配，返回空
	return nil
}
```

### RouterFull

基于RouterRadix额外添加了部分功能，允许对变量和通配符匹配时增校验函数。

### RouterHost

在匹配时无法获得对应的Host，需要修复。

### RouterMethodStd

默认RouterMethod的实现

```golang
// 默认路由器方法注册实现
type RouterMethodStd struct {
	RouterCore
	ControllerParseFunc
	prefix		string
	tags		string
}
```

Group方法返回一个组路由注册，原理是对每次注册添加前缀路径，如果有组默认参数就分离重新组合一下。

将路由器方法的路由器核心和控制器解析函数传递给新的标准路由器方法实现。

```golang
func (m *RouterMethodStd) Group(path string) RouterMethod {
	// 将路径前缀和路径参数分割出来
	args := strings.Split(path, " ")
	prefix := args[0]
	tags := path[len(prefix):]

	// 如果路径是'/*'或'/'结尾，则移除后缀。
	// '/*'为路由结尾，不可为路由前缀
	// '/'不可为路由前缀，会导致出现'//'
	if len(prefix) > 0 && prefix[len(prefix) - 1] == '*' {
		prefix = prefix[:len(prefix) - 1]
	}
	if len(prefix) > 0 && prefix[len(prefix) - 1] == '/' {
		prefix = prefix[:len(prefix) - 1]
	}

	// 构建新的路由方法配置器
	return &RouterMethodStd{
		RouterCore:	m.RouterCore,
		ControllerParseFunc:	m.ControllerParseFunc,
		prefix:		m.prefix + prefix,
		tags:		tags + m.tags,
	}
}
```

路由器方法和注册请求处理者、请求处理中间件、MVC控制器。

对于请求处理者和请求处理中间件，将组路由信息添加给注册，然后使用路由器核心添加。

MVC控制器注册，需要使用控制器解析函数将控制器解析对对应的路由器配置(RouterConfig),然后路由器配置将路由注入给对应的路由器方法。

**控制器解析函数见控制器实现**

```golang
func (m *RouterMethodStd) AddHandler(method ,path string, hs ...HandlerFunc) RouterMethod {
	if len(hs) > 0 {
		m.registerHandlers(method, path, hs)
	}
	return m
}

func (m *RouterMethodStd) AddMiddleware(method ,path string, hs ...HandlerFunc) RouterMethod {
	if len(hs) > 0 {
		m.RegisterMiddleware(method, m.prefix + path, hs)
	}
	return m
}

func (m *RouterMethodStd) AddController(cs ...Controller) RouterMethod {
	for _, c := range cs {
		// controllerRegister(m, c)
		config, err := m.ControllerParseFunc(c)
		if err == nil {
			config.Inject(m)
		}else {
			fmt.Println(err)
		}
	}
	return m
}
```

最后是各类方法和eudore.Handler对象的注册。

```golang
func (m *RouterMethodStd) registerHandlers(method ,path string, hs HandlerFuncs) {
	m.RouterCore.RegisterHandler(method, m.prefix + path + m.tags, hs)
}

// Router Register handler
func (m *RouterMethodStd) Any(path string, h ...Handler) {
	m.registerHandlers(MethodAny, path, handlesToFunc(h))
}

func handlesToFunc(hs []Handler) HandlerFuncs {
	h := make(HandlerFuncs, len(hs))
	for i, _ := range hs {
		h[i] = hs[i].Handle
	}
	return h
}

// RouterRegister handle func
func (m *RouterMethodStd) AnyFunc(path string, h ...HandlerFunc) {
	m.registerHandlers(MethodAny, path, h)
}
```