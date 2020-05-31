package eudore_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/eudore/eudore"
	"github.com/kr/pretty"
)

type (
	config struct {
		Full  *A          `alias:"full"`
		Array A1          `alias:"array"`
		Map   M1          `alias:"map"`
		Face  interface{} `alias:"face"`
		Base  B1          `alias:"base"`
		Time  time.Time
	}
	A struct {
		A map[int]B
	}
	B struct {
		B map[int]*C
	}
	C struct {
		C D
	}
	D struct {
		D *E
	}
	E struct {
		E  string
		E2 string      `alias:"e2"`
		E3 interface{} `alias:"e3"`
	}

	M1 struct {
		M2 map[string]int
		M3 map[string]E
		M4 map[string]*E
		M5 map[string]interface{}
	}
	A1 struct {
		A2 []string
		A3 []int
		A4 []E
		A5 []*E
		A6 []interface{}
	}
	B1 struct {
		B2 int
		B3 string
		B4 interface{}
		B5 []int
		BB BB
		BI sort.Interface
	}
	BB struct {
		BB1 int
		BB2 string
	}
)

type SortBytes []byte

func (p SortBytes) Len() int           { return len(p) }
func (p SortBytes) Less(i, j int) bool { return p[i] < p[j] }
func (p SortBytes) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func equal(a, b interface{}) bool {
	a1 := strings.Split(fmt.Sprintf("%# v", pretty.Formatter(a)), "\n")
	a2 := strings.Split(fmt.Sprintf("%# v", pretty.Formatter(b)), "\n")
	sort.Strings(a1)
	sort.Strings(a2)
	if len(a1) != len(a2) {
		return false
	}
	for i := 0; i < len(a1); i++ {
		// 字符串乱序比较
		if a1[i] != a2[i] {
			b1 := SortBytes(a1[i])
			b2 := SortBytes(a2[i])
			sort.Sort(b1)
			sort.Sort(b2)
			if string(b1) != string(b2) {
				return false
			}
		}
	}
	return true
}

func TestSetMap2(t *testing.T) {
	var conf = &config{}
	t.Log(eudore.Set(conf, "map.M2.2", "9991"))
	t.Log(eudore.Set(conf, "map.M2.3", "9992"))
	t.Log(eudore.Set(conf, "map.M3.2.E", "9993"))
	t.Log(eudore.Set(conf, "map.M3.2.E2", "e2"))
	t.Log(eudore.Set(conf, "map.M3.3.E", "9994"))
	t.Log(eudore.Set(conf, "map.M3.3.E2", "9995"))
	t.Log(eudore.Set(conf, "map.M4.2.E", "9996"))
	t.Log(eudore.Set(conf, "map.M4.2.E2", "9997"))
	t.Log(eudore.Set(conf, "Map.M4.3.E", "ug9"))
	t.Log(eudore.Set(conf, "Map.M4.3.e2", "88"))
	t.Log(eudore.Set(conf, "map.M5.E", "9998"))
	t.Log(eudore.Set(conf, "map.M5.E2", "9999"))

	t.Logf("struct: %# v\n", pretty.Formatter(conf))

	result := &config{
		Map: M1{
			M2: map[string]int{
				"2": 9991,
				"3": 9992,
			},
			M3: map[string]E{
				"2": {E: "9993", E2: "e2"},
				"3": {E: "9994", E2: "9995"},
			},
			M4: map[string]*E{
				"2": {E: "9996", E2: "9997"},
				"3": {E: "ug9", E2: "88"},
			},
			M5: map[string]interface{}{
				"E":  "9998",
				"E2": "9999",
			},
		},
	}
	if !equal(conf, result) {
		panic("ne")
	}
}

func TestSetArray2(t *testing.T) {
	var conf = &config{}
	t.Log(eudore.Set(conf, "array.A2.+", "11"))
	t.Log(eudore.Set(conf, "array.A2.1", "222"))
	t.Log(eudore.Set(conf, "array.A2.1", "2333"))
	t.Log(eudore.Set(conf, "array.A2.3", "44"))
	t.Log(eudore.Set(conf, "array.A4.3.E", "2"))
	t.Log(eudore.Set(conf, "array.A4.3.E2", "2"))
	t.Log(eudore.Set(conf, "array.A4.+.E", "2"))
	t.Log(eudore.Set(conf, "array.A4.4.E2", "2"))
	t.Log(eudore.Set(conf, "array.A5.3.E", "2"))
	t.Log(eudore.Set(conf, "array.A5.3.E2", "2"))
	t.Log(eudore.Set(conf, "array.A6.3.E2", "2"))
	t.Log(eudore.Set(conf, "array.A6.3.E3", "2"))
	t.Logf("struct: %# v\n", pretty.Formatter(conf))

	result := &config{
		Array: A1{
			A2: []string{"11", "2333", "", "44"},
			A4: []E{{}, {}, {}, {E: "2", E2: "2"}, {E: "2", E2: "2"}},
			A5: []*E{nil, nil, nil, {E: "2", E2: "2"}},
			A6: []interface{}{nil, nil, nil, map[string]interface{}{
				"E2": "2",
				"E3": "2",
			}},
		},
	}
	if !equal(conf, result) {
		panic("ne")
	}
}

func TestSetBase2(t *testing.T) {
	var conf = &config{}
	t.Log(eudore.Set(nil, "base.B2", "2"))
	var i map[string]interface{}
	t.Log(eudore.Set(i, "base.B2", "2"))
	t.Log(eudore.Set(conf, "Time.ext", 222))
	t.Log(eudore.Set(99, "base.B2", "2"))
	t.Log(eudore.Set(conf, "base.B2.1", "2"))
	t.Log(eudore.Set(conf, "base.BI.x", SortBytes("abc")))

	t.Log(eudore.Set(conf, "base.B2", "2"))
	t.Log(eudore.Set(conf, "base.B3", 2))
	t.Log(eudore.Set(conf, "base.B4.2", "2"))
	t.Log(eudore.Set(conf, "base.B5", []int{1, 2, 3, 4}))
	t.Log(eudore.Set(conf, "base.BI", SortBytes("abc")))
	t.Logf("struct: %# v\n", pretty.Formatter(conf))

	result := &config{
		Base: B1{
			B2: 2,
			B3: "2",
			B4: map[string]interface{}{
				"2": "2",
			},
			B5: []int{1, 2, 3, 4},
			BI: SortBytes("abc"),
		},
	}
	if !equal(conf, result) {
		panic("ne")
	}
}

func TestSetFull2(t *testing.T) {
	var conf = &config{}
	t.Log(eudore.Set(conf, "full.A.0.B.1.C.D.E", "999"))
	t.Log(eudore.Set(conf, "full.A.0.B.1.C.D.e2", "999"))
	t.Log(eudore.Set(conf, "full.A.0.B.1.C.D.e3", 988))
	t.Logf("struct: %# v\n", pretty.Formatter(conf))

	result := &config{
		Full: &A{
			A: map[int]B{
				0: {
					B: map[int]*C{
						1: {
							C: D{
								D: &E{
									E:  "999",
									E2: "999",
									E3: 988,
								},
							},
						},
					},
				},
			},
		},
	}
	if !equal(conf, result) {
		panic("ne")
	}

}

func TestSetFace2(t *testing.T) {
	var conf = &config{}
	conf.Face = &E{}
	t.Log(eudore.Set(conf, "face.E2", "2"))
	t.Log(eudore.Set(conf, "face.E3", []int{1, 2, 3, 4}))
	t.Log(eudore.Set(conf, "face.E3.0", 0))
	t.Log(eudore.Set(conf, "face.E3.6", 9))
	t.Log(eudore.Set(conf, "face.E3.+", 9))
	t.Log(eudore.Set(conf, "face.E5", []int{1, 2, 3, 4}))
	t.Log(eudore.Set(conf, "face.E3", map[string]string{
		"ss": "ss",
	}))
	t.Logf("struct: %# v\n", pretty.Formatter(conf))

	result := &config{
		Face: &E{
			E2: "2",
			E3: map[string]string{"ss": "ss"},
		},
	}
	if !equal(conf, result) {
		panic("ne")
	}

}

func TestSetZ2(t *testing.T) {
	var conf interface{} = map[string]int{}
	t.Log(eudore.Set(conf, "", 1))
	t.Logf("struct: %# v\n", pretty.Formatter(conf))

	var result map[string]int
	if !equal(conf, result) {
		panic("ne")
	}
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

func TestSetWithString2(t *testing.T) {
	data := &withString{
		Unsafe: unsafe.Pointer(t),
	}
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

	t.Log(eudore.GetWithTags(data, "mapnil.a", eudore.DefaultConvertTags))
	t.Log(eudore.GetWithTags(data, "mapptr.a", eudore.DefaultConvertTags))
	t.Log(eudore.GetWithTags(data, "slicenil.a", eudore.DefaultConvertTags))
	t.Log(eudore.GetWithTags(data, "sliceint.+", eudore.DefaultConvertTags))

	var dm map[string]interface{}
	eudore.ConvertTo(data, &dm)
	dm["nnn"] = 666
	eudore.ConvertTo(dm, data)
	eudore.ConvertMap(data)
	eudore.ConvertMapString(data)
	eudore.ConvertMap(66)
	eudore.ConvertMapString(66)
}

func TestGetBase2(t *testing.T) {
	var config = &config{
		Base: B1{
			B2: 2,
			B3: "22",
			B4: map[int]string{
				1: "11",
				2: "22",
			},
			B5: []int{1, 2, 3},
		},
	}

	t.Log(eudore.GetWithTags(nil, "base.B2", eudore.DefaultConvertTags))
	t.Log(eudore.GetWithTags(config, "", eudore.DefaultConvertTags))
	t.Log(eudore.GetWithTags(config, "Full.A", eudore.DefaultConvertTags))
	t.Log(eudore.GetWithTags(config, "base.B2.xx", eudore.DefaultConvertTags))
	t.Log(eudore.GetWithTags(config, "Time.ext", eudore.DefaultConvertTags))
	t.Log(eudore.GetWithTags(config, "Time.ext2", eudore.DefaultConvertTags))

	t.Log(eudore.Get(config, "base.B2"))
	t.Log(eudore.Get(config, "base.B3"))
	t.Log(eudore.Get(config, "base.B4"))
	t.Log(eudore.Get(config, "base.B4.1"))
	t.Log(eudore.Get(config, "base.B5.1"))
	t.Log(eudore.Get(config, "base.B5"))
	t.Logf("struct: %# v\n", pretty.Formatter(config))
}

func TestConvertMap2(t *testing.T) {
	var b interface{} = &B1{
		B2: 2,
		B3: "33",
		B4: map[string]interface{}{
			"i": 1,
			"s": "ss",
			"st": &E{
				E:  "ee",
				E2: "e2",
				E3: 3,
			},
		},
		B5: []int{1, 2, 3},
	}
	t.Logf("struct: %# v\n", pretty.Formatter(eudore.ConvertMap(b)))
	t.Logf("struct: %# v\n", pretty.Formatter(eudore.ConvertMapString(b)))
}

func TestConvertStruct2(t *testing.T) {
	c := map[interface{}]interface{}{
		"B2": 2,
		"B3": 2,
		"B4": 4,
		"B5": []int{0, 1, 5},
		"BB": map[interface{}]interface{}{
			"BB1": "1",
			"BB2": "2",
		},
	}
	var b = &B1{}
	eudore.ConvertTo(c, b)
	t.Logf("struct: %# v\n", pretty.Formatter(b))

	var a []int //= make([]int, 3)

	eudore.Set(&a, "+", 66)
	eudore.Set(&a, "1", 1)
	t.Logf("struct: %# v\n", pretty.Formatter(a))
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
