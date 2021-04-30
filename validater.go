package eudore

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Validater 接口定义验证器。
type Validater interface {
	RegisterValidations(string, ...interface{})
	Validate(interface{}) error
	ValidateVar(interface{}, string) error
}

// validateInterface 接口定义对象自己验证的方法。
type validateInterface interface {
	Validate() error
}
type validaterBase struct {
	sync.Map
	// 结构体类型 - 验证信息
	ValidateTypes map[reflect.Type]validateBaseFields
	// 验证规则 - 验证类型 - 验证函数
	ValidateFuncs map[string]map[reflect.Type]reflect.Value
	// 验证规则 - 验证生成函数
	ValidateNewFuncs map[string][]interface{}
}
type validateBaseFields []validateBaseField
type validateBaseField struct {
	Index   int
	Value   reflect.Value
	IsImple bool
	Format  string
}

func init() {
	DefaultValidater.RegisterValidations("nozero", validateNozeroInt, validateNozeroString, validateNozeroInterface)
	DefaultValidater.RegisterValidations("isnum", validateIsnumString)
	DefaultValidater.RegisterValidations("min", validateNewMinInt, validateNewMinString)
	DefaultValidater.RegisterValidations("max", validateNewMaxInt, validateNewMaxString)
	DefaultValidater.RegisterValidations("len", validateNewLenString)
	DefaultValidater.RegisterValidations("regexp", validateNewRegexpString)
}

// GetValidateStringFunc 函数获得一个ValidateStringFunc对象。
func GetValidateStringFunc(name string) func(string) bool {
	v, ok := DefaultRouterValidater.(interface {
		GetValidateStringFunc(string) func(string) bool
	})
	if ok {
		return v.GetValidateStringFunc(name)
	}
	return nil
}

// NewValidaterBase 函数创建一个默认的Validater。
func NewValidaterBase() Validater {
	return &validaterBase{
		ValidateTypes:    make(map[reflect.Type]validateBaseFields),
		ValidateFuncs:    make(map[string]map[reflect.Type]reflect.Value),
		ValidateNewFuncs: make(map[string][]interface{}),
	}
}

// RegisterValidations 函数给一个名称注册多个类型的的ValidateFunc或ValidateNewFunc。
//
// ValidateFunc 一个任意参数，返回值类型为为bool。
//
// ValidateNewFunc 一个字符串参数，返回值为interface{}或ValidateFunc，若返回值为interface{}，实际调用返回值类型为nil或ValidateFunc。
func (v *validaterBase) RegisterValidations(name string, fns ...interface{}) {
	for _, fn := range fns {
		v.registerValidateFunc(name, fn)
	}
}

// registerValidateFunc 函数注册一个ValidateFunc或ValidateNewFunc
func (v *validaterBase) registerValidateFunc(name string, fn interface{}) {
	iType := reflect.TypeOf(fn)
	if iType.Kind() != reflect.Func || iType.NumIn() != 1 || iType.NumOut() != 1 {
		return
	}

	// ValidateFunc
	if iType.Out(0) == typeBool {
		if v.ValidateFuncs[name] == nil {
			v.ValidateFuncs[name] = make(map[reflect.Type]reflect.Value)
		}
		v.ValidateFuncs[name][iType.In(0)] = reflect.ValueOf(fn)
	}

	// ValidateNewFunc
	if iType.In(0) == typeString && (iType.Out(0) == typeInterface || checkValidateFunc(iType.Out(0))) {
		v.ValidateNewFuncs[name] = append(v.ValidateNewFuncs[name], fn)
	}
}

func checkValidateFunc(iType reflect.Type) bool {
	if iType.Kind() != reflect.Func {
		return false
	}
	if iType.NumIn() != 1 || iType.NumOut() != 1 {
		return false
	}
	if iType.Out(0) != typeBool {
		return false
	}
	return true
}

func (v *validaterBase) Validate(i interface{}) error {
	// 检测接口
	vf, ok := i.(validateInterface)
	if ok {
		return vf.Validate()
	}

	iValue := reflect.Indirect(reflect.ValueOf(i))
	if iValue.Kind() != reflect.Struct {
		return nil
	}
	vfs, err := v.ParseValidateFields(iValue.Type())
	if err != nil {
		return err
	}

	// 匹配验证器规则
	for _, i := range vfs {
		field := iValue.Field(i.Index)
		// 调用Validater接口
		if i.IsImple {
			if field.IsNil() {
				return fmt.Errorf(i.Format, "field is nil")
			}
			err := field.Interface().(validateInterface).Validate()
			if err != nil {
				return fmt.Errorf(i.Format, err)
			}
			continue
		}
		// 反射调用Validater检测函数
		out := i.Value.Call([]reflect.Value{field})
		if !out[0].Bool() {
			return fmt.Errorf(i.Format, field.Interface())
		}
	}
	return nil
}

func (v *validaterBase) ParseValidateFields(iType reflect.Type) (validateBaseFields, error) {
	data, ok := v.Load(iType)
	if ok {
		return data.(validateBaseFields), nil
	}

	var vfs validateBaseFields
	for i := 0; i < iType.NumField(); i++ {
		field := iType.Field(i)
		tags := field.Tag.Get("validate")
		if field.Type.Implements(typeValidateInterface) && tags == "" {
			vfs = append(vfs, validateBaseField{
				Index:   i,
				IsImple: true,
				Format:  fmt.Sprintf("validate %s.%s field '%s' type is '%s', check Validate method error: %%v", iType.PkgPath(), iType.Name(), field.Name, field.Type),
			})
			continue
		}

		for _, tag := range strings.Split(tags, ",") {
			if tag == "" {
				continue
			}
			fn := v.GetValidateFunc(tag, field.Type)
			if !fn.IsValid() {
				return nil, fmt.Errorf("validate %s.%s field %s not create rule %s", iType.PkgPath(), iType.Name(), field.Name, tag)
			}
			vfs = append(vfs, validateBaseField{
				Index:  i,
				Value:  fn,
				Format: fmt.Sprintf("validate %s.%s field %s check %%#v rule %s fatal", iType.PkgPath(), iType.Name(), field.Name, tag),
			})
		}
	}

	v.Store(iType, vfs)
	return vfs, nil
}

func (v *validaterBase) ValidateVar(i interface{}, tag string) error {
	iType := reflect.TypeOf(i)
	fn := v.GetValidateFunc(tag, iType)
	if !fn.IsValid() {
		return fmt.Errorf("validate variable %s %#v not create rule %s", iType.Kind(), i, tag)
	}
	out := fn.Call([]reflect.Value{reflect.ValueOf(i)})
	if !out[0].Bool() {
		return fmt.Errorf("validate variable %s %#v check rule %s fatal", iType.Kind(), i, tag)
	}
	return nil
}

func (v *validaterBase) GetValidateStringFunc(name string) func(string) bool {
	rfn := v.GetValidateFunc(name, typeString)
	if rfn.IsValid() {
		switch fn := rfn.Interface().(type) {
		case func(string) bool:
			return fn
		case func(interface{}) bool:
			return func(str string) bool {
				return fn(str)
			}
		}
	}
	return nil
}

func (v *validaterBase) GetValidateFunc(name string, iType reflect.Type) reflect.Value {
	fns, ok := v.ValidateFuncs[name]
	if ok {
		fn, ok := fns[iType]
		if ok {
			return fn
		}
		fn, ok = fns[typeInterface]
		if ok {
			return fn
		}
	}
	pos := strings.IndexByte(name, ':')
	if pos == -1 {
		return reflect.ValueOf(nil)
	}
	fn := v.GetValidateNewFunc(name[:pos], name[pos+1:], iType)
	if fn.IsValid() {
		return fn
	}
	return v.GetValidateNewFunc(name[:pos], name[pos+1:], typeInterface)
}

func (v *validaterBase) GetValidateNewFunc(name, args string, iType reflect.Type) reflect.Value {
	fns, ok := v.ValidateNewFuncs[name]
	if !ok {
		return reflect.ValueOf(nil)
	}
	vargs := []reflect.Value{reflect.ValueOf(args)}
	for _, fn := range fns {
		newfn := reflect.Indirect(reflect.ValueOf(fn).Call(vargs)[0])
		if newfn.Type() == typeInterface {
			newfn = newfn.Elem()
		}
		if !newfn.IsValid() {
			continue
		}
		if checkValidateFunc(newfn.Type()) && newfn.Type().In(0) == iType {
			name = name + ":" + args
			if v.ValidateFuncs[name] == nil {
				v.ValidateFuncs[name] = make(map[reflect.Type]reflect.Value)
			}
			v.ValidateFuncs[name][iType] = newfn
			return newfn
		}
	}
	return reflect.ValueOf(nil)
}

// validateNozeroString 函数验证一个字符串是否为空
func validateNozeroString(str string) bool {
	return str != ""
}

// validateNozeroInt 函数验证一个int是否为零
func validateNozeroInt(num int) bool {
	return num != 0
}

// validateNozeroInterface 函数验证一个对象是否为零值，使用reflect.Value.IsZero函数实现。
func validateNozeroInterface(i interface{}) bool {
	return !checkValueIsZero(reflect.ValueOf(i))
}

// validateIsnumString 函数验证一个字符串是否为数字。
func validateIsnumString(str string) bool {
	_, err := strconv.Atoi(str)
	return err == nil
}

// validateNewMinInt 函数生成一个验证int最小值的验证函数。
func validateNewMinInt(str string) interface{} {
	min, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return nil
	}
	intmin := int(min)
	return func(num int) bool {
		return num >= intmin
	}
}

// validateNewMinString 函数生成一个验证string最小值的验证函数。
func validateNewMinString(str string) interface{} {
	min, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return nil
	}
	intmin := int(min)
	return func(arg string) bool {
		num, err := strconv.Atoi(arg)
		if err != nil {
			return false
		}
		return num >= intmin
	}
}

// validateNewMaxInt 函数生成一个验证int最大值的验证函数。
func validateNewMaxInt(str string) interface{} {
	max, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return nil
	}
	intmax := int(max)
	return func(num int) bool {
		return num <= intmax
	}
}

// validateNewMaxString 函数生成一个验证string最大值的验证函数。
func validateNewMaxString(str string) interface{} {
	max, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return nil
	}
	intmax := int(max)
	return func(arg string) bool {
		num, err := strconv.Atoi(arg)
		if err != nil {
			return false
		}
		return num <= intmax
	}
}

// validateNewLenString 函数生一个验证字符串长度'>','<','='指定长度的验证函数。
func validateNewLenString(str string) interface{} {
	var flag string
	for _, i := range []string{">", "<", "=", ""} {
		if strings.HasPrefix(str, i) {
			flag = i
			str = str[len(i):]
			break
		}
	}

	length, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return nil
	}
	intlength := int(length)
	switch flag {
	case ">":
		return func(s string) bool {
			return len(s) > intlength
		}
	case "<":
		return func(s string) bool {
			return len(s) < intlength
		}
	default:
		return func(s string) bool {
			return len(s) == intlength
		}
	}
}

// validateNewRegexpString 函数生成一个正则检测字符串的验证函数。
func validateNewRegexpString(str string) interface{} {
	re, err := regexp.Compile(str)
	if err != nil {
		return nil
	}
	// 返回正则匹配校验函数
	return func(arg string) bool {
		return re.MatchString(arg)
	}
}
