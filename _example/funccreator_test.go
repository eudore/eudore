package eudore_test

import (
	"context"
	"testing"
	"time"

	"github.com/eudore/eudore"
)

func TestFuncCreator(t *testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyFuncCreator, eudore.NewFuncCreator())

	fc := eudore.NewFuncCreatorWithContext(app)
	t.Log(fc.RegisterFunc("zero", func() {}))
	t.Log(fc.CreateFunc(eudore.FuncCreateString, "zero1"))
	t.Log(fc.CreateFunc(eudore.FuncCreateString, "zero=1"))
	t.Log(fc.CreateFunc(eudore.FuncCreateKind(0), "zero"))

	fc.(interface{ Metadata() any }).Metadata()
}

func mustCreate(i any, _ error) any {
	return i
}

func TestFuncCreatorRun(t *testing.T) {
	fc := eudore.NewFuncCreatorWithContext(context.Background())
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "zero")).(func(string) bool)("zero")
	mustCreate(fc.CreateFunc(eudore.FuncCreateAny, "zero")).(func(any) bool)("zero")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "nozero")).(func(string) bool)("zero")
	mustCreate(fc.CreateFunc(eudore.FuncCreateAny, "nozero")).(func(any) bool)("zero")
	mustCreate(fc.CreateFunc(eudore.FuncCreateAny, "nozero")).(func(any) bool)(time.Time{})

	mustCreate(fc.CreateFunc(eudore.FuncCreateInt, "min=0")).(func(int) bool)(0)
	mustCreate(fc.CreateFunc(eudore.FuncCreateInt, "max=0")).(func(int) bool)(0)
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "min=0")).(func(string) bool)("0")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "min=0")).(func(string) bool)("x0")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "max=0")).(func(string) bool)("0")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "max=0")).(func(string) bool)("x0")
	fc.CreateFunc(eudore.FuncCreateInt, "min=0")
	fc.CreateFunc(eudore.FuncCreateUint, "min=0")
	fc.CreateFunc(eudore.FuncCreateFloat, "min=0")
	fc.CreateFunc(eudore.FuncCreateBool, "min=0")
	fc.CreateFunc(eudore.FuncCreateInt, "min=x0")
	fc.CreateFunc(eudore.FuncCreateInt, "max=x0")
	fc.CreateFunc(eudore.FuncCreateString, "min=x0")
	fc.CreateFunc(eudore.FuncCreateString, "max=x0")

	mustCreate(fc.CreateFunc(eudore.FuncCreateInt, "equal=0")).(func(int) bool)(0)
	mustCreate(fc.CreateFunc(eudore.FuncCreateInt, "enum=1,2,3")).(func(int) bool)(0)
	mustCreate(fc.CreateFunc(eudore.FuncCreateInt, "enum=1,2,3")).(func(int) bool)(1)
	mustCreate(fc.CreateFunc(eudore.FuncCreateInt, "enum=1,2,3,4,5,6,7,8,9")).(func(int) bool)(0)
	fc.CreateFunc(eudore.FuncCreateInt, "equal=x0")
	fc.CreateFunc(eudore.FuncCreateInt, "enum=x0")

	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "len=0")).(func(string) bool)("0")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "len!=0")).(func(string) bool)("0")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "len>0")).(func(string) bool)("0")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "len<0")).(func(string) bool)("0")
	mustCreate(fc.CreateFunc(eudore.FuncCreateAny, "len=0")).(func(any) bool)("0")
	mustCreate(fc.CreateFunc(eudore.FuncCreateAny, "len>0")).(func(any) bool)("0")
	mustCreate(fc.CreateFunc(eudore.FuncCreateAny, "len<0")).(func(any) bool)("0")
	mustCreate(fc.CreateFunc(eudore.FuncCreateAny, "len=0")).(func(any) bool)(true)
	mustCreate(fc.CreateFunc(eudore.FuncCreateAny, "len>0")).(func(any) bool)(true)
	mustCreate(fc.CreateFunc(eudore.FuncCreateAny, "len<0")).(func(any) bool)(true)
	mustCreate(fc.CreateFunc(eudore.FuncCreateAny, "after:20180801")).(func(any) bool)(time.Now())
	mustCreate(fc.CreateFunc(eudore.FuncCreateAny, "after:20180801")).(func(any) bool)(nil)
	mustCreate(fc.CreateFunc(eudore.FuncCreateAny, "before:20180801")).(func(any) bool)(time.Now())
	mustCreate(fc.CreateFunc(eudore.FuncCreateAny, "before:20180801")).(func(any) bool)(nil)
	fc.CreateFunc(eudore.FuncCreateString, "len=x0")
	fc.CreateFunc(eudore.FuncCreateAny, "len=x0")
	fc.CreateFunc(eudore.FuncCreateAny, "after:")
	fc.CreateFunc(eudore.FuncCreateAny, "before:")

	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "num")).(func(string) bool)("0")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "integer")).(func(string) bool)("0")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "integer")).(func(string) bool)("x0")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "domain")).(func(string) bool)("www.eudore.cn")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "mail")).(func(string) bool)("postmaster@eudore.cn")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "mail")).(func(string) bool)("eudore.cn")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "phone")).(func(string) bool)("15824681234")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "phone")).(func(string) bool)("+86 15824681234")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "phone")).(func(string) bool)("010-32221234")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "phone")).(func(string) bool)("xx010-32221234")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "regexp=^\\d+$")).(func(string) bool)("123456")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "patten=123456")).(func(string) bool)("123456")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "patten=*")).(func(string) bool)("123456")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "patten!=a*")).(func(string) bool)("123456")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "patten!=a*b")).(func(string) bool)("a")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "patten=a*b")).(func(string) bool)("axb")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "prefix=123")).(func(string) bool)("123456")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "count=1,a")).(func(string) bool)("1aa23456")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "count>1,a")).(func(string) bool)("1aa23456")
	mustCreate(fc.CreateFunc(eudore.FuncCreateString, "count<1,a")).(func(string) bool)("1aa23456")
	fc.CreateFunc(eudore.FuncCreateString, "regexp=^[($")
	fc.CreateFunc(eudore.FuncCreateString, "count=")
	fc.CreateFunc(eudore.FuncCreateString, "count=x,x")

	mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "default")).(func(string) string)("zero")
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetInt, "default")).(func(int) int)(0)
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetUint, "default")).(func(uint) uint)(0)
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetFloat, "default")).(func(float64) float64)(0)
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetAny, "default")).(func(any) any)("zero")
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "value=str")).(func(string) string)("zero")
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetAny, "value="))
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetAny, "value=20060102")).(func(any) any)("zero")
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetAny, "value=20060102")).(func(any) any)(time.Now())
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetInt, "add=10")).(func(int) int)(0)
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetInt, "add=x"))
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetAny, "add="))
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetAny, "add=10h")).(func(any) any)("")
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetAny, "add=10h")).(func(any) any)(time.Now())
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "now:20060102")).(func(string) string)("now")
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "now:xx")).(func(string) string)("now")
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetAny, "now")).(func(any) any)(time.Time{})
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetAny, "now")).(func(any) any)("now")
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "replace=10,AA,aa")).(func(string) string)("AAAAA")
	mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "trim=1")).(func(string) string)("1234  ")

	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidesurname")).(func(string) string)("A4"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidesurname")).(func(string) string)("eudore"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidesurname")).(func(string) string)("eudore org"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidesurname")).(func(string) string)("世界"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidename")).(func(string) string)("A4"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidename")).(func(string) string)("eudore"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidename")).(func(string) string)("eudore org"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidename")).(func(string) string)("世界"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidemail")).(func(string) string)("postmaster@eudore.cn"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidemail")).(func(string) string)("master@eudore.cn"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidemail")).(func(string) string)("root@eudore.cn"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidemail")).(func(string) string)("eudore.cn"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidephone")).(func(string) string)("15824681234"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidephone")).(func(string) string)("+86 15824681234"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidephone")).(func(string) string)("010-32221234"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hidephone")).(func(string) string)("xx010-32221234"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "hide")).(func(string) string)("pass"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "len")).(func(string) string)("123456"))
	t.Log(mustCreate(fc.CreateFunc(eudore.FuncCreateSetString, "md5")).(func(string) string)("123456"))
	fc.CreateFunc(eudore.FuncCreateSetInt, "value=str")
}

func TestFuncCreatorExpr(t *testing.T) {
	k := eudore.FuncCreateString
	fc := eudore.NewFuncCreatorExpr()
	fc.RegisterFunc("init")

	t.Log(mustCreate(fc.CreateFunc(k, "NOT zero")).(func(string) bool)("0"))
	t.Log(mustCreate(fc.CreateFunc(k, "NOT(zero)")).(func(string) bool)(""))
	t.Log(mustCreate(fc.CreateFunc(k, "len>7 AND domain")).(func(string) bool)("eudore.cn"))
	t.Log(mustCreate(fc.CreateFunc(k, "len>7 \r\nAND(contains=xx ss)")).(func(string) bool)("xx"))
	t.Log(mustCreate(fc.CreateFunc(k, "(len>7 AND contains=xxss) OR mail OR phone")).(func(string) bool)("xxssxxss"))
	t.Log(mustCreate(fc.CreateFunc(k, "len>7 AND(contains=xxss OR mail) OR phone")).(func(string) bool)("xxssxxs"))
	t.Log(mustCreate(fc.CreateFunc(k, "len>7 AND contains=xxss OR(mail OR phone)")).(func(string) bool)("xxss@eudore.cn"))

	for i := 0; i < 8; i++ {
		fc.CreateFunc(eudore.FuncCreateKind(i), "NOT zero1")
	}
	fc.CreateFunc(k, "mail AND zero1")
	fc.CreateFunc(k, "mail OR zero1")
	fc.CreateFunc(k, " ")
	fc.CreateFunc(k, "( )")
	fc.CreateFunc(k, "(zero sss")
	fc.CreateFunc(k, "(zero sss) AND")
	fc.CreateFunc(k, "(zero sss) AND ( )")
	fc.CreateFunc(k, " zero")
	fc.CreateFunc(k, "(zero)")
	fc.CreateFunc(k, "NOT\r\nzero")
	fc.CreateFunc(k, "NOT(zero)")
	fc.CreateFunc(k, " len>7 AND NOT(contains=xx ss)")
	fc.CreateFunc(k, " len>7 AND(contains=xx ss) AND mail")
	fc.CreateFunc(k, "(len>7 AND contains=xxss) OR mail OR phone")
	fc.CreateFunc(k, " len>7 AND contains=xxss  OR(mail OR phone)")
	fc.CreateFunc(k, " len>7 AND(contains=xxss OR mail) OR phone")

	fc.List()

	// print meta
	meta, ok := fc.(interface{ Metadata() any }).Metadata().(eudore.MetadataFuncCreator)
	if ok {
		for _, expr := range meta.Exprs {
			t.Log("expr:", expr)
		}
		for _, err := range meta.Errors {
			t.Log("err:", err)
		}
	}
}
