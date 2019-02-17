package cast

import (
	"fmt"
	"strconv"
)


func GetInt(i interface{}) int {
	return GetDefultInt(i, 0)
}

func GetDefultInt(i interface{},n int) int {
	if v, ok := i.(int); ok {
		return v
	}	
	if v, err := strconv.Atoi(GetDefaultString(i, "")); err == nil {
		return v
	}
	return n
}

func GetString(i interface{}) string {
	return GetDefaultString(i, "")
}

func GetDefaultString(i interface{},str string) string {
	if v, ok := i.(string); ok {
		return v
	}
	if v, ok := i.(fmt.Stringer); ok {
		return v.String()
	}
	return str
}

func GetDefaultFloat64(i interface{}, f float64) float64 {
	if v, ok := i.(float64); ok {
		return v
	}
	return f
}


func GetDefaultInt64(i interface{}, n int64) int64 {
	if v, ok := i.(int64); ok {
		return v
	}
	if v, ok := i.(float64); ok {
		return int64(v)
	}
	return n
}