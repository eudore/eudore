package eudore

import (
	"fmt"
	"reflect"
	"unsafe"
)

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
	if reflect.TypeOf(target).Kind() != reflect.Ptr {
		return ErrConverterInputDataNotPtr
	}

	c := &convertMapping{
		Tags: tags,
		Refs: make(map[unsafe.Pointer]reflect.Value),
	}
	return c.convertToData(reflect.ValueOf(source), reflect.ValueOf(target))
}

type convertMapping struct {
	Tags []string
	Refs map[unsafe.Pointer]reflect.Value
}

func getValuePointer(iValue reflect.Value) unsafe.Pointer {
	val := *(*innerValue)(unsafe.Pointer(&iValue))
	return val.ptr
}

type innerValue struct {
	_    *int
	ptr  unsafe.Pointer
	flag uintptr
}

func (c *convertMapping) convertToData(sValue reflect.Value, tValue reflect.Value) error {
	switch sValue.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Interface:
		if !sValue.IsNil() && tValue.CanSet() {
			ref, ok := c.Refs[getValuePointer(sValue)]
			if ok {
				if ref.Type().ConvertibleTo(tValue.Type()) {
					tValue.Set(ref.Convert(tValue.Type()))
					return nil
				}
			}
		}
	}

	skind := sValue.Kind()
	tkind := tValue.Kind()
	switch {
	case checkValueIsZero(sValue):
		return nil
	case sValue.Kind() == reflect.Interface:
		c.Refs[getValuePointer(sValue)] = tValue
		return c.convertToData(sValue.Elem(), tValue)
	case tValue.Kind() == reflect.Interface:
		if tValue.IsNil() {
			newValue := reflect.New(sValue.Type()).Elem()
			if newValue.Type().ConvertibleTo(tValue.Type()) {
				err := c.convertToData(sValue, newValue)
				if err == nil {
					tValue.Set(newValue.Convert(tValue.Type()))
				}
				return err
			}
		} else {
			return c.convertToData(sValue, tValue.Elem())
		}
	case sValue.Kind() == reflect.Ptr:
		c.Refs[getValuePointer(sValue)] = tValue
		return c.convertToData(sValue.Elem(), tValue)
	case tValue.Kind() == reflect.Ptr:
		if tValue.IsNil() {
			newValue := reflect.New(tValue.Type().Elem())
			err := c.convertToData(sValue, newValue.Elem())
			if err == nil {
				tValue.Set(newValue)
			}
			return err
		}
		return c.convertToData(sValue, tValue.Elem())
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
	default:
		return setWithValueData(sValue, tValue)
	}
	return nil
}

func (c *convertMapping) convertToMapToMap(sValue reflect.Value, tValue reflect.Value) {
	tType := tValue.Type()
	if tValue.IsNil() {
		tValue.Set(reflect.MakeMap(tType))
	}

	// TODO: map to map
	// c.Refs[getValuePointer(sValue)] = tValue
	for _, key := range sValue.MapKeys() {
		mapvalue := reflect.New(tType.Elem()).Elem()
		if err := c.convertToData(sValue.MapIndex(key), mapvalue); err == nil {
			tValue.SetMapIndex(key, mapvalue)
		}
	}

}

func (c *convertMapping) convertToMapToStruct(sValue reflect.Value, tValue reflect.Value) {
	tType := tValue.Type()
	for _, key := range sValue.MapKeys() {
		index := getStructIndexOfTags(tType, fmt.Sprint(key.Interface()), c.Tags)
		if index == -1 || !tValue.Field(index).CanSet() {
			continue
		}
		c.convertToData(sValue.MapIndex(key), tValue.Field(index))
	}
}

func (c *convertMapping) convertToStructToMap(sValue reflect.Value, tValue reflect.Value) {
	sType := sValue.Type()
	tType := tValue.Type()
	if tValue.IsNil() {
		tValue.Set(reflect.MakeMap(tType))
	}
	for i := 0; i < sType.NumField(); i++ {
		if checkValueIsZero(sValue.Field(i)) || !sValue.Field(i).CanSet() {
			continue
		}

		mapvalue := reflect.New(tType.Elem()).Elem()
		if err := c.convertToData(sValue.Field(i), mapvalue); err == nil {
			tValue.SetMapIndex(reflect.ValueOf(sType.Field(i).Name), mapvalue)
		}
	}
}

func (c *convertMapping) convertToStructToStruct(sValue reflect.Value, tValue reflect.Value) {
	sType := sValue.Type()
	tType := tValue.Type()
	for i := 0; i < sType.NumField(); i++ {
		if checkValueIsZero(sValue.Field(i)) {
			continue
		}

		index := getStructIndexOfTags(tType, sType.Field(i).Name, c.Tags)
		if index == -1 || !tValue.Field(index).CanSet() {
			continue
		}
		c.convertToData(sValue.Field(i), tValue.Field(index))
	}
}

func (c *convertMapping) convertToSlice(sValue reflect.Value, tValue reflect.Value) {
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
		c.convertToData(sValue.Index(i), tValue.Index(i))
	}
}
