//go:build go1.16
// +build go1.16

package eudore

import (
	"embed"
	"net/http"
	"strings"
)

func init() {
	formartErr := "error: %v"
	formartErr13 := "error: %w"
	ErrFormatRouterStdAddController = strings.Replace(ErrFormatRouterStdAddController, formartErr, formartErr13, 1)
	ErrFormatRouterStdAddHandlerExtend = strings.Replace(ErrFormatRouterStdAddHandlerExtend, formartErr, formartErr13, 1)
	ErrFormatRouterStdRegisterHandlersRecover = strings.Replace(ErrFormatRouterStdRegisterHandlersRecover, formartErr, formartErr13, 1)

	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendFuncEmbed)
}

func NewExtendFuncEmbed(path string, f embed.FS) HandlerFunc {
	return NewHandlerEmbedFunc(f, strings.Split(getRouteParam(path, "dir"), ";")...)
}

// NewHandlerEmbedFunc 函数使用embed.FS和指定目录文件处理响应，依次寻找dirs多个目录是否存在文件，否则使用embed.FS作为默认FS返回响应。
func NewHandlerEmbedFunc(f embed.FS, dirs ...string) HandlerFunc {
	var h fileSystems
	for i := range dirs {
		if dirs[i] != "" {
			h = append(h, http.Dir(dirs[i]))
		}
	}
	h = append(h, http.FS(f))
	return func(ctx Context) {
		file, err := h.Open(ctx.GetParam("*"))
		if err != nil {
			ctx.Fatal(err)
			return
		}
		stat, _ := file.Stat()
		// embed.FS的ModTime()为空无法使用缓存，设置为启动时间使用304缓存机制。
		modtime := stat.ModTime()
		if modtime.IsZero() {
			modtime = DefaultEmbedTime
		}
		if ctx.Request().Header.Get(HeaderCacheControl) == "" {
			ctx.SetHeader(HeaderCacheControl, DefaultEmbedCacheControl)
		}
		http.ServeContent(ctx.Response(), ctx.Request(), stat.Name(), modtime, file)
	}
}

// 组合多个http.FileSystem
type fileSystems []http.FileSystem

func (fs fileSystems) Open(name string) (file http.File, err error) {
	for _, f := range fs {
		// 依次打开多个http.FileSystem返回一个成功打开的数据。
		file, err = f.Open(name)
		if err == nil {
			return
		}
	}
	return
}
