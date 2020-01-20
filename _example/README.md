# example

本部分为演示例子目录功能导航，详细文档查看[wike文档](https://github.com/eudore/eudore/wiki)或者[源码](https://github.com/eudore/eudore)。

- Application
	- [Core](appCore.go)
	- [Core监听代码自动编译重启](appCoreNotify.go)
	- [Eudore](appEudore.go)
	- [Eudore使用全局中间件](appEudoreGlobalMiddleware.go)
	- [Eudore处理信号](appEudoreSignal.go)
	- [Eudore注册静态文件路由](appEudoreStatic.go)
	- [Eudore监听代码自动编译重启](appEudoreNotify.go)
	- [Eudore后台启动](appEudoreDaemon.go)
	- [Eudore启动命令解析](appEudoreCommand.go)
	- [自定义app](appExtend.go)
- Config
	- [解析命令行参数](configArgs.go)
	- [解析环境变量](configEnvs.go)
	- [Eudore配置](configEudore.go)
	- [map配置](configMap.go)
	- [map差异化配置](configMapMods.go)
	- [eudore差异化配置](configEudoreMods.go)
	- [读取文件配置](configReadFile.go)
	- [读取http远程配置](configReadHttp.go)
- Logger
	- [LoggerInit](loggerInit.go)
	- [LoggerStd](loggerStd.go)
- Server
	- [使用https](serverHttps.go)
	- [eudore server启动服务](serverEudore.go)
	- [fastcgi启动服务](serveFcgi.go)
	- [服务监听](serverListen.go)
- Router
	- [组路由和中间件](routerGroupAndMiddleware.go)
	- [路由参数](routerParams.go)
	- [Raidx路由器](routerRadix.go)
	- [Full路由器](routerFull.go)
	- [Host路由器](routerHost.go)
	- [radix树](radixtre.go)
- Context
	- [Request Info](contextRequestInfo.go)
	- [Response Write](contextResponsWrite.go)
	- [请求上下文日志](contextLogger.go)
	- [Bind Body](contextBindBody.go)
	- [Bind Form](contextBindForm.go)
	- [Bind Url](contextBindUrl.go)
	- [Bind并校验结构体数据](contextBindValid.go)
	- [Header](contextHeader.go)
	- [Cookie](contextCookie.go)
	- [Params](contexParams.go)
	- [Form](contexForm.go)
	- [Redirect](contextRedirect.go)
	- [Push](contextPush.go)
	- [Render](contextRender.go)
	- [Send Json](contextRenderJson.go)
	- [Send Template](contextRenderTemplate.go)
- Context处理扩展
	- [处理ContextData扩展](handlerContextData.go)
	- [处理自定义函数类型](handlerFunc.go)
	- [处理自定义请求上下文](handlerMyContext.go)
	- [新增函数处理扩展](handlerAddExtend.go)
	- [Rpc式请求](handlerRpc.go)
	- [map Rpc式请求](handlerRpcMap.go)
	- [使用jwt](handlerJwt.go)
- Controller
	- [基础控制器](controllerBase.go)
	- [单例控制器](controllerSingleton.go)
	- [视图控制器](controllerView.go)
	- [控制器组合路由](controllerComposeRoute.go)
	- [控制器组合方法](controllerComposeMethod.go	)
	- [控制器自定义参数](controllerParams.go)
	- [控制器只读属性](controllerReadFields.go)
	- Controller Handler扩展
- Middlewar
	- [自定义中间件处理函数](middlewareHandle.go)
	- [熔断器及管理后台](middlewareBreaker.go)
	- [BasicAuth](middlewareBasicAuth.go)
	- [CORS跨域资源共享](middlewareCors.go)
	- [gzip压缩](middlewareGzip.go)
	- [限流](middlewareRate.go)
	- [异常捕捉](middlewareRevover.go)
	- [请求超时](middlewareTimeout.go)
	- [访问日志](middlewareLogger.go)
- Ram
	- [Acl权限控制](ramAcl.go)
	- [Rbac权限控制](ramRbac.go)
	- [Pbac权限控制](ramPbacl.go)
	- [混合权限控制](ramAll.go)
	- [自定义ram处理请求](ramHandle.go)
	- [控制器生成action参数](ramControllerAction.go)
- Session
	- [map保存session](sessionMap.go)
	- [数据库保存session](sessionSql.go)
- Websocket
	- [使用github.com/gobwas/ws库](websocketGobwas.go)
	- [使用github.com/gorilla/websocket库](websocketGorilla.go)
- tool
	- [转换对象成map](toolConvertMap.go)
	- [对象转换](toolConvertTo.go)
	- [基于路径读写对象](toolGetSet.go)
	- [结构体和变量校验](toolValidate.go)
- 组件
	- [pprof](componentPprof.go)
	- [expvar](componentExpver.go)
	- [http代理实现](componentProxy.go)
	- [运行时对象数据显示](componentShow.go)
	- [httptest组件](componentHttpTest.go)
	- SRI值自动设置
	- 自动http2 push
	- api模拟工具
	- 生成对象帮助信息