//go:build go1.16
// +build go1.16

package eudore

import (
	"embed"
	"net/http"
	"strings"
	"time"
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

// EmbedTime 设置http返回embed文件的最后修改时间，使用http 304缓冲，默认为服务启动时间。
//
// 如果服务存在多副本部署，通过设置相同的值保持多副本间的版本一样。
var DefaultEmbedTime time.Time

// NewHandlerEmbedFunc 函数使用embed.FS和指定目录文件处理响应，依次寻找dirs多个目录是否存在文件，否则使用embed.FS作为默认FS返回响应。
func NewHandlerEmbedFunc(f embed.FS, dirs ...string) HandlerFunc {
	var h fileSystems
	for i := range dirs {
		if dirs[i] != "" {
			h = append(h, http.Dir(dirs[i]))
		}
	}
	h = append(h, http.FS(f))
	now := time.Now()
	if !DefaultEmbedTime.IsZero() {
		now = DefaultEmbedTime
	}
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
			modtime = now
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
