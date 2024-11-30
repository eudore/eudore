package eudore

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// NewHandlerDataValidate function creates a data Validateation function.
//
// Data implements the `Validate(context.Context) error` method and
// calls the method for Validateation.
func NewHandlerDataValidate() HandlerDataFunc {
	type Validater interface {
		Validate(ctx context.Context) error
	}
	return func(ctx Context, data any) error {
		ver, ok := data.(Validater)
		if ok {
			return ver.Validate(ctx.Context())
		}

		v := reflect.Indirect(reflect.ValueOf(data))
		switch v.Kind() {
		case reflect.Slice, reflect.Array:
			c := ctx.Context()
			for i := 0; i < v.Len(); i++ {
				ver, ok := v.Index(i).Interface().(Validater)
				if ok {
					err := ver.Validate(c)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
}

// NewHandlerDataValidateStruct method creates structure property validator.
//
// Use [reflect] to get [DefaultHandlerValidateTag] from the struct as the field
// validation rule.
//
// Get [FuncCreator] from [context.Context] to create a validation function.
//
// Allowed types are struct []struct []*struct []interface.
func NewHandlerDataValidateStruct(c context.Context) HandlerDataFunc {
	vf := &validateStruct{FuncCreator: NewFuncCreatorWithContext(c)}
	return func(_ Context, data any) error {
		return vf.validate(reflect.Indirect(reflect.ValueOf(data)))
	}
}

type validateStruct struct {
	sync.Map
	FuncCreator FuncCreator
}

type validateStructValue struct {
	Index  int
	Omit   bool
	Func   FuncRunner
	Format string
}

func (vf *validateStruct) validate(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Struct:
		return vf.validateStructs(v)
	case reflect.Slice, reflect.Array:
		// []struct []*struct []any
		for i := 0; i < v.Len(); i++ {
			field := reflect.Indirect(v.Index(i))
			if field.Kind() == reflect.Struct {
				err := vf.validateStructs(field)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (vf *validateStruct) validateStructs(v reflect.Value) error {
	fields, err := vf.parseFields(v.Type())
	if err != nil {
		return err
	}

	// match rules
	for _, i := range fields {
		field := v.Field(i.Index)
		if i.Omit && field.IsZero() {
			continue
		}
		if !i.Func.RunPtr(field) {
			return fmt.Errorf(i.Format, field.Interface())
		}
	}
	return nil
}

func (vf *validateStruct) parseFields(iType reflect.Type) (
	[]validateStructValue, error,
) {
	data, ok := vf.Load(iType)
	if ok {
		switch val := data.(type) {
		case []validateStructValue:
			return val, nil
		case error:
			return nil, val
		}
	}

	var fields []validateStructValue
	for i := 0; i < iType.NumField(); i++ {
		t := iType.Field(i)
		tags, omit := cutOmit(t.Tag.Get(DefaultHandlerValidateTag))
		if tags == "-" {
			continue
		}

		if t.Anonymous {
			et := t.Type
			if et.Kind() == reflect.Ptr {
				et = et.Elem()
			}
			if et.Kind() == reflect.Struct {
				f, err := vf.parseFields(et)
				if err != nil {
					vf.Store(iType, err)
					return nil, err
				}
				fields = append(fields, f...)
				continue
			}
		}

		for _, tag := range splitValidateTag(tags) {
			kind := NewFuncCreateKindWithType(t.Type)
			fn, err := vf.FuncCreator.CreateFunc(kind, tag)
			if err != nil {
				err = fmt.Errorf(ErrHandlerDataValidateCreateRule,
					iType.PkgPath(), iType.Name(), t.Name, tag, err)
				vf.Store(iType, err)
				return nil, err
			}

			val := validateStructValue{
				Index: i, Omit: omit, Func: FuncRunner{kind, fn},
				Format: fmt.Sprintf(ErrHandlerDataValidateCheckFormat,
					iType.PkgPath(), iType.Name(), t.Name, tag),
			}
			fields = append(fields, val)
		}
	}

	vf.Store(iType, fields)
	return fields, nil
}

func splitValidateTag(s string) []string {
	var strs []string
	var last int
	var block int
	for i := range s {
		switch s[i] {
		case '(':
			block++
		case ')':
			block--
		case ',':
			if block == 0 {
				strs = append(strs, s[last:i])
				last = i + 1
			}
		}
	}
	strs = append(strs, s[last:])
	for i, str := range strs {
		if str != "" && str[0] == '(' && str[len(str)-1] == ')' {
			strs[i] = str[1 : len(str)-1]
		}
	}
	strs = sliceFilter(strs, func(s string) bool { return s != "" })
	return strs
}

type filterRule struct {
	FuncCreator
	Rules sync.Map
}

// FilterRule defines Filter data matching objects and rules.
//
// Package and Name define the filter structure object name,
// you can use '*' fuzzy matching.
//
// Checks and Modifys define data matching and modification behavior functions.
// If Modifys is not defined, the entire data object will be empty.
type FilterRule struct {
	Package string   `alias:"package" json:"package" xml:"package" yaml:"package"`
	Name    string   `alias:"name" json:"name" xml:"name" yaml:"name"`
	Checks  []string `alias:"checks" json:"checks" xml:"checks" yaml:"checks"`
	Modifys []string `alias:"modifys" json:"modifys" xml:"modifys" yaml:"modifys"`
}

// NewHandlerDataFilter function creates [FilterRule] filter function.
//
// Get [ContextKeyFilterRules] from [context.Context] to load filter rules.
func NewHandlerDataFilter(c context.Context) HandlerDataFunc {
	filter := &filterRule{FuncCreator: NewFuncCreatorWithContext(c)}
	return func(ctx Context, data any) error {
		switch rule := ctx.Value(ContextKeyFilterRules).(type) {
		case []string:
			filter.filte(reflect.ValueOf(data), &FilterRule{Checks: rule})
		case *FilterRule:
			filter.filte(reflect.ValueOf(data), rule)
		case []FilterRule:
			for i := range rule {
				filter.filte(reflect.ValueOf(data), &rule[i])
			}
		}
		return nil
	}
}

func (f *filterRule) filte(v reflect.Value, rule *FilterRule) {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if !v.IsNil() {
			f.filte(v.Elem(), rule)
		}
	case reflect.Struct:
		if v.CanSet() && rule.matchName(v.Type()) {
			f.filteData(v, rule)
		}
	case reflect.Slice, reflect.Array:
		eType := v.Type().Elem()
		for eType.Kind() == reflect.Ptr {
			eType = eType.Elem()
		}

		if eType.Kind() == reflect.Struct {
			if rule.matchName(eType) {
				for i := 0; i < v.Len(); i++ {
					f.filteData(v.Index(i), rule)
				}
			}
		} else {
			for i := 0; i < v.Len(); i++ {
				f.filte(v.Index(i), rule)
			}
		}
	}
}

func (d *FilterRule) matchName(iType reflect.Type) bool {
	return matchStarWithEmpty(d.Package, iType.PkgPath()) &&
		matchStarWithEmpty(d.Name, iType.Name())
}

func (f *filterRule) filteData(v reflect.Value, rule *FilterRule) {
	if f.filteRules(v, rule.Checks, 0) {
		if len(rule.Modifys) == 0 {
			v.Set(reflect.Zero(v.Type()))
		} else {
			f.filteRules(v, rule.Modifys, FuncCreateNumber)
		}
	}
}

func (f *filterRule) filteRules(v reflect.Value, rules []string,
	kind FuncCreateKind,
) bool {
	for _, rule := range rules {
		key, val, _ := strings.Cut(rule, "=")
		field, err := getValue(v, key, []string{"filter", "alias"}, false)
		if err != nil {
			continue
		}

		k := NewFuncCreateKindWithType(field.Type())
		if k == FuncCreateInvalid {
			continue
		}
		k += kind

		r, ok := f.Rules.Load(k.String() + rule)
		if !ok {
			fn, err := f.FuncCreator.CreateFunc(k, val)
			if err == nil {
				r = &FuncRunner{Kind: k, Func: fn}
			} else {
				r = &FuncRunner{}
			}
			f.Rules.Store(k.String()+rule, r)
		}

		b := r.(*FuncRunner).RunPtr(field)
		if !b {
			return false
		}
	}
	return true
}

// matchStar pattern matching object, allowing the use of patterns with '*'.
func matchStarWithEmpty(patten, obj string) bool {
	if patten == "" {
		return true
	}
	parts := strings.Split(patten, "*")
	if len(parts) < 2 {
		return patten == obj
	}
	if !strings.HasPrefix(obj, parts[0]) {
		return false
	}
	for _, i := range parts {
		if i == "" {
			continue
		}
		pos := strings.Index(obj, i)
		if pos == -1 {
			return false
		}
		obj = obj[pos+len(i):]
	}
	return true
}
