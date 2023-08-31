package eudore_test

/*
goos: linux
goarch: amd64
cpu: Intel(R) Xeon(R) Gold 6133 CPU @ 2.50GHz
BenchmarkReflectTypeNameString-2    	48612525	        27.50 ns/op	       0 B/op	       0 allocs/op
BenchmarkReflectTypeNameInt-2       	49561389	        25.95 ns/op	       0 B/op	       0 allocs/op
BenchmarkReflectTypeEqualString-2   	211048983	        4.910 ns/op	       0 B/op	       0 allocs/op
BenchmarkReflectTypeEqualInt-2      	215496802	        5.405 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	command-line-arguments	6.036s
*/

import (
	"reflect"
	"testing"
)

const (
	String = "string"
	Int    = 0
)

func BenchmarkReflectTypeNameString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = reflect.TypeOf(String).String() == reflect.TypeOf(String).String()
	}
}

func BenchmarkReflectTypeNameInt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = reflect.TypeOf(Int).String() == reflect.TypeOf(Int).String()
	}
}

func BenchmarkReflectTypeEqualString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = reflect.TypeOf(String) == reflect.TypeOf(String)
	}
}

func BenchmarkReflectTypeEqualInt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = reflect.TypeOf(Int) == reflect.TypeOf(Int)
	}
}
