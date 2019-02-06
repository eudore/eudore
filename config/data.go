package config

import (
	// "os"
	"io"
	"fmt"
	"bytes"
	"strconv"
	"strings"
	"reflect"
	"encoding/json"
	"eudore/util/calculate"
	"gopkg.in/yaml.v2"
)

// Test function, serialized output in Json format.
//
// 测试函数，Json格式序列化输出。
func Json(args ...interface{}) {
	for _, i := range args {
		indent, err := json.MarshalIndent(&i, "", "\t")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(indent))
	}
}

// Output a help message for the interface{} object, and each structure member uses the description tag as the description.
//
// 输出一个interface{}对象的帮助信息，每个结构体成员使用description tag为描述信息。
//
// Help(c, "  --", os.Stdio)
func Help(p interface{}, prefix string, w io.Writer) error {
	getDesc(p, prefix, w)
	return nil
}

func getDesc(p interface{}, prefix string, w io.Writer) {
	pt := reflect.TypeOf(p).Elem()
	pv := reflect.ValueOf(p).Elem()
	for i := 0; i < pt.NumField(); i++ {
		sv := pv.Field(i)
		if c :=  pt.Field(i).Tag.Get("description");c != "" {
			fmt.Printf("%s%s\t%s\n",prefix, strings.ToLower(pt.Field(i).Name), c)
		}
		if c := pt.Field(i).Tag.Get("help");c == "-" {
			continue
		}
		if sv.Type().Kind() == reflect.Ptr {
			// is null
			if sv.Elem().Kind() == reflect.Invalid {
				sv.Set(reflect.New(sv.Type().Elem()))	
			}
			sv=sv.Elem()
		}
		if sv.Kind() == reflect.Struct {
			getDesc(sv.Addr().Interface(),fmt.Sprintf("%s%s.",prefix,strings.ToLower(pt.Field(i).Name)), w)
		}
	}
}

// Set the data of the specified path of interface{} to be the value resolution object.
// The current function is not implemented, the data is converted to yaml format and then deserialized.
//
// 设置一个interface{}指定路径的数据为value解析对象。
// 当前函数未实现，会将数据转换成yaml格式然后反序列化。
func SetData(p interface{}, key, value string) error {
	var data bytes.Buffer
	keys := strings.Split(strings.ToLower(key), ".")
	end := len(keys) - 1
	for i, c := range keys {
		data.WriteString(strings.Repeat(" ", i * 2))
		data.WriteString(c)
		data.WriteString(": ")
		if i != end {
			data.WriteString("\n")
		}
	}
	data.WriteString(value)
	return yaml.Unmarshal(data.Bytes(), p)
}

// Read the data of an interface{} execution path and return the interface{} type.
// If the argument is an array, you can use the mathematical expression starting with '#' to represent the index,
// the built-in length variable "len", and the negative index refers to the reverse order.
// 
// 读取一个interface{}执行路径的数据，返回interface{}类型。
// 如果参数是数组，可以使用'#'开头的数学表达式表示索引，内置长度变量“len”，负数索引指倒序。
//
// config.GetData(data, "num.#len-2*1.name")
func GetData(p interface{}, key string) (interface{}, error) {
	return getdata(p, strings.Split(key, "."))
}

func getdata(p interface{}, keys []string) (interface{}, error) {
	len := len(keys) - 1
	for i, key := range keys {
		pv := reflect.ValueOf(p)
		for pv.Kind() == reflect.Ptr || pv.Kind() == reflect.Interface {
			pv = pv.Elem()
		}
		// get attr
		switch pv.Kind() {
		case reflect.Struct:
			f, ok := pv.Type().FieldByName(strings.Title(key))
			if !ok {
				// error
				return pv.Interface(), fmt.Errorf("struct not found attr.")
			}
			pv = pv.Field(f.Index[0])
		case reflect.Array, reflect.Slice:
			index, err := getindex(key, pv.Len())
			if err != nil {
				return p, err
			}
			pv = pv.Index(index)
		case reflect.Map:
			k, err := getTypeValue(pv.Type().Key(), key)
			if err != nil {
				return p, err
			}
			pv = pv.MapIndex(k)
			if !pv.IsValid() {
				return p, fmt.Errorf("key %s is not found.", key)
			}
		default:
			return pv.Interface(), fmt.Errorf("type error %v,not found %s", pv.Kind(), key)
		}
		// is last element
		if i == len {
			return pv.Interface(), nil
		}
		// is null
		switch pv.Kind() {
		case reflect.Ptr:
			if pv.IsNil() {
				fmt.Println("+new struct ptr")
				pv.Set(reflect.New(pv.Type().Elem()))	
			}
			pv = pv.Elem()
		case reflect.Map:
			if pv.IsNil() {
				fmt.Println("+new map")
				pv.Set(reflect.MakeMap(pv.Type()))
			}
		case reflect.Struct, reflect.Array, reflect.Slice:
		case reflect.Invalid:
			return pv, errType
		default:
			fmt.Println("--------",key, pv.Kind())
		}
		// next
		if pv.CanAddr() {
			p = pv.Addr().Interface()	
		}else {
			p = pv.Interface()
		}
	}
	return p, nil
}

//
func getindex(key string, length int) (i int ,err error) {
	if len(key) == 0 || length == 0 {
		return -1, fmt.Errorf("key is nil or len is zero.")
	}
	if key[0] == 35 {
		key = strings.Replace(key[1:], "len", fmt.Sprint(length), -1)
		i, err = calculate.Result(key)
	}else {
		i, err = strconv.Atoi(key)	
	}
	
	if i < 0 {
		i = i + length
	}
	if i >= length {
		err = fmt.Errorf("slice index out of range,len is %d, current index is %d.", length, i)
	}
	return i, err
}

func getTypeValue(t reflect.Type,s string) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Bool:
		if s == "" {
			return reflect.ValueOf(true), nil
		}else {
			rb,_ := strconv.ParseBool(s)
			return reflect.ValueOf(rb), nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return reflect.Zero(t), errType
		}
		return reflect.ValueOf(n).Convert(t),nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return reflect.Zero(t), errType
		}
		return reflect.ValueOf(n).Convert(t),nil
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, t.Bits())
		if err != nil {
			return reflect.Zero(t), errType
		}
		return reflect.ValueOf(n).Convert(t),nil
	case reflect.String:
		return reflect.ValueOf(s),nil
	}
	return reflect.Zero(t), errType
}