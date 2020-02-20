package eudore

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type (
	// ValidateStringFunc 函数由RouterFull使用，检测字符串规则。
	ValidateStringFunc func(string) bool
	// Validater 接口定义验证器。
	Validater interface {
		RegisterValidations(string, ...interface{})
		Validate(interface{}) error
		ValidateVar(interface{}, string) error
	}
	// validateFace 接口定义对象自己验证的方法。
	validateFace interface {
		Validate() error
	}
	validaterBase struct {
		ValidateMutex    sync.RWMutex
		ValidateTypes    map[reflect.Type]validateBaseFields
		ValidateFuncs    map[string]map[reflect.Type]interface{}
		ValidateNewFuncs map[string]map[reflect.Type]interface{}
	}
	validateBaseFields []validateBaseField
	validateBaseField  struct {
		Index   int
		Value   reflect.Value
		IsImple bool
		Format  string
	}
)

var (
	typeBool         = reflect.TypeOf((*bool)(nil)).Elem()
	typeString       = reflect.TypeOf((*string)(nil)).Elem()
	typeInterface    = reflect.TypeOf((*interface{})(nil)).Elem()
	typeValidateFace = reflect.TypeOf((*validateFace)(nil)).Elem()
	// DefaultValidater 定义默认的验证器
	DefaultValidater = NewvalidaterBase()
	// DefaultRouterValidater 为RouterFull提供生成ValidateStringFunc功能,需要实现interface{GetValidateStringFunc(string) ValidateStringFunc}接口。
	DefaultRouterValidater = DefaultValidater
)

func init() {
	RegisterValidations("nonzero", validateNozeroInt, validateNozeroString, validateNozeroInterface)
	RegisterValidations("isnum", validateIsnumString)
	RegisterValidations("min", validateNewMinInt, validateNewMinString)
	RegisterValidations("max", validateNewMaxInt, validateNewMaxString)
	RegisterValidations("len", validateNewLenString)
	RegisterValidations("regexp", validateNewRegexpString)
}

// RegisterValidations 函数给DefaultValidater注册验证函数。
func RegisterValidations(name string, fns ...interface{}) {
	DefaultValidater.RegisterValidations(name, fns...)
}

// Validate 函数使用DefaultValidater验证一个对象。
func Validate(i interface{}) error {
	return DefaultValidater.Validate(i)
}

// ValidateVar 函数使用DefaultValidater验证一个变量。
func ValidateVar(i interface{}, tag string) error {
	return DefaultValidater.ValidateVar(i, tag)
}

// GetValidateStringFunc 函数获得一个ValidateStringFunc对象。
func GetValidateStringFunc(name string) ValidateStringFunc {
	v, ok := DefaultRouterValidater.(interface {
		GetValidateStringFunc(string) ValidateStringFunc
	})
	if ok {
		return v.GetValidateStringFunc(name)
	}
	return nil
}

// NewvalidaterBase 函数创建一个默认的Validater。
func NewvalidaterBase() Validater {
	return &validaterBase{
		ValidateMutex:    sync.RWMutex{},
		ValidateTypes:    make(map[reflect.Type]validateBaseFields),
		ValidateFuncs:    make(map[string]map[reflect.Type]interface{}),
		ValidateNewFuncs: make(map[string]map[reflect.Type]interface{}),
	}
}

// RegisterValidations 函数给一个名称注册多个类型的的ValidateFunc或ValidateNewFunc。
//
// ValidateFunc 一个任意参数，返回值类型为为bool或error。
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
			v.ValidateFuncs[name] = make(map[reflect.Type]interface{})
		}
		v.ValidateFuncs[name][iType.In(0)] = fn
	}

	// ValidateNewFunc
	if iType.In(0) == typeString && (checkValidateFunc(iType.Out(0)) || iType.Out(0) == typeInterface) {
		if v.ValidateNewFuncs[name] == nil {
			v.ValidateNewFuncs[name] = make(map[reflect.Type]interface{})
		}
		if iType.Out(0) == typeInterface {
			v.ValidateNewFuncs[name][iType.Out(0)] = fn
		} else {
			v.ValidateNewFuncs[name][iType.Out(0).In(0)] = fn
		}
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
	vf, ok := i.(validateFace)
	if ok {
		return vf.Validate()
	}

	iValue := reflect.ValueOf(i)
	for iValue.Type().Kind() == reflect.Ptr || iValue.Type().Kind() == reflect.Interface {
		iValue = iValue.Elem()
	}
	iType := iValue.Type()
	if iType.Kind() != reflect.Struct {
		return nil
	}
	vfs, err := v.ParseValidateFields(iType)
	if err != nil {
		return err
	}

	// 匹配验证器规则
	for _, i := range vfs {
		field := iValue.Field(i.Index)
		// 调用Validater接口
		if i.IsImple {
			field.Interface().(Validater).Validate(i)
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
	v.ValidateMutex.RLock()
	vfs, ok := v.ValidateTypes[iType]
	v.ValidateMutex.RUnlock()
	if ok {
		return vfs, nil
	}

	v.ValidateMutex.Lock()
	defer v.ValidateMutex.Unlock()
	for i := 0; i < iType.NumField(); i++ {
		field := iType.Field(i)
		tags := field.Tag.Get("validate")
		for _, tag := range strings.Split(tags, ",") {
			fn := v.GetValidateFunc(tag, field.Type)
			if fn == nil {
				return nil, fmt.Errorf("validate %s.%s field %s not create rule %s", iType.PkgPath(), iType.Name(), field.Name, tag)
			}
			vf := validateBaseField{
				Index:   i,
				Value:   reflect.ValueOf(fn),
				IsImple: field.Type.Implements(typeValidateFace),
			}
			vf.Format = fmt.Sprintf("validate %s.%s field %s check %%#v rule %s fatal", iType.PkgPath(), iType.Name(), field.Name, tag)
			vfs = append(vfs, vf)
		}
	}
	v.ValidateTypes[iType] = vfs
	return vfs, nil
}

func (v *validaterBase) ValidateVar(i interface{}, tag string) error {
	iType := reflect.TypeOf(i)
	fn := v.GetValidateFunc(tag, iType)
	if fn == nil {
		return fmt.Errorf("validate variable %s %#v not create rule %s", iType.Kind(), i, tag)
	}
	fValue := reflect.ValueOf(fn)
	out := fValue.Call([]reflect.Value{reflect.ValueOf(i)})
	if fValue.Type().Out(0) == typeBool {
		if !out[0].Bool() {
			return fmt.Errorf("validate variable %s %#v check rule %s fatal", iType.Kind(), i, tag)
		}
	} else {
		if !out[0].IsNil() {
			return fmt.Errorf("validate variable %s %#v check rule %s fatal, return error: %v", iType.Kind(), i, tag, out[0].Interface())
		}
	}
	return nil
}

func (v *validaterBase) GetValidateStringFunc(name string) ValidateStringFunc {
	rfn := v.GetValidateFunc(name, typeString)
	switch fn := rfn.(type) {
	case func(string) bool:
		return fn
	case ValidateStringFunc:
		return fn
	case func(interface{}) bool:
		return func(str string) bool {
			return fn(str)
		}
	default:
		return nil
	}
}

func (v *validaterBase) GetValidateFunc(name string, iType reflect.Type) interface{} {
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
	if pos != -1 {
		fType, fn := v.GetValidateNewFunc(name[:pos], name[pos+1:], iType)
		fmt.Println(name, fn, fType)
		if fn != nil && (fType == iType || fType == typeInterface) {
			if v.ValidateFuncs[name] == nil {
				v.ValidateFuncs[name] = make(map[reflect.Type]interface{})
			}
			v.ValidateFuncs[name][fType] = fn
			return fn
		}
	}
	return nil
}

func (v *validaterBase) GetValidateNewFunc(name, args string, iType reflect.Type) (reflect.Type, interface{}) {
	fns, ok := v.ValidateNewFuncs[name]
	if !ok {
		return nil, nil
	}
	fn, ok := fns[iType]
	if !ok {
		fn, ok = fns[typeInterface]
	}
	if !ok {
		return nil, nil
	}
	rfn := reflect.ValueOf(fn).Call([]reflect.Value{reflect.ValueOf(args)})[0]
	if rfn.IsNil() {
		return nil, nil
	}
	if rfn.Type() == typeInterface {
		rfn = rfn.Elem()
	}
	return rfn.Type().In(0), rfn.Interface()
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
	for _, i := range []string{">", "<", "", "="} {
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
	case "", "=":
		return func(s string) bool {
			return len(s) == intlength
		}
	}
	return nil

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
