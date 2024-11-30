package eudore_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	. "github.com/eudore/eudore"
)

func TestUtilContextKey(t *testing.T) {
	t.Log(
		NewContextKey("debug-key"),
		TimeDuration(0),
		GetStringRandom(32),
		GetStringDuration(0),
		GetStringDuration(time.Second),
		GetStringByAny(GetStringByAny),
		NewRouter(nil).AddHandlerExtend(1, 2, 3),
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
	}
	for _, err := range errs {
		if err == nil {
			continue
		}
		err.Error()
		u, ok := err.(interface{ Unwrap() error })
		if ok {
			u.Unwrap()
		}
		s, ok := err.(interface{ Status() int })
		if ok {
			s.Status()
		}
		c, ok := err.(interface{ Code() int })
		if ok {
			c.Code()
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
	w.GetAny("")
	w.GetBool("int")
	w.GetInt("int")
	w.GetInt64("int")
	w.GetUint("int")
	w.GetUint64("int")
	w.GetFloat32("int")
	w.GetFloat64("int")
	w.GetString("int")
}

func TestUtilGetAnyValue(t *testing.T) {
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

func TestUtilGetSetValue(t *testing.T) {
	type Field struct {
		Index int    `alias:"index"`
		Name  string `alias:"name"`
	}
	type config struct {
		Name    string `alias:"name"`
		int     int    `alias:"int`
		ano     string
		Ptr     *Field          `alias:"ptr"`
		Array   [4]Field        `alias:"array"`
		Slice   []Field         `alias:"slice"`
		Map     map[int]string  `alias:"map"`
		Any     any             `alias:"any"`
		Context context.Context `alias:"context"`
		*Field
	}

	get := func(i any, key string) any {
		val, err := GetAnyByPathWithTag(i, key, nil, false)
		if err != nil {
			return err
		}
		return val
	}
	set := func(i any, key string, val any) error {
		return SetAnyByPath(i, key, val)
	}
	data := new(config)

	SetAnyByPathWithTag(data, "ano", "ano field", nil, true)
	GetAnyByPathWithTag(data, "ano", nil, true)
	GetAnyByPathWithTag(data, "int", nil, true)
	GetAnyByPathWithValue(data, "int", nil, true)
	GetAnyByPath(data, "int")

	get(nil, "")
	get(data, "ptr.key")
	get(data, "name.num")
	get(data, "null")
	get(data, "int")
	get(data, "map.0")
	get(data, "slice.0")
	get(data, "index")

	set(data, "", 0)
	set(*data, "name", 0)
	set(data, "name.null", 0)
	set(data, "ptr.null", 0)
	set(data, "int", 0)
	set(data, "context.4", 0)
	set(data, "array.x", 0)
	set(data, "slice.x", 0)
	set(data, "map.xs", 0)
	set(data, "index", "x")
	set(data, "index", 11)

	set(data, "ptr.index", 12)
	set(data, "array.0.index", 13)
	set(data, "array.-1.index", 14)
	set(data, "slice.5.index", 15)
	set(data, "slice.[].index", 16)
	set(data, "slice.-1.index", 17)
	set(data, "any.8", 18)
	set(data, "any.9", 19)
	set(data, "map.9", "map9 hello")
	set(data, "map.9", "map9 hello")

	get(data, "map.xs")
	get(data, "map.0")
	get(data, "map.9")
	get(data, "array.x.index")
	get(data, "array.-1.index")
	get(data, "array.0.index")
	get(data, "index")
}

func TestUtilSetWithValue(t *testing.T) {
	type time2 time.Time
	type config struct {
		Ptr     *time.Duration `alias:"ptr"`
		Slice   []int          `alias:"slice"`
		Int     int            `alias:"int"`
		Uint    uint           `alias:"uint"`
		Bool    bool           `alias:"bool"`
		Float   float64        `alias:"float"`
		Complex complex64      `alias:"complex"`
		Time    time.Time      `alias:"time"`
		Time2   time2          `alias:"time2"`
		Struct  struct{}       `alias:"struct"`
		Bytes   []byte         `alias:"bytes"`
		Runes   []rune         `alias:"runes"`
		Any     any            `alias:"any"`
		Face    json.Marshaler `alias:"face"`
		Chan    chan int       `alias:"chan"`
		ano     string
	}

	data := new(config)
	SetAnyByPath(data, "ptr", t)
	SetAnyByPath(data, "ptr", time.Second)
	d := TimeDuration(time.Second)
	SetAnyByPath(data, "ptr", &d)
	SetAnyByPath(data, "ptr", "12x")
	SetAnyByPath(data, "ptr", "12s")

	SetAnyByPath(data, "slice", "12s")
	SetAnyByPath(data, "slice", "12")
	SetAnyByPath(data, "slice", []string{"1", "2", "3"})
	SetAnyByPath(data, "slice", []string{"a", "x", "c"})

	SetAnyByPath(data, "int", "")
	SetAnyByPath(data, "uint", "")
	SetAnyByPath(data, "bool", "")
	SetAnyByPath(data, "float", "")
	SetAnyByPath(data, "complex", "")
	SetAnyByPath(data, "complex", "0+x")
	SetAnyByPath(data, "complex", "0i+x")
	SetAnyByPath(data, "time", "2018")
	SetAnyByPath(data, "chan", "2018")
	SetAnyByPath(data, "int", "1")
	SetAnyByPath(data, "uint", "1")
	SetAnyByPath(data, "bool", "1")
	SetAnyByPath(data, "float", "1")
	SetAnyByPath(data, "complex", "1i")
	SetAnyByPath(data, "time", "20180801")
	SetAnyByPath(data, "time2", "20180801")
	SetAnyByPath(data, "bytes", "bytes")
	SetAnyByPath(data, "runes", "runes")
	SetAnyByPath(data, "any", "any")
	SetAnyByPath(data, "face", "any")
	SetAnyByPath(data, "struct", "struct")
	SetAnyByPathWithTag(data, "ano", time.Now(), nil, true)

	type M struct {
		M1 map[string]any       `alias:"m1"`
		M2 map[*string]any      `alias:"m2"`
		M3 map[LoggerLevel]any  `alias:"m3"`
		M4 map[TimeDuration]any `alias:"m4"`
		M5 map[any]any          `alias:"m5"`
	}

	m := &M{}
	SetAnyByPath(m, "m1.1", "1")
	SetAnyByPath(m, "m2.3", "1")
	SetAnyByPath(m, "m3.ERROR", "1")
	SetAnyByPath(m, "m4.4s", "1")
	SetAnyByPath(m, "m5.5", "1")

	type Cycle struct {
		*Cycle
	}
	c := &Cycle{}
	c.Cycle = c
	SetAnyByPathWithTag(c, "name", "eudore", nil, false)
	GetAnyByPathWithTag(c, "name", nil, false)
}
