package main

/*
ControllerAutoRoute将控制器方法转换成处理函数注册。

ControllerInjectAutoRoute function generates routing rules based on the controller rules, and the usage method is converted into a processing function to support routers.

Routing group: If the'ControllerGroup(string) string' method is implemented, the routing group is returned; if the routing parameter ParamControllerGroup is included, it is used; otherwise, the controller name is used to turn the path.

Routing path: Convert the method with the request method as the prefix to the routing method and path, and then use the map[method]path returned by the'ControllerRoute() map[string]string' method to overwrite the routing path.

Method conversion rules: The method prefix must be a valid request method (within RouterAllMethod), the remaining path is converted to a path, ByName is converted to variable matching/:name, and the last By of the method path is converted to /*;
The return path of ControllerRoute is'-' and the method is ignored. The first character is'', which means it is a path append parameter.

Routing parameters: If you implement the'ControllerParam(string, string, string) string' method to return routing parameters, otherwise use "controllername=%s.%s controllermethod=%s".

Controller combination: If the controller combines other objects, only the methods of the object whose name suffix is ​​Controller are reserved, and other methods with embedded properties will be ignored.

ControllerInjectAutoRoute 函数基于控制器规则生成路由规则，使用方法转换成处理函数支持路由器。

路由组: 如果实现'ControllerGroup(string) string'方法返回路由组；如果包含路由参数ParamControllerGroup则使用;否则使用控制器名称驼峰转路径。

路由路径: 将请求方法为前缀的方法转换成路由方法和路径，然后使用'ControllerRoute() map[string]string'方法返回的map[method]path覆盖路由路径。

方法转换规则: 方法前缀必须是有效的请求方法(RouterAllMethod之内)，剩余路径驼峰转路径，ByName转换成变量匹配/:name,方法路径最后一个By转换成/*;
ControllerRoute返回路径为'-'则忽略方法，第一个字符为' '表示为路径追加参数。

路由参数: 如果实现'ControllerParam(string, string, string) string'方法返回路由参数，否则使用"controllername=%s.%s controllermethod=%s"。

控制器组合: 如果控制器组合了其他对象，仅保留名称后缀为Controller的对象的方法，其他嵌入属性的方法将被忽略。
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.AddController(new(autoController))

	app.Listen(":8088")
	app.Run()
}

type autoController struct {
	eudore.ControllerAutoRoute
}

// Any 方法注册 Any /*路径。
func (*autoController) Any(ctx eudore.Context) {
	ctx.Info("autoController Any")
}

// Get 方法注册 Get /*路径。
func (*autoController) GetBy(ctx eudore.Context) interface{} {
	ctx.Debug("get")
	return "get autoController"
}

// GetInfoById 方法注册GET /info/:id 路由路径。
func (*autoController) GetInfoById(ctx eudore.Context) interface{} {
	return ctx.GetParam("id")
}

// Help 方法定义一个控制器本身的方法。
func (*autoController) Help(ctx eudore.Context) {}

func (*autoController) GetData(ctx eudore.Context) {}

// ControllerRoute 方法返回控制器路由推导修改信息。
func (*autoController) ControllerRoute() map[string]string {
	return map[string]string{
		// 修改Path方法的路由注册,路径为空就是忽略改方法
		"GetData": "-",
		"Help":    "/help",
		// 给Any方法自动生成的路径添加参数iaany
		"Any": " isany=1",
	}
}
