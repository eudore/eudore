package eudore_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/eudore/eudore"
)

func TestUtilContextKey(t *testing.T) {
	t.Log(eudore.NewContextKey("debug-key"))
}

func TestUtilTimeDuration(t *testing.T) {
	var data eudore.TimeDuration
	t.Log(json.Unmarshal([]byte(`"12s"`), &data), data)
	t.Log(json.Unmarshal([]byte(`12000000000`), &data), data)
	t.Log(json.Unmarshal([]byte(`"x"`), &data), data)
	b, _ := json.Marshal(data)
	t.Log(string(b))
	t.Log(data)
}

func TestUtilGetWarp(t *testing.T) {
	app := eudore.NewApp()
	eudore.NewGetWarp(func(key string) interface{} {
		return app.Get(key)
	})
	eudore.NewGetWarpWithApp(app).GetBool("key")
	eudore.NewGetWarpWithMapString(map[string]interface{}{
		"key": true,
	}).GetBool("key")
	eudore.NewGetWarpWithConfig(app.Config).GetBool("key")
	eudore.NewGetWarpWithObject(app).GetBool("key")

	data := map[string]interface{}{
		"int":       1,
		"int8":      int8(2),
		"float":     3.4,
		"string":    "warp string",
		"bytes":     []byte("warp bytes"),
		"arrayint":  []int{1, 2, 3},
		"arraystr":  []string{"a", "b", "c"},
		"arraybyte": []byte{'a', 'b', 'c'},
	}

	warp := eudore.NewGetWarpWithObject(data)
	t.Logf("%#v", warp.GetAny(""))
	t.Log(warp.GetInt("int"))
	t.Log(warp.GetInt64("int"))
	t.Log(warp.GetUint("int"))
	t.Log(warp.GetUint64("int"))
	t.Log(warp.GetFloat32("int"))
	t.Log(warp.GetFloat64("int"))
	t.Log(warp.GetString("int"))
}

func TestUtilGetAnyValue(t *testing.T) {
	t.Log(eudore.GetAnyDefault("default", ""))
	t.Log(eudore.GetAnyDefault("", ""))
	t.Log(eudore.GetAnyDefault("", "default string"))
	t.Log(eudore.GetAnyDefaults("default", ""))
	t.Log(eudore.GetAnyDefaults("", ""))
	t.Log(eudore.GetAnyDefaults("", "default string"))

	t.Log(eudore.GetAny("", "default string"))
	t.Log(eudore.GetAny[int](nil))
	t.Log(eudore.GetAny[int](12))
	t.Log(eudore.GetAny[int](uint(12)))
	t.Log(eudore.GetAny[int]("12"))
	t.Log(eudore.GetAny[string](12))
	t.Log(eudore.GetAny[int64](time.Second))

	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[string]("string")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[bool]("true")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[bool]("false")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[time.Time]("20180801")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[time.Duration]("200h")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[int]("12")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[int8]("12")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[int16]("12")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[int32]("12")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[int64]("12")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[uint]("12")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[uint8]("12")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[uint16]("12")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[uint32]("12")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[uint64]("12")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[float32]("12")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[float64]("12")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[complex64]("1+2i")))
	t.Log(eudore.GetStringByAny(eudore.GetAnyByString[complex128]("1+2i")))
	t.Log(eudore.GetStringByAny([]byte("bytes")))
	t.Log(eudore.GetStringByAny(eudore.GetStringByAny))
	t.Log(eudore.GetStringByAny(""))
	t.Log(eudore.GetStringByAny("", "0"))
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
		val, err := eudore.GetAnyByPathWithTag(i, key, nil, false)
		if err != nil {
			return err
		}
		return val
	}
	set := func(i any, key string, val any) error {
		return eudore.SetAnyByPath(i, key, val)
	}
	data := new(config)

	t.Log(eudore.SetAnyByPathWithTag(data, "ano", "ano field", nil, true))
	t.Log(eudore.GetAnyByPathWithTag(data, "ano", nil, true))
	t.Log(eudore.GetAnyByPathWithTag(data, "int", nil, true))
	t.Log(eudore.GetAnyByPathWithValue(data, "int", nil, true))
	t.Log(eudore.GetAnyByPath(data, "int"))
	// t.Log(eudore.GetAnyByPath(data, "ptr"))

	get(nil, "")
	t.Logf("%#v", get(data, ""))
	t.Log(get(data, "ptr.key"))
	t.Log(get(data, "name.num"))
	t.Log(get(data, "null"))
	t.Log(get(data, "int"))
	t.Log(get(data, "map.0"))
	t.Log(get(data, "slice.0"))
	t.Log(get(data, "index"))

	t.Log(set(data, "", 0))
	t.Log(set(*data, "name", 0))
	t.Log(set(data, "name.null", 0))
	t.Log(set(data, "ptr.null", 0))
	t.Log(set(data, "int", 0))
	t.Log(set(data, "context.4", 0))
	t.Log(set(data, "array.x", 0))
	t.Log(set(data, "slice.x", 0))
	t.Log(set(data, "map.xs", 0))
	t.Log(set(data, "index", "x"))
	t.Log(set(data, "index", 11))

	t.Log(set(data, "ptr.index", 12))
	t.Log(set(data, "array.0.index", 13))
	t.Log(set(data, "array.-1.index", 14))
	t.Log(set(data, "slice.5.index", 15))
	t.Log(set(data, "slice.[].index", 16))
	t.Log(set(data, "slice.-1.index", 17))
	t.Log(set(data, "any.8", 18))
	t.Log(set(data, "any.9", 19))
	t.Log(set(data, "map.9", "map9 hello"))
	t.Log(set(data, "map.9", "map9 hello"))

	t.Log(get(data, "map.xs"))
	t.Log(get(data, "map.0"))
	t.Log(get(data, "map.9"))
	t.Log(get(data, "array.x.index"))
	t.Log(get(data, "array.-1.index"))
	t.Log(get(data, "array.0.index"))
	t.Log(get(data, "index"))
	t.Logf("%#v", get(data, ""))
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
	eudore.SetAnyByPath(data, "ptr", t)
	eudore.SetAnyByPath(data, "ptr", time.Second)
	t.Logf("%p", data.Ptr)
	d := eudore.TimeDuration(time.Second)
	eudore.SetAnyByPath(data, "ptr", &d)
	t.Logf("%p %s", data.Ptr, d)
	eudore.SetAnyByPath(data, "ptr", "12x")
	eudore.SetAnyByPath(data, "ptr", "12s")
	t.Logf("%p %s", data.Ptr, d)

	eudore.SetAnyByPath(data, "slice", "12s")
	eudore.SetAnyByPath(data, "slice", "12")
	eudore.SetAnyByPath(data, "slice", []string{"1", "2", "3"})
	eudore.SetAnyByPath(data, "slice", []string{"a", "x", "c"})

	eudore.SetAnyByPath(data, "int", "")
	eudore.SetAnyByPath(data, "uint", "")
	eudore.SetAnyByPath(data, "bool", "")
	eudore.SetAnyByPath(data, "float", "")
	eudore.SetAnyByPath(data, "complex", "")
	eudore.SetAnyByPath(data, "complex", "0+x")
	eudore.SetAnyByPath(data, "complex", "0i+x")
	t.Log(eudore.SetAnyByPath(data, "time", "2018"))
	t.Log(eudore.SetAnyByPath(data, "chan", "2018"))

	t.Log(eudore.SetAnyByPath(data, "int", "1"))
	t.Log(eudore.SetAnyByPath(data, "uint", "1"))
	t.Log(eudore.SetAnyByPath(data, "bool", "1"))
	t.Log(eudore.SetAnyByPath(data, "float", "1"))
	t.Log(eudore.SetAnyByPath(data, "complex", "1i"))
	t.Log(eudore.SetAnyByPath(data, "time", "20180801"))
	t.Log(eudore.SetAnyByPath(data, "time2", "20180801"))
	t.Log(eudore.SetAnyByPath(data, "bytes", "bytes"))
	t.Log(eudore.SetAnyByPath(data, "runes", "runes"))
	t.Log(eudore.SetAnyByPath(data, "any", "any"))
	t.Log(eudore.SetAnyByPath(data, "face", "any"))
	t.Log(eudore.SetAnyByPath(data, "struct", "struct"))
	t.Log(eudore.SetAnyByPathWithTag(data, "ano", time.Now(), nil, true))
	t.Logf("%#v", eudore.GetAnyByPath(data, ""))

	type M struct {
		M1 map[string]any              `alias:"m1"`
		M2 map[*string]any             `alias:"m2"`
		M3 map[eudore.LoggerLevel]any  `alias:"m3"`
		M4 map[eudore.TimeDuration]any `alias:"m4"`
		M5 map[any]any                 `alias:"m5"`
	}

	m := &M{}
	t.Log(eudore.SetAnyByPath(m, "m1.1", "1"))
	t.Log(eudore.SetAnyByPath(m, "m2.3", "1"))
	t.Log(eudore.SetAnyByPath(m, "m3.ERROR", "1"))
	t.Log(eudore.SetAnyByPath(m, "m4.4s", "1"))
	t.Log(eudore.SetAnyByPath(m, "m5.5", "1"))
	t.Logf("%#v", m)

	type Cycle struct {
		*Cycle
	}
	c := &Cycle{}
	c.Cycle = c
	t.Log(eudore.SetAnyByPathWithTag(c, "name", "eudore", nil, false))
	t.Log(eudore.GetAnyByPathWithTag(c, "name", nil, false))
}
