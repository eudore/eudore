package eudore

type (
	// ContextData 扩展Context对象，加入获取数据类型转换。
	ContextData struct {
		Context
	}
)

// NewExtendContextData 转换ContextData处理函数为Context处理函数。
func NewExtendContextData(fn func(ContextData)) HandlerFunc {
	return func(ctx Context) {
		fn(ContextData{Context: ctx})
	}
}

// GetParamBool 获取参数转换成bool类型。
func (ctx *ContextData) GetParamBool(key string) bool {
	return GetStringDefaultBool(ctx.GetParam(key), false)
}

// GetParamDefaultBool 获取参数转换成bool类型，转换失败返回默认值。
func (ctx *ContextData) GetParamDefaultBool(key string, b bool) bool {
	return GetStringDefaultBool(ctx.GetParam(key), b)
}

// GetParamInt 获取参数转换成int类型。
func (ctx *ContextData) GetParamInt(key string) int {
	return GetStringDefaultInt(ctx.GetParam(key), 0)
}

// GetParamDefaultInt 获取参数转换成int类型，转换失败返回默认值。
func (ctx *ContextData) GetParamDefaultInt(key string, i int) int {
	return GetStringDefaultInt(ctx.GetParam(key), i)
}

// GetParamInt64 获取参数转换成int64类型。
func (ctx *ContextData) GetParamInt64(key string) int64 {
	return GetStringDefaultInt64(ctx.GetParam(key), 0)
}

// GetParamDefaultInt64 获取参数转换成int64类型，转换失败返回默认值。
func (ctx *ContextData) GetParamDefaultInt64(key string, i int64) int64 {
	return GetStringDefaultInt64(ctx.GetParam(key), i)
}

// GetParamFloat32 获取参数转换成int32类型。
func (ctx *ContextData) GetParamFloat32(key string) float32 {
	return GetStringDefaultFloat32(ctx.GetParam(key), 0)
}

// GetParamDefaultFloat32 获取参数转换成int32类型，转换失败返回默认值。
func (ctx *ContextData) GetParamDefaultFloat32(key string, f float32) float32 {
	return GetStringDefaultFloat32(ctx.GetParam(key), f)
}

// GetParamFloat64 获取参数转换成float64类型。
func (ctx *ContextData) GetParamFloat64(key string) float64 {
	return GetStringDefaultFloat64(ctx.GetParam(key), 0)
}

// GetParamDefaultFloat64 获取参数转换成float64类型，转换失败返回默认值。
func (ctx *ContextData) GetParamDefaultFloat64(key string, f float64) float64 {
	return GetStringDefaultFloat64(ctx.GetParam(key), f)
}

// GetParamDefaultString 获取一个参数，如果为空字符串返回默认值。
func (ctx *ContextData) GetParamDefaultString(key, str string) string {
	return GetStringDefault(ctx.GetParam(key), str)
}

// GetHeaderBool 获取header转换成bool类型。
func (ctx *ContextData) GetHeaderBool(key string) bool {
	return GetStringDefaultBool(ctx.GetHeader(key), false)
}

// GetHeaderDefaultBool 获取header转换成bool类型，转换失败返回默认值。
func (ctx *ContextData) GetHeaderDefaultBool(key string, b bool) bool {
	return GetStringDefaultBool(ctx.GetHeader(key), b)
}

// GetHeaderInt 获取header转换成int类型。
func (ctx *ContextData) GetHeaderInt(key string) int {
	return GetStringDefaultInt(ctx.GetHeader(key), 0)
}

// GetHeaderDefaultInt 获取header转换成int类型，转换失败返回默认值。
func (ctx *ContextData) GetHeaderDefaultInt(key string, i int) int {
	return GetStringDefaultInt(ctx.GetHeader(key), i)
}

// GetHeaderInt64 获取header转换成int64类型。
func (ctx *ContextData) GetHeaderInt64(key string) int64 {
	return GetStringDefaultInt64(ctx.GetHeader(key), 0)
}

// GetHeaderDefaultInt64 获取header转换成int64类型，转换失败返回默认值。
func (ctx *ContextData) GetHeaderDefaultInt64(key string, i int64) int64 {
	return GetStringDefaultInt64(ctx.GetHeader(key), i)
}

// GetHeaderFloat32 获取header转换成float32类型。
func (ctx *ContextData) GetHeaderFloat32(key string) float32 {
	return GetStringDefaultFloat32(ctx.GetHeader(key), 0)
}

// GetHeaderDefaultFloat32 获取header转换成float32类型，转换失败返回默认值。
func (ctx *ContextData) GetHeaderDefaultFloat32(key string, f float32) float32 {
	return GetStringDefaultFloat32(ctx.GetHeader(key), f)
}

// GetHeaderFloat64 获取header转换成float64类型。
func (ctx *ContextData) GetHeaderFloat64(key string) float64 {
	return GetStringDefaultFloat64(ctx.GetHeader(key), 0)
}

// GetHeaderDefaultFloat64 获取header转换成float64类型，转换失败返回默认值。
func (ctx *ContextData) GetHeaderDefaultFloat64(key string, f float64) float64 {
	return GetStringDefaultFloat64(ctx.GetHeader(key), f)
}

// GetHeaderDefaultString 获取header，如果为空字符串返回默认值。
func (ctx *ContextData) GetHeaderDefaultString(key, str string) string {
	return GetStringDefault(ctx.GetHeader(key), str)
}

// GetQueryBool 获取uri参数值转换成bool类型。
func (ctx *ContextData) GetQueryBool(key string) bool {
	return GetStringDefaultBool(ctx.GetQuery(key), false)
}

// GetQueryDefaultBool 获取uri参数值转换成bool类型，转换失败返回默认值。
func (ctx *ContextData) GetQueryDefaultBool(key string, b bool) bool {
	return GetStringDefaultBool(ctx.GetQuery(key), b)
}

// GetQueryInt 获取uri参数值转换成int类型。
func (ctx *ContextData) GetQueryInt(key string) int {
	return GetStringDefaultInt(ctx.GetQuery(key), 0)
}

// GetQueryDefaultInt 获取uri参数值转换成int类型，转换失败返回默认值。
func (ctx *ContextData) GetQueryDefaultInt(key string, i int) int {
	return GetStringDefaultInt(ctx.GetQuery(key), i)
}

// GetQueryInt64 获取uri参数值转换成int64类型。
func (ctx *ContextData) GetQueryInt64(key string) int64 {
	return GetStringDefaultInt64(ctx.GetQuery(key), 0)
}

// GetQueryDefaultInt64 获取uri参数值转换成int64类型，转换失败返回默认值。
func (ctx *ContextData) GetQueryDefaultInt64(key string, i int64) int64 {
	return GetStringDefaultInt64(ctx.GetQuery(key), i)
}

// GetQueryFloat32 获取url参数值转换成float32类型。
func (ctx *ContextData) GetQueryFloat32(key string) float32 {
	return GetStringDefaultFloat32(ctx.GetQuery(key), 0)
}

// GetQueryDefaultFloat32 获取url参数值转换成float32类型，转换失败返回默认值。
func (ctx *ContextData) GetQueryDefaultFloat32(key string, f float32) float32 {
	return GetStringDefaultFloat32(ctx.GetQuery(key), f)
}

// GetQueryFloat64 获取url参数值转换成float64类型。
func (ctx *ContextData) GetQueryFloat64(key string) float64 {
	return GetStringDefaultFloat64(ctx.GetQuery(key), 0)
}

// GetQueryDefaultFloat64 获取url参数值转换成float64类型，转换失败返回默认值。
func (ctx *ContextData) GetQueryDefaultFloat64(key string, f float64) float64 {
	return GetStringDefaultFloat64(ctx.GetQuery(key), f)
}

// GetQueryDefaultString 获取一个uri参数的值，如果为空字符串返回默认值。
func (ctx *ContextData) GetQueryDefaultString(key, str string) string {
	return GetStringDefault(ctx.GetQuery(key), str)
}

// GetCookieBool 获取一个cookie的转换成bool类型。
func (ctx *ContextData) GetCookieBool(key string) bool {
	return GetStringDefaultBool(ctx.GetCookie(key), false)
}

// GetCookieDefaultBool 获取一个cookie的转换成bool类型，转换失败返回默认值
func (ctx *ContextData) GetCookieDefaultBool(key string, b bool) bool {
	return GetStringDefaultBool(ctx.GetCookie(key), b)
}

// GetCookieInt 获取一个cookie的转换成int类型。
func (ctx *ContextData) GetCookieInt(key string) int {
	return GetStringDefaultInt(ctx.GetCookie(key), 0)
}

// GetCookieDefaultInt 获取一个cookie的转换成int类型，转换失败返回默认值
func (ctx *ContextData) GetCookieDefaultInt(key string, i int) int {
	return GetStringDefaultInt(ctx.GetCookie(key), i)
}

// GetCookieInt64 获取一个cookie的转换成int64类型。
func (ctx *ContextData) GetCookieInt64(key string) int64 {
	return GetStringDefaultInt64(ctx.GetCookie(key), 0)
}

// GetCookieDefaultInt64 获取一个cookie的转换成int64类型，转换失败返回默认值
func (ctx *ContextData) GetCookieDefaultInt64(key string, i int64) int64 {
	return GetStringDefaultInt64(ctx.GetCookie(key), i)
}

// GetCookieFloat32 获取一个cookie的转换成float32类型。
func (ctx *ContextData) GetCookieFloat32(key string) float32 {
	return GetStringDefaultFloat32(ctx.GetCookie(key), 0)
}

// GetCookieDefaultFloat32 获取一个cookie的转换成float32类型，转换失败返回默认值
func (ctx *ContextData) GetCookieDefaultFloat32(key string, f float32) float32 {
	return GetStringDefaultFloat32(ctx.GetCookie(key), f)
}

// GetCookieFloat64 获取一个cookie的转换成float64类型。
func (ctx *ContextData) GetCookieFloat64(key string) float64 {
	return GetStringDefaultFloat64(ctx.GetCookie(key), 0)
}

// GetCookieDefaultFloat64 获取一个cookie的转换成float64类型，转换失败返回默认值
func (ctx *ContextData) GetCookieDefaultFloat64(key string, f float64) float64 {
	return GetStringDefaultFloat64(ctx.GetCookie(key), f)
}

// GetCookieDefaultString 获取一个cookie的值，如果为空字符串返回默认值。
func (ctx *ContextData) GetCookieDefaultString(key, str string) string {
	return GetStringDefault(ctx.GetCookie(key), str)
}
