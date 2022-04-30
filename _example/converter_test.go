package eudore_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/eudore/eudore"
	"github.com/kr/pretty"
)

type (
	config017 struct {
		InterfaceType context.Context
		Ptr           *config017
		InterfaceNone interface{}
		MapString     map[string]*Param017
		MapInt        map[int]string
		Struct        Param017
		Array         [4]string
		Slice         []string
		Chan          chan int `json:"-"`
	}
	Param017 struct {
		Key   string
		Value string
		none  string
	}
	config018 struct {
		InterfaceType context.Context
		Ptr           *config018
		InterfaceNone interface{}
		MapString     map[string]*Param018
		MapInt        map[int]string
		Struct        Param018
		Array         [4]string
		Slice         []string
		Chan          chan int `json:"-"`
	}
	Param018 struct {
		Key  string
		none string
	}
)

func TestConvertValueSet(t *testing.T) {
	config := &config017{}
	// setValue
	eudore.Set(config, "Chan.1", "value")
	// setInterface
	eudore.Set(config, "InterfaceType.Key", "value")
	eudore.Set(config, "InterfaceNone.Key", "value")
	// setStruct
	eudore.Set(config, "Struct.Key", "value")
	eudore.Set(config, "Struct.none", "value")
	eudore.Set(config, "Struct.no", "value")
	// setMap
	eudore.Set(config, "MapString.Key.Key", "value1")
	eudore.Set(config, "MapString.Key.Key", "value2")
	// setArray
	eudore.Set(config, "Array.0", "000")
	eudore.Set(config, "Array.4", "4")
	// setSlice
	eudore.Set(config, "Slice.0", "000")
	eudore.Set(config, "Slice.4", "4")
	eudore.Set(config, "Slice.x", "x")
	// error
	eudore.Set(nil, "Slice.x", "x")
	eudore.Set(TestConvertValueSet, "Slice.x", "x")

	body, err := json.Marshal(config)
	t.Log(string(body), err)
	if string(body) != `{"InterfaceType":null,"Ptr":null,"InterfaceNone":{"Key":"value"},"MapString":{"Key":{"Key":"value2","Value":""}},"MapInt":null,"Struct":{"Key":"value","Value":""},"Array":["000","","",""],"Slice":["000","","","","4","x"]}` {
		panic("check result")
	}
}

func TestConvertValueGet(t *testing.T) {
	config := &config017{}
	// getValue
	eudore.Get(config, "InterfaceNone.Key")
	eudore.Get(config, "Chan.1")
	// getStruct
	config.Struct = Param017{
		Key:  "value",
		none: "value",
	}
	eudore.Get(config, "Struct.Key")
	eudore.Get(config, "Struct.none")
	eudore.Get(config, "Struct.no")
	// getMap
	eudore.Get(config, "MapString.Key")
	config.MapString = map[string]*Param017{
		"Key": {Key: "String"},
	}
	eudore.Get(config, "MapString.Key")
	eudore.Get(config, "MapString.Key2")
	config.MapInt = map[int]string{
		1: "int",
	}
	eudore.Get(config, "MapInt.Key")
	// getSlice
	eudore.Get(config, "Slice.x")
	config.Slice = []string{"000", "", "", "", "4"}
	eudore.Get(config, "Slice.0")
	eudore.Get(config, "Slice.4")
	eudore.Get(config, "Slice.x")

	// error
	eudore.Get(nil, "Slice.x")
	eudore.GetWithTags(config, "Slice.0", []string{"alias"}, false)
}

func TestConvertMappingMapString(t *testing.T) {
	config := &config017{}
	config.InterfaceNone = config
	config.MapString = map[string]*Param017{
		"Key": {Key: "String"},
		"nil": nil,
	}
	eudore.ConvertMapString(config)
	eudore.ConvertMap(config)

	var data map[string]interface{}
	eudore.ConvertTo(config, &data)

	config2 := &config017{}
	eudore.ConvertTo(&data, config2)
}

func TestConvertMappingTo(t *testing.T) {
	config := &config017{}
	config.Ptr = config
	config.InterfaceNone = config
	config.MapString = map[string]*Param017{
		"Key": {Key: "String"},
		"nil": nil,
	}
	config.MapInt = map[int]string{
		1: "A",
		2: "B",
	}
	config.Struct = Param017{Key: "name", Value: "018"}
	config.Array = [4]string{"str1", "str2", "str3", ""}
	config.Slice = []string{"Slice1", "Slice2", "Slice3", ""}
	config.InterfaceType = context.WithValue(context.Background(), "key", "name")

	config2 := &config017{}
	conv(config, config2)

	var config3 map[string]interface{}
	conv(config, &config3)

	var config4 map[string]interface{}
	conv(&config3, &config4)

	var config5 map[string]interface{}
	config5 = make(map[string]interface{})
	conv(config3, &config5)

	config6 := &config017{}
	config3["none"] = "none"
	conv(&config3, config6)

	config7 := &config018{}
	config7.InterfaceType = context.Background()
	conv(config, config7)

	eudore.ConvertTo(config, config4)
	eudore.ConvertTo(config, nil)
	eudore.ConvertTo(nil, config4)
}

func conv(a, b interface{}) {
	eudore.ConvertTo(a, b)
	// 1.13 not json
	// return
	fmt.Printf("%# v\n", a)
	body, err := json.Marshal(a)
	fmt.Printf("%s %v\n", body, err)

	fmt.Printf("%# v\n", b)
	body, err = json.Marshal(b)
	fmt.Printf("%s %v\n", body, err)
	fmt.Println("------------------")
}

type BB struct {
	BB1 int
	BB2 string
}

type withString struct {
	String     string                      `alias:"string"`
	Bytes      []byte                      `alias:"bytes"`
	Int        int                         `alias:"int"`
	Int8       int8                        `alias:"int8"`
	Int16      int16                       `alias:"int16"`
	Int32      int32                       `alias:"int32"`
	Int64      int64                       `alias:"int64"`
	Uint       uint                        `alias:"uint"`
	Uint8      uint8                       `alias:"uint8"`
	Uint16     uint16                      `alias:"uint16"`
	Uint32     uint32                      `alias:"uint32"`
	Uint64     uint64                      `alias:"uint64"`
	Bool       bool                        `alias:"bool"`
	Float32    float32                     `alias:"float32"`
	Float64    float64                     `alias:"float64"`
	Complex64  complex64                   `alias:"complex64"`
	Complex128 complex128                  `alias:"complex128"`
	Time       time.Time                   `alias:"time"`
	Time2      Time020                     `alias:"time2"`
	Duration   time.Duration               `alias:"duration"`
	Interface  interface{}                 `alias:"interface"`
	Struct     BB                          `alias:"struct"`
	StructNil  BB                          `alias:"structnil"`
	Map        map[interface{}]interface{} `alias:"map"`
	MapNil     map[interface{}]interface{} `alias:"mapnil"`
	MapString  map[string]interface{}      `alias:"mapstring"`
	MapPtr     map[*int]interface{}        `alias:"mapptr"`
	SliceInt   []int                       `alias:"sliceint"`
	SliceNil   []int                       `alias:"slicenil"`
	ArrayInt   [10]int                     `alias:"arrayint"`
	ArrayInt2  [10]int                     `alias:"arrayint2"`
	Ptr        *int                        `alias:"ptr"`
	PtrMap     *map[string]interface{}     `alias:"ptrmap"`
	Unsafe     unsafe.Pointer              `alias:"unsafe"`
}

type Time020 time.Time

func TestSetWithString2(t *testing.T) {
	data := &withString{
		Unsafe: unsafe.Pointer(t),
	}
	t.Log(eudore.Set(data, "string", TestSetWithString2))
	t.Log(eudore.Set(data, "string", []byte("666s")))
	t.Log(eudore.Set(data, "bytes", "[]byte"))
	t.Log(eudore.Set(data, "int", ""))
	t.Log(eudore.Set(data, "int", []int{1, 2, 3}))
	t.Log(eudore.Set(data, "int", []byte("123")))
	t.Log(eudore.Set(data, "int", "1"))
	t.Log(eudore.Set(data, "int8", "2"))
	t.Log(eudore.Set(data, "int16", "3"))
	t.Log(eudore.Set(data, "int32", "4"))
	t.Log(eudore.Set(data, "int64", "5"))
	t.Log(eudore.Set(data, "uint", ""))
	t.Log(eudore.Set(data, "uint", "1"))
	t.Log(eudore.Set(data, "uint8", "2"))
	t.Log(eudore.Set(data, "uint16", "3"))
	t.Log(eudore.Set(data, "uint32", "4"))
	t.Log(eudore.Set(data, "uint64", "5"))
	t.Log(eudore.Set(data, "uint64", 6))
	t.Log(eudore.Set(data, "bool", ""))
	t.Log(eudore.Set(data, "bool", "true"))
	t.Log(eudore.Set(data, "float32", ""))
	t.Log(eudore.Set(data, "float32", "16"))
	t.Log(eudore.Set(data, "float64", "32"))
	t.Log(eudore.Set(data, "complex64", "1+b"))
	t.Log(eudore.Set(data, "complex64", "a+b"))
	t.Log(eudore.Set(data, "complex64", "(1.2"))
	t.Log(eudore.Set(data, "complex128", "1.2+3.4i"))
	t.Log(eudore.Set(data, "duration", ""))
	t.Log(eudore.Set(data, "duration", "3m"))
	t.Log(eudore.Set(data, "duration", "30s"))
	t.Log(eudore.Set(data, "duration", "30sxx"))
	t.Log(eudore.Set(data, "time", ""))
	t.Log(eudore.Set(data, "time", "2018-08-12"))
	t.Log(eudore.Set(data, "time2", "2018-08-12"))
	t.Log(eudore.Set(data, "interface", "eface"))
	t.Log(eudore.Set(data, "struct", `{"BB1":22,"BB2":"22"}`))
	t.Log(eudore.Set(data, "map.eface", "str"))
	t.Log(eudore.Set(data, "mapptr.2", "str"))
	t.Log(eudore.Set(data, "mapstring", `{"a":1,"b":2}`))
	t.Log(eudore.Set(data, "sliceint", "[1]"))
	t.Log(eudore.Set(data, "sliceint.2", "12"))
	t.Log(eudore.Set(data, "arrayint.2", "12"))
	t.Log(eudore.Set(data, "arrayint.12", "12"))
	t.Log(eudore.Set(data, "ptr", "1234"))
	t.Log(eudore.Set(data, "interface", data))
	t.Log(eudore.Set(data, "unsafe", "666"))
	t.Log(eudore.Set(data, "unsafe", 666))
	t.Log(eudore.Set(data, "unsafe", unsafe.Pointer(t)))
	t.Logf("%p\n", t)
	t.Logf("struct: %# v\n", pretty.Formatter(data))

	t.Log(eudore.GetWithTags(data, "mapnil.a", eudore.DefaultConvertTags, false))
	t.Log(eudore.GetWithTags(data, "mapptr.a", eudore.DefaultConvertTags, false))
	t.Log(eudore.GetWithTags(data, "slicenil.a", eudore.DefaultConvertTags, false))
	t.Log(eudore.GetWithTags(data, "sliceint.+", eudore.DefaultConvertTags, false))

	var dm map[string]interface{}
	eudore.ConvertTo(data, &dm)
	t.Logf("map[string]interface{}: %# v\n", pretty.Formatter(dm))
	dm["nnn"] = 666
	eudore.ConvertTo(dm, data)
	eudore.ConvertMap(data)
	eudore.ConvertMapString(data)
	eudore.ConvertMap(66)
	eudore.ConvertMapString(66)

	var ii interface{}
	t.Log(reflect.Indirect(reflect.ValueOf(ii)).Kind())
}

func Benchmark_iterator(b *testing.B) {
	b.StopTimer() //调用该函数停止压力测试的时间计数

	//做一些初始化的工作,例如读取文件数据,数据库连接之类的,
	//这样这些时间不影响我们测试函数本身的性能

	b.StartTimer() //重新开始时间
	b.ReportAllocs()
	type Stu1 struct {
		Name string
		Age  int
		HIgh bool
		sex  string
	}
	type Stu2 struct {
		Name string
		Age  int
		HIgh bool
		sex  string
		Stu1 Stu1
	}
	type Stu3 struct {
		Name string
		Age  int
		HIgh bool
		sex  string
		Stu2 Stu2
	}
	stu := Stu3{
		Name: "张三3",
		Age:  183,
		HIgh: true,
		sex:  "男3",
		Stu2: Stu2{
			Name: "张三2",
			Age:  182,
			HIgh: true,
			sex:  "男2",
			Stu1: Stu1{
				Name: "张三1",
				Age:  181,
				HIgh: true,
				sex:  "男1",
			},
		},
	}
	for i := 0; i < b.N; i++ {
		eudore.ConvertMap(stu)
	}
}
