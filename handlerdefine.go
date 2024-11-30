package eudore

// defines built-in converts functions.

import (
	"net/http"
	"reflect"
	"runtime"
)

func getCallerName(i any) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// NewHandlerFunc function converts func().
func NewHandlerFunc(fn func()) HandlerFunc {
	return func(Context) {
		fn()
	}
}

// NewHandlerFuncAny function converts func() any.
func NewHandlerFuncAny(fn func() any) HandlerFunc {
	name := getCallerName(fn)
	return func(ctx Context) {
		data := fn()
		if ctx.Response().Size() == 0 {
			err := ctx.Render(data)
			if err != nil {
				ctx.WithField(ParamCaller, name).Fatal(err)
			}
		}
	}
}

// NewHandlerFuncError function converts func() error and handles error.
func NewHandlerFuncError(fn func() error) HandlerFunc {
	name := getCallerName(fn)
	return func(ctx Context) {
		err := fn()
		if err != nil {
			ctx.WithField(ParamCaller, name).Fatal(err)
		}
	}
}

// NewHandlerFuncRenderError function converts func() (any, error), handles data Render and error.
func NewHandlerFuncAnyError(fn func() (any, error)) HandlerFunc {
	name := getCallerName(fn)
	return func(ctx Context) {
		data, err := fn()
		if err == nil && ctx.Response().Size() == 0 {
			err = ctx.Render(data)
		}
		if err != nil {
			ctx.WithField(ParamCaller, name).Fatal(err)
		}
	}
}

// NewHandlerFuncContextAny function converts func(Context) any to handle data Render.
func NewHandlerFuncContextAny(fn func(Context) any) HandlerFunc {
	name := getCallerName(fn)
	return func(ctx Context) {
		data := fn(ctx)
		if ctx.Response().Size() == 0 {
			err := ctx.Render(data)
			if err != nil {
				ctx.WithField(ParamCaller, name).Fatal(err)
			}
		}
	}
}

// The NewHandlerFuncContextError function converts func(Context) error and handles the returned error.
func NewHandlerFuncContextError(fn func(Context) error) HandlerFunc {
	name := getCallerName(fn)
	return func(ctx Context) {
		err := fn(ctx)
		if err != nil {
			ctx.WithField(ParamCaller, name).Fatal(err)
		}
	}
}

// NewHandlerFuncContextAnyError function converts func(Context) (any, error), handles data Render and error.
func NewHandlerFuncContextAnyError(fn func(Context) (any, error)) HandlerFunc {
	name := getCallerName(fn)
	return func(ctx Context) {
		data, err := fn(ctx)
		if err == nil && ctx.Response().Size() == 0 {
			err = ctx.Render(data)
		}
		if err != nil {
			ctx.WithField(ParamCaller, name).Fatal(err)
		}
	}
}

func NewHandlerFuncContextType[T any](fn func(Context, T)) HandlerFunc {
	name := getCallerName(fn)
	return func(ctx Context) {
		req := new(T)
		err := ctx.Bind(req)
		if err != nil {
			ctx.WithField(ParamCaller, name).Fatal(err)
			return
		}

		fn(ctx, *req)
	}
}

func NewHandlerFuncContextTypeError[T any](fn func(Context, T) error) HandlerFunc {
	name := getCallerName(fn)
	return func(ctx Context) {
		req := new(T)
		err := ctx.Bind(req)
		if err != nil {
			ctx.WithField(ParamCaller, name).Fatal(err)
			return
		}

		err = fn(ctx, *req)
		if err != nil {
			ctx.WithField(ParamCaller, name).Fatal(err)
		}
	}
}

func NewHandlerFuncContextTypeAny[T any](fn func(Context, T) any) HandlerFunc {
	name := getCallerName(fn)
	return func(ctx Context) {
		req := new(T)
		err := ctx.Bind(req)
		if err != nil {
			ctx.WithField(ParamCaller, name).Fatal(err)
			return
		}

		data := fn(ctx, *req)
		if ctx.Response().Size() == 0 {
			err := ctx.Render(data)
			if err != nil {
				ctx.WithField(ParamCaller, name).Fatal(err)
			}
		}
	}
}

func NewHandlerFuncContextTypeAnyError[T any](fn func(Context, T) (any, error)) HandlerFunc {
	name := getCallerName(fn)
	return func(ctx Context) {
		req := new(T)
		err := ctx.Bind(req)
		if err != nil {
			ctx.WithField(ParamCaller, name).Fatal(err)
			return
		}

		data, err := fn(ctx, *req)
		if err == nil && ctx.Response().Size() == 0 {
			err = ctx.Render(data)
		}
		if err != nil {
			ctx.WithField(ParamCaller, name).Fatal(err)
		}
	}
}

// The NewHandlerAnyContextTypeAnyError function can match all extension objects.
//
// When the function form is func(Context, Request) (Response, error),
// it returns [HandlerFunc], otherwise it returns nil.
//
// The types of Request can be map, struct or struct pointer.
// Use [reflect] to create Request and call function.
//
// This extension function is not recommended.
func NewHandlerAnyContextTypeAnyError(fn any) HandlerFunc {
	v := reflect.Indirect(reflect.ValueOf(fn))
	iType := v.Type()
	if iType.Kind() != reflect.Func {
		return nil
	}
	if iType.NumIn() != 2 || iType.In(0) != typeContext {
		return nil
	}
	if iType.NumOut() != 2 || iType.Out(1) != typeError {
		return nil
	}
	typeIn := iType.In(1)
	kindIn := typeIn.Kind()
	typenew := iType.In(1)
	// check request type
	switch typeIn.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Struct:
	default:
		return nil
	}
	if typenew.Kind() == reflect.Ptr {
		typenew = typenew.Elem()
	}

	name := getCallerName(fn)
	return func(ctx Context) {
		// create a request and initialize
		req := reflect.New(typenew)
		err := ctx.Bind(req.Interface())
		if err != nil {
			ctx.Fatal(err)
			return
		}
		if kindIn != reflect.Ptr {
			req = req.Elem()
		}

		// reflect call execution function.
		vals := v.Call([]reflect.Value{reflect.ValueOf(ctx), req})

		// check call err.
		err, ok := vals[1].Interface().(error)
		if ok {
			ctx.WithField(ParamCaller, name).Fatal(err)
			return
		}

		// render the returned data.
		err = ctx.Render(vals[0].Interface())
		if err != nil {
			ctx.Fatal(err)
		}
	}
}

// NewHandlerFuncRenderError function converts func(Context, map[string]any) (any, error),
// Bind request parameters to map and handles data Render and error.
//
// Compared with [NewHandlerAnyContextTypeAnyError], no [reflect] call is used.
//
// This extension function is not recommended.
func NewHandlerFuncContextMapAnyError(fn func(Context, map[string]any) (any, error)) HandlerFunc {
	name := getCallerName(fn)
	return func(ctx Context) {
		req := make(map[string]any)
		err := ctx.Bind(&req)
		if err != nil {
			ctx.Fatal(err)
			return
		}
		resp, err := fn(ctx, req)
		if err == nil && ctx.Response().Size() == 0 {
			err = ctx.Render(resp)
		}
		if err != nil {
			ctx.WithField(ParamCaller, name).Fatal(err)
		}
	}
}

// NewHandlerHTTPFunc1 function converts [http.HandlerFunc] type.
func NewHandlerHTTPFunc1(fn http.HandlerFunc) HandlerFunc {
	return func(ctx Context) {
		fn(ctx.Response(), ctx.Request())
	}
}

// NewHandlerHTTPFunc2 function converts func([http.ResponseWriter], *[http.Request]) type.
func NewHandlerHTTPFunc2(fn func(http.ResponseWriter, *http.Request)) HandlerFunc {
	return func(ctx Context) {
		fn(ctx.Response(), ctx.Request())
	}
}

// NewHandlerNetHTTP function converts [http.Handler] type.
func NewHandlerHTTPHandler(h http.Handler) HandlerFunc {
	return func(ctx Context) {
		h.ServeHTTP(ctx.Response(), ctx.Request())
	}
}
