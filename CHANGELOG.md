# Change Log

Next
- ClientOption
- FuncCreator
- EventHub

[2024年11月30日]
- App		新增使用Mutex锁全部Values。
- Logger	新增WriterAsync实现日志异步写入。
- Config	移除NewConfigParseHelp实现，Parse添加timeout。
- Client	新增ClientHook处理RoundTripper包裹，使用Hook重新实现部分功能。
- Database	移除数据库相关设计。
- Router	移除路由反注册功能和读写锁路由实现。
- Router.checkMethod	修复检查NotFound方法判断大写值。
- Router.LoggerKind		新增使用~值修改当前输出日志类型。
- RouterMux				修改逻辑实现减少代码量，更新路径切割括号。
- RouterHost			修改host映射方法，将host pattern添加带路径之前。
- Context.Context		修改Context使用上下文拷贝允许脱离请求使用，增长读写锁。
- Context.WriteStatue	修改Render Status使用独立方法。
- HandlerData			新增组合式处理，可组合Validate和Filter。
- NewHandlerDataRenders	修改Render失败时可重试下一个Render。
- NewHandlerDataRenderTemplates	新增模板渲染方式，实现io/fs加载和自动重新加载。
- HandlerExtender		新增4个泛型下扩展函数，移除部分不常用扩展函数。
- HandlerExtenderTree	修改radix实现与router.Middlewares一致。
- ControllerAutoRoute	修改注册方法根据控制器方法和路由器方法排序，显示实现可用控制器接口。
- LoggerFormatter		修复数组格式化异常。
- Event				新增SSE客户端和服务端。
- EventHub			新增Hub处理SSE和WS消息分发。
- middleware		新增DefaultPage配置各项中间件默认响应页面。
- middleware		新增中间件实现Timeout、BodySize、Health、Metadata、ServerTiming。
- middleware		更新cors、referer、rewrute与routerHost、routerMux使用一致radix实现。
- middleware		修改black、breaker、cache、rate使用Option初始化。
- middleware/black		新增ipv6匹配实现和ip格式检查，启用api时使用读写锁。
- middleware/cache		移除自定义存储接口，简化Accept内容。
- middleware/compress	修复SSE判断条件，修复错误循环，移除对压缩权值处理。
- middleware/logger		修改默认参数和动态参数格式。
- middleware/pprof		更新1.21 pprof格式。
- middleware/rate		新增RateRequest返回限流状态Header。
- daemon		修改daemon配置函数参数。
- \_example		更新并合并减少example内容。


[2023年8月31日](https://github.com/eudore/eudore/tree/b68a5f9199b93ce260d14ce0b1f1b30d56ce2359)
- go.mod	升级go版本依赖从1.9到1.20，增加error embed 泛型等新版本特性支持。
- github	新增action配置添加lint和codecov。
- LoggerStd		修改为Hook结构增强扩展，添加回收文件、过滤日志、彩色Level等功能。
- Client		新增ClietOption/ClintBody，修改请求构造方法。
- RouterStd		新增Group时参数loggerkind时修改router日志输出级别，加入Metadata接口实现。
- HandlerData	validate使用新fc实现避免反射，完成filter实现过滤或修改数据。
- FuncCreator	使用泛型重构减少反射使用，额外扩展新函数规则，允许使用逻辑关系式。
- Context		FormValues调用parseForm解析方法修改，不将PostForm和Form复制数据。
- ConvertTo		移除To/ToMap等转换函数，Get/Set函数优化异常处理。
- GetAny		修改GetAny相关函数使用泛型实现，重命名移除多余函数。
- HandlerExtender	默认扩展函数重命名。
- ResponseWriter	添加WriteString和Unwrap实现。
- NewFileSystems	处理Dir和Embed的http混合文件对象。
- NewConfigParseEnvFile	配置解析env文件。
- NewConfigParseArgs	保存未处理的命令行参数。
- ServerListenConfig	使用DefaultServerListen启动监听。
- middleware/cache		新增对Accept/Accept-Encoding/304支持。
- middleware/compress	新增选择压缩方法，忽略小Body和已压缩Mime。
- middleware/bodylimit	修改忽略NoBody，使用http.MaxBytesReader限制body长度。
- daemon				重估启动命令、后台启动、信号处理、热重启，不进行单位测试覆盖。

[2022年10月31日](https://github.com/eudore/eudore/tree/de9fd1ea1b653ba6e4f9bb5c108733e3142cadf6)
- App		优化运行输出日志
- Client	完整重构
- Logger	修复Sync方法，更新其他组件日志
- Config	合并实现方法
- Render/BindProtobuf	无需proto文件进行编码
- NewContextMessage	新增函数返回请求上下文消息，复用message。
- middleware/gzip	使用自定义压缩函数，可以使用br压缩。
- middleware/look	使用自定义data获取函数。

[2022年4月30日](https://github.com/eudore/eudore/tree/b80422e67f5c9907967e36e577d23220793a6c9c)
- App和Context	生命周期管理
- HandlerExtender	允许使用扩展函数获得路由路径
- DataHandlerFunc	合并Bind Validate Filte Render
- Client	移入App组合
- Server	实现ServeConn方法
- ConvertTo	重构实现
- example	重写单位测试，不将example转换成测试文件。


[2021年10月31日](https://github.com/eudore/eudore/tree/46e8a335592dfd8b498d5f681f2f385204b80a1a)
- RouterStd			添加others方法，清理deleteRoute参数。
- ControllerAutoRoute	简化调整规则。
- middleware/look	解析Accept Header为format值，模板内容优化。
- ConfigParseFunc	ConfigParseFunc重构
- ResponseWriter	WriteHeader将延时写入
- contextBase	细节调整
- policy		增加401
- httptest		修复响应对象并非读写

[2021年8月31日](https://github.com/eudore/eudore/commit/627e6de1fa64c45873c70f86637efa2decc5763f)
- Controller	简化内容保留ControllerAutoRoute。
- middleware/admin	ui重构。
- middleware/look	重构优化数据格式，完善功能。
- middleware/pprof	重构优化数据格式。
- component/ram 移除包，pbac功能转移到policy。

[2021年6月30日](https://github.com/eudore/eudore/tree/266448deb4ed7b48ab7003032d6c8d79f41b962e)
- endpoint/gorm gorm单model控制器。
- endpoint/gorm	gorm日志接入。
- endpoint/tracer	代码无入侵全链路日志双写。
- endpoint/prometheus	采集http请求数据。
- policy	重构权限，支持pbac、rbac、数据权限。
- ConvertRows	处理sql结果绑定。
- middleware/bodylimit	新增bodylimit限制中间件。
- middleware/header		新增header写入中间件。
- middleware/requestid	自定义id函数加入参数。

[2021年4月30日](https://github.com/eudore/eudore/tree/b795b83986f06ab03d47b30d5a2f966cb3413a7b)
- endpoint		提供无入侵链路日志
- Controller	修改参数获取接口
- Controller	修改name映射为多级，方法按照字母排序。
- ConfigParseHelp	分析结构体生成帮助信息。
- RouterHost	支持对host端口处理。
- middleware/cache	新增自定义缓冲key设置。

[2021年1月31日](https://github.com/eudore/eudore/tree/6ba7a5f6603407a09ffbefbb6a830b8926afe247)
- LoggerStd	输出不使用fields嵌套属性，Fields使用切片存储。
- ContextBase	优化Form Querys Context存储使用net/http属性。
- RenderJSON	对基本类型将自动封装一层结构，非json Accept使用indent格式化。
- Config.ParseOption	修改参数类型，使用[]ConfigParseFunc传递。
- Controller	优化控制器获取路由规则函数。

[2021年1月17日](https://github.com/eudore/eudore/tree/d4c9edf68ee3a71bfe7947b5adb4659a3ae3b3d0)
- middleware/rate	重构为RateRequest和RateSpeed用于限流和限速。
- component/pprof	重构为middleware.NewLookFunc和middleare.NewPprofController()。
- middleware/cache	新增数据缓存中间件，同时具有singleflight特性。
- middleware/singleflight 数据缓存删除

[2020年12月31日](https://github.com/eudore/eudore/tree/b4c8ae5d45c01177330fb754341d612db7c36f2d)
- RouterCoreStd	将先匹配路径后匹配方法，正确处理405。
- ControllerError 新增用于New控制器时返回错误自动去处理。
- ServerGrace	正式移除热重启Server功能，华而不实。

[2020年11月15日](https://github.com/eudore/eudore/tree/2cc49fe8c2301f6d73f3ed1c99ce3b68e533c8b5)
 - Logger	移除Logout接口、重构Logger实现方式、抽象LoggerStd
 - Router	移除RouterCoreRadix，重命名RouterCoreFull为RouterCoreStd
 - ContextBase	SetLogger方法不会自动WithFiles(nil)设置Logger属性

[2020年10月31日](https://github.com/eudore/eudore/tree/f919218094d32cf319d4ad9a1f33a260b8450014)
 - ContextBase、ConfigMap、ConfigEudore、LoggerInit、LoggerStd、ServerStd、ServerFcgi、RouterCoreRadix、RouterCoreFull、RouterCoreDebug、RouterCoreHost、RouterCoreLock、ResponseWriterHTTP不再可导出，不再显示在godoc索引中
 - RouterStd checkMethod新增允许trace和connnet，优化printerror时堆栈信息，ControllerFuncExtend扩展会提示控制器方法类型
 - RouterMethod	接口删除，方法合并到Router接口中,删除OptionsFunc方法,RouterAllMethod中不再包含Optionns方法
 - RouterRadix、RouterFull 新增Trace和Connect方法存储节点
 - HandlerFunc String方法输出名称二次修复，优化名称存储方法
 - Config	配置键移除默认前缀keys
 - Controller	修改路由组合规则，仅组合xxxController这样以控制器为后缀的对象允许组合路由
 - ControllerSingleton	取消各种参数控制
 - ControllerAutoRoute	新增自动路由控制器，用于自动注册restful路由规则
 - middleware/breaker	完成重构
 - middleare/logger		状态码小于500才输出Error级日志
 - httptest	增加AddBasicAuth方法设置basicauth信息
 - component/exmaple-appTunnel	新增隧道代理演示

[2020年9月16日](https://github.com/eudore/eudore/tree/10bd82aaa71fc68dfed2d57b5f9ffc45ecdfb8b4)
- App.Run   使用CancelError属性保存cannel结束时的error并返回
- Logger    修改返回深拷贝使用方法为WithFields(nil)
- ServerListenConfig    新增属性Certificate保存启动https的证书信息
- middleware/black  优化结构数据存储及算法 耗时减少1/8 1421ns到1248ns
- component/ram 允许直接判断权限是否允许

[2020年7月31日](https://github.com/eudore/eudore/tree/8cca525455dc48d2b4aaf6f4ceb3faf1251b239c)
- Context 允许设置Logger方法设置基础Fields
- RouterStd   AddHandler添加'TEST'方法输出debug信息
- RouterCoreRadix 修改路径切割方法，加入块模式
- RouterCoreRadix RouterCoreFull允许删除路由规则
- RouterCoreHost  重构 移动component/router/host到主包
- RouterCoreDebug 重构 移动component/router/debug到主包
- RouterCoreLock  新增 用于对路由器进行并发操作
- util    重组 统一类型转换和GetWarp的使用
- middleware/dump 修复无法移除关闭的连接
- component/ram     测试覆盖率完成
- component/ram/pbac/condition 优化条件结构
- component/httptest  单位测试覆盖及优化

[2020年6月30日](https://github.com/eudore/eudore/tree/df89a634ce080e46d9ff822c5da68679570dc2d0)
- App     新增AddMiddleware方法允许添加路由前全局请求中间件
- Context Reset方法修改参数，使用http.ResponseWriter初始化
- Context Next方法新增特性在Next最后一次执行完毕后自动Put到sync.Pool
- Context Context方法重命名为GetContext，解决一些Context组合下冲突
- Context GetHandler方法新增，用于执行一些特殊操作
- HandlerFunc 修复不具有函数可比性导致的方法名称混乱
- Router  AddMiddleware将按照添加顺序排序，不再按照路由路径顺序
- Logger  优化输出日志调用位置，可以封装后正确输出位置
- Logger  WithField使用参数logout返回副本
- Params  取消接口，方便数据range
- middleware/cors    修复Add Access-Control-Allow-Origin错误 修复validateOrigin参数前缀错误 优化add headers性能
- middleware/timeout 优化逻辑 新增支持panic信息传递 修复pool回收ctx异常修复
- middleware/recover 新增支持timeout panic抛出调用栈
- middleware/black   新增高性能黑名单匹配实现
- middleware/singleflight    新增实现
- middleware/csrf    新增实现 自定义cookie选项 自定义key值
- middleware/rate    新增重构实现 内置令牌桶 新增拥有context.Deadline时Wait直到拥有令牌
- middleware/rewrite 新增高性能路由重写中间件实现
- middleware/referer 新增高性能referer检查中间件实现
- middleware/router  新增路由器中间件，基于路径匹配执行额外的处理函数
- middleware/routerRewrite   新增基于Router中间件实现Rewrite中间件
- middleware/context 新增ContextWarp中间件，修改后续函数使用的Context对象
- middleware/admin   管理后台
- component/httptest    新增支持gzip response
- component/httptest    获取Cookie值
- component/httptest    设置tls模拟请求

[2020年5月31日](https://github.com/eudore/eudore/tree/9ee797e6c7e0a23bb04e18795fdccb11d120f907)
- 单元测试覆盖100%
- App.Validater	新增Validater属性保存校验器，不再使用全局单例
- ConfigEudore	如果配置实现configRLocker接口实现锁，会使用配置对象的锁
- LoggerStd		优化io输入,支持日志切割清理等功能
- Context.Validate	方法新增
- ServerGrace	从主包移除到component/server/grace
- ServerStdConfig	方法修改时间类型为TimeDuration，优化Set和json方法使用
- validate	更新使用方法，不再使用单例
- GetWrap	使用map[string]interface{}创建getwrap
- ConvertTo	优化对象接口和指针对象转换处理
- component/pprof 优化look显示属性

[2020年4月30日](https://github.com/eudore/eudore/tree/d283ed31f0579d4015bd141afd47936c9ad4ef28)
- 主包代码行缩减至6195无依赖，单元测试覆盖率提升到98.4%剩余热重启和runtime部分。
- 修改type定义使用，不再使用type(...)语法，方法grep查找定义
- App 重构app、删除core和eudore修改扩展方案
- Config	重构配置加载函数
- middleware/basicauth	新增保存验证通过的用户名
- middleware/ram	从中间件移到到组件component/ram
- component/httptest    支持cookie、发送网络请求、ws实现
- component/pprof   实现godoc跳转，默认使用GOROOR启动一个内置godoc
- component/pprof	新增按照路径输出对象属性
- component/show    重构并移到到pprof中
- component/expvar  删除移到到pprof中
- component/session 删除实现改为适配gorilla和beego两种session
- component/notify  优化编译和启动逻辑
- example 全部同步更新。
