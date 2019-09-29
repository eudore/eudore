# Eudore

eudore最初目标是一个库完成web开发，除了了db部分 目前主站开发额外库使用了pg驱动pq，将需要使用的功能部分相互整合起来以便更好使用。例如配置和日志就是先解析配置还是先初始化日志的问题，配置解析及之前需要输出一些日志，但是日志库初始化需要配置才行，eudore处理的方法的解决方案使用loginit保存起来，然后初始化后处理。

**eudore最大目标是可扩展**，可以根据自己需要去修改一些框架内容，而不用fork来修改源码实现，可以方便与master更新同步，所以从18年8月的初始版本就设置了各种借口以便扩展，在后续过程中不停地学习、思考、修改来优化各种实现，思考各类场景和功能需要eu该如何巧妙实现，然后修改接口及实现。

**高性能和易用是再持续修改中实现**，路由库最初参考bone实现，后来在wangsir关于radix相关资料提点下以及echo使用的benchmark例子测试结果太惨下重新实现，  httprouter代码复杂度太高难以看懂，在使用2小时补充radiz算法代码基础后 自己独立实现了erouter库的思路，完成新了路由匹配，解决了httprouter的没有匹配优先级问题，由于实现思路简单，简单扩展出带正则的路由实现routetfull，后续完成一些细节优化后性能达到httprouter匹配api的90%。  再次学习使用benchmark和pprof性能之后优化中间件机制 思考旧机制问题后子路由功能使用gin一样的执行机制，且一直都算属于最优解(完全新内存配置、静态化)。处理函数扩展机制也属于长时间思考偶尔出现的灵感，然后benckmark测githubapi，发现性能下降不到1%  然后正式加入，ctx和app两个都可以由用户组合实现对象扩展，例如ContextData这样加入类型转换这样，大幅增加易用性，并解决了Context实际是接口还是结构体的选择问题。

log部分受logrus和zap影响，个人习惯logrus这样filed那样，简单好用，zap的字符串拼接给予了进一步压缩性能空间 ，以前运行扩uber的zap性能benckmark测试 ，通过logrus对比发现性能还不错，使用json拼接实现log优化。

由于出现ctx处理函数扩展，把cache virw session这样非必要移除了ctx，如果需要使用可以使用扩展ctx或数即可，简单且更方便。

eudore中影响运行性能的router middleware logger完成性能优化，ctx完成扩展易用性大幅增加切性能丢失忽略不计，实现高性能高易用兼顾。Config部分是初始化使用不影响使用性能，http部分可以兼容net/http、fasthttp、自定义http三种server选择，性能只会可能上升，最不济就是net/http，binder和renderer就是一个函数可以不用，每部分没有出现性能瓶颈，最差和其他库差不多吧。

**eudore第四特点属于简单**，指实现简单，在理解思路后源码阅读简单，eudore函数实现基本接口即是逻辑，一般函数内部实现逻辑少，函数名就可以描述内容，实现细节少，因为有自由扩展  乱七八遭的一些东西不用管，一些默认兼容也都不要，约定大于实现，例如core就是一个很简单的app实现，方法少属性少代码少(整个文件不足百行)，但是功能基本都有，Core app主要是简单，Eudore app主要是功能齐全。几个机制多思考就好，erouter实现算法不懂可以忽略理解接口就好，中间件允许机制和ctx扩展机制不懂就别用了，反正可以跑起来。主要代码一共6k，其中2k在写各种辅助函数(类型转换、封装函数、cp标准库兼容)，主要阅读、coreapp、context、handler这个四个文件即可，其他看看接口定义就好了，其他函数一般用名称猜内容就好。

eudore mvc实现思路就是闭包形成HandlerFunc，然后使用路由器注册即可，默认的mvc和其他mvc一样使用反射来调用执行，eudore mvc的Controller是一个接口，所有行为也是自定义的，Init和Release方法就可以是控制器before和after逻辑，Inject方法实现控制器路由注入即可。

eudore实现未参考或遵循任何设计模式，设计从心，不停思考如何可以实现更多的扩展，将整体思路优化到流畅，并保证功能和任意扩展。

eudore最大的特点是解耦，除Application以外其他对象均为接口，每个对象都具有明确语义，Application是最顶级对象可以通过组合方式实现重写(参考Core)，其他对象为接口定义直接重新实现，或组合接口实现部分重写。

各部分定义：

| 名称 | 作用 | 定义 |
| ------------ | ------------ | ------------ |
| [Application](application_zh.md) | 运行对象主体 | app.go core.go eudore.go |
| [Context](context_zh.md) | 请求处理上下文 | context.go contextextend.go |
| [Router](router_zh.md) | 请求路由选择 | router.go routerradix.go routerfull.go |
| [Handler](handler_zh.md) | 请求处理函数 ||  handler.go |
| [Middleware](middleware_zh.md) | 多Handler组合运行 | handler.go |
| Logger | App和Ctx日志输出 | logger.go loggerstd.go |
| Server | http Server启动 | server.go |
| [Config](config_zh.md) | 配置数据管理 | config.go |
| [Controller](controller_zh.md) | 解析执行控制器 | controller.go |
| Bind | 请求数据反序列化 | bind.go |
| Render | 响应数据序列化 | render.go |
| Lintener | 监听和热重启 |listener.go|
| Http | 实现http的一些处理| http.go|
| NetHttp | 兼容nethttp及转换 | nethttp.go |
| Const | 常量 |const.go |
| Error | 错误 | error.go |
| Converter | 反射读写数据 | converter.go |

# Q & A

## eudore的完整学习方式？

先了解从整体上了解每种组件的作用，然后一部分的源码阅读,eudore从core.go文件开始阅读，之后每部分可以随意看。

## 设计难点？

划分每个接口的功能，每个接口的方法尽量通用，接口即为逻辑，接口实现可能会有些小细节，但是主要逻辑就是函数名称。

## 发展方向

日期：2019年8月16日 logger重写提升效率、更新或整理老旧扩展部分、文档完善
日期：2019年9月1日 view封装、session完善、httptest、文档完善
日期：2019年9月7日 工具相关优化完善

## 命名规则

Core、Eudore等app对象变量名称统一为app，Context及衍生对象变量名称统一为ctx。基础控制器命名为ControllerXxxx，应用控制器为XxxxController，这些控制器变量命名为ctl。