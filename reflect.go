package eudore

/*
获取和设置一个对象的属性
map和结构体相互转换
*/

import (
	"fmt"
	"time"
	"errors"
	"reflect"
	"strings"
	"strconv"
	"encoding/json"
)

// Set the properties of an object.
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
// 设置一个对象的属性。
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
func Set(i interface{}, key string, val interface{}) (interface{}, error) {
	if i == nil {
		return i, fmt.Errorf("input value is nil.")
	}
	newValue := reflect.New(reflect.TypeOf(i)).Elem()
	newValue.Set(reflect.ValueOf(i))
	err := setValue(newValue.Type(), newValue, strings.Split(key, "."), val)
	return newValue.Interface(), err
	// return setValue(reflect.TypeOf(i), reflect.ValueOf(i), strings.Split(key, "."), val)
}

func setValue(iType reflect.Type, iValue reflect.Value, key []string, val interface{}) error {
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
	return fmt.Errorf("not setValue type is %s, key: %v,val: %s", iType.Kind(), key ,val)
}

func setPtr(iType reflect.Type, iValue reflect.Value, key []string, val interface{}) (err error) {
	// 将空指针赋值
	if iValue.IsNil() {
		iValue.Set(reflect.New(iType.Elem()))
	}
	// 对指针解除引用，设置值
	return setValue(iType.Elem(), iValue.Elem(), key ,val)
}

func setInterface(iType reflect.Type, iValue reflect.Value, key []string, val interface{}) (err error) {
	if iValue.Elem().Kind() == reflect.Invalid {
		iValue.Set(reflect.ValueOf(make(map[string]interface{})))
	}
	if len(key) == 0 {
		return setWithInterface(iType.Kind(), iValue, val)
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

func setMap(iType reflect.Type, iValue reflect.Value, key []string, val interface{}) (err error) {
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
		err = setWithInterface(mapValueType.Kind(), mapvalue, val)
	}else {
		err = setValue(mapValueType, mapvalue, key[1:], val)
	}
	// 将修改后的mapvalue重新赋值给map
	if err == nil {
		iValue.SetMapIndex(mapKey, mapvalue)
	}
	return 
}

func setSlice(iType reflect.Type, iValue reflect.Value, key []string, val interface{}) (err error) {
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
			iValue.Set(reflect.AppendSlice(reflect.MakeSlice(iType, 0, index + 1), iValue))
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
		err = setWithInterface(arrayType.Kind(), newValue, val)	
	}else {
		err = setValue(arrayType, newValue, key[1:], val)
	}
	
	if err == nil {
		if index == -1 {
			iValue.Set(reflect.Append(iValue, newValue))
		}else{
			iValue.Index(index).Set(newValue)
		}
		
	}
	return
}

func getArrayIndex(key string) int {
	i, err := strconv.Atoi(key)
	if err != nil {
		return -1
	}
	return i
}

func setStruct(iType reflect.Type, iValue reflect.Value, key []string, val interface{}) error {
	var index int = getStructField(iType, key[0])
	if index == -1 {
		return errors.New("struct not field " + key[0])
	}
	typeField := iType.Field(index)
	structField := iValue.Field(index)
	if !structField.CanSet() {
		return fmt.Errorf("struct %s field %s not set, use pointer type. ", iValue.Type().String(), key[0])
	}

	if len(key) == 1 {
		return setWithInterface(typeField.Type.Kind(), structField, val)
	}
	return setValue(typeField.Type, structField, key[1:], val)
}

func getStructField(iType reflect.Type, name string) int {
	for i := 0; i < iType.NumField(); i++ {
		typeField := iType.Field(i)
		if typeField.Name == name || typeField.Tag.Get("set") == name {
			return i
		}
	}
	return -1
}

// A more path to get an attribute from an object.
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
	return getValue(reflect.TypeOf(i), reflect.ValueOf(i), strings.Split(key, "."))
}

func getValue(iType reflect.Type, iValue reflect.Value, key []string) interface{} {
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

func getPtr(iType reflect.Type, iValue reflect.Value, key []string) interface{} {
	if iValue.IsNil() {
		return nil
	}
	return getValue(iType.Elem(), iValue.Elem(), key)
}

func getStruct(iType reflect.Type, iValue reflect.Value, key []string) interface{} {
	// 查找key对应的属性索引，不存在返回-1。
	var index int = getStructField(iType, key[0])
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

func getSlice(iType reflect.Type, iValue reflect.Value, key []string) interface{} {
	// 检测数组是否为空
	if iValue.IsNil() {
		return nil
	}
	// 检测索引是否存在
	index := getArrayIndex(key[0])
	if index < 0 || iValue.Len() <= index  {
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


func ConvertStruct(i interface{}, m interface{}) error {
	return getConvertStruct(reflect.TypeOf(i), reflect.ValueOf(i), reflect.TypeOf(m), reflect.ValueOf(m))
}

func getConvertStruct(iType reflect.Type, iValue reflect.Value, mType reflect.Type, mValue reflect.Value) error {
	fmt.Println(iType.Kind())
	switch iType.Kind() {
	case reflect.Ptr:
		if iValue.IsNil() {
			iValue.Set(reflect.New(iType.Elem()))
		}
		return getConvertStruct(iValue.Elem().Type(), iValue.Elem(), mType, mValue)
	case reflect.Struct:
		return convertMapToStruct(iType, iValue, mType, mValue)
	}
	return nil
}

func convertMapToStruct(iType reflect.Type, iValue reflect.Value, mType reflect.Type, mValue reflect.Value) error {
	for _, key := range mValue.MapKeys() {
		index := getStructField(iType, getTypeName(key))
		fmt.Println(index)
		if index != -1 {
			setWithValue(iType.Field(index).Type.Kind(), iValue.Field(index), mValue.MapIndex(key))
		}
	}
	return nil
}

func getTypeName(iValue reflect.Value) string {
	switch iValue.Kind() {
	case reflect.String:
		return iValue.String()
	case reflect.Interface, reflect.Ptr:
		return getTypeName(iValue.Elem())
	default:
		return fmt.Sprint(iValue.Interface())
	}
}

func setWithValue(iTypeKind reflect.Kind, iValue reflect.Value, val reflect.Value) {
	fmt.Println(iTypeKind, val.Kind())
	if val.Type().AssignableTo(iValue.Type()) {
		iValue.Set(val)
		return
	}
	if val.Kind() == reflect.Interface {
		setWithValue(iTypeKind, iValue, val.Elem())
	}
}



func ConvertMapString(i interface{}) map[string]interface{} {
	val, ok := getConvertMapString(reflect.TypeOf(i), reflect.ValueOf(i)).(map[string]interface{})
	if ok {
		return val
	}
	return nil
}

func getConvertMapString(iType reflect.Type, iValue reflect.Value) interface{} {
	fmt.Println(iType.Kind(), iType)
	switch iType.Kind() {
	case reflect.Interface:
		if iValue.Elem().Kind() == reflect.Invalid {
			return iValue.Interface()
		}
		return getConvertMapString(iValue.Elem().Type(), iValue.Elem())
	case reflect.Ptr: 
		if iValue.IsNil() {
			return iValue.Interface()
		}
		return getConvertMapString(iValue.Elem().Type(), iValue.Elem())
	case reflect.Map:
		val := make(map[string]interface{})
		convertMapToMapString(iType, iValue, val)
		return val
	case reflect.Struct:
		val := make(map[string]interface{})
		convertStructToMapString(iType, iValue, val)
		return val
	}
	return iValue.Interface()
}

func convertStructToMapString(iType reflect.Type, iValue reflect.Value, val map[string]interface{}) {
	for i := 0; i < iType.NumField(); i++ {
		fieldKey := iType.Field(i)
		fieldValue := iValue.Field(i)		
		val[fieldKey.Name] = getConvertMapString(fieldValue.Type(), fieldValue)
	}
}

func convertMapToMapString(iType reflect.Type, iValue reflect.Value, val map[string]interface{}) {
	for _, key := range iValue.MapKeys() {
		v := iValue.MapIndex(key)
		val[fmt.Sprint(key.Interface())] = getConvertMapString(v.Type(), v)
	}
}


func ConvertMap(i interface{}) map[interface{}]interface{} {
	val, ok := getConvertMap(reflect.TypeOf(i), reflect.ValueOf(i)).(map[interface{}]interface{})
	if ok {
		return val
	}
	return nil
}

func getConvertMap(iType reflect.Type, iValue reflect.Value) interface{} {
	switch iType.Kind() {
	case reflect.Interface:
		if iValue.Elem().Kind() == reflect.Invalid {
			return iValue.Interface()
		}
		return getConvertMap(iValue.Elem().Type(), iValue.Elem())
	case reflect.Ptr: 
		if iValue.IsNil() {
			return iValue.Interface()
		}
		return getConvertMap(iValue.Elem().Type(), iValue.Elem())
	case reflect.Map:
		val := make(map[interface{}]interface{})
		convertMapToMap(iType, iValue, val)
		return val
	case reflect.Struct:
		val := make(map[interface{}]interface{})
		convertStructToMap(iType, iValue, val)
		return val
	}
	return iValue.Interface()
}

func convertStructToMap(iType reflect.Type, iValue reflect.Value, val map[interface{}]interface{}) {
	for i := 0; i < iType.NumField(); i++ {
		fieldKey := iType.Field(i)
		fieldValue := iValue.Field(i)		
		val[fieldKey.Name] = getConvertMap(fieldValue.Type(), fieldValue)
	}
}

func convertMapToMap(iType reflect.Type, iValue reflect.Value, val map[interface{}]interface{}) {
	for _, key := range iValue.MapKeys() {
		v := iValue.MapIndex(key)
		val[key.Interface()] = getConvertMap(v.Type(), v)
	}
}





func setWithInterface(iTypeKind reflect.Kind, iValue reflect.Value, val interface{}) error {
	if reflect.TypeOf(val).AssignableTo(iValue.Type()) {
		iValue.Set(reflect.ValueOf(val))
		return nil
	}
	if reflect.TypeOf(val).Kind() == reflect.String {
		return setWithString(iTypeKind, iValue, val.(string))
	}
	if iTypeKind == reflect.String {
		iValue.SetString(fmt.Sprint(val))
	}	
	return nil
}

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
	case reflect.Slice, reflect.Array, reflect.Struct, reflect.Map:
		 json.Unmarshal([]byte(val), iValue.Addr().Interface())
	case reflect.Interface:
		iValue.Set(reflect.ValueOf(val))
	case reflect.Ptr:
		if !iValue.Elem().IsValid() {
			iValue.Set(reflect.New(iValue.Type().Elem()))
		}
		return setWithInterface(iValue.Elem().Kind(), iValue.Elem(), val)
	default:
		return errors.New("Unknown type " + iTypeKind.String())
	}
	return nil
}

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