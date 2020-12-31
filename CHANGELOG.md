# Change Log

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
 - COnfig	配置键移除默认前缀keys
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
- GetWarp	使用map[string]interface{}创建getwarp
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