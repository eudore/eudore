package eudore_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	. "github.com/eudore/eudore"
)

func TestUtilContextKey(*testing.T) {
	fmt.Sprint(
		NewContextKey("debug-key"),
		TimeDuration(0),
		GetStringRandom(32),
		GetStringDuration(0),
		GetStringDuration(time.Second),
		GetStringByAny(GetStringByAny),
	)
}

func TestUtilError(t *testing.T) {
	errs := []error{
		NewErrorWithStatusCode(nil, 500, 1004),
		NewErrorWithStatusCode(context.Canceled, 500, 1004),
		NewErrorWithStatus(nil, 500),
		NewErrorWithStatus(context.Canceled, 0),
		NewErrorWithStatus(context.Canceled, 500),
		NewErrorWithCode(nil, 1004),
		NewErrorWithCode(context.Canceled, 0),
		NewErrorWithCode(context.Canceled, 1004),
		NewErrorWithWrapped(context.Canceled, "msg"),
		NewErrorWithStack(context.Canceled, nil),
		NewErrorWithDepth(context.Canceled, 0),
		NewRouter(nil).AddHandlerExtend(1, 2, 3),
	}
	for _, err := range errs {
		if err == nil {
			continue
		}
		err.Error()
		e1, ok := err.(interface{ Unwrap() error })
		if ok {
			e1.Unwrap()
		}
		e2, ok := err.(interface{ Unwrap() []error })
		if ok {
			e2.Unwrap()
		}
		e3, ok := err.(interface{ Status() int })
		if ok {
			e3.Status()
		}
		e4, ok := err.(interface{ Code() int })
		if ok {
			e4.Code()
		}
		e5, ok := err.(interface{ Stack() []string })
		if ok {
			e5.Stack()
		}
	}
}

func TestUtilTimeDuration(t *testing.T) {
	datas := []struct {
		data string
		time TimeDuration
		err  string
	}{
		{`"12s"`, TimeDuration(12000000000), ""},
		{`12000000000`, TimeDuration(12000000000), ""},
		{`"x"`, 0, "invalid duration value: 'x'"},
	}
	for i := range datas {
		var v TimeDuration
		err := json.Unmarshal([]byte(datas[i].data), &v)
		if (err != nil && err.Error() != datas[i].err) || v != datas[i].time {
			t.Error(datas[i], err)
		}
		v.MarshalText()
	}
}

func TestUtilGetWrap(t *testing.T) {
	app := NewApp()
	NewGetWrapWithApp(app).GetAny("")
	NewGetWrapWithMapString(map[string]any{"key": true}).GetAny("")
	NewGetWrapWithConfig(app.Config)
	NewGetWrapWithObject(app)

	w := NewGetWrapWithObject(map[string]any{"int": 1})
	vals := []any{
		w.GetAny("int"), 1,
		w.GetBool("int"), true,
		w.GetInt("int"), 1,
		w.GetInt64("int"), int64(1),
		w.GetUint("int"), uint(1),
		w.GetUint64("int"), uint64(1),
		w.GetFloat32("int"), float32(1),
		w.GetFloat64("int"), float64(1),
		w.GetString("int"), "1",
	}
	for i := 0; i < len(vals); i += 2 {
		if vals[i] != vals[i+1] {
			t.Error(i, vals[i])
		}
	}
}

func TestUtilGetAnyValue(t *testing.T) {
	DefaultValueTimeLocation = time.UTC
	defer func() {
		DefaultValueTimeLocation = time.Local
	}()
	vals := []any{
		GetAnyDefault("default", ""), "default",
		GetAnyDefault("", "default"), "default",
		GetAnyDefault("", ""), "",
		GetAnyDefaults("default", ""), "default",
		GetAnyDefaults("", "default"), "default",
		GetAnyDefaults("", ""), "",
		GetAny("", "default"), "default",

		GetAny[int](nil), 0,
		GetAny[int](12), 12,
		GetAny[int](uint(12)), 12,
		GetAny[int]("12"), 12,
		GetAny[string](12), "12",
		GetAny[int64](time.Second), int64(1000000000),

		GetAny[bool](true), true,
		GetAny[bool]("true"), true,
		GetAny[bool](uint(1)), true,
		GetAny[bool](float32(2)), true,
		GetAny[bool](float64(4)), true,
		GetAny[bool]([]any{8}), true,
		GetAny[bool](t), false,

		GetStringByAny(GetAnyByString[string]("string")), "string",
		GetStringByAny(GetAnyByString[bool]("true")), "true",
		GetStringByAny(GetAnyByString[bool]("false")), "false",
		GetStringByAny(GetAnyByString[time.Time]("20180801")), "2018-08-01 00:00:00 +0000 UTC",
		GetStringByAny(GetAnyByString[time.Duration]("200h")), "200h0m0s",
		GetStringByAny(GetAnyByString[int]("12")), "12",
		GetStringByAny(GetAnyByString[int8]("12")), "12",
		GetStringByAny(GetAnyByString[int16]("12")), "12",
		GetStringByAny(GetAnyByString[int32]("12")), "12",
		GetStringByAny(GetAnyByString[int64]("12")), "12",
		GetStringByAny(GetAnyByString[uint]("12")), "12",
		GetStringByAny(GetAnyByString[uint8]("12")), "12",
		GetStringByAny(GetAnyByString[uint16]("12")), "12",
		GetStringByAny(GetAnyByString[uint32]("12")), "12",
		GetStringByAny(GetAnyByString[uint64]("12")), "12",
		GetStringByAny(GetAnyByString[float32]("12")), "12",
		GetStringByAny(GetAnyByString[float64]("12")), "12",
		GetStringByAny(GetAnyByString[complex64]("1+2i")), "(1+2i)",
		GetStringByAny(GetAnyByString[complex128]("1+2i")), "(1+2i)",
		GetStringByAny([]byte("bytes")), "bytes",
		GetStringByAny(""), "",
		GetStringByAny("", "0"), "0",
	}

	for i := 0; i < len(vals); i += 2 {
		if vals[i] != vals[i+1] {
			t.Error(i, vals[i])
		}
	}
}

func TestUtilSetValue(t *testing.T) {
	DefaultValueTimeLocation = time.UTC
	defer func() {
		DefaultValueTimeLocation = time.Local
	}()
	type time2 time.Time
	type C struct {
		name string
	}
	type S struct {
		Int       int
		Uint      uint
		Bool      bool
		Complex   complex64
		Float     float64
		Duration  time.Duration
		Duration2 TimeDuration
		Time      time.Time
		Time2     time2
		String    string
		Struct    struct {
			*C
			anonymou int
		} `alias:"struct"`
		Func    func()
		Ptr     *string
		Any     any
		Context context.Context
		Slice   []int

		MapS map[*string]string
		MapE map[any]string
		MapI map[fmt.Stringer]string
	}

	vals := []any{
		"Int", "", "0",
		"Int", "1", "1",
		"Int", "x", "strconv.ParseInt: parsing \"x\": invalid syntax",
		"Uint", "", "0",
		"Uint", "1", "1",
		"Uint", "x", "strconv.ParseUint: parsing \"x\": invalid syntax",
		"Bool", "", "true",
		"Bool", "1", "true",
		"Bool", "x", "strconv.ParseBool: parsing \"x\": invalid syntax",
		"Complex", "", "(0+0i)",
		"Complex", "(1+1)", "(1+1i)",
		"Complex", "1", "(1+0i)",
		"Complex", "0+x", "strconv.ParseFloat: parsing \"x\": invalid syntax",
		"Complex", "x+0", "strconv.ParseFloat: parsing \"x\": invalid syntax",
		"Float", "", "0",
		"Float", "1", "1",
		"Float", "x", "strconv.ParseFloat: parsing \"x\": invalid syntax",
		"Duration", "", "0s",
		"Duration", "100ns", "100ns",
		"Duration", "x", "time: invalid duration \"x\"",
		"Duration2", "", "0s",
		"Duration2", "100ns", "100ns",
		"Time", "20060102", "2006-01-02 00:00:00 +0000 UTC",
		"Time", "20060102x", "parsing time \"20060102x\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"0102x\" as \"-\"",
		"Time2", "20060102", "{0 63271756800 <nil>}",
		"String", "String", "String",
		"Struct", "String", "the SetValueString unknown type struct { *eudore_test.C; anonymou int }",
		"Func", "String", "the SetValueString unknown type func()",

		"Int", 1, "1",
		"Int", uint(1), "1",
		"Slice", "1", "[1]",
		"Slice", "x", "strconv.ParseInt: parsing \"x\": invalid syntax",
		"String", true, "true",
		"Int", true, "the SetValuePtr method type bool cannot be assigned to type int",
		"Ptr.ptr", "String", "look value type string path 'ptr': value not found",
		"Ptr", "String", "String",
		"Ptr", &DefaultGodocServer, DefaultGodocServer,
		"Any", &DefaultGodocServer, DefaultGodocServer,

		"MapS.str", "String", "",
		"MapE.str", "String", "String",
		"MapI.str", "String", "parse map key 'str' error: the SetValueString unknown type fmt.Stringer",

		"Any", nil, "",
		"Any.1", "1", "1",
		"Any.2", "2", "2",
		"Context.3", "3", "look type interface value context.Context: value is nil",
		"Context", context.Background(), "context.Background",
		"struct.anonymou", "1", "look struct type struct { *eudore_test.C; anonymou int } field 'anonymou': struct field unexported",
		"struct.notfound", "1", "look struct type struct { *eudore_test.C; anonymou int } field 'notfound': struct field not found",
		"struct.name", "name", "look struct type struct { *eudore_test.C; anonymou int } field 'name': struct field not found",
		"Slice", nil, "[]",
		"Slice.+", "1", "parse slice index '+', len is 0 error: strconv.Atoi: parsing \"+\": invalid syntax",
		"Slice.-1", "2", "parse slice index '-1', len is 0 error: slice index out of range",
		"Slice.[]", "3", "",
		"Slice.-1", "4", "4",
		"Slice.4", "xx", "strconv.ParseInt: parsing \"xx\": invalid syntax",
	}

	s := &S{}
	GetAnyByPath(nil, "Any", nil)
	SetAnyByPath(nil, "Any", nil, nil)
	SetAnyByPath(S{}, "Any", nil, nil)
	GetAnyByPath(s, "Any.x", nil)
	SetAnyByPath(reflect.ValueOf(s), "struct.name", "name", nil)
	GetAnyByPath(reflect.ValueOf(s), "struct.name", nil)
	for i := 0; i < len(vals); i += 3 {
		err := SetAnyByPath(s, vals[i].(string), vals[i+1], nil)
		if err == nil {
			var str string
			val, err := GetValueByPath(s, vals[i].(string), nil)
			val = reflect.Indirect(val)
			for val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			if err == nil && val.Kind() != reflect.Invalid {
				str = fmt.Sprint(val.Interface())
			}
			if str != vals[i+2].(string) {
				t.Log(i/3, vals[i], str, "!=", vals[i+2].(string))
			}
		} else {
			if err.Error() != vals[i+2].(string) {
				t.Log(i/3, vals[i], err)
			}
		}
	}
}
