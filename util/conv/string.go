package conv

import (
	"strconv"
)

type String string


func (s String) GetInt() int {
	return s.GetDefultInt(0)
}

func (s String) GetDefultInt(n int) int {
	if v, err := strconv.Atoi(s.GetString()); err == nil {
		return v
	}
	return n
}

func (s String) GetString() string {
	return string(s)
}

func (s String) GetDefaultFloat64(f float64) float64 {
	if v, err := strconv.ParseFloat(s.GetString(), 64); err == nil {
		return v
	}
	return f
}


func (s String)  GetDefaultInt64(n int64) int64 {
	if v, err := strconv.ParseInt(s.GetString(), 10, 64); err == nil {
		return v
	}
	return n
}