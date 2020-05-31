package eudore_test

/*
goos: linux
goarch: amd64
BenchmarkHead1-2   	30000000	        58.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkHead2-2   	10000000	       220 ns/op	      32 B/op	       1 allocs/op
PASS
ok  	command-line-arguments	4.284s
*/

import ( 
	"testing"
	"net/textproto"
) 
func BenchmarkHead1(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		textproto.CanonicalMIMEHeaderKey("Content-Disposition")
	}
}
func BenchmarkHead2(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		textproto.CanonicalMIMEHeaderKey("content-disposition")
	}
}