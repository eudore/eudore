// +build !go1.11

package pprof

import (
	"net/http"
)

// NewGodoc 函数创建一个空的godoc处理对象，在go1.11以下的版本不支持godoc server。
func NewGodoc(string) http.Handler {
	return nil
}
