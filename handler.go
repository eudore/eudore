package eudore

import (
	"errors"
	"fmt"
	iofs "io/fs"
	"math"
	"net/http"
	"os"
	filepath "path"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"unsafe"
)

// HandlerFunc is a function that processes a Context.
//
// HandlerFunc 是处理一个Context的函数。
type HandlerFunc func(Context)

// HandlerFuncs is a collection of HandlerFunc, representing multiple request processing functions.
//
// HandlerFuncs 是HandlerFunc的集合，表示多个请求处理函数。
type HandlerFuncs []HandlerFunc

// HandlerFuncs is a collection of HandlerFunc, representing multiple request processing functions.
//
// HandlerEmpty 函数定义一个空的请求上下文处理函数。
func HandlerEmpty(Context) {
	// Do nothing because empty handler does not process entries.
}

// HandlerRouter403 function defines the default 403 processing.
//
// HandlerRouter403 函数定义默认403处理。
func HandlerRouter403(ctx Context) {
	const page404 string = "403 forbidden"
	ctx.WriteHeader(StatusForbidden)
	_ = ctx.Render(page404)
}

// HandlerRouter404 function defines the default 404 processing.
//
// HandlerRouter404 函数定义默认404处理。
func HandlerRouter404(ctx Context) {
	const page404 string = "404 page not found"
	ctx.WriteHeader(StatusNotFound)
	_ = ctx.Render(page404)
}

// HandlerRouter405 function defines the default 405 processing and returns Allow and X-Match-Route Header.
//
// HandlerRouter405 函数定义默认405处理,返回Allow和X-Match-Route Header。
func HandlerRouter405(ctx Context) {
	const page405 string = "405 method not allowed"
	ctx.SetHeader(HeaderAllow, ctx.GetParam(ParamAllow))
	ctx.SetHeader(HeaderXEudoreRoute, ctx.GetParam(ParamRoute))
	ctx.WriteHeader(StatusMethodNotAllowed)
	_ = ctx.Render(page405)
}

// HandlerMetadata 函数返回从contextKey获取metadata，可以使用路由参数*或name指定key。
func HandlerMetadata(ctx Context) {
	name := GetAnyByString(ctx.GetParam("*"), ctx.GetParam("name"))
	if name != "" {
		meta := anyMetadata(ctx.Value(NewContextKey(name)))
		if meta != nil {
			_ = ctx.Render(meta)
		} else {
			HandlerRouter404(ctx)
		}
		return
	}

	keys := ctx.Value(ContextKeyAppKeys).([]any)
	metas := make(map[string]any, len(keys))
	for i := range keys {
		meta := anyMetadata(ctx.Value(keys[i]))
		if meta != nil {
			metas[fmt.Sprint(keys[i])] = meta
		}
	}
	_ = ctx.Render(metas)
}

// NewHandlerFuncsFilter 函数过滤掉多个请求上下文处理函数中的空对象。
func NewHandlerFuncsFilter(hs HandlerFuncs) HandlerFuncs {
	var size int
	for _, h := range hs {
		if h != nil {
			size++
		}
	}
	if size == len(hs) {
		return hs[:size:size]
	}

	// 返回新过滤空的处理函数。
	nhs := make(HandlerFuncs, 0, size)
	for _, h := range hs {
		if h != nil {
			nhs = append(nhs, h)
		}
	}
	return nhs
}

// NewHandlerFuncsCombine function merges two HandlerFuncs into one. The default maximum length is now 63, which exceeds panic.
//
// Used to reconstruct the slice and prevent the appended data from being confused.
//
// HandlerFuncsCombine 函数将两个HandlerFuncs合并成一个，默认现在最大长度63，超过过panic。
//
// 用于重构切片，防止切片append数据混乱。
func NewHandlerFuncsCombine(hs1, hs2 HandlerFuncs) HandlerFuncs {
	// if nil
	if len(hs1) == 0 {
		return hs2[:len(hs2):len(hs2)]
	}
	if len(hs2) == 0 {
		return hs1[:len(hs1):len(hs1)]
	}
	// combine
	size := len(hs1) + len(hs2)
	if size >= DefaultContextMaxHandler {
		panic(fmt.Errorf("HandlerFuncsCombine: too many handlers %d", size))
	}
	hs := make(HandlerFuncs, size)
	copy(hs, hs1)
	copy(hs[len(hs1):], hs2)
	return hs
}

type reflectValue struct {
	_    *uintptr
	ptr  uintptr
	flag uintptr
}

// getFuncPointer 函数获取一个reflect值的地址作为唯一标识id。
func getFuncPointer(v reflect.Value) uintptr {
	val := *(*reflectValue)(unsafe.Pointer(&v))
	return val.ptr
}

// SetHandlerAliasName 函数设置一个函数处理对象原始名称，如果扩展未生成名称，使用此值。
//
// 在handlerExtendBase对象和ControllerInjectSingleton函数中使用到，用于传递控制器函数名称。
func SetHandlerAliasName(i any, name string) {
	if name == "" {
		return
	}
	v, ok := i.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(i)
	}
	val := *(*reflectValue)(unsafe.Pointer(&v))
	names := contextAliasName[val.ptr]
	index := int(val.flag >> 10)
	if len(names) <= index {
		newnames := make([]string, index+1)
		copy(newnames, names)
		names = newnames
		contextAliasName[val.ptr] = names
	}
	names[index] = name
}

func getHandlerAliasName(v reflect.Value) string {
	val := *(*reflectValue)(unsafe.Pointer(&v))
	names := contextAliasName[val.ptr]
	index := int(val.flag >> 10)
	if index < len(names) {
		return names[index]
	}
	return ""
}

// SetHandlerFuncName function sets the name of a request context handler.
//
// Note: functions are not comparable, the method names of objects are overwritten by other method names.
//
// SetHandlerFuncName 函数设置一个请求上下文处理函数的名称。
//
// 注意：函数不具有可比性，对象的方法的名称会被其他方法名称覆盖。
func SetHandlerFuncName(i HandlerFunc, name string) {
	if name == "" {
		return
	}
	contextSaveName[getFuncPointer(reflect.ValueOf(i))] = name
}

// String method implements the fmt.Stringer interface and implements the output function name.
//
// String 方法实现fmt.Stringer接口，实现输出函数名称。
func (h HandlerFunc) String() string {
	rh := reflect.ValueOf(h)
	ptr := getFuncPointer(rh)
	name, ok := contextFuncName[ptr]
	if ok {
		return name
	}
	name, ok = contextSaveName[ptr]
	if ok {
		return name
	}
	return runtime.FuncForPC(rh.Pointer()).Name()
}

// NewHandlerStatic 函数使用多个any值创建混合静态文件处理函数。
func NewHandlerStatic(dirs ...any) HandlerFunc {
	return NewHandlerHTTPFileSystem(NewFileSystems(dirs...))
}

// NewHandlerEmbed 函数创建iofs.FS扩展函数。
func NewHandlerEmbed(fs iofs.FS) HandlerFunc {
	return NewHandlerHTTPFileSystem(NewFileSystems(fs))
}

// NewHandlerHTTPFileSystem 函数创建http.FileSystem扩展函数。
func NewHandlerHTTPFileSystem(fs http.FileSystem) HandlerFunc {
	return func(ctx Context) {
		path := filepath.Join(ctx.GetParam(ParamPrefix), ctx.GetParam("*"))
		if path == "" {
			path = "."
		}
		file, err := fs.Open(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				HandlerRouter404(ctx)
			} else if errors.Is(err, os.ErrPermission) {
				HandlerRouter403(ctx)
			}
			ctx.Error(err)
			return
		}
		defer file.Close()

		stat, _ := file.Stat()
		// embed.FS的ModTime()为空无法使用缓存，设置为默认时间使用304缓存机制。
		modtime := stat.ModTime()
		if modtime.IsZero() {
			modtime = DefaultHandlerEmbedTime
		}

		switch {
		case !stat.IsDir():
			if ctx.Response().Header().Get(HeaderCacheControl) == "" {
				ctx.SetHeader(HeaderCacheControl, DefaultHandlerEmbedCacheControl)
			}
			http.ServeContent(ctx.Response(), ctx.Request(), stat.Name(), modtime, file)
		case GetAnyByString[bool](ctx.GetParam(ParamAutoIndex)):
			ctx.SetHeader(HeaderCacheControl, "no-cache")
			ctx.SetHeader(HeaderLastModified, modtime.UTC().Format(http.TimeFormat))
			handlerStaticDirs(ctx, "/"+ctx.GetParam("*"), file)
		default:
			ctx.WriteHeader(StatusNotFound)
		}
	}
}

func handlerStaticDirs(ctx Context, path string, file http.File) {
	files, err := file.Readdir(-1)
	if err != nil {
		ctx.Fatal(err)
		return
	}

	type fileInfo struct {
		Name       string `alias:"name" json:"name" xml:"name" yaml:"name"`
		Size       int64  `alias:"size" json:"size" xml:"size" yaml:"size"`
		SizeFormat string `alias:"sizeformat" json:"sizeformat" xml:"sizeformat" yaml:"sizeformat"`
		ModTime    string `alias:"modtime" json:"modtime" xml:"modtime" yaml:"modtime"`
		UnixTime   int64  `alias:"unixtime" json:"unixtime" xml:"unixtime" yaml:"unixtime"`
		IsDir      bool   `alias:"isdir" json:"isdir" xml:"isdir" yaml:"isdir"`
	}
	infos := make([]fileInfo, len(files))
	for i := range files {
		infos[i] = fileInfo{
			Name:       files[i].Name(),
			Size:       files[i].Size(),
			SizeFormat: formatSize(files[i].Size()),
			ModTime:    files[i].ModTime().Format("1/2/06, 3:04:05 PM"),
			UnixTime:   files[i].ModTime().Unix(),
			IsDir:      files[i].IsDir(),
		}
	}
	sort.Slice(infos, func(i, j int) bool {
		if infos[i].IsDir == infos[j].IsDir {
			return infos[i].Name < infos[j].Name
		}
		return infos[i].IsDir
	})

	if ctx.GetParam(ParamTemplate) == "" {
		ctx.SetParam(ParamTemplate, DefaultTemplateNameStaticIndex)
	}
	_ = ctx.Render(struct {
		Path   string
		Files  []fileInfo
		Upload bool
	}{path, infos, GetAnyByString[bool](ctx.GetParam("upload"))})
}

func formatSize(n int64) string {
	if n < 1024 {
		return strconv.FormatInt(n, 10) + " B"
	}
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	e := math.Floor(math.Log(float64(n)) / math.Log(1024))
	v := float64(n) / math.Pow(2, e*10)
	if v < 100 {
		return fmt.Sprintf("%.1f %s", v, sizes[int(e)])
	}
	return fmt.Sprintf("%.0f %s", v, sizes[int(e)])
}

// Combine multiple http.FileSystem
//
// 组合多个http.FileSystem。
type fileSystems []http.FileSystem

// The NewFileSystems function creates a hybrid http.FileSystem object that returns the first file from multiple http.FileSystems.
//
// If the type is string and the path exists, it will be converted to http.Dir;
// If the type is embed.FS converted to http.FS;
// If the type is http.FileSystem, add it directly.
//
// NewFileSystems 函数创建一个混合http.FileSystem对象，从多个http.FileSystem返回首个文件。
//
// 如果类型为string且路径存在将转换成http.Dir;
// 如果类型为embed.FS转换成http.FS;
// 如果类型为http.FileSystem直接追加。
func NewFileSystems(dirs ...any) http.FileSystem {
	var fs fileSystems
	for i := range dirs {
		switch dir := dirs[i].(type) {
		case string:
			_, err := os.Stat(dir)
			if err == nil {
				fs = append(fs, http.Dir(dir))
			}
		case iofs.FS:
			fs = append(fs, http.FS(dir))
		case fileSystems:
			fs = append(fs, dir...)
		case http.FileSystem:
			fs = append(fs, dir)
		}
	}
	if len(fs) == 1 {
		return fs[0]
	}
	return fs
}

// Open 方法从多个http.FileSystem返回首个文件。
func (fs fileSystems) Open(name string) (file http.File, err error) {
	err = os.ErrNotExist
	for _, f := range fs {
		file, err = f.Open(name)
		if err == nil {
			return file, nil
		}
	}
	return
}
