package middleware

/*
重要Cookie设置SameSite属性也可以防止阅览器请求CSRF攻击。
*/

import (
	"math/rand"
	"net/http"
	"strings"

	"github.com/eudore/eudore"
)

// NewCsrfFunc 函数创建一个Csrf处理函数，key指定请求带有crsf参数的关键字，cookie是csrf设置cookie的基本详细。
//
// key value:
//
// - "csrf"
//
// - "query: csrf"
//
// - "header: X-CSRF-Token"
//
// - "form: csrf"
//
// - func(ctx eudore.Context) string {return ctx.Query("csrf")}
//
// - nil
//
// cookie value:
//
// - "csrf"
//
// - http.Cookie{Name: "csrf"}
//
// - nil
func NewCsrfFunc(key, cookie interface{}) eudore.HandlerFunc {
	keyfunc := getCsrfTokenFunc(key)
	basecookie := getCsrfBaseCookie(cookie)
	return func(ctx eudore.Context) {
		key := ctx.GetCookie(basecookie.Name)
		if key == "" {
			key = getRandomToken()
			newcookie := basecookie
			newcookie.Value = key
			ctx.SetCookie(&newcookie)
		}
		ctx.SetParam("csrf", key)
		switch ctx.Method() {
		case eudore.MethodGet, eudore.MethodHead, eudore.MethodOptions, eudore.MethodTrace:
			return
		}
		if keyfunc(ctx) != key {
			ctx.WriteHeader(eudore.StatusBadRequest)
			ctx.WriteString("invalid csrf token " + key)
			ctx.End()
		}
	}
}

// getCsrfBaseCookie 函数创建应该CSRF基础Cookie。
func getCsrfBaseCookie(cookie interface{}) http.Cookie {
	switch val := cookie.(type) {
	case http.Cookie:
		return val
	case *http.Cookie:
		return *val
	case string:
		return http.Cookie{Name: val}
	default:
		return http.Cookie{Name: "_csrf"}
	}
}

// getCsrfTokenFunc 函数根据key一个csrf token获取函数。
//
// 如果key是字符串类型通过query、header、form前缀返回对应获得token方法；如果key为func(eudore.Context) string类型直接返回；否在返回默认函数。
func getCsrfTokenFunc(key interface{}) func(eudore.Context) string {
	switch val := key.(type) {
	case string:
		switch {
		case strings.HasPrefix(val, "query:"):
			val = strings.TrimSpace(val[6:])
			return func(ctx eudore.Context) string {
				return ctx.GetQuery(val)
			}
		case strings.HasPrefix(val, "header:"):
			val = strings.TrimSpace(val[7:])
			return func(ctx eudore.Context) string {
				return ctx.GetHeader(val)
			}
		case strings.HasPrefix(val, "form:"):
			val = strings.TrimSpace(val[5:])
			return func(ctx eudore.Context) string {
				if strings.Index(ctx.GetHeader(eudore.HeaderContentType), eudore.MimeMultipartForm) == -1 {
					return ""
				}
				return ctx.FormValue(val)
			}
		}
	case func(eudore.Context) string:
		return val
	}
	return func(ctx eudore.Context) string {
		return ctx.GetQuery("csrf")
	}
}

// getRandomToken 函数创建一个随机字符串。
func getRandomToken() string {
	letters := []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY")
	result := make([]rune, 32)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}
