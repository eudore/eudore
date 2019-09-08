package eudore

/*
功能1：获取和设置一个对象的属性
func Get(i interface{}, key string) interface{}
func Set(i interface{}, key string, val interface{}) (interface{}, error)

功能2：map和结构体相互转换
func ConvertMap(i interface{}) map[interface{}]interface{}
func ConvertMapString(i interface{}) map[string]interface{}
func ConvertTo(source interface{}, target interface{}) error

*/

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	defaultConvertTag = "set"
)

// Seter 定义对象set属性的方法。
type Seter interface {
	Set(string, interface{}) error
}

// Set the properties of an object. If the target implements the Seter interface, call the Set method.
//
// The path will be split using '.' and then look for the path in turn.
//
// When the object type selected in the path is ptr, it will check if it is empty. If the object is empty, it will be initialized by default.
//
// When the object type selected in the path is interface{}, if the object is empty, it will be initialized to map[string]interface{}, otherwise the value will be judged according to the value type.
//
// When the object type selected in the path is array, the path is converted to an object index to set the array element. If it cannot be converted, the element is appended.
//
// When the object type in the path is selected as a struct, the attribute name and the attribute tag 'set' are used to match when selecting the attribute.
//
// If the type of the value is a string, it will be converted according to the target type set.
//
// If the target type is a string, the value is output as a string and then assigned.
//
// If the target type is an array, map, or struct, the json deserializes the set object.
//
// If the target type passed in is an array, map, or struct, the json deserializes the set object.
//
// 设置一个对象的属性,如果目标实现Seter接口，调用Set方法。
//
// 路径将使用'.'分割，然后依次寻找路径。
//
// 当路径中选择对象类型为ptr时,会检查是否为空，对象为空会默认进行初始化。
//
// 当路径中选择对象类型为interface{}时,如果对象为空会初始化为map[string]interface{},否则按值类型来判断下一步操作。
//
// 当路径中选择对象类型为array时,路径会转换成对象索引来设置数组元素，无法转换则追加元素。
//
// 当路径中选择对象类型为struct时,选择属性时会使用属性名称和属性标签'set'来匹配。
//
// 如果值的类型是字符串，会根据设置的目标类型来转换。
//
// 如果目标类型是字符串，将会值输出成字符串然后赋值。
//
// 如果目标类型是数组、map、结构体，会使用json反序列化设置对象。
//
// 如果传入的目标类型是数组、map、结构体，会使用json反序列化设置对象。
func Set(i interface{}, key string, val interface{}) (interface{}, error) {
	if i == nil {
		return i, ErrConverterInputDataNil
	}
	seter, ok := i.(Seter)
	if ok {
		err := seter.Set(key, val)
		if err == nil || err != ErrSeterNotSupportField {
			return i, err
		}
	}
	// 将对象转换成可取地址的reflect.Value。
	newValue := reflect.New(reflect.TypeOf(i)).Elem()
	newValue.Set(reflect.ValueOf(i))
	// 对字符串路径进行切割
	err := setValue(newValue.Type(), newValue, strings.Split(key, "."), val)
	return newValue.Interface(), err
}

// 递归设置一个目标的路径为key属性的值为val
func setValue(iType reflect.Type, iValue reflect.Value, key []string, val interface{}) error {
	// 不同类型调用不同的方法设置
	switch iType.Kind() {
	case reflect.Ptr:
		return setPtr(iType, iValue, key, val)
	case reflect.Struct:
		return setStruct(iType, iValue, key, val)
	case reflect.Map:
		return setMap(iType, iValue, key, val)
	case reflect.Array, reflect.Slice:
		return setSlice(iType, iValue, key, val)
	case reflect.Interface:
		return setInterface(iType, iValue, key, val)
	}
	return fmt.Errorf(ErrFormatConverterSetTypeError, iType.Kind(), key, val)
}

// 设置指针情况
func setPtr(iType reflect.Type, iValue reflect.Value, key []string, val interface{}) (err error) {
	// 将空指针赋值
	if iValue.IsNil() {
		iValue.Set(reflect.New(iType.Elem()))
	}
	// 对指针解除引用，然后设置值
	return setValue(iType.Elem(), iValue.Elem(), key, val)
}

// 设置接口类型
func setInterface(iType reflect.Type, iValue reflect.Value, key []string, val interface{}) (err error) {
	// 如果路径匹配完直接设置该对象，不确定是否有效，未测试。
	if len(key) == 0 {
		return setWithInterface(iType, iValue, val)
	}
	// 如果是空接口，初始化为map[string]interface{}类型
	if iValue.Elem().Kind() == reflect.Invalid {
		iValue.Set(reflect.ValueOf(make(map[string]interface{})))
	}
	// 创建一个可取地址的临时变量，并设置值用于下一步设置。
	newValue := reflect.New(iValue.Elem().Type()).Elem()
	newValue.Set(iValue.Elem())
	err = setValue(iValue.Elem().Type(), newValue, key, val)
	// 将修改后的值重新赋值给对象
	if err == nil {
		iValue.Set(newValue)
	}
	return err
}

// 处理map
func setMap(iType reflect.Type, iValue reflect.Value, key []string, val interface{}) (err error) {
	// 对空map初始化
	if iValue.IsNil() {
		iValue.Set(reflect.MakeMap(iType))
	}

	// 获取map key/value的类型
	mapKeyType := iType.Key()
	mapValueType := iType.Elem()

	// 创建map需要匹配的key
	mapKey := reflect.New(mapKeyType).Elem()
	setWithString(mapKeyType.Kind(), mapKey, key[0])

	// 获得map的value
	mapvalue := reflect.New(mapValueType).Elem()
	// 如果map存在key对应的值，则设置给mapvalue
	mapvalueTemp := iValue.MapIndex(mapKey)
	if mapvalueTemp.Kind() != reflect.Invalid {
		mapvalue.Set(mapvalueTemp)
	}

	// 设置map value
	if len(key) == 1 {
		err = setWithInterface(mapValueType, mapvalue, val)
	} else {
		err = setValue(mapValueType, mapvalue, key[1:], val)
	}
	// 将修改后的mapvalue重新赋值给map
	if err == nil {
		iValue.SetMapIndex(mapKey, mapvalue)
	}
	return
}

// 处理数组和切片
func setSlice(iType reflect.Type, iValue reflect.Value, key []string, val interface{}) (err error) {
	// 空切片初始化，默认长度2
	if iValue.IsNil() {
		iValue.Set(reflect.MakeSlice(iType, 0, 2))
	}
	// 创建新元素的类型和值
	arrayType := iType.Elem()
	newValue := reflect.New(arrayType).Elem()
	index := getArrayIndex(key[0])
	if index != -1 {
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
	// 设置新元素的值
	if len(key) == 1 {
		err = setWithInterface(arrayType, newValue, val)
	} else {
		err = setValue(arrayType, newValue, key[1:], val)
	}

	// 设置属性成功，将新的值传回给数组。
	if err == nil {
		if index == -1 {
			iValue.Set(reflect.Append(iValue, newValue))
		} else {
			iValue.Index(index).Set(newValue)
		}
	}
	return
}

// 获取字符串索引转换为整数
func getArrayIndex(key string) int {
	i, err := strconv.Atoi(key)
	if err != nil {
		return -1
	}
	return i
}

// 处理结构体设置属性
func setStruct(iType reflect.Type, iValue reflect.Value, key []string, val interface{}) error {
	// 查找属性是结构体的第几个属性。
	var index = getStructFieldOfTag(iType, key[0], defaultConvertTag)
	// 未找到直接返回错误。
	if index == -1 {
		return fmt.Errorf(ErrFormatConverterSetStructNotField, key[0])
	}

	// 获取结构体的属性
	typeField := iType.Field(index)
	structField := iValue.Field(index)

	// 设置属性的值
	if len(key) == 1 {
		if !structField.CanSet() {
			return fmt.Errorf(ErrFormatConverterNotCanset, key[0], iValue.Type().String())
		}
		return setWithInterface(typeField.Type, structField, val)
	}
	return setValue(typeField.Type, structField, key[1:], val)
}

// Get method A more path to get an attribute from an object.
//
// The path will be split using '.' and then look for the path in turn.
//
// Structure attributes can use the structure tag 'set' to match attributes.
//
// Returns a null value if the match fails.
//
// 更具路径来从一个对象获得一个属性。
//
// 路径将使用'.'分割，然后依次寻找路径。
//
// 结构体属性可以使用结构体标签'set'来匹配属性。
//
// 如果匹配失败直接返回空值。
func Get(i interface{}, key string) interface{} {
	if i == nil {
		return nil
	}
	return getValue(reflect.TypeOf(i), reflect.ValueOf(i), strings.Split(key, "."))
}

// 从目标类型获取字符串路径的属性
func getValue(iType reflect.Type, iValue reflect.Value, key []string) interface{} {
	// fmt.Println("getValue:", iType, iValue.Type(), key)
	if len(key) == 0 {
		return iValue.Interface()
	}
	switch iType.Kind() {
	case reflect.Ptr:
		return getPtr(iType, iValue, key)
	case reflect.Struct:
		return getStruct(iType, iValue, key)
	case reflect.Map:
		return getMap(iType, iValue, key)
	case reflect.Array, reflect.Slice:
		return getSlice(iType, iValue, key)
	case reflect.Interface:
		return getInterface(iType, iValue, key)
	}
	return nil
}

// 处理指针对象的读取
func getPtr(iType reflect.Type, iValue reflect.Value, key []string) interface{} {
	// 空指针返回空
	if iValue.IsNil() {
		return nil
	}
	return getValue(iType.Elem(), iValue.Elem(), key)
}

// 处理结构体对象的读取
func getStruct(iType reflect.Type, iValue reflect.Value, key []string) interface{} {
	// 查找key对应的属性索引，不存在返回-1。
	var index = getStructFieldOfTag(iType, key[0], defaultConvertTag)
	if index == -1 {
		return nil
	}
	// 获取key对应结构的属性。
	typeField := iType.Field(index)
	structField := iValue.Field(index)
	if len(key) > 1 {
		return getValue(typeField.Type, structField, key[1:])
	}
	return structField.Interface()
}

// 处理map读取属性
func getMap(iType reflect.Type, iValue reflect.Value, key []string) interface{} {
	// 检测map是否为空
	if iValue.IsNil() {
		return nil
	}
	// 获取map key/value的类型
	mapKeyType := iType.Key()
	mapValueType := iType.Elem()

	// 创建map需要的key
	mapKey := reflect.New(mapKeyType).Elem()
	setWithString(mapKeyType.Kind(), mapKey, key[0])

	// 获得map的value, 如果值无效则返回空。
	mapvalue := iValue.MapIndex(mapKey)
	if mapvalue.Kind() == reflect.Invalid {
		return nil
	}

	// 设置key为匹配完
	if len(key) > 1 {
		return getValue(mapValueType, mapvalue, key[1:])
	}
	return mapvalue.Interface()
}

// 处理数组切片读取属性
func getSlice(iType reflect.Type, iValue reflect.Value, key []string) interface{} {
	// 检测数组是否为空
	if iValue.IsNil() {
		return nil
	}
	// 检测索引是否存在
	index := getArrayIndex(key[0])
	if index < 0 || iValue.Len() <= index {
		return nil
	}
	// 获取索引的值
	newValue := iValue.Index(index)
	// 如果key未匹配完，则继续查找
	if len(key) > 1 {
		return getValue(iType.Elem(), newValue, key[1:])
	}
	return newValue.Interface()
}

// 处理接口读取属性
func getInterface(iType reflect.Type, iValue reflect.Value, key []string) interface{} {
	// 检测接口是否为空
	if iValue.Elem().Kind() == reflect.Invalid {
		return nil
	}
	// 对接口保存的类型解引用
	if len(key) > 0 {
		return getValue(iValue.Elem().Type(), iValue.Elem(), key)
	}
	return iValue.Elem().Interface()
}

// ConvertMapString 将一个结构体转换成map[string]interface{}
//
// 其他map转map[string]interface{}未测试。
func ConvertMapString(i interface{}) map[string]interface{} {
	val, ok := convertMapString(reflect.TypeOf(i), reflect.ValueOf(i)).(map[string]interface{})
	if ok {
		return val
	}
	return nil
}

// 将一个map或结构体对象转换成map[string]interface{}返回。
func convertMapString(iType reflect.Type, iValue reflect.Value) interface{} {
	// fmt.Println(iType.Kind(), iType)
	switch iType.Kind() {
	// 接口类型解除引用
	case reflect.Interface:
		// 空接口直接返回
		if iValue.Elem().Kind() == reflect.Invalid {
			return iValue.Interface()
		}
		return convertMapString(iValue.Elem().Type(), iValue.Elem())
	// 指针类型解除引用
	case reflect.Ptr:
		// 空指针直接返回
		if iValue.IsNil() {
			return iValue.Interface()
		}
		return convertMapString(iValue.Elem().Type(), iValue.Elem())
	// 将map转换成map[string]interface{}
	case reflect.Map:
		val := make(map[string]interface{})
		convertMapstrngMapToMapString(iType, iValue, val)
		return val
	// 将结构体转换成map[string]interface{}
	case reflect.Struct:
		val := make(map[string]interface{})
		convertMapstringStructToMapString(iType, iValue, val)
		return val
	}
	// 其他类型直接返回
	return iValue.Interface()
}

// 结构体转换成map[string]interface{}
func convertMapstringStructToMapString(iType reflect.Type, iValue reflect.Value, val map[string]interface{}) {
	// 遍历结构体的属性
	for i := 0; i < iType.NumField(); i++ {
		fieldKey := iType.Field(i)
		fieldValue := iValue.Field(i)
		// map设置键位结构体的名称，值为结构体值转换，基本类型会直接返回。
		// 未支持接口和标签
		val[fieldKey.Name] = convertMapString(fieldValue.Type(), fieldValue)
	}
}

// 将map转换成map[string]interface{}
func convertMapstrngMapToMapString(iType reflect.Type, iValue reflect.Value, val map[string]interface{}) {
	// 遍历map的全部keys
	for _, key := range iValue.MapKeys() {
		v := iValue.MapIndex(key)
		// 设置新map的键为原map的字符串输出，未支持接口转换
		// 设置新map的值为原map匹配的值的转换，值为基本类型会直接返回。
		val[fmt.Sprint(key.Interface())] = convertMapString(v.Type(), v)
	}
}

// ConvertMap 将一个结构体转换成map[interface{}]interface{}
//
// 其他map转map[interface{}]interface{}未测试。
func ConvertMap(i interface{}) map[interface{}]interface{} {
	val, ok := convertMap(reflect.TypeOf(i), reflect.ValueOf(i)).(map[interface{}]interface{})
	if ok {
		return val
	}
	return nil
}

// 将一个map或结构体对象转换成map[interface{}]interface{}返回。
func convertMap(iType reflect.Type, iValue reflect.Value) interface{} {
	switch iType.Kind() {
	case reflect.Interface:
		if iValue.Elem().Kind() == reflect.Invalid {
			return iValue.Interface()
		}
		return convertMap(iValue.Elem().Type(), iValue.Elem())
	case reflect.Ptr:
		if iValue.IsNil() {
			return iValue.Interface()
		}
		return convertMap(iValue.Elem().Type(), iValue.Elem())
	case reflect.Map:
		val := make(map[interface{}]interface{})
		convertMapMapToMap(iType, iValue, val)
		return val
	case reflect.Struct:
		val := make(map[interface{}]interface{})
		convertMapStructToMap(iType, iValue, val)
		return val
	}
	return iValue.Interface()
}

// 结构体转换成map[interface{}]interface{}
func convertMapStructToMap(iType reflect.Type, iValue reflect.Value, val map[interface{}]interface{}) {
	// 遍历结构体的属性
	for i := 0; i < iType.NumField(); i++ {
		fieldKey := iType.Field(i)
		fieldValue := iValue.Field(i)
		// map设置键位结构体的名称，值为结构体值转换，基本类型会直接返回。
		// 未支持接口和标签
		val[fieldKey.Name] = convertMap(fieldValue.Type(), fieldValue)
	}
}

// 将map转换成map[interface{}]interface{}
func convertMapMapToMap(iType reflect.Type, iValue reflect.Value, val map[interface{}]interface{}) {
	// 遍历map的全部keys
	for _, key := range iValue.MapKeys() {
		v := iValue.MapIndex(key)
		// 设置新map的键为原map的字符串输出，未支持接口转换
		// 设置新map的值为原map匹配的值的转换，值为基本类型会直接返回。
		val[key.Interface()] = convertMap(v.Type(), v)
	}
}

// ConvertTo 将一个对象属性复制给另外一个对象。
func ConvertTo(source interface{}, target interface{}) error {
	if source == nil {
		return ErrConverterInputDataNil
	}
	if target == nil {
		return ErrConverterTargetDataNil
	}
	return convertTo(reflect.TypeOf(source), reflect.ValueOf(source), reflect.TypeOf(target), reflect.ValueOf(target))
}

func convertTo(sType reflect.Type, sValue reflect.Value, tType reflect.Type, tValue reflect.Value) error {
	var kind int

	// 检测目标类型并解除引用和初始化空对象
	switch tType.Kind() {
	case reflect.Ptr:
		if tValue.IsNil() {
			tValue.Set(reflect.New(tType.Elem()))
		}
		return convertTo(sType, sValue, tValue.Elem().Type(), tValue.Elem())
	case reflect.Interface:
		// 如果目标类型无法解引用，直接转换
		if tValue.Elem().Kind() == reflect.Invalid {
			return setWithValue(sType, sValue, tType, tValue)
		}
		return convertTo(sType, sValue, tValue.Elem().Type(), tValue.Elem())
	case reflect.Map:
		if tValue.IsNil() {
			tValue.Set(reflect.MakeMap(tType))
		}
		kind = kind | 0x01
	case reflect.Struct:
		kind = kind | 0x02
	}
	// 检测源类型并解除引用
	switch sType.Kind() {
	case reflect.Ptr:
		if sValue.IsNil() {
			return ErrConverterInputDataNil
		}
		return convertTo(sValue.Elem().Type(), sValue.Elem(), tType, tValue)
	case reflect.Interface:
		if sValue.Elem().Kind() == reflect.Invalid {
			return ErrConverterInputDataNil
		}
		return convertTo(sValue.Elem().Type(), sValue.Elem(), tType, tValue)
	case reflect.Map:
		if sValue.IsNil() {
			return ErrConverterInputDataNil
		}
		kind = kind | 0x10
	case reflect.Struct:
		kind = kind | 0x20
	}

	// fmt.Println(kind, sType.Kind(), tType.Kind())
	// 根据数据类型执行转换
	switch kind {
	case 0x11:
		convertToMapToMap(sType, sValue, tType, tValue)
	case 0x21:
		convertToStructToMap(sType, sValue, tType, tValue)
	case 0x12:
		convertToMapToStruct(sType, sValue, tType, tValue)
	case 0x22:
		convertToStructToStruct(sType, sValue, tType, tValue)
	default:
		setWithValue(sType, sValue, tType, tValue)
	}
	return nil
}

func convertToMapToMap(sType reflect.Type, sValue reflect.Value, tType reflect.Type, tValue reflect.Value) {
	for _, key := range sValue.MapKeys() {
		mapvalue := reflect.New(tType.Elem()).Elem()
		if err := convertTo(sValue.MapIndex(key).Type(), sValue.MapIndex(key), tType.Elem(), mapvalue); err == nil {
			tValue.SetMapIndex(key, mapvalue)
		}
		// tValue.SetMapIndex(key, sValue.MapIndex(key))
	}
}

func convertToMapToStruct(sType reflect.Type, sValue reflect.Value, tType reflect.Type, tValue reflect.Value) {
	for _, key := range sValue.MapKeys() {
		index := getStructFieldOfTag(tType, fmt.Sprint(key.Interface()), defaultConvertTag)
		if index == -1 {
			continue
		}
		convertTo(sValue.MapIndex(key).Type(), sValue.MapIndex(key), tType.Field(index).Type, tValue.Field(index))
		// tValue.Field(index).Set(sValue.MapIndex(key))
	}
}

func convertToStructToStruct(sType reflect.Type, sValue reflect.Value, tType reflect.Type, tValue reflect.Value) {
	for i := 0; i < sType.NumField(); i++ {
		if checkValueIsZero(sValue.Field(i)) {
			continue
		}
		index := getStructFieldOfTag(tType, sType.Field(i).Name, defaultConvertTag)
		if index == -1 {
			continue
		}
		convertTo(sType.Field(i).Type, sValue.Field(i), tType.Field(index).Type, tValue.Field(index))
		// tValue.Field(index).Set(sValue.Field(i))
	}
}

func convertToStructToMap(sType reflect.Type, sValue reflect.Value, tType reflect.Type, tValue reflect.Value) {
	for i := 0; i < sType.NumField(); i++ {
		if checkValueIsZero(sValue.Field(i)) {
			continue
		}
		mapvalue := reflect.New(tType.Elem()).Elem()
		if err := convertTo(sType.Field(i).Type, sValue.Field(i), tType.Elem(), mapvalue); err == nil {
			tValue.SetMapIndex(reflect.ValueOf(sType.Field(i).Name), mapvalue)
		}
		// tValue.SetMapIndex(reflect.ValueOf(sType.Field(i).Name), sValue.Field(i))
	}
}

// 通过字符串获取结构体属性的索引
func getStructFieldOfTag(iType reflect.Type, name, tag string) int {
	// 遍历匹配
	for i := 0; i < iType.NumField(); i++ {
		typeField := iType.Field(i)
		// 字符串为结构体名称或结构体属性的set标签的值，则匹配返回索引。
		if typeField.Name == name || typeField.Tag.Get(tag) == name {
			return i
		}
	}
	return -1
}

// checkValueIsZero 函数检测一个值是否为空。
func checkValueIsZero(iValue reflect.Value) bool {
	switch iValue.Type().Kind() {
	case reflect.Bool:
		return !iValue.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return iValue.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return iValue.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return math.Float64bits(iValue.Float()) == 0
	case reflect.Complex64, reflect.Complex128:
		c := iValue.Complex()
		return math.Float64bits(real(c)) == 0 && math.Float64bits(imag(c)) == 0
	case reflect.Array:
		for i := 0; i < iValue.Len(); i++ {
			if !checkValueIsZero(iValue.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return iValue.IsNil()
	case reflect.String:
		return iValue.Len() == 0
	case reflect.Struct:
		for i := 0; i < iValue.NumField(); i++ {
			if !checkValueIsZero(iValue.Field(i)) {
				return false
			}
		}
		return true
	// 无效作为空值返回，忽略后续处理。
	case reflect.Invalid:
		return true
	default:
		panic(fmt.Errorf(ErrFormatConverterCheckZeroUnknownType, iValue.Type().Kind().String()))
	}
}

func getValueString(iType reflect.Type, iValue reflect.Value) string {
	switch iType.Kind() {
	case reflect.Array, reflect.Slice, reflect.Struct, reflect.Map:
		b, err := json.Marshal(iValue.Interface())
		if err == nil {
			return string(b)
		}
	case reflect.Interface, reflect.Ptr:
		return getValueString(iType.Elem(), iValue.Elem())
	case reflect.String:
		return iValue.String()
	default:
		return fmt.Sprint(iValue.Interface())
	}
	return ""
}

// 将一个interface{}赋值给对象
func setWithInterface(iType reflect.Type, iValue reflect.Value, val interface{}) error {
	return setWithValue(reflect.TypeOf(val), reflect.ValueOf(val), iType, iValue)
}

func setWithValue(sType reflect.Type, sValue reflect.Value, tType reflect.Type, tValue reflect.Value) error {
	if sType.AssignableTo(tType) {
		tValue.Set(sValue)
		return nil
	}
	switch sType.Kind() {
	case reflect.String:
		return setWithString(tType.Kind(), tValue, sValue.String())
	case reflect.Array, reflect.Slice:
		err := setWithValue(sType.Elem(), sValue.Index(0), tType, tValue)
		if err == nil {
			return nil
		}
	case reflect.Ptr:
		return setWithValue(sType.Elem(), sValue.Elem(), tType, tValue)
	}
	return setWithString(tType.Kind(), tValue, getValueString(sType, sValue))
}

// 将一个字符串赋值给对象
func setWithString(iTypeKind reflect.Kind, iValue reflect.Value, val string) error {
	switch iTypeKind {
	case reflect.Int:
		return setIntField(val, 0, iValue)
	case reflect.Int8:
		return setIntField(val, 8, iValue)
	case reflect.Int16:
		return setIntField(val, 16, iValue)
	case reflect.Int32:
		return setIntField(val, 32, iValue)
	case reflect.Int64:
		switch iValue.Interface().(type) {
		case time.Duration:
			return setTimeDuration(val, iValue)
		}
		return setIntField(val, 64, iValue)
	case reflect.Uint:
		return setUintField(val, 0, iValue)
	case reflect.Uint8:
		return setUintField(val, 8, iValue)
	case reflect.Uint16:
		return setUintField(val, 16, iValue)
	case reflect.Uint32:
		return setUintField(val, 32, iValue)
	case reflect.Uint64:
		return setUintField(val, 64, iValue)
	case reflect.Bool:
		return setBoolField(val, iValue)
	case reflect.Float32:
		return setFloatField(val, 32, iValue)
	case reflect.Float64:
		return setFloatField(val, 64, iValue)
	case reflect.String:
		iValue.SetString(val)
	case reflect.Slice, reflect.Array:
		return json.Unmarshal([]byte(val), iValue.Addr().Interface())
	case reflect.Struct:
		switch iValue.Interface().(type) {
		case time.Time:
			t, err := time.Parse(time.RFC3339, val)
			if err != nil {
				return err
			}
			iValue.Set(reflect.ValueOf(t))
			return nil
		}
		return json.Unmarshal([]byte(val), iValue.Addr().Interface())
	case reflect.Map:
		return json.Unmarshal([]byte(val), iValue.Addr().Interface())
	// 目标类型是字符串直接设置
	case reflect.Interface:
		iValue.Set(reflect.ValueOf(val))
	// 目标类型是指针进行解引用然后赋值。
	case reflect.Ptr:
		if !iValue.Elem().IsValid() {
			iValue.Set(reflect.New(iValue.Type().Elem()))
		}
		return setWithString(iValue.Elem().Kind(), iValue.Elem(), val)
	default:
		return fmt.Errorf(ErrFormatConverterSetStringUnknownType, iTypeKind.String())
	}
	return nil
}

// 设置int类型属性
func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

// 设置无符号整形属性
func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

// 设置布尔类型属性
func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

// 设置浮点类型属性
func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

// 设置时间类型，未支持未使用
func setTimeDuration(val string, ivalue reflect.Value) error {
	t, err := time.ParseDuration(val)
	if err != nil {
		return err
	}

	ivalue.Set(reflect.ValueOf(t))
	return nil
}

// 设置时间类型，未支持未使用
func setTimeField(val string, structField reflect.StructField, value reflect.Value) error {
	timeFormat := structField.Tag.Get("time_format")
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	if val == "" {
		value.Set(reflect.ValueOf(time.Time{}))
		return nil
	}

	l := time.Local
	if isUTC, _ := strconv.ParseBool(structField.Tag.Get("time_utc")); isUTC {
		l = time.UTC
	}

	if locTag := structField.Tag.Get("time_location"); locTag != "" {
		loc, err := time.LoadLocation(locTag)
		if err != nil {
			return err
		}
		l = loc
	}

	t, err := time.ParseInLocation(timeFormat, val, l)
	if err != nil {
		return err
	}

	value.Set(reflect.ValueOf(t))
	return nil
}
