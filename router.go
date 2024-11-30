package eudore

// The Router object is used to define the router match.

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

const (
	routerLoggerHandler = 1 << (iota)
	routerLoggerController
	routerLoggerMiddleware
	routerLoggerExtend
	routerLoggerError
	routerLoggerMetadata
)

// The Router interface provides wrap registration behavior,
// setting route [Params], Group router, Middleware, [HandlerExtender],
// and [Controller].
type Router interface {
	RouterCore

	// Group method returns a new group [Router].
	//
	// Group routing will completely copy Params and Middlewares,
	// and HandlerExtender will wrap the parent.
	//
	// The [Router] [Logger] kind will be modified when the route parameter
	// 'loggerkind' is present.
	//
	// example: app.Group(" loggerkind=~handler")
	Group(path string) Router

	// The Params method returns the current Router [Params],
	// which can be modified.
	Params() *Params

	// AddHandler method adds a new route,
	// [HandlerExtender] will convert any type Handlers into []HandlerFunc.
	//
	// The registration method allows adding multiple methods separated by ',',
	// but must be defined in [DefaultRouterAllMethod] or the value is
	// ANY TEST 404 405 NotFound MethodNotAllowed.
	//
	// Use the current path to match the middleware [HandlerFuncs],
	// and then add it before the Handler.
	//
	// The method is TEST, which will output the debug information related to
	// the route registration,
	// but will not perform the registration behavior.
	AddHandler(method string, path string, fn ...any) error

	// AddController method registers the [Controller], and the controller
	// determines the routing registration behavior.
	AddController(ctls ...Controller) error

	// AddMiddleware adds multiple middleware [HandlerFuncs] to the [Router],
	// [HandlerExtender] will convert any type Handlers into []HandlerFunc.
	//
	// If the first parameter is a string type, it is used as a Group route.
	AddMiddleware(fn ...any) error

	// AddHandlerExtend method adds an extension function to the
	// current [Router].
	// [HandlerExtender] will convert any type Handlers into []HandlerFunc.
	//
	// Make the Router's built-in [HandlerExtender] call RegisterExtender.
	//
	// If the first parameter is a string type, it is used as a Group route.
	AddHandlerExtend(fn ...any) error

	// Register the Any method route, [HandlerExtender] will convert any type
	// Handlers into []HandlerFunc.
	// alias app.AddHandler(eudore.MethodAny, method, path).
	//
	// Any method route will be overwritten by the specified method route,
	// but not vice versa.
	//
	// The Any method includes Get Post Put Delete Head Patch and is defined in
	// the global variable [DefaultRouterAnyMethod].
	AnyFunc(path string, fn ...any)
	// refer AnyFunc
	GetFunc(path string, fn ...any)
	// refer AnyFunc
	PostFunc(path string, fn ...any)
	PutFunc(path string, fn ...any)
	DeleteFunc(path string, fn ...any)
	HeadFunc(path string, fn ...any)
	PatchFunc(path string, fn ...any)
}

// RouterCore interface implements registration and matching of routes
//
// RouterCore implements route matching details.
type RouterCore interface {
	// Register a route path. If used directly,
	// the wrap behavior of Router will be ignored.
	//
	// It is recommended to use the [Router.AddHandler] method to add routes.
	HandleFunc(method string, path string, fn []HandlerFunc)

	// Match a request path and method. If it does not match,
	// return [StatusNotFound].
	// If the method is not allowed, return [StatusMethodNotAllowed].
	Match(method string, path string, params *Params) []HandlerFunc
}

// routerStd default [Router] registration implementation.
type routerStd struct {
	RouterCore      `alias:"routercore"`
	HandlerExtender `alias:"handlerextender"`
	Middlewares     *middlewareTree `alias:"middlewares"`
	GroupParams     Params          `alias:"params"`
	Logger          Logger          `alias:"logger"`
	LoggerKind      int             `alias:"loggerkind"`
	MethodAll       []string        `alias:"methodall"`
	Meta            *MetadataRouter `alias:"meta"`
}

type MetadataRouter struct {
	Health       bool       `json:"health" protobuf:"1,name=health" yaml:"health"`
	Name         string     `json:"name" protobuf:"2,name=name" yaml:"name"`
	Core         any        `json:"core" protobuf:"3,name=core" yaml:"core"`
	AllMethod    any        `json:"allMethod" protobuf:"4,name=allMethod" yaml:"allMethod"`
	Errors       []string   `json:"errors,omitempty" protobuf:"5,name=errors" yaml:"errors,omitempty"`
	Methods      []string   `json:"methods" protobuf:"6,name=methods" yaml:"methods"`
	Paths        []string   `json:"paths" protobuf:"7,name=paths" yaml:"paths"`
	Params       []Params   `json:"params" protobuf:"8,name=params" yaml:"params"`
	HandlerNames [][]string `json:"handlerNames" protobuf:"9,name=handlerNames" yaml:"handlerNames"`
}

// NewRouter method uses a [RouterCore] to create a [Router] object,
// [NewRouterCoreMux] is used by default.
func NewRouter(core RouterCore) Router {
	if core == nil {
		core = NewRouterCoreMux()
	}
	return &routerStd{
		RouterCore: core,
		HandlerExtender: NewHandlerExtenderWrap(
			NewHandlerExtenderTree(), DefaultHandlerExtender,
		),
		Middlewares: &middlewareTree{},
		GroupParams: Params{ParamRoute, ""},
		Logger:      DefaultLoggerNull,
		LoggerKind:  getRouterLoggerKind(0, DefaultRouterLoggerKind),
		MethodAll:   DefaultRouterAllMethod,
		Meta: &MetadataRouter{
			Name:      "eudore.routerStd",
			AllMethod: DefaultRouterAllMethod,
		},
	}
}

// The Mount method causes routerStd to mount the [context.Context].
//
// Get [ContextKeyApp] or [ContextKeyLogger] from [context.Context] as [Logger];
// Get [ContextKeyHandlerExtender] from [context.Context] as [HandlerExtender].
func (r *routerStd) Mount(ctx context.Context) {
	for _, key := range [...]any{ContextKeyApp, ContextKeyLogger} {
		log, ok := ctx.Value(key).(Logger)
		if ok {
			r.Logger = log
			break
		}
	}

	he, ok := ctx.Value(ContextKeyHandlerExtender).(HandlerExtender)
	if ok {
		r.HandlerExtender = NewHandlerExtenderWrap(NewHandlerExtenderTree(), he)
	}
	anyMount(ctx, r.RouterCore)
}

// The Unmount method causes routerStd to unload the [context.Context].
func (r *routerStd) Unmount(ctx context.Context) {
	anyUnmount(ctx, r.RouterCore)
	r.Logger = DefaultLoggerNull
}

// The Metadata method returns the Metadata of RouterCore.
func (r *routerStd) Metadata() any {
	r.Meta.Health = len(r.Meta.Errors) == 0
	r.Meta.Core = anyMetadata(r.RouterCore)
	if r.Meta.Core == nil {
		r.Meta.Core = fmt.Sprintf("%T", r.RouterCore)
	}
	return *r.Meta
}

func (r *routerStd) Group(path string) Router {
	params := NewParamsRoute(path)
	kind := params.Get(ParamLoggerKind)
	if kind != "" {
		params.Del(ParamLoggerKind)
	}

	// Copy the data and build a new router
	return &routerStd{
		RouterCore: r.RouterCore,
		HandlerExtender: NewHandlerExtenderWrap(
			NewHandlerExtenderTree(), r.HandlerExtender,
		),
		Middlewares: r.Middlewares.clone(),
		Logger:      r.Logger,
		LoggerKind:  getRouterLoggerKind(r.LoggerKind, kind),
		MethodAll:   r.MethodAll,
		GroupParams: combineParams(r.GroupParams.Clone(), params),
		Meta:        r.Meta,
	}
}

func (r *routerStd) Params() *Params {
	return &r.GroupParams
}

// The combineParams method merges the params data for route merging.
func combineParams(p1, p2 Params) Params {
	p1[1] += p2[1]
	for i := 2; i < len(p2); i += 2 {
		p1 = p1.Set(p2[i], p2[i+1])
	}
	return p1
}

func (r *routerStd) AddHandler(method, path string, hs ...any) error {
	return r.addHandler(strings.ToUpper(method), path, hs...)
}

var formatTestInfo = "test handlers params is %s, " +
	"split path to: ['%s'], " +
	"match middlewares is: %v, " +
	"register handlers is: %v."

// The addHandler method converts the handler into [HandlerFuncs],
// adds the request middleware corresponding to the routing path,
// and calls the [RouterCore] object to register the routing method.
func (r *routerStd) addHandler(method, path string, hs ...any) (err error) {
	defer func() {
		// [NewRouterCoreMux] panics when registering unknown validation rules,
		// or panics when registering other custom routes.
		if rerr := recover(); rerr != nil {
			err = fmt.Errorf(ErrRouterAddHandlerRecover, method, path, rerr)
			r.getLoggerError(err, 0).
				WithField(ParamDepth, DefaultLoggerDepthKindStack).
				Error(err)
		}
	}()

	params := combineParams(r.GroupParams.Clone(), NewParamsRoute(path))
	path = params.Get(ParamRoute)
	fullpath := params.String()
	// If the method is 404 or 405, the [ParamRoute] is empty
	if len(fullpath) > 6 && fullpath[:6] == "route=" {
		fullpath = fullpath[6:]
	}

	depth := getRouterDepthWithFunc(2, 8, ".AddController")
	handlers, err := r.newHandlerFuncs(path, hs, depth+1)
	if err != nil {
		return err
	}

	// If the registration method is TEST, then output routerStd debug info.
	if method == "TEST" {
		r.getLogger(routerLoggerHandler, depth).Debugf(
			formatTestInfo,
			params.String(), strings.Join(getSplitPath(path), "', '"),
			r.Middlewares.Lookup(path), handlers,
		)
		return nil
	}
	r.getLogger(routerLoggerHandler, depth).Info("register handler:",
		method, strings.TrimPrefix(params.String(), "route="), handlers,
	)
	if handlers != nil {
		handlers = NewHandlerFuncsCombine(r.Middlewares.Lookup(path), handlers)
	}

	// Handle multiple methods
	var errs mulitError
	for _, method := range strings.Split(method, ",") {
		method = strings.TrimSpace(method)
		if checkMethod(r.MethodAll, method) {
			r.RouterCore.HandleFunc(method, fullpath, handlers)
			if r.getLogger(routerLoggerMetadata, 0) != DefaultLoggerNull {
				addMetadataRouter(r.Meta, method, fullpath, handlers)
			}
		} else {
			err := fmt.Errorf(ErrRouterAddHandlerMethodInvalid,
				method, fullpath,
			)
			errs.Handle(err)
			r.getLoggerError(err, depth).Error(err)
		}
	}
	return errs.Unwrap()
}

func checkMethod(all []string, method string) bool {
	switch method {
	case "ANY", "404", "405", "NOTFOUND", "METHODNOTALLOWED":
		return true
	}
	for _, allMethod := range all {
		if allMethod == method {
			return true
		}
	}
	return false
}

// The newHandlerFuncs method creates []HandlerFunc based on the path and
// multiple parameters.
//
// first calls the current [HandlerExtender.NewHandlerFuncs] to
// create multiple function handlers.
// If it returns null, it will be created from the superior [HandlerExtender].
func (r *routerStd) newHandlerFuncs(path string, handlers []any, depth int,
) ([]HandlerFunc, error) {
	var hs []HandlerFunc
	var errs mulitError
	// Conversion handlers
	for i, fn := range handlers {
		handler := r.HandlerExtender.CreateHandlers(path, fn)
		if len(handler) > 0 {
			hs = NewHandlerFuncsCombine(hs, handler)
		} else if _, ok := handlers[i].(HandlerFunc); !ok {
			err := fmt.Errorf(ErrRouterHandlerFuncsUnregisterType,
				path, i, reflect.TypeOf(fn).String(),
			)
			errs.Handle(err)
			r.getLoggerError(err, depth).Error(err)
		}
	}
	return hs, errs.Unwrap()
}

func (r *routerStd) AddController(controllers ...Controller) error {
	var errs mulitError
	for _, controller := range controllers {
		name := getControllerPathName(controller)
		route := strings.TrimPrefix(r.GroupParams.String(), "route=")
		log := r.getLogger(routerLoggerController, 1)
		if route != "" {
			log.Info("register controller:", route, name)
		} else {
			log.Info("register controller:", name)
		}
		err := controller.Inject(controller, r)
		if err != nil {
			err = fmt.Errorf(ErrRouterAddController, name, err)
			errs.Handle(err)
			r.getLoggerError(err, 1).Error(err)
		}
	}
	return errs.Unwrap()
}

// The getControllerPathName function gets the name of the [Controller].
func getControllerPathName(ctl Controller) string {
	u, ok := ctl.(interface{ Unwrap() Controller })
	if ok {
		ctl = u.Unwrap()
	}
	cType := reflect.Indirect(reflect.ValueOf(ctl)).Type()
	return fmt.Sprintf("%s.%s", cType.PkgPath(), cType.Name())
}

func (r *routerStd) AddMiddleware(hs ...any) error {
	path := r.GroupParams.Get("route")
	if len(hs) > 1 {
		route, ok := hs[0].(string)
		if ok {
			path += route
			hs = hs[1:]
		}
	}
	if len(hs) == 0 {
		return nil
	}

	depth := getRouterDepthWithFunc(1, 4, "(*App).AddMiddleware")
	handlers, err := r.newHandlerFuncs(path, hs, depth+1)
	if err != nil {
		return err
	}

	r.Middlewares.Insert(path, handlers)
	r.RouterCore.HandleFunc("Middlewares", path, handlers)
	log := r.getLogger(routerLoggerMiddleware, depth)
	if path != "" {
		log.Info("register middleware:", path, handlers)
	} else {
		log.Info("register middleware:", handlers)
	}
	return nil
}

func (r *routerStd) AddHandlerExtend(handlers ...any) error {
	if len(handlers) == 0 {
		return nil
	}

	path := r.GroupParams.Get("route")
	if len(handlers) > 1 {
		route, ok := handlers[0].(string)
		if ok {
			path += route
			handlers = handlers[1:]
		}
	}

	var errs mulitError
	for _, handler := range handlers {
		err := r.HandlerExtender.RegisterExtender(path, handler)
		if err != nil {
			err = fmt.Errorf(ErrRouterAddHandlerExtender, path, err)
			errs.Handle(err)
			r.getLoggerError(err, 1).Error(err)
		} else {
			v := reflect.ValueOf(handler)
			if v.Kind() == reflect.Func {
				name := runtime.FuncForPC(v.Pointer()).Name()
				r.getLogger(routerLoggerExtend, 1).Info("register extend:",
					name, v.Type().In(0).String(),
				)
			}
		}
	}
	return errs.Unwrap()
}

func (r *routerStd) AnyFunc(path string, h ...any) {
	_ = r.addHandler(MethodAny, path, h...)
}

func (r *routerStd) GetFunc(path string, h ...any) {
	_ = r.addHandler(MethodGet, path, h...)
}

func (r *routerStd) PostFunc(path string, h ...any) {
	_ = r.addHandler(MethodPost, path, h...)
}

func (r *routerStd) PutFunc(path string, h ...any) {
	_ = r.addHandler(MethodPut, path, h...)
}

func (r *routerStd) DeleteFunc(path string, h ...any) {
	_ = r.addHandler(MethodDelete, path, h...)
}

func (r *routerStd) HeadFunc(path string, h ...any) {
	_ = r.addHandler(MethodHead, path, h...)
}

func (r *routerStd) PatchFunc(path string, h ...any) {
	_ = r.addHandler(MethodPatch, path, h...)
}

func (r *routerStd) getLogger(kind int, depth int) Logger {
	if r.LoggerKind&kind == kind {
		return r.Logger.WithField(ParamDepth, depth)
	}
	return DefaultLoggerNull
}

func (r *routerStd) getLoggerError(err error, depth int) Logger {
	r.Meta.Errors = append(r.Meta.Errors, err.Error())
	return r.getLogger(routerLoggerError, depth)
}
