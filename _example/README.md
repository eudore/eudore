# example

本部分为演示例子目录功能导航,保存eudore、middleware、policy实现的功能演示，eudore只有没实现的功能，没有无法实现的功能，详细文档查看[wiki文档](https://github.com/eudore/eudore/wiki)或者[源码](https://github.com/eudore/eudore),仅保证eudore和middleware两个库的稳定性，Alpha的演示库不稳定。

单元测试： `CGO_ENABLED=1 go test -v -timeout=2m -race -cover -coverpkg='github.com/eudore/eudore,github.com/eudore/eudore/middleware,github.com/eudore/eudore/policy' *_test.go
`

go version go1.20.1 linux/amd64	coverage: 100.0% of statements in github.com/eudore/eudore, github.com/eudore/eudore/middleware, github.com/eudore/eu
dore/policy

因版本变化未全部更新。

- Application
	- [New](appNew.go)
	- [静态文件](appStatic.go)
	- [全局请求中间件](appMiddleware.go)
	- [自定义app](appExtend.go)
- Config
	- [map存储配置](configMap.go)
	- [结构体存储配置](configEudore.go)
	- [eudore配置读写锁](configEudoreLocker.go)
	- [eudore字符串路径层次访问属性](configEudorePath.go)
	- [自定义配置解析函数](configOption.go)
	- [自定义读取http远程配置](configReadHttp.go)
	- [读取json文件配置](configReadFile.go)
	- [解析命令行参数](configArgs.go)
	- [解析环境变量](configEnvs.go)
	- [使用结构体描述生成帮助信息](configEudoreHelp.go)
- Logger
	- [初始化日志LoggerInit](loggerInit.go)
	- [LoggerStd](loggerStd.go)
	- [日志切割](loggerStdRotate.go)
	- [日志清理](loggerStdClean.go)
	- [写入Elastic](loggerElastic.go)
	- [日志脱敏](loggerSensitive.go)
	- [logrus库适配](loggerLogrus.go)
- Client
- Server
	- [设置超时](serverStd.go)
	- [服务监听](serverListen.go)
	- [使用https](serverHttps.go)
	- [双向https](serverMutualTLS.go)
	- [fastcgi启动服务](serverFcgi.go)
- Router
	- [组路由](routerGroup.go)
	- [组路由和中间件](routerMiddleware.go)
	- [路由参数](routerParams.go)
	- [Any方法注册](routerAny.go)
	- [自定义路由方法](routerMethod.go)
	- [Std路由器](routerStd.go)
	- [Host路由器](routerHost.go)
	- [路由器注册调试](routerDebug.go)
	- [路由器注册移除](routerDelete.go)
	- [路由器核心简化](routerCore.go)
	- [radix树](routerRadix.go)
- Controller
	- [路由控制器](controllerAutoRoute.go)
	- [控制器组合](controllerCompose.go)
	- [控制器自定义参数](controllerParams.go)
	- [控制器错误处理](controllerError.go)
- Context
	- [Request Info](contextRequestInfo.go)
	- [Response Write](contextResponsWrite.go)
	- [请求上下文日志](contextLogger.go)
	- [Bind Body](contextBindBody.go)
	- [Bind Form](contextBindForm.go)
	- [Bind Url](contextBindUrl.go)
	- [Bind Header](contextBindHeader.go)
	- [Bind并校验结构体数据](contextBindValid.go)
	- [Query url参数](contextQuerys.go)
	- [Header](contextHeader.go)
	- [Cookie](contextCookie.go)
	- [Params](contextParams.go)
	- [Form](contextForm.go)
	- [Redirect](contextRedirect.go)
	- [Push](contextPush.go)
	- [Render](contextRender.go)
	- [Send Json](contextRenderJson.go)
	- [Send Template](contextRenderTemplate.go)
	- [文件上传](contextUpload.go)
	- [设置额外数据](contextValue.go)
- HandlerExtender
	- [默认处理](handlerDefault.go)
	- [处理ContextData扩展](handlerContextData.go)
	- [处理自定义函数类型](handlerFunc.go)
	- [处理自定义请求上下文](handlerMyContext.go)
	- [新增函数处理扩展](handlerAddExtend.go)
	- [路径匹配扩展](handlerTree.go)
	- [分级匹配扩展](handlerWarp.go)
	- [Rpc式请求](handlerRpc.go)
	- [Rpc式map请求](handlerRpcMap.go)
	- [使用embed](handlerEmbed.go)
	- [使用jwt](handlerJwt.go)
- HandlerData
	- FuncCreator 
- Middleware
	- [Admin中间件管理后台](middlewareAdmin.go)
	- [BasicAuth](middlewareBasicAuth.go)
	- [BodyLimit](middlewareBodyLimit.go)
	- [Black黑名单](middlewareBlack.go)
	- [Breaker熔断器](middlewareBreaker.go)
	- [Cache数据缓存](middlewareCache.go)
	- [Cache数据缓存自定义存储](middlewareCacheStore.go)
	- [ContextWarp](middlewareContextWarp.go)
	- [CORS跨域资源共享](middlewareCors.go)
	- [CSRF](middlewareCsrf.go)
	- [Dump捕捉请求信息](middlewareCsrf.go)
	- [Gzip压缩](middlewareGzip.go)
	- [Header写入响应](middlewareHeader.go)
	- [Header外部过滤](middlewareHeaderFilte.go)
	- [Logger访问日志](middlewareLogger.go)
	- [LoggerLevel设置请求独立的日志级别](middlewareLoggerLevel.go)
	- [Look查看对象数据](middlewareLook.go)
	- [Pprof](middlewarePprof.go)
	- [Rate限流](middlewareRateRequest.go)
	- [Rate限速](middlewareRateSpeed.go)
	- [Recover异常捕捉](middlewareRecover.go)
	- [Referer检查](middlewareReferer.go)
	- [RequestID添加](middlewareRequestID.go)
	- [Rewrite路径重写](middlewareRewrite.go)
	- [Router匹配](middlewareRouter.go)
	- [RouterRewrite](middlewareRouterRewrite.go)
	- [Timeout请求超时](middlewareTimeout.go)
	- [自定义中间件处理函数](middlewareHandle.go)
- Daemon
	- [后台启动](appDaemon.go)
	- [命令管理进程](appCommand.go)
	- [热重启](appRestart.go)
	- [重新加载配置](appReload.go)
- Policy(Alpha)
	- [Pbac](policyPbac.go)
	- [Rbac](policyRbac.go)
	- [数据权限](policyData.go)
	- [策略限制条件](policyCondition.go)
	- [策略数据表达式](policyExpression.go)
	- [控制器生成action参数](policyControllerAction.go)
 - Httptest(Alpha)
	- [发送请求](httptestRequest.go)
	- [构造多种body](httptestBody.go)
	- [使用cookie](httptestCookies.go)
	- [测试websocket](httptestWebsocket.go)
- Session
	- [gorilla session](sessionGorilla.go)
	- [beego session](sessionBeego.go)
- Websocket
	- [使用github.com/gobwas/ws库](websocketGobwas.go)
	- [使用github.com/gorilla/websocket库](websocketGorilla.go)
- net/http
	- [中间件 黑名单](nethttpBlack.go)
	- [中间件 路径重写](nethttpRewrite.go)
	- [中间件 BasicAuth](nethttpBasicAuth.go)
	- [中间件 限流](nethttpRateRequest.go)
- other
	- [反向代理](otherProxy.go)
	- [隧道代理](otherTunnel.go)
	- [http客户端简化实现](otherHttpClient.go)
	- [http服务端简化实现](otherHttpServer.go)
	- [http服务端简化](otherHttpServer2.go)
	- [监听代码自动编译重启](otherNotify.go)(Alpha)


