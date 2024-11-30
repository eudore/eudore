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

// HandlerFunc is a function that processes a [Context].
type HandlerFunc func(Context)

// HandlerFuncs is a collection of [HandlerFunc], representing multiple
// request processing functions.
type HandlerFuncs = []HandlerFunc

// The HandlerEmpty function is an empty handler.
func HandlerEmpty(Context) {
	// Do nothing because empty handler does not process entries.
}

// HandlerRouter403 function defines the [StatusForbidden] processing.
func HandlerRouter403(ctx Context) {
	const page404 = "403 Forbidden"
	ctx.WriteStatus(StatusForbidden)
	_ = ctx.Render(page404)
}

// HandlerRouter404 function defines the [StatusNotFound] processing.
func HandlerRouter404(ctx Context) {
	const page404 = "404 Not Found"
	ctx.WriteStatus(StatusNotFound)
	_ = ctx.Render(page404)
}

// HandlerRouter405 function defines the [StatusMethodNotAllowed] processing
// and returns [HeaderAllow] and [HeaderXEudoreRoute] Header.
func HandlerRouter405(ctx Context) {
	const page405 = "405 Method Not Allowed"
	ctx.SetHeader(HeaderAllow, ctx.GetParam(ParamAllow))
	ctx.SetHeader(HeaderXEudoreRoute, ctx.GetParam(ParamRoute))
	ctx.WriteStatus(StatusMethodNotAllowed)
	_ = ctx.Render(page405)
}

// The NewHandlerFuncsFilter function filters out nil objects in [HandlerFuncs].
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

	// Return the filtered processing function.
	nhs := make(HandlerFuncs, 0, size)
	for _, h := range hs {
		if h != nil {
			nhs = append(nhs, h)
		}
	}
	return nhs
}

// NewHandlerFuncsCombine function merges two [HandlerFuncs] into one.
// The default max length is [DefaultContextMaxHandler], which exceeds panic.
//
// Used to reconstruct the slice and prevent the slice append data from
// being confused.
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
		panic(fmt.Errorf(ErrHandlerFuncsCombineTooMany, size))
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

// The getFuncPointer function get the address of [HandlerFunc]
// as the unique identifier id.
func getFuncPointer(v reflect.Value) uintptr {
	val := *(*reflectValue)(unsafe.Pointer(&v))
	return val.ptr
}

// The SetHandlerAliasName function sets the original name of extension object.
//
// Used in the [NewHandlerExtenderBase] object and [ControllerInjectAutoRoute]
// function to pass the controller function name.
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

// SetHandlerFuncName function sets the name of a [HandlerFunc].
//
// Note: functions are not comparable, the method names of objects are
// overwritten by other method names.
func SetHandlerFuncName(i HandlerFunc, name string) {
	if name == "" {
		return
	}
	contextSaveName[getFuncPointer(reflect.ValueOf(i))] = name
}

// String method implements the [fmt.Stringer] interface
// and output [HandlerFunc] name.
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

// The NewHandlerFileSystems function uses multiple any values to create
// an [http.FileSystem] handler for static files.
//
// refer [NewHandlerFileSystem] and [NewFileSystems].
func NewHandlerFileSystems(dirs ...any) HandlerFunc {
	return NewHandlerFileSystem(NewFileSystems(dirs...))
}

// The NewHandlerFileEmbed function creates the [iofs.FS] extension function.
//
// refer [NewHandlerFileSystem].
func NewHandlerFileEmbed(fs iofs.FS) HandlerFunc {
	return NewHandlerFileSystem(NewFileSystems(fs))
}

// The NewHandlerFileSystem function creates an [http.FileSystem] extension
// function.
//
// Open the file path as [ParamPrefix] join ctx.GetParam("*").
//
// If the file is a directory and [ParamAutoIndex] is true,
// display the directory index page.
func NewHandlerFileSystem(fs http.FileSystem) HandlerFunc {
	embedTime := DefaultHandlerEmbedTime
	cacheControl := DefaultHandlerEmbedCacheControl
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
		// // The ModTime() of embed.FS is empty and cannot use cache.
		// Set it to the default time and use the 304 cache mechanism.
		modtime := stat.ModTime()
		if modtime.IsZero() {
			modtime = embedTime
		}

		switch {
		case !stat.IsDir():
			w := ctx.Response()
			if w.Header().Get(HeaderCacheControl) == "" {
				w.Header().Add(HeaderCacheControl, cacheControl)
			}
			http.ServeContent(w, ctx.Request(), stat.Name(), modtime, file)
		case GetAnyByString[bool](ctx.GetParam(ParamAutoIndex)):
			h := ctx.Response().Header()
			h.Set(HeaderCacheControl, "no-cache")
			h.Set(HeaderLastModified, modtime.UTC().Format(http.TimeFormat))
			handlerStaticDirs(ctx, "/"+ctx.GetParam("*"), file)
		default:
			ctx.WriteHeader(StatusNotFound)
		}
	}
}

type fileInfo struct {
	Name       string `json:"name" protobuf:"1,name" yaml:"name"`
	Size       int64  `json:"size" protobuf:"2,size"  yaml:"size"`
	SizeFormat string `json:"sizeFormat" protobuf:"3,sizeFormat" yaml:"sizeFormat"`
	ModTime    string `json:"modTime" protobuf:"4,modTime" yaml:"modTime"`
	UnixTime   int64  `json:"unixTime" protobuf:"5,unixTime" yaml:"unixTime"`
	IsDir      bool   `json:"isDir" protobuf:"6,isDir" yaml:"isDir"`
}

func handlerStaticDirs(ctx Context, path string, file http.File) {
	files, err := file.Readdir(-1)
	if err != nil {
		ctx.Fatal(err)
		return
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
		ctx.SetParam(ParamTemplate, DefaultHandlerEmbedTemplateName)
	}
	_ = ctx.Render(struct {
		Path  string
		Files []fileInfo
	}{path, infos})
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

// Combine multiple [http.FileSystem].
type fileSystems []http.FileSystem

// The NewFileSystems function creates a hybrid [http.FileSystem] object
// that returns the first [http.File] from multiple [http.FileSystems].
//
// If the type is string and path exists, it will be converted to [http.Dir];
// If the type is [iofs.FS] or [embed.FS] converted to [http.FS];
// If the type is [http.FileSystem], add it directly.
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

// The Open method returns the first [http.File] from multiple
// [http.FileSystems].
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
