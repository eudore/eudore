package eudore_test

/*
goos: linux
goarch: amd64
cpu: Intel(R) Xeon(R) Gold 6133 CPU @ 2.50GHz
BenchmarkReflectFuncString-2      	 3243481	       401.2 ns/op	      25 B/op	       2 allocs/op
BenchmarkReflectFuncInt-2         	 2878328	       376.2 ns/op	      25 B/op	       2 allocs/op
BenchmarkReflectFuncInterface-2   	 2321782	       476.6 ns/op	      41 B/op	       3 allocs/op
BenchmarkRunFuncString-2          	630790024	       2.236 ns/op	       0 B/op	       0 allocs/op
BenchmarkRunFuncInt-2             	511247667	       2.575 ns/op	       0 B/op	       0 allocs/op
BenchmarkRunFuncInterface-2       	615269007	       1.986 ns/op	       0 B/op	       0 allocs/op
BenchmarkRunFuncStringKind-2      	541799686	       2.139 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	command-line-arguments	10.828s
*/

import (
	"reflect"
	"testing"
)

var (
	ReflectString    = reflect.ValueOf(stringIsZero)
	ReflectInt       = reflect.ValueOf(intIsZero)
	ReflectInterface = reflect.ValueOf(interfaceIsZero)
	FuncString       = stringIsZero
	FuncInt          = intIsZero
	FuncInterface    = interfaceIsZero
)

func stringIsZero(string) bool         { return true }
func intIsZero(int) bool               { return true }
func interfaceIsZero(interface{}) bool { return true }

func BenchmarkReflectFuncString(b *testing.B) {
	args := []reflect.Value{reflect.ValueOf("0")}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ReflectString.Call(args)
	}
}

func BenchmarkReflectFuncInt(b *testing.B) {
	args := []reflect.Value{reflect.ValueOf(0)}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ReflectInt.Call(args)
	}
}

func BenchmarkReflectFuncInterface(b *testing.B) {
	args := []reflect.Value{reflect.ValueOf(0)}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ReflectInterface.Call(args)
	}
}

func BenchmarkRunFuncString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		FuncString("0")
	}
}

func BenchmarkRunFuncInt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		FuncInt(0)
	}
}

func BenchmarkRunFuncInterface(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		FuncInterface(0)
	}
}

func BenchmarkRunFuncStringKind(b *testing.B) {
	var fn interface{} = stringIsZero
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		fn.(func(string) bool)("0")
	}
}
