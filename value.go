package eudore

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type value struct {
	Tags     []string
	Keys     []string
	Index    int
	All      bool
	Set      bool
	Value    any
	Pointers []uintptr
	Pindex   int
}

// GetAnyByPath method A more path to get an attribute from an object.
//
// The path will be split using '.' and then look for the path in turn.
//
// Structure attributes can use the structure tag 'alias' to match attributes.
//
// Returns a null value if the match fails.
//
// 根据路径来从一个对象获得一个属性。
//
// 路径将使用'.'分割，然后依次寻找路径。
//
// 结构体属性可以使用结构体标签'alias'来匹配属性。
//
// 如果匹配失败直接返回空值。
func GetAnyByPath(i any, key string) any {
	val, err := getValue(i, key, nil, false)
	if err != nil {
		return nil
	}
	return val.Interface()
}

// GetAnyByPathWithTag 函数和GetAnyByPath函数相同，可以额外设置tags，同时会返回error。
func GetAnyByPathWithTag(i any, key string, tags []string, all bool) (any, error) {
	val, err := getValue(i, key, tags, all)
	if err != nil {
		return nil, err
	}
	if all {
		val = reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem()
	}
	return val.Interface(), nil
}

// GetAnyByPathWithValue 函数和Get函数相同，可以允许查找私有属性并返回reflect.Value。
func GetAnyByPathWithValue(i any, key string, tags []string, all bool) (reflect.Value, error) {
	return getValue(i, key, tags, all)
}

func getValue(i any, key string, tags []string, all bool) (reflect.Value, error) {
	val, ok := i.(reflect.Value)
	if !ok {
		val = reflect.ValueOf(i)
	}
	if i == nil {
		return val, ErrValueInputDataNil
	}
	if key == "" {
		return val, nil
	}
	if tags == nil {
		tags = DefaultValueGetSetTags
	}
	v := &value{
		Tags: tags,
		Keys: strings.Split(key, "."),
		All:  all,
	}
	v.Pointers = make([]uintptr, 0, len(v.Keys))
	return v.getValue(val)
}

// 从目标类型获取字符串路径的属性。
func (v *value) getValue(iValue reflect.Value) (reflect.Value, error) {
	if len(v.Keys) == v.Index {
		return iValue, nil
	}
	if v.HasPointer(iValue) {
		return iValue, v.newError(ErrFormatValueAnonymousField, iValue)
	}
	switch iValue.Kind() {
	case reflect.Ptr, reflect.Interface:
		if iValue.IsNil() {
			return iValue, v.newError(ErrFormatValueTypeNil, iValue)
		}
		return v.getValue(iValue.Elem())
	case reflect.Struct:
		return v.getStruct(iValue)
	case reflect.Map:
		return v.getMap(iValue)
	case reflect.Array, reflect.Slice:
		return v.getSlice(iValue)
	}
	return iValue, v.newError(ErrFormatValueNotField, iValue, v.Keys[v.Index])
}

// 处理结构体对象的读取。
func (v *value) getStruct(iValue reflect.Value) (reflect.Value, error) {
	field := getStructFieldOfTags(iValue, v.Keys[v.Index], v.Tags)
	if field.Kind() == reflect.Invalid {
		iType := iValue.Type()
		for i := 0; i < iType.NumField(); i++ {
			if iType.Field(i).Anonymous {
				v2, err := v.getValue(iValue.Field(i))
				if err == nil {
					return v2, nil
				}
			}
		}

		return iValue, v.newError(ErrFormatValueNotField, iValue, v.Keys[v.Index])
	}

	if field.CanInterface() || v.All {
		v.Index++
		defer func() { v.Index-- }()
		return v.getValue(field)
	}
	return iValue, v.newError(ErrFormatValueStructUnexported, iValue, v.Keys[v.Index])
}

// 处理map读取属性。
func (v *value) getMap(iValue reflect.Value) (reflect.Value, error) {
	// 检测map是否为空
	if iValue.IsNil() {
		return iValue, v.newError(ErrFormatValueTypeNil, iValue)
	}
	// 创建map需要的key
	mapKey := reflect.New(iValue.Type().Key()).Elem()
	err := setValueString(mapKey, v.Keys[v.Index])
	if err != nil {
		return iValue, v.newError(ErrFormatValueMapIndexInvalid, iValue, v.Keys[v.Index])
	}

	// 获得map的value, 如果值无效则返回空。
	mapvalue := iValue.MapIndex(mapKey)
	if mapvalue.Kind() == reflect.Invalid {
		return iValue, v.newError(ErrFormatValueMapValueInvalid, iValue, v.Keys[v.Index])
	}
	v.Index++
	defer func() { v.Index-- }()
	return v.getValue(mapvalue)
}

// 处理数组切片读取属性。
func (v *value) getSlice(iValue reflect.Value) (reflect.Value, error) {
	// 检测切片是否为空
	if iValue.Kind() == reflect.Slice && iValue.IsNil() {
		return iValue, v.newError(ErrFormatValueTypeNil, iValue)
	}
	// 检测索引是否存在
	index, err := strconv.Atoi(v.Keys[v.Index])
	if err != nil || iValue.Len() <= index || iValue.Len() < -index {
		return iValue, v.newError(ErrFormatValueArrayIndexInvalid, iValue, v.Keys[v.Index], iValue.Len())
	} else if index < 0 {
		index += iValue.Len()
	}
	v.Index++
	defer func() { v.Index-- }()
	return v.getValue(iValue.Index(index))
}

// The SetAnyByPath function sets the properties of an object, and the object must be a pointer type.
//
// The path will be separated using '.', and then the path will be searched for in sequence.
//
// When the object type selected in the path is ptr, it will be checked to see if it is empty.
// If the object is empty, it will be initialized by default.
//
// When the object type selected in the path is any,
// if the object is empty, it will be initialized to map[string]any,
// otherwise the next operation will be determined based on the value type.
//
// When the object type selected in the path is array,
// the path will be converted into an object index to set the array elements,
// and if the index is [], the elements will be appended.
//
// When the object type selected in the path is struct,
// the attribute name and attribute label 'alias' will be used to match when selecting attributes.
//
// If the value type is a string, it will be converted according to the set target type.
//
// If the target type is a string, the value will be output as a string and then assigned.
//
// SetAnyByPath 函数设置一个对象的属性，改对象必须是指针类型。
//
// 路径将使用'.'分割，然后依次寻找路径。
//
// 当路径中选择对象类型为ptr时，会检查是否为空，对象为空会默认进行初始化。
//
// 当路径中选择对象类型为any时，如果对象为空会初始化为map[string]any，
// 否则按值类型来判断下一步操作。
//
// 当路径中选择对象类型为array时，路径会转换成对象索引来设置数组元素，索引为[]则追加元素。
//
// 当路径中选择对象类型为struct时，选择属性时会使用属性名称和属性标签'alias'来匹配。
//
// 如果值的类型是字符串，会根据设置的目标类型来转换。
//
// 如果目标类型是字符串，将会值输出成字符串然后赋值。
func SetAnyByPath(i any, key string, val any) error {
	return SetAnyByPathWithTag(i, key, val, nil, false)
}

// SetAnyByPathWithTag 函数和SetAnyByPath函数相同，可以额外设置tags。
func SetAnyByPathWithTag(i any, key string, val any, tags []string, all bool) error {
	if i == nil || key == "" {
		return ErrValueInputDataNil
	}
	iValue, ok := i.(reflect.Value)
	if !ok {
		iValue = reflect.ValueOf(i)
	}
	// 检测目标是指针类型。
	if iValue.Kind() != reflect.Ptr {
		return ErrValueInputDataNotPtr
	}
	if tags == nil {
		tags = DefaultValueGetSetTags
	}
	v := &value{
		Tags:  tags,
		Keys:  strings.Split(key, "."),
		All:   all,
		Set:   true,
		Value: val,
	}
	v.Pointers = make([]uintptr, 0, len(v.Keys))
	return v.setValue(iValue)
}

func (v *value) setValue(iValue reflect.Value) error {
	if len(v.Keys) == v.Index {
		err := setValuePtr(reflect.ValueOf(v.Value), iValue)
		if err != nil {
			v.Index--
			err = v.newError("%s", iValue, err)
			v.Index++
		}
		return err
	}
	if v.HasPointer(iValue) {
		return v.newError(ErrFormatValueAnonymousField, iValue)
	}
	switch iValue.Kind() {
	case reflect.Ptr:
		if iValue.IsNil() {
			return v.setMake(iValue, reflect.New(iValue.Type().Elem()))
		}
		return v.setValue(iValue.Elem())
	case reflect.Interface:
		return v.setInterface(iValue)
	case reflect.Struct:
		return v.setStruct(iValue)
	case reflect.Map:
		return v.setMap(iValue)
	case reflect.Slice:
		return v.setSlice(iValue)
	case reflect.Array:
		return v.setArray(iValue)
	}

	return v.newError(ErrFormatValueNotField, iValue, v.Keys[v.Index])
}

func (v *value) setMake(iValue, newValue reflect.Value) error {
	err := v.setValue(newValue)
	if err == nil {
		iValue.Set(newValue)
	}
	return err
}

// 处理接口类型。
func (v *value) setInterface(iValue reflect.Value) (err error) {
	// 如果是空接口，初始化为map[string]any类型
	if iValue.IsNil() {
		if iValue.Type() != typeAny {
			return v.newError(ErrFormatValueTypeNil, iValue)
		}
		return v.setMake(iValue, reflect.ValueOf(make(map[string]any)))
	}
	// 创建一个可取地址的临时变量，并设置值用于下一步设置。
	newValue := reflect.New(iValue.Elem().Type()).Elem()
	newValue.Set(iValue.Elem())
	err = v.setValue(newValue)
	// 将修改后的值重新赋值给对象
	if err == nil {
		iValue.Set(newValue)
	}
	return err
}

// 处理结构体设置属性。
func (v *value) setStruct(iValue reflect.Value) error {
	field := getStructFieldOfTags(iValue, v.Keys[v.Index], v.Tags)
	if field.Kind() == reflect.Invalid {
		iType := iValue.Type()
		for i := 0; i < iType.NumField(); i++ {
			if iType.Field(i).Anonymous {
				err := v.setValue(iValue.Field(i))
				if err == nil {
					return nil
				}
			}
		}

		return v.newError(ErrFormatValueNotField, iValue, v.Keys[v.Index])
	}

	if !field.CanSet() {
		if v.All {
			field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
		} else {
			return v.newError(ErrFormatValueStructNotCanset, iValue, v.Keys[v.Index])
		}
	}
	v.Index++
	defer func() { v.Index-- }()
	return v.setValue(field)
}

// 处理map。
func (v *value) setMap(iValue reflect.Value) error {
	iType := iValue.Type()
	// 对空map初始化
	if iValue.IsNil() {
		return v.setMake(iValue, reflect.MakeMap(iType))
	}

	// 创建map需要匹配的key
	mapKey := reflect.New(iType.Key()).Elem()
	err := setValueString(mapKey, v.Keys[v.Index])
	if err != nil {
		return v.newError(ErrFormatValueMapIndexInvalid, iValue, v.Keys[v.Index])
	}

	newValue := reflect.New(iType.Elem()).Elem()
	mapvalue := iValue.MapIndex(mapKey)
	if mapvalue.Kind() != reflect.Invalid {
		newValue.Set(mapvalue)
	}

	v.Index++
	defer func() { v.Index-- }()
	err = v.setValue(newValue)
	// 将修改后的mapvalue重新赋值给map
	if err == nil {
		iValue.SetMapIndex(mapKey, newValue)
	}
	return err
}

func (v *value) setArray(iValue reflect.Value) error {
	index, err := strconv.Atoi(v.Keys[v.Index])
	if err != nil || iValue.Len() <= index || iValue.Len() < -index {
		return v.newError(ErrFormatValueArrayIndexInvalid, iValue, v.Keys[v.Index], iValue.Len())
	} else if index < 0 {
		index += iValue.Len()
	}
	v.Index++
	defer func() { v.Index-- }()
	return v.setValue(iValue.Index(index))
}

// 处理数组和切片。
func (v *value) setSlice(iValue reflect.Value) error {
	iType := iValue.Type()
	// 处理空切片
	if iValue.IsNil() {
		iValue.Set(reflect.MakeSlice(iType, 0, 4))
		err := v.setSlice(iValue)
		if err != nil {
			iValue.Set(reflect.Zero(iType))
		}
		return err
	}

	// 解析index
	index, err := strconv.Atoi(v.Keys[v.Index])
	switch {
	case (err != nil && v.Keys[v.Index] != "[]") || iValue.Len() < -index:
		return v.newError(ErrFormatValueArrayIndexInvalid, iValue, v.Keys[v.Index], iValue.Len())
	case index < 0:
		index += iValue.Len()
	case v.Keys[v.Index] == "[]":
		index = -1
	}

	// 创建新元素的类型和值
	newValue := reflect.New(iType.Elem()).Elem()
	if index > -1 {
		// 新建数组替换原数组扩容
		if iValue.Cap() <= index {
			iValue.Set(reflect.AppendSlice(reflect.MakeSlice(iType, 0, index+1), iValue))
		}
		// 对数组长度扩充，新元素添加空值
		if iValue.Len() <= index {
			iValue.SetLen(index + 1)
		}
		// 将原数组值设置给newValue
		newValue.Set(iValue.Index(index))
	}

	v.Index++
	defer func() { v.Index-- }()
	err = v.setValue(newValue)
	if err == nil {
		if index > -1 {
			iValue.Index(index).Set(newValue)
		} else {
			iValue.Set(reflect.Append(iValue, newValue))
		}
	}
	return err
}

func (v *value) HasPointer(iValue reflect.Value) bool {
	kind := iValue.Kind()
	if kind < reflect.Map || kind > reflect.Slice {
		return false
	}

	ptr := iValue.Pointer()
	if v.Pointers != nil && v.Index != v.Pindex {
		v.Pindex = v.Index
		v.Pointers = nil
	}

	for _, p := range v.Pointers {
		if p == ptr {
			return true
		}
	}
	v.Pointers = append(v.Pointers, ptr)
	return false
}

func (v *value) newError(f string, iValue reflect.Value, args ...any) error {
	m := "get"
	if v.Set {
		m = "set"
	}

	err := fmt.Errorf(fmt.Sprintf("%s type %s ", iValue.Kind(), iValue.Type())+f, args...)
	return fmt.Errorf(ErrFormatValueError, m, strings.Join(v.Keys[:v.Index+1], "."), err)
}

// 通过字符串获取结构体属性的索引。
func getStructFieldOfTags(iValue reflect.Value, name string, tags []string) reflect.Value {
	iType := iValue.Type()
	for i := 0; i < iType.NumField(); i++ {
		typeField := iType.Field(i)
		// 字符串为结构体名称或结构体属性标签的值，则匹配返回索引。
		if typeField.Name == name {
			return iValue.Field(i)
		}
		for _, tag := range tags {
			if typeField.Tag.Get(tag) == name {
				return iValue.Field(i)
			}
		}
	}
	return reflect.Value{}
}

// getIndirectAllValue 函数获得解除引用的全部类型和值。
func getIndirectAllValue(iValue reflect.Value) (types []reflect.Type, values []reflect.Value) {
	for {
		types = append(types, iValue.Type())
		values = append(values, iValue)
		switch iValue.Kind() {
		case reflect.Ptr, reflect.Interface:
			if iValue.IsNil() {
				return
			}
			iValue = iValue.Elem()
		default:
			return
		}
	}
}

func setValuePtr(sValue reflect.Value, tValue reflect.Value) error {
	if sValue.Kind() == reflect.Ptr || sValue.Kind() == reflect.Interface ||
		tValue.Kind() == reflect.Ptr || tValue.Kind() == reflect.Interface {
		stypes, svalues := getIndirectAllValue(sValue)
		ttypes, tvalues := getIndirectAllValue(tValue)
		for i, ttype := range ttypes {
			for j, stype := range stypes {
				// 转换接口类型、相同类型、type别名类型
				if stype.ConvertibleTo(ttype) && tvalues[i].CanSet() {
					return setValueData(svalues[j], tvalues[i])
				}
			}
		}
		sValue = svalues[len(svalues)-1]
		tValue = tvalues[len(tvalues)-1]

		// 目标类型如果是空指针，则尝试进行初始化并转换
		if tValue.Kind() == reflect.Ptr && tValue.IsNil() {
			newValue := reflect.New(tValue.Type().Elem())
			err := setValuePtr(sValue, newValue)
			if err == nil {
				tValue.Set(newValue)
			}
			return err
		}
	}
	return setValueData(sValue, tValue)
}

func setValueData(sValue reflect.Value, tValue reflect.Value) error {
	sType := sValue.Type()
	tType := tValue.Type()
	switch {
	case sType == tType:
		tValue.Set(sValue)
		return nil
	case sType.ConvertibleTo(tType):
		tValue.Set(sValue.Convert(tType))
		return nil
	case tValue.Kind() == reflect.Slice:
		newValue := reflect.New(tValue.Type().Elem()).Elem()
		err := setValueData(sValue, newValue)
		if err == nil {
			tValue.Set(reflect.Append(tValue, newValue))
		}
		return err
	case sType.Kind() == reflect.String:
		return setValueString(tValue, strings.TrimSpace(sValue.String()))
	case tType.Kind() == reflect.String:
		tValue.SetString(fmt.Sprintf("%+v", sValue.Interface()))
		return nil
	}
	return fmt.Errorf(ErrFormatValueSetWithValue, sValue.Type().String(), tValue.Type().String())
}

var bitSizes = [...]int{0, 0, 0, 8, 16, 32, 64, 0, 8, 16, 32, 64, 32, 64, 32, 64}

// 使用字符串设置对象的值。
//
//nolint:cyclop,gocyclo
func setValueString(v reflect.Value, s string) error {
	var err error
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err = setIntField(v, s)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = setUintField(v, s)
	case reflect.Bool:
		err = setBoolField(v, s)
	case reflect.Float32, reflect.Float64:
		err = setFloatField(v, s)
	case reflect.Complex64, reflect.Complex128:
		err = setComplexField(v, s)
	case reflect.String:
		v.SetString(s)
		return nil
	case reflect.Ptr:
		if v.IsNil() {
			newValue := reflect.New(v.Type().Elem())
			err := setValueString(newValue, s)
			if err == nil {
				v.Set(newValue)
			}
			return err
		}
		return setValueString(v.Elem(), s)
	case reflect.Interface:
		if v.IsNil() && v.Type() == typeAny {
			v.Set(reflect.ValueOf(s))
			return nil
		}
		return setValueString(v.Elem(), s)
	case reflect.Struct:
		if v.Type().ConvertibleTo(typeTimeTime) {
			return setTimeField(v, s)
		}
		return fmt.Errorf(ErrFormatValueSetStringUnknownType, v.Kind().String())
	default:
		return fmt.Errorf(ErrFormatValueSetStringUnknownType, v.Kind().String())
	}

	if err != nil {
		p := reflect.New(v.Type())
		e, ok := p.Interface().(encoding.TextUnmarshaler)
		if ok {
			err = e.UnmarshalText([]byte(s))
		}
		if err == nil {
			v.Set(p.Elem())
		}
	}
	return err
}

func setIntField(field reflect.Value, str string) error {
	if str == "" {
		str = "0"
	}
	intVal, err := strconv.ParseInt(str, 10, bitSizes[int(field.Kind())])
	if err == nil {
		field.SetInt(intVal)
	} else if field.Type() == typeTimeDuration {
		var t time.Duration
		if t, err = time.ParseDuration(str); err == nil {
			field.SetInt(int64(t))
		}
	}
	return err
}

func setUintField(field reflect.Value, str string) error {
	if str == "" {
		str = "0"
	}
	uintVal, err := strconv.ParseUint(str, 10, bitSizes[int(field.Kind())])
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(field reflect.Value, str string) error {
	if str == "" {
		field.SetBool(true)
		return nil
	}
	boolVal, err := strconv.ParseBool(str)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setComplexField(field reflect.Value, str string) error {
	str = strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(str, "("), "i"), ")")
	pos := strings.Index(str, "+")
	if pos == -1 {
		pos = len(str)
		str += "+0"
	}

	read, err := strconv.ParseFloat(str[:pos], bitSizes[int(field.Kind())])
	if err != nil {
		return err
	}
	image, err := strconv.ParseFloat(str[pos+1:], bitSizes[int(field.Kind())])
	if err != nil {
		return err
	}

	field.SetComplex(complex(read, image))
	return nil
}

func setFloatField(field reflect.Value, str string) error {
	if str == "" {
		str = "0.0"
	}
	floatVal, err := strconv.ParseFloat(str, bitSizes[int(field.Kind())])
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

// TimeParse 方法通过解析内置支持的时间格式。
func setTimeField(field reflect.Value, str string) (err error) {
	var t time.Time
	for i, f := range DefaultValueParseTimeFormats {
		if DefaultValueParseTimeFixed[i] && len(str) != len(f) {
			continue
		}
		t, err = time.Parse(f, str)
		if err == nil {
			if field.Type() != typeTimeTime {
				field.Set(reflect.ValueOf(t).Convert(field.Type()))
			} else {
				field.Set(reflect.ValueOf(t))
			}
			return
		}
	}
	return
}
