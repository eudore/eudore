package eudore

/*
功能1：获取和设置一个对象的属性
func Get(i interface{}, key string) interface{}
func Set(i interface{}, key string, val interface{}) (interface{}, error)

功能2：map和结构体相互转换
func ConvertMap(i interface{}) map[interface{}]interface{}
func ConvertMapString(i interface{}) map[string]interface{}
func ConvertStruct(i interface{}, m interface{}) error
func ConvertTo(source interface{}, target interface{}) error

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
// If the target type passed in is an array, map, or struct, the json deserializes the set object.
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
//
// 如果传入的目标类型是数组、map、结构体，会使用json反序列化设置对象。
func Set(i interface{}, key string, val interface{}) (interface{}, error) {
	if i == nil {
		return i, fmt.Errorf("input value is nil.")
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
	// fmt.Println("setValue:", iType, key, val)
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
	return fmt.Errorf("not setValue type is %s, key: %v,val: %s", iType.Kind(), key ,val)
}

// 设置指针情况
func setPtr(iType reflect.Type, iValue reflect.Value, key []string, val interface{}) (err error) {
	// 将空指针赋值
	if iValue.IsNil() {
		iValue.Set(reflect.New(iType.Elem()))
	}
	// 对指针解除引用，然后设置值
	return setValue(iType.Elem(), iValue.Elem(), key ,val)
}

// 设置接口类型
func setInterface(iType reflect.Type, iValue reflect.Value, key []string, val interface{}) (err error) {
	// 如果路径匹配完直接设置该对象，不确定是否有效，未测试。
	if len(key) == 0 {
		return setWithInterface(iType.Kind(), iValue, val)
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
	
	// 设置属性成功，将新的值传回给数组。
	if err == nil {
		if index == -1 {
			iValue.Set(reflect.Append(iValue, newValue))
		}else{
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
	var index int = getStructField(iType, key[0])
	// 未找到直接返回错误。
	if index == -1 {
		return errors.New("struct not field " + key[0])
	}
	
	// 获取结构体的属性
	typeField := iType.Field(index)
	structField := iValue.Field(index)

	// 设置属性的值
	if len(key) == 1 {
		if !structField.CanSet() {
			return fmt.Errorf("struct %s field %s not set, use pointer type. ", iValue.Type().String(), key[0])
		}
		return setWithInterface(typeField.Type.Kind(), structField, val)
	}
	return setValue(typeField.Type, structField, key[1:], val)
}

// 通过字符串获取结构体属性的索引
func getStructField(iType reflect.Type, name string) int {
	// 遍历匹配
	for i := 0; i < iType.NumField(); i++ {
		typeField := iType.Field(i)
		// 字符串为结构体名称或结构体属性的set标签的值，则匹配返回索引。
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


// 将map转换成结构体对象。
//
// 第一参数i为结构体，第二参数m为map。
//
// 如果输入结构体是不可取地址的例如结构体、数组，可能修改后对象改变，会返回新的值，其他情况下返回值等于输出的对象。
func ConvertStruct(i interface{}, m interface{}) (interface{}, error) {
	if i == nil || m == nil {
		return i, fmt.Errorf("input value is nil.")
	}
	// 将结构体转换成可取地址的reflect.Value。
	newValue := reflect.New(reflect.TypeOf(i)).Elem()
	newValue.Set(reflect.ValueOf(i))
	err := getConvertStruct(newValue.Type(), newValue, reflect.TypeOf(m), reflect.ValueOf(m))
	return newValue.Interface(), err
}

// 将map m转换成结构体i。
func getConvertStruct(iType reflect.Type, iValue reflect.Value, mType reflect.Type, mValue reflect.Value) error {
	// 判断map的类型，是指针和接口接触引用，是map进行转换。
	switch mType.Kind() {
	case reflect.Ptr:
		// 检查指针是否有效
		if mValue.IsNil() {
			return nil
		}
		return getConvertStruct(iType, iValue, mValue.Elem().Type(), mValue.Elem())
	case reflect.Interface:
		// 检查接口是否有效
		if mValue.Elem().Kind() == reflect.Invalid {
			return nil
		}
		return getConvertStruct(iType, iValue, mValue.Elem().Type(), mValue.Elem())
	case reflect.Slice, reflect.Array:
		if iType.Kind() == reflect.Slice {
			var err error
			length := mValue.Len()
			if iValue.IsNil() {
				iValue.Set(reflect.MakeSlice(iType, length, length))
			}
			for i := 0; i < length; i++ {
				err = getConvertStruct(iType.Elem(), iValue.Index(i), mType.Elem(), mValue.Index(i))
				if err != nil {
					return err
				}
			}
			return nil
		}
	case reflect.Map:
		// 判断i的类型，是指针和接口接触引用，是结构体进行转换。
		switch iType.Kind() {
		case reflect.Ptr:
			if iValue.IsNil() {
				iValue.Set(reflect.New(iType.Elem()))
			}
			return getConvertStruct(iValue.Elem().Type(), iValue.Elem(), mType, mValue)
		case reflect.Interface:
			if iValue.Elem().Kind() == reflect.Invalid {
				return nil
			}
			return getConvertStruct(iValue.Elem().Type(), iValue.Elem(), mType, mValue)
		case reflect.Struct:
			return convertMapToStruct(iType, iValue, mType, mValue)
		}
	}
	// 无法处理的转换，直接类型转换。
	return setWithInterface(iType.Kind(), iValue, mValue.Interface())
}

// 执行map属性转换成结构体属性
func convertMapToStruct(iType reflect.Type, iValue reflect.Value, mType reflect.Type, mValue reflect.Value) error {
	var tmp reflect.Value
	for _, key := range mValue.MapKeys() {
		// 查找结构体是否该属性
		index := getStructField(iType, getTypeName(key))
		// 找到后就设置值
		if index != -1 {
			tmp = mValue.MapIndex(key)
			getConvertStruct(iType.Field(index).Type, iValue.Field(index), tmp.Type(), tmp)
		}
	}
	return nil
}

// 获取类型的名称
func getTypeName(iValue reflect.Value) string {
	switch iValue.Kind() {
	// 字符串直接返回值
	case reflect.String:
		return iValue.String()
	// 对指针和接口接触引用
	case reflect.Interface, reflect.Ptr:
		return getTypeName(iValue.Elem())
	// 默认返回字符串输出，未实现接口获得名称
	default:
		return fmt.Sprint(iValue.Interface())
	}
}

// 将结构体转换成map
//
// 功能与ConvertStruct相同，但不递归处理多层。
func ConvertStructOnce(i interface{}, m interface{}) (interface{}, error) {
	if i == nil || m == nil {
		return i, fmt.Errorf("input value is nil.")
	}
	// 将结构体转换成可取地址的reflect.Value。
	newValue := reflect.New(reflect.TypeOf(i)).Elem()
	newValue.Set(reflect.ValueOf(i))
	err := getConvertStructOnce(newValue.Type(), newValue, reflect.TypeOf(m), reflect.ValueOf(m))
	return newValue.Interface(), err
}

// 将map m转换成结构体i。
func getConvertStructOnce(iType reflect.Type, iValue reflect.Value, mType reflect.Type, mValue reflect.Value) error {
	// 判断map的类型，是指针和接口接触引用，是map进行转换。
	switch mType.Kind() {
	case reflect.Ptr:
		// 检查指针是否有效
		if mValue.IsNil() {
			return nil
		}
		return getConvertStructOnce(iType, iValue, mValue.Elem().Type(), mValue.Elem())
	case reflect.Interface:
		// 检查接口是否有效
		if mValue.Elem().Kind() == reflect.Invalid {
			return nil
		}
		return getConvertStructOnce(iType, iValue, mValue.Elem().Type(), mValue.Elem())
	case reflect.Map:
		// 判断i的类型，是指针和接口接触引用，是结构体进行转换。
		switch iType.Kind() {
		case reflect.Ptr:
			if iValue.IsNil() {
				iValue.Set(reflect.New(iType.Elem()))
			}
			return getConvertStructOnce(iValue.Elem().Type(), iValue.Elem(), mType, mValue)
		case reflect.Interface:
			if iValue.Elem().Kind() == reflect.Invalid {
				return nil
			}
			return getConvertStructOnce(iValue.Elem().Type(), iValue.Elem(), mType, mValue)
		case reflect.Struct:
			return convertMapToStructOnce(iType, iValue, mType, mValue)
		}
	}
	// 无法处理的转换，直接类型转换。
	return nil
}

// 执行map属性转换成结构体属性
func convertMapToStructOnce(iType reflect.Type, iValue reflect.Value, mType reflect.Type, mValue reflect.Value) error {
	for _, key := range mValue.MapKeys() {
		// 查找结构体是否该属性
		index := getStructField(iType, getTypeName(key))
		// 找到后就设置值
		if index != -1 {
			setWithInterface(iType.Field(index).Type.Kind(), iValue.Field(index),  mValue.MapIndex(key).Interface())
		}
	}
	return nil
}

// 将一个结构体转换成map[string]interface{}
//
// 其他map转map[string]interface{}未测试。 
func ConvertMapString(i interface{}) map[string]interface{} {
	val, ok := getConvertMapString(reflect.TypeOf(i), reflect.ValueOf(i)).(map[string]interface{})
	if ok {
		return val
	}
	return nil
}

// 将一个map或结构体对象转换成map[string]interface{}返回。
func getConvertMapString(iType reflect.Type, iValue reflect.Value) interface{} {
	fmt.Println(iType.Kind(), iType)
	switch iType.Kind() {
	// 接口类型解除引用
	case reflect.Interface:
		// 空接口直接返回
		if iValue.Elem().Kind() == reflect.Invalid {
			return iValue.Interface()
		}
		return getConvertMapString(iValue.Elem().Type(), iValue.Elem())
	// 指针类型解除引用
	case reflect.Ptr: 
		// 空指针直接返回
		if iValue.IsNil() {
			return iValue.Interface()
		}
		return getConvertMapString(iValue.Elem().Type(), iValue.Elem())
	// 将map转换成map[string]interface{}
	case reflect.Map:
		val := make(map[string]interface{})
		convertMapToMapString(iType, iValue, val)
		return val
	// 将结构体转换成map[string]interface{}
	case reflect.Struct:
		val := make(map[string]interface{})
		convertStructToMapString(iType, iValue, val)
		return val
	}
	// 其他类型直接返回
	return iValue.Interface()
}

// 结构体转换成map[string]interface{}
func convertStructToMapString(iType reflect.Type, iValue reflect.Value, val map[string]interface{}) {
	// 遍历结构体的属性
	for i := 0; i < iType.NumField(); i++ {
		fieldKey := iType.Field(i)
		fieldValue := iValue.Field(i)		
		// map设置键位结构体的名称，值为结构体值转换，基本类型会直接返回。
		// 未支持接口和标签
		val[fieldKey.Name] = getConvertMapString(fieldValue.Type(), fieldValue)
	}
}

// 将map转换成map[string]interface{}
func convertMapToMapString(iType reflect.Type, iValue reflect.Value, val map[string]interface{}) {
	// 遍历map的全部keys
	for _, key := range iValue.MapKeys() {
		v := iValue.MapIndex(key)
		// 设置新map的键为原map的字符串输出，未支持接口转换
		// 设置新map的值为原map匹配的值的转换，值为基本类型会直接返回。
		val[fmt.Sprint(key.Interface())] = getConvertMapString(v.Type(), v)
	}
}


// 将一个结构体转换成map[interface{}]interface{}
//
// 其他map转map[interface{}]interface{}未测试。
func ConvertMap(i interface{}) map[interface{}]interface{} {
	val, ok := getConvertMap(reflect.TypeOf(i), reflect.ValueOf(i)).(map[interface{}]interface{})
	if ok {
		return val
	}
	return nil
}

// 将一个map或结构体对象转换成map[interface{}]interface{}返回。
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

// 结构体转换成map[interface{}]interface{}
func convertStructToMap(iType reflect.Type, iValue reflect.Value, val map[interface{}]interface{}) {
	// 遍历结构体的属性
	for i := 0; i < iType.NumField(); i++ {
		fieldKey := iType.Field(i)
		fieldValue := iValue.Field(i)		
		// map设置键位结构体的名称，值为结构体值转换，基本类型会直接返回。
		// 未支持接口和标签
		val[fieldKey.Name] = getConvertMap(fieldValue.Type(), fieldValue)
	}
}

// 将map转换成map[interface{}]interface{}
func convertMapToMap(iType reflect.Type, iValue reflect.Value, val map[interface{}]interface{}) {
	// 遍历map的全部keys	
	for _, key := range iValue.MapKeys() {
		v := iValue.MapIndex(key)
		// 设置新map的键为原map的字符串输出，未支持接口转换
		// 设置新map的值为原map匹配的值的转换，值为基本类型会直接返回。
		val[key.Interface()] = getConvertMap(v.Type(), v)
	}
}







func ConvertTo(source interface{}, target interface{}) error {
	if source == nil || target == nil {
		return fmt.Errorf("input value is nil.")
	}
	return convertTo(reflect.TypeOf(source), reflect.ValueOf(source), reflect.TypeOf(target), reflect.ValueOf(target))
}

func convertTo(sType reflect.Type, sValue reflect.Value, tType reflect.Type, tValue reflect.Value) error {
	var kind int = 0

	// 检测源类型并接触引用
	switch sType.Kind() {
	case reflect.Ptr:
		if sValue.IsNil() {
			return nil
		}
		return convertTo(sValue.Elem().Type(), sValue.Elem(), tType, tValue)
	case reflect.Interface:
		if sValue.Elem().Kind() == reflect.Invalid {
			return nil
		}
		return convertTo(sValue.Elem().Type(), sValue.Elem(), tType, tValue)
	case reflect.Map:
		if sValue.IsNil() {
			return nil
		}
		kind = kind | 0x10
	case reflect.Struct:
		kind = kind | 0x20
	}

	// 检测目标类型并接触引用和初始化空对象
	switch tType.Kind() {
	case reflect.Ptr:
		if tValue.IsNil() {
			tValue.Set(reflect.New(tType.Elem()))
		}
		return convertTo(sType, sValue, tValue.Elem().Type(), tValue.Elem())
	case reflect.Interface:
		// if tValue.Elem().Kind() != reflect.Invalid {
		// 	return convertTo(sType, sValue, tValue.Elem().Type(), tValue.Elem())
		// }
		if isUseType(tValue) {
			return convertTo(sType, sValue, tValue.Elem().Type(), tValue.Elem())	
		}
	case reflect.Map:
		if tValue.IsNil() {
			tValue.Set(reflect.MakeMap(tType))
		}
		kind = kind | 0x01
	case reflect.Struct:
		kind = kind | 0x02
	}

	fmt.Println(kind, sType.Kind(), tType.Kind())
	// 更具数据类型执行转换
	switch kind {
	case 0x11:
		convertMapToMap1(sType, sValue, tType, tValue)
	case 0x21:
		convertStructToMap1(sType, sValue, tType, tValue)
	case 0x12:
		convertMapToStruct1(sType, sValue, tType, tValue)
	case 0x22:
		convertStructToStruct1(sType, sValue, tType, tValue)
	default:
		convertSetValue1(sType, sValue, tType, tValue)
	}
	return nil
}

func isUseType(iValue reflect.Value) bool {
	for iValue.Kind() == reflect.Interface || iValue.Kind() == reflect.Ptr {
		iValue = iValue.Elem()
	}
	return iValue.Kind() == reflect.Struct || iValue.Kind() == reflect.Map
}

// 检测一个值是否为空
func checkValueNil(iValue reflect.Value) bool {
	switch iValue.Type().Kind() {
	case reflect.Bool:
		return iValue.Bool() == false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return iValue.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return iValue.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return iValue.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return iValue.Complex() == 0
	case reflect.Func, reflect.Ptr, reflect.Interface:	
		return iValue.IsNil()
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return iValue.Len() == 0
	}
	return true
}

func convertMapToMap1(sType reflect.Type, sValue reflect.Value, tType reflect.Type, tValue reflect.Value) {
	for _, key := range sValue.MapKeys() {
		mapvalue := reflect.New(tType.Elem()).Elem()
		if err := convertTo(sValue.MapIndex(key).Type(), sValue.MapIndex(key), tType.Elem(), mapvalue); err == nil {
			tValue.SetMapIndex(key, mapvalue)
		}
		// tValue.SetMapIndex(key, sValue.MapIndex(key))
	}
}

func convertMapToStruct1(sType reflect.Type, sValue reflect.Value, tType reflect.Type, tValue reflect.Value) {
	for _, key := range sValue.MapKeys() {
		index := getStructField(tType, fmt.Sprint(key.Interface()))
		if index == -1 {
			continue
		}
		convertTo(sValue.MapIndex(key).Type(), sValue.MapIndex(key), tType.Field(index).Type, tValue.Field(index))
		// tValue.Field(index).Set(sValue.MapIndex(key))
	}
}

func convertStructToStruct1(sType reflect.Type, sValue reflect.Value, tType reflect.Type, tValue reflect.Value) {
	for i := 0; i < sType.NumField(); i++ {
		if checkValueNil(sValue.Field(i)) {
			continue
		}
		index := getStructField(tType, sType.Field(i).Name)
		if index == -1 {
			continue
		}
		convertTo(sType.Field(i).Type, sValue.Field(i), tType.Field(index).Type, tValue.Field(index))
		// tValue.Field(index).Set(sValue.Field(i))
	}
}

func convertStructToMap1(sType reflect.Type, sValue reflect.Value, tType reflect.Type, tValue reflect.Value) {
	for i := 0; i < sType.NumField(); i++ {
		if checkValueNil(sValue.Field(i)) {
			continue
		}
		mapvalue := reflect.New(tType.Elem()).Elem()
		if err := convertTo(sType.Field(i).Type, sValue.Field(i), tType.Elem(), mapvalue); err == nil {
			tValue.SetMapIndex(reflect.ValueOf(sType.Field(i).Name), mapvalue)
		}
		// tValue.SetMapIndex(reflect.ValueOf(sType.Field(i).Name), sValue.Field(i))
	}
}

func convertSetValue1(sType reflect.Type, sValue reflect.Value, tType reflect.Type, tValue reflect.Value) error {
	// fmt.Println("convertSetValue1", sValue.Interface())
	if sType.AssignableTo(tType) {
		tValue.Set(sValue)
		return nil
	}
	if sType.Kind() == reflect.String {
		return setWithString(tType.Kind(), tValue, sValue.String())
	}
	if tType.Kind() == reflect.String {
		tValue.SetString(getValueString(sType, sValue))
		return nil
	}
	return setWithString(tType.Kind(), tValue, getValueString(sType, sValue))
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
	// case reflect.Invalid, reflect.Chan, reflect.Func:
	// 	return ""
	case reflect.String:
		return iValue.String()
	default:
		return fmt.Sprint(iValue.Interface())
	}
	return ""
}

// 将一个interface{}赋值给对象
func setWithInterface(iTypeKind reflect.Kind, iValue reflect.Value, val interface{}) error {
	// 类型可以转换直接设置
	if reflect.TypeOf(val).AssignableTo(iValue.Type()) {
		iValue.Set(reflect.ValueOf(val))
		return nil
	}
	// 值的类型是字符串，进行类型转换。
	if reflect.TypeOf(val).Kind() == reflect.String {
		return setWithString(iTypeKind, iValue, val.(string))
	}
	// 目标类型是字符串直接输出interface{}对象。
	// 为支持接口。
	if iTypeKind == reflect.String {
		iValue.SetString(fmt.Sprint(val))
	}	
	return nil
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
	// 目标类型是数组、切片、结构体、map使用json反序列化解析。
	case reflect.Slice, reflect.Array, reflect.Struct, reflect.Map:
		 json.Unmarshal([]byte(val), iValue.Addr().Interface())
	// 目标类型是字符串直接设置
	case reflect.Interface:
		iValue.Set(reflect.ValueOf(val))
	// 目标类型是指针进行解引用然后赋值。
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
