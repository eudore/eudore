package config

import (
	"fmt"
	"strings"
	"reflect"
)

// Use sep to split str into two strings.
func split2(str string,sep string) (string, string) {
	pos := strings.IndexByte(str, sep[0])
	if pos == -1 {
		return "", ""
	}
	return str[:pos], str[pos + len(sep):]
}


func AllKeyVal(p interface{}) ([]string, []interface{}, error) {
	pt := reflect.TypeOf(p)
	pv := reflect.ValueOf(p)
	for pv.Kind() == reflect.Ptr || pv.Kind() == reflect.Interface {
		pt = pt.Elem()
		pv = pv.Elem()
	}
	if pv.Kind() == reflect.Struct {
		// get p name and value
		num := pv.Type().NumField()
		names := make([]string, num)
		values := make([]interface{}, num)
		for i := 0; i < num; i++ {
			names[i] = pt.Field(i).Name
			values[i] = pv.Field(i).Interface()
		}
		return names, values, nil
	}
	if pv.Kind() == reflect.Map {
		num := pv.Len()
		names := make([]string, num)
		values := make([]interface{}, num)
		for i, key := range pv.MapKeys() {
			fmt.Println(key)
			names[i] = key.String()
			values[i] = pv.MapIndex(key).Interface()
		}
		return names, values, nil
	}
	return nil, nil, fmt.Errorf("Type is %s, unable to read key-value data.", pv.Kind().String())
}