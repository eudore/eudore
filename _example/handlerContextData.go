package main

/*
ContextData额外增加了数据类型转换方法。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AddHandlerExtend(NewExtendContextData)
	app.AnyFunc("/*", func(ctx ContextData) {
		var id int = ctx.GetQueryInt("id")
		ctx.WriteString("hello eudore core")
		ctx.Infof("id is %d", id)
	})
	app.GetFunc("/params/:key", func(ctx ContextData) {
		ctx.Debugf("bool: %#v", ctx.GetParamBool("key"))
		ctx.Debugf("int: %#v", ctx.GetParamInt("key"))
		ctx.Debugf("int: %#v", ctx.GetParamInt("key", -101))
		ctx.Debugf("int64: %#v", ctx.GetParamInt64("key"))
		ctx.Debugf("int64: %#v", ctx.GetParamInt64("key", -164))
		ctx.Debugf("float32: %#v", ctx.GetParamFloat32("key"))
		ctx.Debugf("float32: %#v", ctx.GetParamFloat32("key", -132))
		ctx.Debugf("float64: %#v", ctx.GetParamFloat64("key"))
		ctx.Debugf("float64: %#v", ctx.GetParamFloat64("key", 0))
		ctx.Debugf("string: %#v", ctx.GetParamString("keysss", "default string"))
	})
	app.GetFunc("/header", func(ctx ContextData) {
		ctx.Debugf("bool: %#v", ctx.GetHeaderBool("key"))
		ctx.Debugf("int: %#v", ctx.GetHeaderInt("key"))
		ctx.Debugf("int: %#v", ctx.GetHeaderInt("key", -101))
		ctx.Debugf("int64: %#v", ctx.GetHeaderInt64("key"))
		ctx.Debugf("int64: %#v", ctx.GetHeaderInt64("key", -164))
		ctx.Debugf("float32: %#v", ctx.GetHeaderFloat32("key"))
		ctx.Debugf("float32: %#v", ctx.GetHeaderFloat32("key", -132))
		ctx.Debugf("float64: %#v", ctx.GetHeaderFloat64("key"))
		ctx.Debugf("float64: %#v", ctx.GetHeaderFloat64("key", 0))
		ctx.Debugf("string: %#v", ctx.GetHeaderString("keysss", "default string"))
	})
	app.GetFunc("/query", func(ctx ContextData) {
		ctx.Debugf("bool: %#v", ctx.GetQueryBool("key"))
		ctx.Debugf("int: %#v", ctx.GetQueryInt("key"))
		ctx.Debugf("int: %#v", ctx.GetQueryInt("key", -101))
		ctx.Debugf("int64: %#v", ctx.GetQueryInt64("key"))
		ctx.Debugf("int64: %#v", ctx.GetQueryInt64("key", -164))
		ctx.Debugf("float32: %#v", ctx.GetQueryFloat32("key"))
		ctx.Debugf("float32: %#v", ctx.GetQueryFloat32("key", -132))
		ctx.Debugf("float64: %#v", ctx.GetQueryFloat64("key"))
		ctx.Debugf("float64: %#v", ctx.GetQueryFloat64("key", 0))
		ctx.Debugf("string: %#v", ctx.GetQueryString("keysss", "default string"))
	})
	app.GetFunc("/cookie", func(ctx ContextData) {
		ctx.Debugf("bool: %#v", ctx.GetCookieBool("key"))
		ctx.Debugf("int: %#v", ctx.GetCookieInt("key"))
		ctx.Debugf("int: %#v", ctx.GetCookieInt("key", -101))
		ctx.Debugf("int64: %#v", ctx.GetCookieInt64("key"))
		ctx.Debugf("int64: %#v", ctx.GetCookieInt64("key", -164))
		ctx.Debugf("float32: %#v", ctx.GetCookieFloat32("key"))
		ctx.Debugf("float32: %#v", ctx.GetCookieFloat32("key", -132))
		ctx.Debugf("float64: %#v", ctx.GetCookieFloat64("key"))
		ctx.Debugf("float64: %#v", ctx.GetCookieFloat64("key", 0))
		ctx.Debugf("string: %#v", ctx.GetCookieString("keysss", "default string"))
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/?id=333").Do().Out()
	client.NewRequest("GET", "/params/333").Do()
	client.NewRequest("GET", "/header").WithHeaderValue("key", "123").Do()
	client.NewRequest("GET", "/query?key=111").Do()
	client.NewRequest("GET", "/cookie").WithHeaderValue("Cookie", "key=1234").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

// ContextData 扩展Context对象，加入获取数据类型转换。
//
// 额外扩展 Get{Param,Header,Query,Cookie}{Bool,Int,Int64,Float32,Float64,String}共4*6=24个数据类型转换方法。
//
// 第一个参数为获取数据的key，第二参数是可变参数列表，返回第一个非零值。
type ContextData struct {
	eudore.Context
}

// NewExtendContextData 转换ContextData处理函数为Context处理函数。
func NewExtendContextData(fn func(ContextData)) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		fn(ContextData{Context: ctx})
	}
}

// GetParamBool 获取参数转换成bool类型。
func (ctx ContextData) GetParamBool(key string) bool {
	return eudore.GetStringBool(ctx.GetParam(key))
}

// GetParamInt 获取参数转换成int类型。
func (ctx ContextData) GetParamInt(key string, nums ...int) int {
	return eudore.GetStringInt(ctx.GetParam(key), nums...)
}

// GetParamInt64 获取参数转换成int64类型。
func (ctx ContextData) GetParamInt64(key string, nums ...int64) int64 {
	return eudore.GetStringInt64(ctx.GetParam(key), nums...)
}

// GetParamFloat32 获取参数转换成int32类型。
func (ctx ContextData) GetParamFloat32(key string, nums ...float32) float32 {
	return eudore.GetStringFloat32(ctx.GetParam(key), nums...)
}

// GetParamFloat64 获取参数转换成float64类型。
func (ctx ContextData) GetParamFloat64(key string, nums ...float64) float64 {
	return eudore.GetStringFloat64(ctx.GetParam(key), nums...)
}

// GetParamString 获取一个参数，如果为空字符串返回默认值。
func (ctx ContextData) GetParamString(key string, strs ...string) string {
	return eudore.GetString(ctx.GetParam(key), strs...)
}

// GetHeaderBool 获取header转换成bool类型。
func (ctx ContextData) GetHeaderBool(key string) bool {
	return eudore.GetStringBool(ctx.GetHeader(key))
}

// GetHeaderInt 获取header转换成int类型。
func (ctx ContextData) GetHeaderInt(key string, nums ...int) int {
	return eudore.GetStringInt(ctx.GetHeader(key), nums...)
}

// GetHeaderInt64 获取header转换成int64类型。
func (ctx ContextData) GetHeaderInt64(key string, nums ...int64) int64 {
	return eudore.GetStringInt64(ctx.GetHeader(key), nums...)
}

// GetHeaderFloat32 获取header转换成float32类型。
func (ctx ContextData) GetHeaderFloat32(key string, nums ...float32) float32 {
	return eudore.GetStringFloat32(ctx.GetHeader(key), nums...)
}

// GetHeaderFloat64 获取header转换成float64类型。
func (ctx ContextData) GetHeaderFloat64(key string, nums ...float64) float64 {
	return eudore.GetStringFloat64(ctx.GetHeader(key), nums...)
}

// GetHeaderString 获取header，如果为空字符串返回默认值。
func (ctx ContextData) GetHeaderString(key string, strs ...string) string {
	return eudore.GetString(ctx.GetHeader(key), strs...)
}

// GetQueryBool 获取uri参数值转换成bool类型。
func (ctx ContextData) GetQueryBool(key string) bool {
	return eudore.GetStringBool(ctx.GetQuery(key))
}

// GetQueryInt 获取uri参数值转换成int类型。
func (ctx ContextData) GetQueryInt(key string, nums ...int) int {
	return eudore.GetStringInt(ctx.GetQuery(key), nums...)
}

// GetQueryInt64 获取uri参数值转换成int64类型。
func (ctx ContextData) GetQueryInt64(key string, nums ...int64) int64 {
	return eudore.GetStringInt64(ctx.GetQuery(key), nums...)
}

// GetQueryFloat32 获取url参数值转换成float32类型。
func (ctx ContextData) GetQueryFloat32(key string, nums ...float32) float32 {
	return eudore.GetStringFloat32(ctx.GetQuery(key), nums...)
}

// GetQueryFloat64 获取url参数值转换成float64类型。
func (ctx ContextData) GetQueryFloat64(key string, nums ...float64) float64 {
	return eudore.GetStringFloat64(ctx.GetQuery(key), nums...)
}

// GetQueryString 获取一个uri参数的值，如果为空字符串返回默认值。
func (ctx ContextData) GetQueryString(key string, strs ...string) string {
	return eudore.GetString(ctx.GetQuery(key), strs...)
}

// GetCookieBool 获取一个cookie的转换成bool类型。
func (ctx ContextData) GetCookieBool(key string) bool {
	return eudore.GetStringBool(ctx.GetCookie(key))
}

// GetCookieInt 获取一个cookie的转换成int类型。
func (ctx ContextData) GetCookieInt(key string, nums ...int) int {
	return eudore.GetStringInt(ctx.GetCookie(key), nums...)
}

// GetCookieInt64 获取一个cookie的转换成int64类型。
func (ctx ContextData) GetCookieInt64(key string, nums ...int64) int64 {
	return eudore.GetStringInt64(ctx.GetCookie(key), nums...)
}

// GetCookieFloat32 获取一个cookie的转换成float32类型。
func (ctx ContextData) GetCookieFloat32(key string, nums ...float32) float32 {
	return eudore.GetStringFloat32(ctx.GetCookie(key), nums...)
}

// GetCookieFloat64 获取一个cookie的转换成float64类型。
func (ctx ContextData) GetCookieFloat64(key string, nums ...float64) float64 {
	return eudore.GetStringFloat64(ctx.GetCookie(key), nums...)
}

// GetCookieString 获取一个cookie的值，如果为空字符串返回默认值。
func (ctx ContextData) GetCookieString(key string, strs ...string) string {
	return eudore.GetString(ctx.GetCookie(key), strs...)
}
