package eudore

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type validateField struct {
	sync.Map
	FuncCreator FuncCreator
}

type validateFieldValue struct {
	Index  int
	Omit   bool
	Func   FuncRunner
	Format string
}

// The NewValidateField method creates a struct property validator.
//
// Get ContextKeyFuncCreator.(FuncCreator) to create a verification function.
// Use the structure tag validate to get the verification function from FuncCreator.
//
// Allowed types are struct []struct []*struct []interface.
//
// Only verify that the field type is string/int/uint/float/bool/any,
// and the int-related numerical type is converted to int and then verified,
// and the precision may be lost.
//
// NewValidateField 方法创建结构体属性校验器。
//
// 获取ContextKeyFuncCreator.(FuncCreator)用于创建校验函数。
// 使用结构体tag validate从FuncCreator获取校验函数。
//
// 允许类型为struct []struct []*struct []interface。
//
// 仅校验字段类型为string/int/uint/float/bool/any，int相关数值类型转换成int后校验，可能精度丢失。
func NewValidateField(ctx context.Context) HandlerDataFunc {
	vf := &validateField{FuncCreator: NewFuncCreatorWithContext(ctx)}
	return func(ctx Context, i any) error {
		c := ctx.GetContext()
		v := reflect.Indirect(reflect.ValueOf(i))
		switch v.Kind() {
		case reflect.Struct:
			return vf.validateFields(c, i, v)
		case reflect.Slice, reflect.Array:
			// []struct []*struct []any
			for i := 0; i < v.Len(); i++ {
				field := v.Index(i)
				for field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
					field = field.Elem()
				}
				if field.Kind() == reflect.Struct {
					err := vf.validateFields(c, v.Index(i), field)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
}

func (vf *validateField) validateFields(c context.Context, i any, v reflect.Value) error {
	fields, err := vf.parseStructFields(v.Type())
	if err != nil {
		return err
	}

	if validater, ok := i.(interface{ Validate(context.Context) error }); ok {
		if err := validater.Validate(c); err != nil {
			return err
		}
	}

	// 匹配验证器规则
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

func (vf *validateField) parseStructFields(iType reflect.Type) ([]validateFieldValue, error) {
	data, ok := vf.Load(iType)
	if ok {
		switch val := data.(type) {
		case []validateFieldValue:
			return val, nil
		case error:
			return nil, val
		}
	}

	var fields []validateFieldValue
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
				f, err := vf.parseStructFields(et)
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
				err = fmt.Errorf(ErrFormatValidateParseFieldError,
					iType.PkgPath(), iType.Name(), t.Name, tag, err.Error())
				vf.Store(iType, err)
				return nil, err
			}

			val := validateFieldValue{
				Index: i, Omit: omit, Func: FuncRunner{kind, fn},
				Format: fmt.Sprintf(ErrFormatValidateErrorFormat,
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

// FilterData defines Filter data matching objects and rules.
//
// Package and Name define the filter structure object name,
// you can use '*' fuzzy matching.
//
// Checks and Modifys define data matching and modification behavior functions.
// If Modifys is not defined, the entire data object will be empty.
//
// FilterData 定义Filter数据匹配对象和规则。
//
// Package和Name定义过滤结构体对象名称，可以使用'*'模糊匹配。
//
// Checks和Modifys定义数据匹配和修改行为函数，如果未定义Modifys将整个数据对象置空。
type FilterData struct {
	Package string   `alias:"package" json:"package" xml:"package" yaml:"package"`
	Name    string   `alias:"name" json:"name" xml:"name" yaml:"name"`
	Checks  []string `alias:"checks" json:"checks" xml:"checks" yaml:"checks"`
	Modifys []string `alias:"modifys" json:"modifys" xml:"modifys" yaml:"modifys"`
}

// NewFilterRules function creates FilterData filter function.
//
// Load filter rules from ctx.Value(ContextKeyFilterRules).
//
// NewFilterRules 函数创建FilterData过滤函数。
//
// 从ctx.Value(ContextKeyFilterRules)加载过滤规则。
func NewFilterRules(c context.Context) HandlerDataFunc {
	filter := &filterRule{FuncCreator: NewFuncCreatorWithContext(c)}
	return func(ctx Context, i any) error {
		switch rule := ctx.Value(ContextKeyFilterRules).(type) {
		case []string:
			filter.filte(reflect.ValueOf(i), &FilterData{Checks: rule})
		case *FilterData:
			filter.filte(reflect.ValueOf(i), rule)
		case []FilterData:
			for i := range rule {
				filter.filte(reflect.ValueOf(i), &rule[i])
			}
		}
		return nil
	}
}

func (f *filterRule) filte(v reflect.Value, data *FilterData) {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if !v.IsNil() {
			f.filte(v.Elem(), data)
		}
	case reflect.Struct:
		if v.CanSet() && data.matchName(v.Type()) {
			f.filteData(v, data)
		}
	case reflect.Slice, reflect.Array:
		eType := v.Type().Elem()
		for eType.Kind() == reflect.Ptr {
			eType = eType.Elem()
		}

		if eType.Kind() == reflect.Struct {
			if data.matchName(eType) {
				for i := 0; i < v.Len(); i++ {
					f.filteData(v.Index(i), data)
				}
			}
		} else {
			for i := 0; i < v.Len(); i++ {
				f.filte(v.Index(i), data)
			}
		}
	}
}

func (d *FilterData) matchName(iType reflect.Type) bool {
	return matchStarWithEmpty(d.Package, iType.PkgPath()) && matchStarWithEmpty(d.Name, iType.Name())
}

func (f *filterRule) filteData(v reflect.Value, data *FilterData) {
	if f.filteRules(v, data.Checks, 0) {
		if len(data.Modifys) == 0 {
			v.Set(reflect.Zero(v.Type()))
		} else {
			f.filteRules(v, data.Modifys, FuncCreateNumber)
		}
	}
}

func (f *filterRule) filteRules(v reflect.Value, rules []string, kind FuncCreateKind) bool {
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

// matchStar 模式匹配对象，允许使用带'*'的模式。
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
