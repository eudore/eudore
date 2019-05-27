package eudore

type (
	ContextData struct {
		Context
	}
	HandlerDataFunc func(ContextData)
)

func HandlerContextExtent(i interface{}) HandlerFunc {
	switch fn := i.(type) {
	case func(ContextData): 
		return func(ctx Context) {
			fn(ContextData{ctx})
		}
	}
	return nil
}

func (fn HandlerDataFunc) Handle(ctx Context) {
	fn(ContextData{ctx})
}

func (ctx *ContextData) GetParamBool(key string) bool {
	return GetStringDefaultBool(ctx.GetParam(key), false)
}

func (ctx *ContextData) GetParamDefaultBool(key string, b bool) bool {
	return GetStringDefaultBool(ctx.GetParam(key), b)
}

func (ctx *ContextData) GetParamInt(key string) int {
	return GetStringDefaultInt(ctx.GetParam(key), 0)
}

func (ctx *ContextData) GetParamDefaultInt(key string, i int) int {
	return GetStringDefaultInt(ctx.GetParam(key), i)
}

func (ctx *ContextData) GetParamFloat32(key string) float32 {
	return GetStringDefaultFloat32(ctx.GetParam(key), 0)
}

func (ctx *ContextData) GetParamDefaultFloat32(key string, f float32)  float32 {
	return GetStringDefaultFloat32(ctx.GetParam(key), f)
}
func (ctx *ContextData) GetParamFloat64(key string) float64 {
	return GetStringDefaultFloat64(ctx.GetParam(key), 0)
}

func (ctx *ContextData) GetParamDefaultFloat64(key string, f float64)  float64 {
	return GetStringDefaultFloat64(ctx.GetParam(key), f)
}

func (ctx *ContextData) GetParamString(key string) string {
	return ctx.GetParam(key)
}

func (ctx *ContextData) GetParamDefaultString(key, str string) string {
	return GetStringDefault(ctx.GetParam(key), str)
}

func (ctx *ContextData) GetHeaderBool(key string) bool {
	return GetStringDefaultBool(ctx.GetHeader(key), false)
}

func (ctx *ContextData) GetHeaderDefaultBool(key string, b bool) bool {
	return GetStringDefaultBool(ctx.GetHeader(key), b)
}

func (ctx *ContextData) GetHeaderInt(key string) int {
	return GetStringDefaultInt(ctx.GetHeader(key), 0)
}

func (ctx *ContextData) GetHeaderDefaultInt(key string, i int) int {
	return GetStringDefaultInt(ctx.GetHeader(key), i)
}

func (ctx *ContextData) GetHeaderFloat32(key string) float32 {
	return GetStringDefaultFloat32(ctx.GetHeader(key), 0)
}

func (ctx *ContextData) GetHeaderDefaultFloat32(key string, f float32)  float32 {
	return GetStringDefaultFloat32(ctx.GetHeader(key), f)
}
func (ctx *ContextData) GetHeaderFloat64(key string) float64 {
	return GetStringDefaultFloat64(ctx.GetHeader(key), 0)
}

func (ctx *ContextData) GetHeaderDefaultFloat64(key string, f float64)  float64 {
	return GetStringDefaultFloat64(ctx.GetHeader(key), f)
}

func (ctx *ContextData) GetHeaderString(key string) string {
	return ctx.GetHeader(key)
}

func (ctx *ContextData) GetHeaderDefaultString(key, str string) string {
	return GetStringDefault(ctx.GetHeader(key), str)
}

func (ctx *ContextData) GetQueryBool(key string) bool {
	return GetStringDefaultBool(ctx.GetQuery(key), false)
}

func (ctx *ContextData) GetQueryDefaultBool(key string, b bool) bool {
	return GetStringDefaultBool(ctx.GetQuery(key), b)
}

func (ctx *ContextData) GetQueryInt(key string) int {
	return GetStringDefaultInt(ctx.GetQuery(key), 0)
}

func (ctx *ContextData) GetQueryDefaultInt(key string, i int) int {
	return GetStringDefaultInt(ctx.GetQuery(key), i)
}

func (ctx *ContextData) GetQueryFloat32(key string) float32 {
	return GetStringDefaultFloat32(ctx.GetQuery(key), 0)
}

func (ctx *ContextData) GetQueryDefaultFloat32(key string, f float32)  float32 {
	return GetStringDefaultFloat32(ctx.GetQuery(key), f)
}
func (ctx *ContextData) GetQueryFloat64(key string) float64 {
	return GetStringDefaultFloat64(ctx.GetQuery(key), 0)
}

func (ctx *ContextData) GetQueryDefaultFloat64(key string, f float64)  float64 {
	return GetStringDefaultFloat64(ctx.GetQuery(key), f)
}

func (ctx *ContextData) GetQueryString(key string) string {
	return ctx.GetQuery(key)
}

func (ctx *ContextData) GetQueryDefaultString(key, str string) string {
	return GetStringDefault(ctx.GetQuery(key), str)
}

func (ctx *ContextData) GetCookieBool(key string) bool {
	return GetStringDefaultBool(ctx.GetCookie(key), false)
}

func (ctx *ContextData) GetCookieDefaultBool(key string, b bool) bool {
	return GetStringDefaultBool(ctx.GetCookie(key), b)
}

func (ctx *ContextData) GetCookieInt(key string) int {
	return GetStringDefaultInt(ctx.GetCookie(key), 0)
}

func (ctx *ContextData) GetCookieDefaultInt(key string, i int) int {
	return GetStringDefaultInt(ctx.GetCookie(key), i)
}

func (ctx *ContextData) GetCookieFloat32(key string) float32 {
	return GetStringDefaultFloat32(ctx.GetCookie(key), 0)
}

func (ctx *ContextData) GetCookieDefaultFloat32(key string, f float32)  float32 {
	return GetStringDefaultFloat32(ctx.GetCookie(key), f)
}
func (ctx *ContextData) GetCookieFloat64(key string) float64 {
	return GetStringDefaultFloat64(ctx.GetCookie(key), 0)
}

func (ctx *ContextData) GetCookieDefaultFloat64(key string, f float64)  float64 {
	return GetStringDefaultFloat64(ctx.GetCookie(key), f)
}

func (ctx *ContextData) GetCookieString(key string) string {
	return ctx.GetCookie(key)
}

func (ctx *ContextData) GetCookieDefaultString(key, str string) string {
	return GetStringDefault(ctx.GetCookie(key), str)
}

func (ctx *ContextData) GetSessionBool(key string) bool {
	return GetDefaultBool(ctx.GetSession().Get(key), false)
}

func (ctx *ContextData) GetSessionDefaultBool(key string, b bool) bool {
	return GetDefaultBool(ctx.GetSession().Get(key), b)
}

func (ctx *ContextData) GetSessionInt(key string) int {
	return GetDefaultInt(ctx.GetSession().Get(key), 0)
}

func (ctx *ContextData) GetSessionDefaultInt(key string, i int) int {
	return GetDefaultInt(ctx.GetSession().Get(key), i)
}

func (ctx *ContextData) GetSessionFloat32(key string) float32 {
	return GetDefaultFloat32(ctx.GetSession().Get(key), 0)
}

func (ctx *ContextData) GetSessionDefaultFloat32(key string, f float32)  float32 {
	return GetDefaultFloat32(ctx.GetSession().Get(key), f)
}
func (ctx *ContextData) GetSessionFloat64(key string) float64 {
	return GetDefaultFloat64(ctx.GetSession().Get(key), 0)
}

func (ctx *ContextData) GetSessionDefaultFloat64(key string, f float64)  float64 {
	return GetDefaultFloat64(ctx.GetSession().Get(key), f)
}

func (ctx *ContextData) GetSessionString(key string) string {
	return GetDefaultString(ctx.GetSession().Get(key), "")
}

func (ctx *ContextData) GetSessionDefaultString(key, str string) string {
	return GetDefaultString(ctx.GetSession().Get(key), str)
}
