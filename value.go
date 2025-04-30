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

// The GetAnyByPath function gets the specified attribute of the object.
//
// The path will be separated by '.' and then searched for in sequence.
//
// The structure uses the tag 'alias' to match the field by default.
func GetAnyByPath(data any, key string, tags []string) (any, error) {
	v := &value{tags: tags}
	err := v.Look(data, strings.Split(key, ".")...)
	if err != nil {
		return nil, err
	}
	return v.value.Interface(), nil
}

// The GetAnyByPath function gets the specified attribute of the object and
// returns the [reflect.Value] type.
//
// refer: [GetAnyByPath].
func GetValueByPath(data any, key string, tags []string) (reflect.Value, error) {
	v := &value{tags: tags}
	err := v.Look(data, strings.Split(key, ".")...)
	if err != nil {
		return reflect.Value{}, err
	}
	return v.value, nil
}

// The GetAnyByPath function sets the specified attribute of the object.
//
// If the value type is string, it will be converted according to the dta type.
//
// refer: [GetAnyByPath].
func SetAnyByPath(data any, key string, val any, tags []string) error {
	v := &value{to: val, tags: tags, allowSet: true}
	return v.Look(data, strings.Split(key, ".")...)
}

type value struct {
	value     reflect.Value
	to        any
	tags      []string
	anonymous map[int][]reflect.Type
	allowSet  bool
	allowAll  bool
}

func (v *value) Look(data any, paths ...string) error {
	if data == nil {
		return ErrValueNil
	}
	val, ok := data.(reflect.Value)
	if !ok {
		val = reflect.ValueOf(data)
	}
	v.allowAll = ok
	v.value = reflect.Indirect(val)
	if v.tags == nil {
		v.tags = DefaultValueGetSetTags
	}

	if v.allowSet && !v.value.CanSet() {
		return ErrValueNotSet
	}
	return v.lookValue(v.value, paths)
}

// Get the string path property from the target type.
//
//nolint:cyclop
func (v *value) lookValue(val reflect.Value, path []string) error {
	for len(path) > 0 && path[0] == "" {
		path = path[1:]
	}
	if len(path) == 0 {
		v.value = val
		if v.allowSet {
			if v.to == nil {
				v.value.Set(reflect.Zero(v.value.Type()))
				return nil
			}
			// return set error
			return setValuePtr(v.value, reflect.ValueOf(v.to))
		}
		return nil
	}

	if !v.allowSet {
		switch val.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice:
			if val.IsNil() {
				s := val.Type().String()
				return fmt.Errorf(ErrValueLookNil, val.Kind(), s, ErrValueNil)
			}
		}
	}
	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return v.lookNil(val, reflect.New(val.Type().Elem()), path)
		}
		return v.lookValue(val.Elem(), path)
	case reflect.Interface:
		return v.lookInterface(val, path)
	case reflect.Struct:
		return v.lookStruct(val, path)
	case reflect.Map:
		return v.lookMap(val, path)
	case reflect.Array, reflect.Slice:
		return v.lookSlice(val, path)
	default:
		return fmt.Errorf(ErrValueLookType, val.Type().String(), path[0], ErrValueNotFound)
	}
}

func (v *value) lookNil(val, next reflect.Value, path []string) error {
	val.Set(next)
	err := v.lookValue(val, path)
	if err != nil {
		val.Set(reflect.Zero(val.Type()))
	}
	return err
}

func (v *value) lookInterface(val reflect.Value, path []string) error {
	// If it is an empty interface, initialize it to map[string]any type
	if val.IsNil() {
		if val.Type() != typeAny {
			s := val.Type().String()
			return fmt.Errorf(ErrValueLookNil, reflect.Interface, s, ErrValueNil)
		}
		return v.lookNil(val, reflect.ValueOf(make(map[string]any)), path)
	}
	return v.lookValue(val.Elem(), path)
}

func (v *value) lookStruct(val reflect.Value, path []string) error {
	field := getStructFieldOfTags(val, path[0], v.tags)
	if field.Kind() == reflect.Invalid {
		iType := val.Type()
		for i := 0; i < iType.NumField(); i++ {
			if iType.Field(i).Anonymous && sliceIndex(v.anonymous[len(path)], iType.Field(i).Type) == -1 {
				if v.anonymous == nil {
					v.anonymous = make(map[int][]reflect.Type)
				}
				v.anonymous[len(path)] = append(v.anonymous[len(path)], iType.Field(i).Type)
				if v.lookValue(val.Field(i), path) == nil {
					return nil
				}
			}
		}

		return fmt.Errorf(ErrValueLookStruct, val.Type(), path[0], ErrValueStructNotField)
	}

	if field.CanInterface() {
		return v.lookValue(field, path[1:])
	}
	if v.allowAll {
		return v.lookValue(reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem(), path[1:])
	}
	return fmt.Errorf(ErrValueLookStruct, val.Type(), path[0], ErrValueStructUnexported)
}

// Get the index of the struct property through the string.
func getStructFieldOfTags(iValue reflect.Value, name string, tags []string) reflect.Value {
	iType := iValue.Type()
	for i := 0; i < iType.NumField(); i++ {
		typeField := iType.Field(i)
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

func (v *value) lookMap(val reflect.Value, path []string) error {
	iType := val.Type()
	if val.IsNil() {
		return v.lookNil(val, reflect.MakeMap(iType), path)
	}

	// Create the key used by the map
	mapKey := reflect.New(iType.Key()).Elem()
	err := setValueString(mapKey, path[0])
	if err != nil {
		return fmt.Errorf(ErrValueParseMapKey, path[0], err)
	}

	mapvalue := val.MapIndex(mapKey)
	if mapvalue.Kind() != reflect.Invalid {
		newValue := reflect.New(iType.Elem()).Elem()
		newValue.Set(mapvalue)

		return v.lookValue(newValue, path[1:])
	} else if !v.allowSet {
		return fmt.Errorf(ErrValueLookMap, val.Type().String(), path[0], ErrValueMapIndexInvalid)
	}

	// Reassign the modified mapvalue to the map
	mapvalue = reflect.New(iType.Elem()).Elem()
	err = v.lookValue(mapvalue, path[1:])
	if err == nil {
		val.SetMapIndex(mapKey, mapvalue)
	}
	return err
}

func (v *value) lookSlice(val reflect.Value, path []string) error {
	index, err := strconv.Atoi(path[0])
	switch {
	case (err != nil && path[0] != "[]"):
		return fmt.Errorf(ErrValueParseSliceIndex, path[0], val.Len(), err)
	case val.Len() < -index:
		return fmt.Errorf(ErrValueParseSliceIndex, path[0], val.Len(), ErrValueSliceIndexOutOfRange)
	case index < 0:
		index += val.Len()
	case path[0] == "[]":
		index = val.Len() + 1
	}

	// Check slice is empty
	iType := val.Type()
	if val.Kind() == reflect.Slice && val.IsNil() {
		return v.lookNil(val, reflect.MakeSlice(iType, index, index), path)
	}

	if index < val.Len() {
		return v.lookValue(val.Index(index), path[1:])
	} else if !v.allowSet {
		return fmt.Errorf(ErrValueLookSlice, iType.String(), path[0], val.Len(), ErrValueSliceIndexOutOfRange)
	}

	// Creates a new element's type and value
	newValue := reflect.New(iType.Elem()).Elem()
	err = v.lookValue(newValue, path[1:])
	if err != nil {
		return err
	}

	// Create a new array to replace the original array and expand the capacity
	if val.Cap() <= index {
		val.Set(reflect.AppendSlice(reflect.MakeSlice(iType, 0, index+1), val))
	}
	// Expand the array length and add null values to new elements
	if val.Len() <= index {
		val.SetLen(index + 1)
	}
	val.Index(index).Set(newValue)
	return nil
}

// The function obtains the full type and value of the dereference.
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

func setValuePtr(tValue, sValue reflect.Value) error {
	if sValue.Kind() == reflect.Ptr || sValue.Kind() == reflect.Interface ||
		tValue.Kind() == reflect.Ptr || tValue.Kind() == reflect.Interface {
		stypes, svalues := getIndirectAllValue(sValue)
		ttypes, tvalues := getIndirectAllValue(tValue)
		for i, ttype := range ttypes {
			for j, stype := range stypes {
				// Convert interface type, same type, type alias type
				if stype.ConvertibleTo(ttype) && tvalues[i].CanSet() {
					return setValueData(tvalues[i], svalues[j])
				}
			}
		}
		sValue = svalues[len(svalues)-1]
		tValue = tvalues[len(tvalues)-1]

		// If the target type is a null pointer, try to initialize and convert
		if tValue.Kind() == reflect.Ptr && tValue.IsNil() {
			newValue := reflect.New(tValue.Type().Elem())
			err := setValuePtr(newValue, sValue)
			if err == nil {
				tValue.Set(newValue)
			}
			return err
		}
	}
	return setValueData(tValue, sValue)
}

func setValueData(tValue, sValue reflect.Value) error {
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
		err := setValueData(newValue, sValue)
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
	return fmt.Errorf(ErrValueSetValuePtr, sValue.Type().String(), tValue.Type().String())
}

var bitSizes = [...]int{0, 0, 0, 8, 16, 32, 64, 0, 8, 16, 32, 64, 32, 64, 32, 64}

// Set the value of an object using a string.
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
	case reflect.Struct:
		if v.Type().ConvertibleTo(typeTimeTime) {
			err = setTimeField(v, s)
		} else {
			err = fmt.Errorf(ErrValueSetStringUnknownType, v.Type().String())
		}
	case reflect.Ptr:
		// only lookMap
		if v.IsNil() {
			newValue := reflect.New(v.Type().Elem())
			err = setValueString(newValue, s)
			if err == nil {
				v.Set(newValue)
			}
		} else {
			err = setValueString(v.Elem(), s)
		}
	case reflect.Interface:
		// only lookMap
		if v.Type() == typeAny {
			v.Set(reflect.ValueOf(s))
		} else {
			err = fmt.Errorf(ErrValueSetStringUnknownType, v.Type().String())
		}
	default:
		err = fmt.Errorf(ErrValueSetStringUnknownType, v.Type().String())
	}

	if err != nil {
		newValue := reflect.New(v.Type())
		e, ok := newValue.Interface().(encoding.TextUnmarshaler)
		if ok {
			err = e.UnmarshalText([]byte(s))
			if err == nil {
				v.Set(newValue.Elem())
			}
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
		field.SetUint(0)
		return nil
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
	if str == "" {
		field.SetComplex(complex(0, 0))
		return nil
	} else if str[0] == '(' && str[len(str)-1] == ')' {
		str = str[1 : len(str)-1]
	}
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
		field.SetFloat(0.0)
		return nil
	}
	floatVal, err := strconv.ParseFloat(str, bitSizes[int(field.Kind())])
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

// The TimeParse method parses the time format supported by the built-in time format.
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
