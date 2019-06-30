package test

import (
	"fmt"
	"github.com/eudore/eudore"
	"github.com/kr/pretty"
	"testing"
)

type (
	ST struct {
		A int
		B int
	}
	ST2 struct {
		A string
		B string
	}
	ST3 struct {
		C *ST
	}
	ST4 struct {
		C *ST2
	}
	ST5 struct {
		C  interface{}
		D1 interface{}
		D2 interface{}
	}
)

func TestConvertTo1(*testing.T) {
	{

		var s ST = ST{A: 1}
		var s1 ST
		var m1 = make(map[string]int)
		eudore.ConvertTo(&s, &s1)
		eudore.ConvertTo(&s, m1)
		fmt.Printf("struct: %# v\n", pretty.Formatter(s1))
		fmt.Printf("struct: %# v\n", pretty.Formatter(m1))
	}
	{
		var m = map[string]int{
			"A": 2,
			"B": 0,
		}
		var s1 ST
		var m1 = make(map[string]int)
		eudore.ConvertTo(m, &s1)
		eudore.ConvertTo(m, m1)
		fmt.Printf("struct: %# v\n", pretty.Formatter(s1))
		fmt.Printf("struct: %# v\n", pretty.Formatter(m1))
	}
}

func TestConvertTo2(*testing.T) {
	{
		var s ST = ST{A: 1}
		var s1 ST2
		var m1 = make(map[string]string)
		eudore.ConvertTo(&s, &s1)
		eudore.ConvertTo(&s, m1)
		fmt.Printf("struct: %# v\n", pretty.Formatter(s1))
		fmt.Printf("struct: %# v\n", pretty.Formatter(m1))
	}
	{
		var m = map[string]string{
			"A": "2",
			"B": "0",
		}
		var s1 ST
		var m1 = make(map[string]int)
		eudore.ConvertTo(m, &s1)
		eudore.ConvertTo(m, m1)
		fmt.Printf("struct: %# v\n", pretty.Formatter(s1))
		fmt.Printf("struct: %# v\n", pretty.Formatter(m1))
	}
}

func TestConvertTo3(*testing.T) {
	{
		var m = map[string]interface{}{
			"C": map[string]string{
				"A": "11",
			},
		}
		var s1 ST3
		var m1 = make(map[string]map[string]string)
		eudore.ConvertTo(m, &s1)
		eudore.ConvertTo(m, m1)
		fmt.Printf("struct: %# v\n", pretty.Formatter(s1))
		fmt.Printf("struct: %# v\n", pretty.Formatter(m1))
	}
	{
		var s ST3 = ST3{
			C: &ST{
				A: 1,
			},
		}
		var s1 ST4
		var m1 = make(map[string]map[string]string)
		eudore.ConvertTo(&s, &s1)
		eudore.ConvertTo(&s, m1)
		fmt.Printf("struct: %# v\n", pretty.Formatter(s1))
		fmt.Printf("struct: %# v\n", pretty.Formatter(m1))
	}
}

func TestConvertTo4(*testing.T) {
	{
		var m = map[string]interface{}{
			"D1": map[string]string{
				"A": "11",
			},
			"D2": &ST{
				A: 2,
			},
		}
		var s1 ST5
		var m1 = make(map[string]map[string]string)
		eudore.ConvertTo(m, &s1)
		eudore.ConvertTo(m, m1)
		fmt.Printf("struct: %# v\n", pretty.Formatter(s1))
		fmt.Printf("struct: %# v\n", pretty.Formatter(m1))
	}
	{
		var s ST5 = ST5{
			C: map[string]string{
				"A": "1",
			},
			D1: map[string]int{
				"A": 2,
			},
			D2: &ST{
				A: 3,
			},
		}
		var s1 ST4
		var m1 = make(map[string]map[string]string)
		var m2 = make(map[string]map[string]interface{})
		var m3 = make(map[string]interface{})
		eudore.ConvertTo(&s, &s1)
		eudore.ConvertTo(&s, m1)
		eudore.ConvertTo(&s, m2)
		eudore.ConvertTo(&s, m3)
		fmt.Printf("struct: %# v\n", pretty.Formatter(s1))
		fmt.Printf("struct: %# v\n", pretty.Formatter(m1))
		fmt.Printf("struct: %# v\n", pretty.Formatter(m2))
		fmt.Printf("struct: %# v\n", pretty.Formatter(m3))
	}
}
