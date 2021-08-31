package eudore

/*
功能1：获取和设置一个对象的属性
func Get(i interface{}, key string) interface{}
func GetWithTags(i interface{}, key string, tags []string) (interface{}, error)
func Set(i interface{}, key string, val interface{}) error
func SetWithTags(i interface{}, key string, val interface{}, tags []string) error

功能2：map和结构体相互转换
func ConvertMap(i interface{}) map[interface{}]interface{}
func ConvertMapString(i interface{}) map[string]interface{}
func ConvertMapStringWithTags(i interface{}, tags []string) map[string]interface{}
func ConvertMapWithTags(i interface{}, tags []string) map[interface{}]interface{}
func ConvertTo(source interface{}, target interface{}) error
func ConvertToWithTags(source interface{}, target interface{}, tags []string) error

功能3：sql结果Rows绑定
func ConvertRows(rows *sql.Rows, i interface{}) error
func ConvertRowsWithTags(rows *sql.Rows, i interface{}, tags []string) error
*/

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// seter 定义对象set属性的方法。
type seter interface {
	Set(string, interface{}) error
}

type getSeter struct {
	all   bool
	index int
	keys  []string
	tags  []string
	Value interface{}
}

type converter struct {
	tags    []string
	results map[reflect.Value]interface{}
}

// Set the properties of an object. The object must be a pointer type. If the target implements the Seter interface, the Set method is called.
//
// The path will be split using '.' and then look for the path in turn.
//
// When the object type selected in the path is ptr, it will check if it is empty. If the object is empty, it will be initialized by default.
//
// When the object type selected in the path is interface{}, if the object is empty, it will be initialized to map[string]interface{}, otherwise the value will be judged according to the value type.
//
// When the object type selected in the path is array, the path is converted to an object index to set the array element. If it cannot be converted, the element is appended.
//
// When the object type in the path is selected as a struct, the attribute name and the attribute tag 'alias' are used to match when selecting the attribute.
//
// If the type of the value is a string, it will be converted according to the target type set.
//
// If the target type is a string, the value is output as a string and then assigned.
//
// If the target type is an array, map, or struct, the json deserializes the set object.
//
// If the target type passed in is an array, map, or struct, the json deserializes the set object.
//
// 设置一个对象的属性,改对象必须是指针类型,如果目标实现Seter接口，调用Set方法。
//
// 路径将使用'.'分割，然后依次寻找路径。
//
// 当路径中选择对象类型为ptr时,会检查是否为空，对象为空会默认进行初始化。
//
// 当路径中选择对象类型为interface{}时,如果对象为空会初始化为map[string]interface{},否则按值类型来判断下一步操作。
//
// 当路径中选择对象类型为array时,路径会转换成对象索引来设置数组元素，无法转换则追加元素。
//
// 当路径中选择对象类型为struct时,选择属性时会使用属性名称和属性标签'alias'来匹配。
//
// 如果值的类型是字符串，会根据设置的目标类型来转换。
//
// 如果目标类型是字符串，将会值输出成字符串然后赋值。
//
// 如果目标类型是数组、map、结构体，会使用json反序列化设置对象。
//
// 如果传入的目标类型是数组、map、结构体，会使用json反序列化设置对象。
func Set(i interface{}, key string, val interface{}) error {
	return SetWithTags(i, key, val, DefaultGetSetTags)
}

// SetWithTags 函数和Set函数相同，可以额外设置tags。
func SetWithTags(i interface{}, key string, val interface{}, tags []string) error {
	if i == nil {
		return ErrConverterInputDataNil
	}
	seter, ok := i.(seter)
	if ok {
		err := seter.Set(key, val)
		if err == nil || err != ErrSeterNotSupportField {
			return err
		}
	}
	iValue := reflect.ValueOf(i)
	switch iValue.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Interface:
		if iValue.IsNil() {
			return ErrConverterInputDataNil
		}
	default:
		return ErrConverterInputDataNotPtr
	}
	if key == "" {
		return ErrConverterInputDataNil
	}
	s := &getSeter{
		keys:  strings.Split(key, "."),
		tags:  tags,
		Value: val,
	}
	return s.setValue(iValue)
}

func (s *getSeter) setValue(iValue reflect.Value) error {
	if len(s.keys) == 0 {
		return setWithValue(reflect.ValueOf(s.Value), iValue)
	}
	switch iValue.Kind() {
	case reflect.Ptr:
		if iValue.IsNil() {
			// 将空指针赋值
			iValue.Set(reflect.New(iValue.Type().Elem()))
		}
		return s.setValue(iValue.Elem())
	case reflect.Interface:
		return s.setInterface(iValue)
	case reflect.Struct:
		return s.setStruct(iValue)
	case reflect.Map:
		return s.setMap(iValue)
	case reflect.Slice:
		return s.setSlice(iValue)
	case reflect.Array:
		return s.setArray(iValue)
	}
	return fmt.Errorf(ErrFormatConverterSetTypeError, iValue.Kind(), s.keys, s.Value)
}

// 设置接口类型
func (s *getSeter) setInterface(iValue reflect.Value) (err error) {
	// 如果是空接口，初始化为map[string]interface{}类型
	if iValue.IsNil() {
		if iValue.Type() != typeInterface {
			return nil
		}
		iValue.Set(reflect.ValueOf(make(map[string]interface{})))
	}
	// 创建一个可取地址的临时变量，并设置值用于下一步设置。
	newValue := reflect.New(iValue.Elem().Type()).Elem()
	newValue.Set(iValue.Elem())
	err = s.setValue(newValue)
	// 将修改后的值重新赋值给对象
	if err == nil {
		iValue.Set(newValue)
	}
	return err
}

// 处理结构体设置属性
func (s *getSeter) setStruct(iValue reflect.Value) error {
	// 查找属性是结构体的第几个属性。
	var index = getStructIndexOfTags(iValue.Type(), s.keys[0], s.tags)
	// 未找到直接返回错误。
	if index == -1 {
		return fmt.Errorf(ErrFormatConverterSetStructNotField, s.keys[0])
	}

	// 获取结构体的属性
	structField := iValue.Field(index)
	if !structField.CanSet() {
		return fmt.Errorf(ErrFormatConverterNotCanset, s.keys[0], iValue.Type().String())
	}
	s.keys = s.keys[1:]
	return s.setValue(structField)
}

// 处理map
func (s *getSeter) setMap(iValue reflect.Value) error {
	iType := iValue.Type()
	// 对空map初始化
	if iValue.IsNil() {
		iValue.Set(reflect.MakeMap(iType))
	}

	// 创建map需要匹配的key
	mapKey := reflect.New(iType.Key()).Elem()
	setWithString(mapKey, s.keys[0])

	newValue := reflect.New(iType.Elem()).Elem()
	mapvalue := iValue.MapIndex(mapKey)
	if mapvalue.Kind() != reflect.Invalid {
		newValue.Set(mapvalue)
	}

	s.keys = s.keys[1:]
	err := s.setValue(newValue)
	// 将修改后的mapvalue重新赋值给map
	if err == nil {
		iValue.SetMapIndex(mapKey, newValue)
	}
	return err
}

func (s *getSeter) setArray(iValue reflect.Value) error {
	index, err := strconv.Atoi(s.keys[0])
	if err != nil || index < 0 || index >= iValue.Len() {
		return fmt.Errorf(ErrFormatConverterSetArrayIndexInvalid, s.keys[0], iValue.Len())
	}
	s.keys = s.keys[1:]
	return s.setValue(iValue.Index(index))
}

// 处理数组和切片
func (s *getSeter) setSlice(iValue reflect.Value) error {
	iType := iValue.Type()
	// 空切片初始化，默认长度2
	if iValue.IsNil() {
		iValue.Set(reflect.MakeSlice(iType, 0, 2))
	}
	// 创建新元素的类型和值
	newValue := reflect.New(iType.Elem()).Elem()
	index, err := strconv.Atoi(s.keys[0])
	if err != nil {
		index = -1
	}
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

	s.keys = s.keys[1:]
	err = s.setValue(newValue)
	if err == nil {
		if index > -1 {
			iValue.Index(index).Set(newValue)
		} else {
			iValue.Set(reflect.Append(iValue, newValue))
		}
	}
	return err
}

// Get method A more path to get an attribute from an object.
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
func Get(i interface{}, key string) interface{} {
	val, err := getValue(i, key, false, DefaultGetSetTags)
	if err != nil {
		return nil
	}
	return val.Interface()
}

// GetWithTags 函数和Get函数相同，可以额外设置tags，同时会返回error。
func GetWithTags(i interface{}, key string, tags []string) (interface{}, error) {
	val, err := getValue(i, key, false, tags)
	if err != nil {
		return nil, err
	}
	return val.Interface(), nil
}

// GetWithValue 函数和Get函数相同，可以允许查找私有属性并返回reflect.Value。
func GetWithValue(i interface{}, key string, all bool) (reflect.Value, error) {
	return getValue(i, key, all, nil)
}

func getValue(i interface{}, key string, all bool, tags []string) (reflect.Value, error) {
	val := reflect.ValueOf(i)
	if i == nil {
		return val, ErrConverterInputDataNil
	}
	if key == "" {
		return val, nil
	}
	s := &getSeter{
		all:  all,
		keys: strings.Split(key, "."),
		tags: tags,
	}
	val, err := s.getValue(val)
	if err != nil {
		return val, err
	}
	return val, nil
}

// 从目标类型获取字符串路径的属性
func (s *getSeter) getValue(iValue reflect.Value) (reflect.Value, error) {
	if len(s.keys) == s.index {
		return iValue, nil
	}
	switch iValue.Kind() {
	case reflect.Ptr, reflect.Interface:
		if iValue.IsNil() {
			return iValue, s.newGetError("is nil ptr or interface")
		}
		return s.getValue(iValue.Elem())
	case reflect.Struct:
		return s.getStruct(iValue)
	case reflect.Map:
		return s.getMap(iValue)
	case reflect.Array, reflect.Slice:
		return s.getSlice(iValue)
	}
	return iValue, s.newGetError("not find sub path")
}

// 处理结构体对象的读取
func (s *getSeter) getStruct(iValue reflect.Value) (reflect.Value, error) {
	// 查找key对应的属性索引，不存在返回-1。
	var index = getStructIndexOfTags(iValue.Type(), s.keys[s.index], s.tags)
	if index == -1 {
		return iValue, s.newGetError("not field")
	}
	// 获取key对应结构的属性。
	structField := iValue.Field(index)
	if structField.CanSet() || s.all {
		s.index++
		return s.getValue(structField)
	}
	return iValue, s.newGetError("field is not CanSet")
}

// 处理map读取属性
func (s *getSeter) getMap(iValue reflect.Value) (reflect.Value, error) {
	// 检测map是否为空
	if iValue.IsNil() {
		return iValue, s.newGetError("is nil map")
	}
	// 创建map需要的key
	mapKey := reflect.New(iValue.Type().Key()).Elem()
	err := setWithString(mapKey, s.keys[s.index])
	if err != nil {
		return iValue, s.newGetError("map key is invalid")
	}

	// 获得map的value, 如果值无效则返回空。
	mapvalue := iValue.MapIndex(mapKey)
	if mapvalue.Kind() == reflect.Invalid {
		return iValue, s.newGetError("map value is invalid")
	}
	s.index++
	return s.getValue(mapvalue)
}

// 处理数组切片读取属性
func (s *getSeter) getSlice(iValue reflect.Value) (reflect.Value, error) {
	// 检测切片是否为空
	if iValue.Kind() == reflect.Slice && iValue.IsNil() {
		return iValue, s.newGetError("is nil slice")
	}
	// 检测索引是否存在
	index, err := strconv.Atoi(s.keys[s.index])
	if err != nil || index < 0 || iValue.Len() <= index {
		return iValue, s.newGetError("slice index is invalid")
	}
	s.index++
	return s.getValue(iValue.Index(index))
}

func (s *getSeter) newGetError(str string) error {
	return fmt.Errorf(ErrFormatConverterGet, strings.Join(s.keys[:s.index+1], "."), str)
}

// ConvertMapString 函数将一个map或struct转换成map[string]interface{}。
func ConvertMapString(i interface{}) map[string]interface{} {
	return ConvertMapStringWithTags(i, DefaultConvertTags)
}

// ConvertMapStringWithTags 函数与ConvertMapString相同，允许使用额外的tags。
func ConvertMapStringWithTags(i interface{}, tags []string) map[string]interface{} {
	c := &converter{
		tags:    tags,
		results: make(map[reflect.Value]interface{}),
	}
	// 其他类型直接返回
	val, ok := c.convertMapString(reflect.ValueOf(i)).(map[string]interface{})
	if ok {
		return val
	}
	return nil
}

// 将一个map或结构体对象转换成map[string]interface{}返回。
func (c *converter) convertMapString(iValue reflect.Value) interface{} {
	result, ok := c.results[iValue]
	if ok {
		return result
	}
	switch iValue.Kind() {
	// 接口类型解除引用
	case reflect.Interface:
		// 空接口直接返回
		if iValue.Elem().Kind() == reflect.Invalid {
			return iValue.Interface()
		}
		return c.convertMapString(iValue.Elem())
	// 指针类型解除引用
	case reflect.Ptr:
		// 空指针直接返回
		if iValue.IsNil() {
			return iValue.Interface()
		}
		return c.convertMapString(iValue.Elem())
	// 将map转换成map[string]interface{}
	case reflect.Map:
		val := make(map[string]interface{})
		c.results[iValue] = val
		c.convertMapstrngMapToMapString(iValue, val)
		return val
	// 将结构体转换成map[string]interface{}
	case reflect.Struct:
		val := make(map[string]interface{})
		c.results[iValue] = val
		c.convertMapstringStructToMapString(iValue, val)
		return val
	}
	// 其他类型直接返回
	return iValue.Interface()
}

// 结构体转换成map[string]interface{}
func (c *converter) convertMapstringStructToMapString(iValue reflect.Value, val map[string]interface{}) {
	iType := iValue.Type()
	// 遍历结构体的属性
	for i := 0; i < iType.NumField(); i++ {
		fieldKey := iType.Field(i)
		fieldValue := iValue.Field(i)
		if fieldValue.CanSet() {
			// map设置键位结构体的名称，值为结构体值转换，基本类型会直接返回。
			val[getStructNameOfTags(fieldKey, c.tags)] = c.convertMapString(fieldValue)
		}
	}
}

// 将map转换成map[string]interface{}
func (c *converter) convertMapstrngMapToMapString(iValue reflect.Value, val map[string]interface{}) {
	// 遍历map的全部keys
	for _, key := range iValue.MapKeys() {
		v := iValue.MapIndex(key)
		// 设置新map的键为原map的字符串输出，未支持接口转换
		// 设置新map的值为原map匹配的值的转换，值为基本类型会直接返回。
		val[fmt.Sprint(key.Interface())] = c.convertMapString(v)
	}
}

// ConvertMap 函数将一个map或struct转换成map[interface{}]interface{}。
func ConvertMap(i interface{}) map[interface{}]interface{} {
	return ConvertMapWithTags(i, DefaultConvertTags)
}

// ConvertMapWithTags 函数与ConvertMap相同，允许使用额外的tags。
func ConvertMapWithTags(i interface{}, tags []string) map[interface{}]interface{} {
	c := &converter{
		tags:    tags,
		results: make(map[reflect.Value]interface{}),
	}
	// 其他类型直接返回
	val, ok := c.convertMap(reflect.ValueOf(i)).(map[interface{}]interface{})
	if ok {
		return val
	}
	return nil
}

// 将一个map或结构体对象转换成map[interface{}]interface{}返回。
func (c *converter) convertMap(iValue reflect.Value) interface{} {
	result, ok := c.results[iValue]
	if ok {
		return result
	}
	switch iValue.Kind() {
	case reflect.Interface:
		if iValue.Elem().Kind() == reflect.Invalid {
			return iValue.Interface()
		}
		return c.convertMap(iValue.Elem())
	case reflect.Ptr:
		if iValue.IsNil() {
			return iValue.Interface()
		}
		return c.convertMap(iValue.Elem())
	case reflect.Map:
		val := make(map[interface{}]interface{})
		c.results[iValue] = val
		c.convertMapMapToMap(iValue, val)
		return val
	case reflect.Struct:
		val := make(map[interface{}]interface{})
		c.results[iValue] = val
		c.convertMapStructToMap(iValue, val)
		return val
	}
	return iValue.Interface()
}

// 结构体转换成map[interface{}]interface{}
func (c *converter) convertMapStructToMap(iValue reflect.Value, val map[interface{}]interface{}) {
	iType := iValue.Type()
	// 遍历结构体的属性
	for i := 0; i < iType.NumField(); i++ {
		fieldKey := iType.Field(i)
		fieldValue := iValue.Field(i)
		if fieldValue.CanSet() {
			// map设置键位结构体的名称，值为结构体值转换，基本类型会直接返回。
			val[getStructNameOfTags(fieldKey, c.tags)] = c.convertMap(fieldValue)
		}
	}
}

// 将map转换成map[interface{}]interface{}
func (c *converter) convertMapMapToMap(iValue reflect.Value, val map[interface{}]interface{}) {
	// 遍历map的全部keys
	for _, key := range iValue.MapKeys() {
		v := iValue.MapIndex(key)
		// 设置新map的键为原map的字符串输出，未支持接口转换
		// 设置新map的值为原map匹配的值的转换，值为基本类型会直接返回。
		val[key.Interface()] = c.convertMap(v)
	}
}

// ConvertTo 将一个对象属性复制给另外一个对象,可转换对象属性会覆盖原值。
func ConvertTo(source interface{}, target interface{}) error {
	return ConvertToWithTags(source, target, DefaultConvertTags)
}

// ConvertToWithTags 函数与ConvertTo相同，允许使用额外的tags。
func ConvertToWithTags(source interface{}, target interface{}, tags []string) error {
	if source == nil {
		return ErrConverterInputDataNil
	}
	if target == nil {
		return ErrConverterTargetDataNil
	}

	// 检测目标是指针类型。
	iValue := reflect.ValueOf(target)
	switch iValue.Kind() {
	case reflect.Ptr:
	case reflect.Map, reflect.Interface:
		if iValue.IsNil() {
			return ErrConverterInputDataNotPtr
		}
	default:
		return ErrConverterInputDataNotPtr
	}

	c := &converter{
		tags: tags,
	}
	return c.convertTo(reflect.ValueOf(source), reflect.ValueOf(target))
}

func (c *converter) convertTo(sValue reflect.Value, tValue reflect.Value) error {
	if sValue.Kind() == reflect.Ptr || sValue.Kind() == reflect.Interface || tValue.Kind() == reflect.Ptr || tValue.Kind() == reflect.Interface {
		stypes, svalues := getIndirectAllValue(sValue)
		ttypes, tvalues := getIndirectAllValue(tValue)
		sValue = svalues[len(svalues)-1]
		tValue = tvalues[len(tvalues)-1]
		for i, ttype := range ttypes {
			for j, stype := range stypes {
				// 转换接口类型、相同类型、type别名类型
				if stype.ConvertibleTo(ttype) && tvalues[i].CanSet() {
					// 如果类型最终指向map或struct则进行最后转换，将map或struct转换成map或struct
					if ttype.Kind() == reflect.Ptr && indirectKindInMapStruct(ttype) && indirectKindInMapStruct(stype) {
						return c.convertTo(sValue, tValue)
					}
					return c.convertToData(svalues[j], tvalues[i])
				}
			}
		}

		// 目标类型如果是空指针，则尝试进行初始化并转换
		if tValue.Kind() == reflect.Ptr && tValue.IsNil() {
			newValue := reflect.New(tValue.Type().Elem())
			err := c.convertTo(sValue, newValue)
			if err == nil {
				tValue.Set(newValue)
			}
			return err
		}
	}
	return c.convertToData(sValue, tValue)
}

func indirectKindInMapStruct(iType reflect.Type) bool {
	for iType.Kind() == reflect.Ptr {
		iType = iType.Elem()
	}
	return iType.Kind() == reflect.Struct || iType.Kind() == reflect.Map
}

func (c *converter) convertToData(sValue reflect.Value, tValue reflect.Value) error {
	skind := sValue.Kind()
	tkind := tValue.Kind()
	// map和struct转换
	switch {
	case skind == reflect.Map && tkind == reflect.Map:
		c.convertToMapToMap(sValue, tValue)
	case skind == reflect.Map && tkind == reflect.Struct:
		c.convertToMapToStruct(sValue, tValue)
	case skind == reflect.Struct && tkind == reflect.Map:
		c.convertToStructToMap(sValue, tValue)
	case skind == reflect.Struct && tkind == reflect.Struct:
		c.convertToStructToStruct(sValue, tValue)
	case (skind == reflect.Slice || skind == reflect.Array) && (tkind == reflect.Slice || tkind == reflect.Array):
		c.convertToSlice(sValue, tValue)
	case checkValueIsZero(sValue):
		return nil
	default:
		setWithValueData(sValue, tValue)
	}
	return nil
}

func (c *converter) convertToSlice(sValue reflect.Value, tValue reflect.Value) {
	num := sValue.Len() - tValue.Len()
	if num > 0 && tValue.CanSet() {
		tValue.Set(reflect.AppendSlice(tValue, reflect.MakeSlice(tValue.Type(), num, num)))
	}
	if num > 0 {
		num = tValue.Len()
	} else {
		num = sValue.Len()
	}
	for i := 0; i < num; i++ {
		c.convertTo(sValue.Index(i), tValue.Index(i))
	}
}

func (c *converter) convertToMapToMap(sValue reflect.Value, tValue reflect.Value) {
	tType := tValue.Type()
	if tValue.IsNil() {
		tValue.Set(reflect.MakeMap(tType))
	}
	for _, key := range sValue.MapKeys() {
		mapvalue := reflect.New(tType.Elem()).Elem()
		if err := c.convertTo(sValue.MapIndex(key), mapvalue); err == nil {
			tValue.SetMapIndex(key, mapvalue)
		}
	}
}

func (c *converter) convertToMapToStruct(sValue reflect.Value, tValue reflect.Value) {
	tType := tValue.Type()
	for _, key := range sValue.MapKeys() {
		index := getStructIndexOfTags(tType, fmt.Sprint(key.Interface()), c.tags)
		if index == -1 || !tValue.Field(index).CanSet() {
			continue
		}
		c.convertTo(sValue.MapIndex(key), tValue.Field(index))
	}
}

func (c *converter) convertToStructToMap(sValue reflect.Value, tValue reflect.Value) {
	sType := sValue.Type()
	tType := tValue.Type()
	if tValue.IsNil() {
		tValue.Set(reflect.MakeMap(tType))
	}
	for i := 0; i < sType.NumField(); i++ {
		if !sValue.Field(i).CanSet() || checkValueIsZero(sValue.Field(i)) {
			continue
		}
		mapvalue := reflect.New(tType.Elem()).Elem()
		if err := c.convertTo(sValue.Field(i), mapvalue); err == nil {
			tValue.SetMapIndex(reflect.ValueOf(sType.Field(i).Name), mapvalue)
		}
	}
}

func (c *converter) convertToStructToStruct(sValue reflect.Value, tValue reflect.Value) {
	sType := sValue.Type()
	tType := tValue.Type()
	for i := 0; i < sType.NumField(); i++ {
		if !sValue.CanSet() || checkValueIsZero(sValue.Field(i)) {
			continue
		}
		index := getStructIndexOfTags(tType, sType.Field(i).Name, c.tags)
		if index == -1 || !tValue.Field(index).CanSet() {
			continue
		}
		c.convertTo(sValue.Field(i), tValue.Field(index))
	}
}

// ConvertRows 函数将*sql.Rows数据解析成指定struct、map、slice。
func ConvertRows(rows *sql.Rows, i interface{}) error {
	return ConvertRowsWithTags(rows, i, DefaultConvertRowsTags)
}

// ConvertRowsWithTags 函数指定tags将*sql.Rows数据解析成指定struct、map、slice。
func ConvertRowsWithTags(rows *sql.Rows, i interface{}, tags []string) error {
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("ConvertRows get columns error: %s", err.Error())
	}
	iValue := reflect.ValueOf(i)
	if iValue.Kind() == reflect.Invalid {
		return fmt.Errorf("ConvertRows target is invalid zero value")
	}
	scaner := &sqlScaner{
		tags:    tags,
		Rows:    rows,
		columns: columns,
		isrow:   true,
	}
	err = scaner.Scan(iValue)
	if err != nil {
		return fmt.Errorf("ConvertRows scan type %s error: %s", reflect.TypeOf(i).String(), err.Error())
	}
	return nil
}

type sqlScaner struct {
	tags         []string
	Rows         *sql.Rows
	columns      []string
	isrow        bool
	structFields []int
	mapValues    []interface{}
}

func (scan *sqlScaner) Scan(iValue reflect.Value) error {
	switch iValue.Kind() {
	case reflect.Slice, reflect.Array:
		scan.isrow = false
		return scan.ScanSlice(iValue)
	case reflect.Struct:
		if iValue.CanAddr() {
			return scan.ScanStruct(iValue)
		}
		return fmt.Errorf("struct %s must can addr", iValue.Type().String())
	case reflect.Map:
		return scan.ScanMap(iValue)
	case reflect.Ptr:
		if iValue.IsNil() {
			if iValue.CanSet() {
				iValue.Set(reflect.New(iValue.Type().Elem()))
			} else {
				return fmt.Errorf("ptr %s is nil and not set", iValue.Type().String())
			}
		}
		return scan.Scan(iValue.Elem())
	case reflect.Interface:
		if iValue.IsNil() {
			return fmt.Errorf("interface %s is nil", iValue.Type().String())
		}
		return scan.Scan(iValue.Elem())
	case reflect.String, reflect.Int, reflect.Uint, reflect.Float32, reflect.Float64, reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if len(scan.columns) == 1 && iValue.CanAddr() {
			if scan.isrow && scan.Rows.Next() {
				defer scan.Rows.Close()
			}
			return scan.Rows.Scan(iValue.Addr().Interface())
		}
		return fmt.Errorf("base type scan columns must is one,this is %d", len(scan.columns))
	}
	return fmt.Errorf("scan invalid type %s", iValue.Type().String())
}

func (scan *sqlScaner) ScanSlice(iValue reflect.Value) error {
	num := iValue.Len()
	if iValue.Kind() == reflect.Slice {
		num = iValue.Cap()
	}
	if num == 0 {
		num = 65536
	}
	for i := 0; scan.Rows.Next() && i < num; i++ {
		if i >= iValue.Len() {
			iValue.Set(reflect.Append(iValue, reflect.New(iValue.Type().Elem()).Elem()))
		}
		err := scan.Scan(iValue.Index(i))
		if err != nil {
			return err
		}
	}
	scan.Rows.Close()
	return scan.Rows.Err()
}

func (scan *sqlScaner) ScanStruct(iValue reflect.Value) error {
	if scan.isrow && scan.Rows.Next() {
		defer scan.Rows.Close()
	}
	if scan.structFields == nil {
		scan.structFields = scan.getStructFiles(iValue.Type(), scan.columns)
	}
	datas := make([]interface{}, len(scan.structFields))
	for i, field := range scan.structFields {
		datas[i] = iValue.Field(field).Addr().Interface()
	}
	return scan.Rows.Scan(datas...)
}

func (scan *sqlScaner) getStructFiles(iType reflect.Type, columns []string) []int {
	fields := make([]int, 0, len(columns))
	for _, column := range columns {
		for i := 0; i < iType.NumField(); i++ {
			field := iType.Field(i)
			if strings.ToLower(field.Name) == column {
				fields = append(fields, i)
				break
			}
			for _, tag := range scan.tags {
				if field.Tag.Get(tag) == column {
					fields = append(fields, i)
					break
				}
			}
		}
	}
	return fields
}

func (scan *sqlScaner) ScanMap(iValue reflect.Value) error {
	if iValue.Type().Key() != typeString {
		return fmt.Errorf("map key type must is string, current key type is %s", iValue.Type().String())
	}
	if iValue.IsNil() {
		if !iValue.CanSet() {
			return fmt.Errorf("map %s is nil and not set", iValue.Type().String())
		}
		iValue.Set(reflect.MakeMap(iValue.Type()))
	}
	if scan.isrow && scan.Rows.Next() {
		defer scan.Rows.Close()
	}
	if scan.mapValues == nil {
		types, _ := scan.Rows.ColumnTypes()
		scan.mapValues = make([]interface{}, len(scan.columns))
		for i := 0; i < len(scan.columns); i++ {
			scan.mapValues[i] = reflect.New(types[i].ScanType()).Interface()
		}
	}
	err := scan.Rows.Scan(scan.mapValues...)
	if err == nil {
		for i := 0; i < len(scan.columns); i++ {
			iValue.SetMapIndex(reflect.ValueOf(scan.columns[i]), reflect.ValueOf(scan.mapValues[i]).Elem())
		}
	}
	return err
}

// checkValueIsZero 函数检测一个值是否为空, 修改go.1.13 refletv.Value.IsZero方法。
func checkValueIsZero(iValue reflect.Value) bool {
	switch iValue.Kind() {
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
	case reflect.String:
		return iValue.Len() == 0
	case reflect.UnsafePointer:
		// 兼容go1.9
		//		if iValue.CanSet(){
		return iValue.Interface() == nil
		//}
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return iValue.IsNil()
	case reflect.Array:
		for i := 0; i < iValue.Len(); i++ {
			if !checkValueIsZero(iValue.Index(i)) {
				return false
			}
		}
	case reflect.Struct:
		for i := 0; i < iValue.NumField(); i++ {
			if !checkValueIsZero(iValue.Field(i)) {
				return false
			}
		}
	}
	return true
}

// 通过字符串获取结构体属性的索引
func getStructIndexOfTags(iType reflect.Type, name string, tags []string) int {
	// 遍历匹配
	for i := 0; i < iType.NumField(); i++ {
		typeField := iType.Field(i)
		// 字符串为结构体名称或结构体属性标签的值，则匹配返回索引。
		if typeField.Name == name {
			return i
		}
		for _, tag := range tags {
			if typeField.Tag.Get(tag) == name {
				return i
			}
		}
	}
	return -1
}

func getStructNameOfTags(field reflect.StructField, tags []string) string {
	for _, tag := range tags {
		name := field.Tag.Get(tag)
		if name != "" {
			return name
		}
	}
	return field.Name
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

func setWithValue(sValue reflect.Value, tValue reflect.Value) error {
	if sValue.Kind() == reflect.Ptr || sValue.Kind() == reflect.Interface || tValue.Kind() == reflect.Ptr || tValue.Kind() == reflect.Interface {
		stypes, svalues := getIndirectAllValue(sValue)
		ttypes, tvalues := getIndirectAllValue(tValue)
		for i, ttype := range ttypes {
			for j, stype := range stypes {
				// 转换接口类型、相同类型、type别名类型
				if stype.ConvertibleTo(ttype) && tvalues[i].CanSet() {
					return setWithValueData(svalues[j], tvalues[i])
				}
			}
		}
		sValue = svalues[len(svalues)-1]
		tValue = tvalues[len(tvalues)-1]

		// 目标类型如果是空指针，则尝试进行初始化并转换
		if tValue.Kind() == reflect.Ptr && tValue.IsNil() {
			newValue := reflect.New(tValue.Type().Elem())
			err := setWithValue(sValue, newValue)
			if err == nil {
				tValue.Set(newValue)
			}
			return err
		}
	}
	return setWithValueData(sValue, tValue)
}

func setWithValueData(sValue reflect.Value, tValue reflect.Value) error {
	sType := sValue.Type()
	tType := tValue.Type()
	switch {
	case sType == tType:
		tValue.Set(sValue)
		return nil
	case sType.Kind() == reflect.String:
		return setWithString(tValue, sValue.String())
	case tType.Kind() == reflect.String:
		tValue.SetString(getWithValueString(sType, sValue))
		return nil
	case sType.ConvertibleTo(tType):
		tValue.Set(sValue.Convert(tType))
		return nil
	case sType.Kind() == reflect.Slice:
		err := setWithValueData(reflect.Indirect(sValue.Index(0)), tValue)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf(ErrFormatConverterSetWithValue, sValue.Type().String(), tValue.Type().String())
}

func getWithValueString(t reflect.Type, v reflect.Value) string {
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		switch t.Elem().Kind() {
		case reflect.String:
			if v.Len() > 0 {
				return v.Index(0).String()
			}
		case reflect.Uint8, reflect.Int32:
			return v.Convert(typeString).String()
		}
	}
	return fmt.Sprintf("%#v", v.Interface())
}

// 将一个字符串赋值给对象
func setWithString(iValue reflect.Value, val string) error {
	val = strings.TrimSpace(val)
	switch iValue.Kind() {
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
	case reflect.Complex64:
		return setComplexField(val, 32, iValue)
	case reflect.Complex128:
		return setComplexField(val, 64, iValue)
	// 目标类型是字符串直接设置
	case reflect.String:
		iValue.SetString(val)
	case reflect.Struct:
		if iValue.Type() == typeTimeTime {
			return setTimeField(val, iValue)
		}
		return json.Unmarshal([]byte(val), iValue.Addr().Interface())
	case reflect.Slice:
		switch iValue.Type().Elem().Kind() {
		case reflect.Uint8, reflect.Int32:
			iValue.Set(reflect.ValueOf(val).Convert(iValue.Type()))
		default:
			return json.Unmarshal([]byte(val), iValue.Addr().Interface())
		}
	case reflect.Array, reflect.Map:
		return json.Unmarshal([]byte(val), iValue.Addr().Interface())
	case reflect.Interface:
		if iValue.Type() == typeInterface {
			iValue.Set(reflect.ValueOf(val))
		}
	// 目标类型是指针进行解引用然后赋值。
	case reflect.Ptr:
		if !iValue.Elem().IsValid() {
			iValue.Set(reflect.New(iValue.Type().Elem()))
		}
		return setWithString(iValue.Elem(), val)
	default:
		return fmt.Errorf(ErrFormatConverterSetStringUnknownType, iValue.Kind().String())
	}
	return nil
}

// 设置int类型属性
func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	// 兼容 time.Duration及衍生类型
	if err != nil && field.Kind() == reflect.Int64 {
		var t time.Duration
		t, err = time.ParseDuration(val)
		if err != nil {
			return err
		}
		intVal = int64(t)
	}
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

// 设置复数
func setComplexField(val string, bitSize int, field reflect.Value) error {
	val = strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(val, "("), "i"), ")")
	pos := strings.Index(val, "+")
	if pos == -1 {
		pos = len(val)
		val += "+0"
	}

	read, err := strconv.ParseFloat(val[:pos], bitSize)
	if err != nil {
		return err

	}
	image, err := strconv.ParseFloat(val[pos+1:], bitSize)
	if err != nil {
		return err
	}

	field.SetComplex(complex(read, image))
	return nil
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

// timeformats 定义允许使用的时间格式。
var timeformats = []string{
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
	"2006-1-02",
	"2006-01-02",
	"15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05.999999999Z07:00",
}

// TimeParse 方法通过解析内置支持的时间格式。
func setTimeField(str string, iValue reflect.Value) (err error) {
	var t time.Time
	for _, f := range timeformats {
		t, err = time.Parse(f, str)
		if err == nil {
			iValue.Set(reflect.ValueOf(t))
			return
		}
	}
	return
}
