package test

import (
	"fmt"
	// "time"
	// "errors"
	// "reflect"
	// "strings"
	// "strconv"
	// "encoding/json"

	"github.com/eudore/eudore"
	"github.com/kr/pretty"
	"testing"
)

type (
	config struct {
		Full  *A          `set:"full"`
		Array A1          `set:"array"`
		Map   M1          `set:"map"`
		Face  interface{} `set:"face"`
		Base  B1          `set:"base"`
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
		E2 string      `set:"e2"`
		E3 interface{} `set:"e3"`
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
	}
	BB struct {
		BB1 int
		BB2 string
	}
)

func TestSetMap(*testing.T) {
	var config = &config{}
	fmt.Println(eudore.Set(config, "map.M2.2", "999"))
	fmt.Println(eudore.Set(config, "map.M2.3", "999"))
	fmt.Println(eudore.Set(config, "map.M3.2.E", "999"))
	fmt.Println(eudore.Set(config, "map.M3.2.E2", "e2"))
	fmt.Println(eudore.Set(config, "map.M3.3.E", "999"))
	fmt.Println(eudore.Set(config, "map.M3.3.E2", "999"))
	fmt.Println(eudore.Set(config, "map.M4.2.E", "999"))
	fmt.Println(eudore.Set(config, "map.M4.2.E2", "999"))
	fmt.Println(eudore.Set(config, "Map.M4.3.E", "ug9"))
	fmt.Println(eudore.Set(config, "Map.M4.3.e2", "88"))
	fmt.Println(eudore.Set(config, "map.M5.E", "999"))
	fmt.Println(eudore.Set(config, "map.M5.E2", "999"))

	fmt.Printf("struct: %# v\n", pretty.Formatter(config))
}
func TestSetArray(*testing.T) {
	var config = &config{}
	fmt.Println(eudore.Set(config, "array.A2.+", "11"))
	fmt.Println(eudore.Set(config, "array.A2.1", "222"))
	fmt.Println(eudore.Set(config, "array.A2.1", "2333"))
	fmt.Println(eudore.Set(config, "array.A2.3", "44"))
	fmt.Println(eudore.Set(config, "array.A4.3.E", "2"))
	fmt.Println(eudore.Set(config, "array.A4.3.E2", "2"))
	fmt.Println(eudore.Set(config, "array.A4.+.E", "2"))
	fmt.Println(eudore.Set(config, "array.A4.4.E2", "2"))
	fmt.Println(eudore.Set(config, "array.A5.3.E", "2"))
	fmt.Println(eudore.Set(config, "array.A5.3.E2", "2"))
	fmt.Println(eudore.Set(config, "array.A6.3.E2", "2"))
	fmt.Println(eudore.Set(config, "array.A6.3.E3", "2"))
	fmt.Printf("struct: %# v\n", pretty.Formatter(config))
}
func TestSetBase(*testing.T) {
	var config = &config{}
	fmt.Println(eudore.Set(config, "base.B2", "2"))
	fmt.Println(eudore.Set(config, "base.B3", 2))
	fmt.Println(eudore.Set(config, "base.B4.2", "2"))
	fmt.Println(eudore.Set(config, "base.B5", []int{1, 2, 3, 4}))
	fmt.Printf("struct: %# v\n", pretty.Formatter(config))
}
func TestSetFull(*testing.T) {
	var config = &config{}
	fmt.Println(eudore.Set(config, "full.A.0.B.1.C.D.E", "999"))
	fmt.Println(eudore.Set(config, "full.A.0.B.1.C.D.e2", "999"))
	fmt.Println(eudore.Set(config, "full.A.0.B.1.C.D.e3", 988))
	fmt.Printf("struct: %# v\n", pretty.Formatter(config))
}
func TestSetFace(*testing.T) {
	var config = &config{}
	config.Face = &E{}
	fmt.Println(eudore.Set(config, "face.E2", "2"))
	fmt.Println(eudore.Set(config, "face.E3", []int{1, 2, 3, 4}))
	fmt.Println(eudore.Set(config, "face.E3.0", 0))
	fmt.Println(eudore.Set(config, "face.E3.6", 9))
	fmt.Println(eudore.Set(config, "face.E3.+", 9))
	fmt.Println(eudore.Set(config, "face.E5", []int{1, 2, 3, 4}))
	// fmt.Println(eudore.Set(config, "face.E3", map[string]string{
	// "ss": "ss",
	// }))
	fmt.Printf("struct: %# v\n", pretty.Formatter(config))
}
func TestSetZ(*testing.T) {
	var config interface{} = map[string]int{}
	fmt.Println(eudore.Set(config, "", 1))
	fmt.Printf("struct: %# v\n", pretty.Formatter(config))
}

func TestGetBase(*testing.T) {
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

	fmt.Println(eudore.Get(config, "base.B2"))
	fmt.Println(eudore.Get(config, "base.B3"))
	fmt.Println(eudore.Get(config, "base.B4"))
	fmt.Println(eudore.Get(config, "base.B4.1"))
	fmt.Println(eudore.Get(config, "base.B5.1"))
	fmt.Println(eudore.Get(config, "base.B5"))
	fmt.Printf("struct: %# v\n", pretty.Formatter(config))
}

func TestConvertMap(*testing.T) {
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
	fmt.Printf("struct: %# v\n", pretty.Formatter(eudore.ConvertMap(b)))
	fmt.Printf("struct: %# v\n", pretty.Formatter(eudore.ConvertMapString(b)))
}

func TestConvertStruct(*testing.T) {
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
	eudore.ConvertStruct(b, c)
	fmt.Printf("struct: %# v\n", pretty.Formatter(b))

	var a []int //= make([]int, 3)

	ii, _ := eudore.Set(a, "+", 66)
	ii, _ = eudore.Set(ii.([]int), "1", 1)
	fmt.Printf("struct: %# v\n", pretty.Formatter(ii))
}
