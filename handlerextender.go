package eudore

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
)

// HandlerExtender defines the extension management that converts any func into
// [HandlerFunc].
//
// It has three implementations: [NewHandlerExtenderBase]
// [NewHandlerExtenderWrap] [NewHandlerExtenderTree].
//
// Generally, [NewHandlerExtender] and [DefaultHandlerExtender] can be used.
type HandlerExtender interface {
	// The RegisterExtender method registers the converts func of HanderFunc.
	//
	// The converts function type must be func(Type) [HandlerFunc] or
	// func(string, Type) HandlerFunc,
	// Type is a func or interface.
	//
	// If you register an interface type,
	// [CreateHandlers] will determine the implementation interface.
	RegisterExtender(path string, fn any) error

	// The CreateHandlers method converts any func to [HandlerFuncs]
	CreateHandlers(path string, data any) []HandlerFunc

	// List method displays all registered extensions
	List() []string
}

type MetadataHandlerExtender struct {
	Health   bool     `json:"health" protobuf:"1,name=health" yaml:"health"`
	Name     string   `json:"name" protobuf:"2,name=name" yaml:"name"`
	Extender []string `json:"extender" protobuf:"3,name=extender" yaml:"extender"`
}

// The NewHandlerExtender function creates [NewHandlerExtenderBase]
// and loads the extended functions in [DefaultHandlerExtenderFuncs].
func NewHandlerExtender() HandlerExtender {
	he := NewHandlerExtenderBase()
	for _, fn := range DefaultHandlerExtenderFuncs {
		_ = he.RegisterExtender("", fn)
	}
	return he
}

// The NewHandlerExtenderWithContext function gets [ContextKeyHandlerExtender]
// from [context.Context], otherwise returns [DefaultHandlerExtender].
func NewHandlerExtenderWithContext(ctx context.Context) HandlerExtender {
	he, ok := ctx.Value(ContextKeyHandlerExtender).(HandlerExtender)
	if ok {
		return he
	}
	return DefaultHandlerExtender
}

// handlerExtenderBase defines returns a basic func extension processing object.
type handlerExtenderBase struct {
	NewType    []reflect.Type
	NewFunc    []reflect.Value
	AnyType    []reflect.Type
	AnyFunc    []reflect.Value
	allowKinds map[reflect.Kind]struct{}
}

// The NewHandlerExtenderBase method creates a basic [HandlerExtender].
//
// Implement registration and creation of [HandlerFunc].
func NewHandlerExtenderBase() HandlerExtender {
	return &handlerExtenderBase{
		allowKinds: mapClone(DefaultHandlerExtenderAllowKind),
	}
}

// RegisterExtender function implements registration converts function.
func (he *handlerExtenderBase) RegisterExtender(_ string, fn any) error {
	iType := reflect.TypeOf(fn)
	// fn value must be a func type
	if iType.Kind() != reflect.Func {
		return ErrHandlerExtenderParamNotFunc
	}

	// Check that the fn type must be func(Type) HandlerFunc or
	// func(string, Type) HandlerFunc,
	if iType.NumIn() != 1 &&
		(iType.NumIn() != 2 || iType.In(0).Kind() != reflect.String) {
		return fmt.Errorf(ErrHandlerExtenderInputParam, iType.String())
	}
	if iType.NumOut() != 1 || iType.Out(0) != typeHandlerFunc {
		return fmt.Errorf(ErrHandlerExtenderOutputParam, iType.String())
	}

	// DefaultHandlerExtendAllowType defines the allowed Kinds.
	_, ok := he.allowKinds[iType.In(iType.NumIn()-1).Kind()]
	if !ok {
		return fmt.Errorf(ErrHandlerExtenderInputParam, iType.String())
	}

	he.NewType = append(he.NewType, iType.In(iType.NumIn()-1))
	he.NewFunc = append(he.NewFunc, reflect.ValueOf(fn))
	if iType.In(iType.NumIn()-1).Kind() == reflect.Interface {
		he.AnyType = append(he.AnyType, iType.In(iType.NumIn()-1))
		he.AnyFunc = append(he.AnyFunc, reflect.ValueOf(fn))
	}
	return nil
}

func (he *handlerExtenderBase) CreateHandlers(p string, d any) []HandlerFunc {
	val, ok := d.(reflect.Value)
	if !ok {
		val = reflect.ValueOf(d)
	}
	return NewHandlerFuncsFilter(he.create(p, val))
}

func (he *handlerExtenderBase) create(p string, v reflect.Value) []HandlerFunc {
	// Basic Types
	switch fn := v.Interface().(type) {
	case func(Context):
		return []HandlerFunc{fn}
	case HandlerFunc:
		return []HandlerFunc{fn}
	case []HandlerFunc:
		return fn
	}
	// Try converts to HandlerFuncs
	fn := he.find(p, v)
	if fn != nil {
		return []HandlerFunc{fn}
	}

	// Try converts slice to HandlerFuncs
	switch v.Type().Kind() {
	case reflect.Slice, reflect.Array:
		var fns []HandlerFunc
		for i := 0; i < v.Len(); i++ {
			hs := he.create(p, reflect.Indirect(v.Index(i)))
			if hs != nil {
				fns = append(fns, hs...)
			}
		}
		return fns
	case reflect.Interface, reflect.Ptr:
		return he.create(p, v.Elem())
	default:
		return nil
	}
}

// findHandlerFunc function finds the extension function of the Type or
// implements the interface.
//
//	First check Type has a directly registered type extension function,
//
// and then check Type implements the registered interface type.
func (he *handlerExtenderBase) find(p string, v reflect.Value) HandlerFunc {
	iType := v.Type()
	for i := range he.NewType {
		if he.NewType[i] == iType {
			h := he.exec(p, he.NewFunc[i], v)
			if h != nil {
				return h
			}
		}
	}
	// Determine the interface type
	for i, iface := range he.AnyType {
		if iType.Implements(iface) {
			h := he.exec(p, he.AnyFunc[i], v)
			if h != nil {
				return h
			}
		}
	}
	return nil
}

// The newHandlerFunc function uses an extension function to
// convert any into [HandlerFunc].
func (he *handlerExtenderBase) exec(p string, fn, v reflect.Value) HandlerFunc {
	var args []reflect.Value
	if fn.Type().NumIn() == 1 {
		args = []reflect.Value{v}
	} else {
		args = []reflect.Value{reflect.ValueOf(p), v}
	}
	result := fn.Call(args)[0]
	if result.IsNil() {
		return nil
	}

	handlerFuncLocker.Lock()
	handlerFuncCreator[result.Pointer()] = fn.Type().In(0)
	handlerFuncLocker.Unlock()
	return result.Interface().(HandlerFunc)
}

const extenderNameFormat = "%s(%s)"

// The List method returns all registered [HandlerFunc] names.
func (he *handlerExtenderBase) List() []string {
	names := make([]string, 0, len(he.NewFunc))
	for i := range he.NewType {
		if he.NewType[i].Kind() != reflect.Interface {
			name := runtime.FuncForPC(he.NewFunc[i].Pointer()).Name()
			names = append(names, fmt.Sprintf(extenderNameFormat,
				name, he.NewType[i].String(),
			))
		}
	}
	for i, iface := range he.AnyType {
		name := runtime.FuncForPC(he.AnyFunc[i].Pointer()).Name()
		names = append(names, fmt.Sprintf(extenderNameFormat,
			name, iface.String(),
		))
	}
	return names
}

func (he *handlerExtenderBase) Metadata() any {
	return MetadataHandlerExtender{
		Health:   true,
		Name:     "eudore.handlerExtenderBase",
		Extender: he.List(),
	}
}

// handlerExtenderWrap defines chained HandlerExtender object.
type handlerExtenderWrap struct {
	data HandlerExtender
	last HandlerExtender
}

// NewHandlerExtenderWrap function creates a chained [HandlerExtender] object.
//
// All objects are registered and created using baseExtender.
// If baseExtender cannot create a [HandlerFuncs],
// use lastExtender to create a [HandlerFuncs].
func NewHandlerExtenderWrap(base, last HandlerExtender) HandlerExtender {
	return &handlerExtenderWrap{
		data: base,
		last: last,
	}
}

func (he *handlerExtenderWrap) RegisterExtender(path string, fn any) error {
	return he.data.RegisterExtender(path, fn)
}

func (he *handlerExtenderWrap) CreateHandlers(path string, data any,
) []HandlerFunc {
	hs := he.data.CreateHandlers(path, data)
	if hs != nil {
		return hs
	}
	return he.last.CreateHandlers(path, data)
}

func (he *handlerExtenderWrap) List() []string {
	return append(he.last.List(), he.data.List()...)
}

func (he *handlerExtenderWrap) Metadata() any {
	return MetadataHandlerExtender{
		Health:   true,
		Name:     "eudore.handlerExtenderWrap",
		Extender: he.List(),
	}
}

// handlerExtenderTree defines [HandlerExtender] based on path matching.
type handlerExtenderTree struct {
	root handlerExtenderNode
}
type handlerExtenderNode = radixNode[*handlerExtenderData, handlerExtenderData]

type handlerExtenderData struct {
	HandlerExtender
}

// The NewHandlerExtenderTree function creates a [HandlerExtender] based
// on path matching.
//
// Group the extension functions by registering the path
// and create a [HandlerFunc] that selects the extension function with the
// longest path.
func NewHandlerExtenderTree() HandlerExtender {
	return &handlerExtenderTree{}
}

func (data *handlerExtenderData) Insert(vals ...any) error {
	if data.HandlerExtender == nil {
		data.HandlerExtender = NewHandlerExtenderBase()
	}
	return data.HandlerExtender.RegisterExtender(vals[0].(string), vals[1])
}

func (he *handlerExtenderTree) RegisterExtender(path string, fn any) error {
	return he.root.insert(path, path, fn)
}

func (he *handlerExtenderTree) CreateHandlers(path string, data any) []HandlerFunc {
	vals := he.root.lookPath(path)
	for i := len(vals) - 1; i >= 0; i-- {
		h := vals[i].CreateHandlers(path, data)
		if h != nil {
			return h
		}
	}
	return nil
}

func (he *handlerExtenderTree) Metadata() any {
	return MetadataHandlerExtender{
		Health:   true,
		Name:     "eudore.handlerExtenderTree",
		Extender: he.List(),
	}
}

// The List method recursively adds path prefixes
// and returns extension function names.
func (he *handlerExtenderTree) List() []string {
	return handlerExtenderList(&he.root, "")
}

func handlerExtenderList(node *handlerExtenderNode, prefix string) []string {
	prefix += node.path
	var names []string
	if node.data != nil {
		names = node.data.List()
		if prefix != "" {
			for i := range names {
				names[i] = prefix + " " + names[i]
			}
		}
	}

	for i := range node.child {
		names = append(names, handlerExtenderList(node.child[i], prefix)...)
	}
	return names
}
