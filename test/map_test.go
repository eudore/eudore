package test

import (
	"testing"
)

/*
n = 30

goos: linux
goarch: amd64
BenchmarkMap-2     	  500000	      2397 ns/op	    1167 B/op	       1 allocs/op
BenchmarkArray-2   	 5000000	       395 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	command-line-arguments	3.599s
*/

var (
	n    int
	data []int
	m    map[int]int
)

func init() {
	n = 30
	m := make(map[int]int)
	data = make([]int, n)
	for i := 0; i < n; i++ {
		data[i] = i
		m[i] = i
	}
}

func BenchmarkMap(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// 创建header
		m := make(map[int]int, n+2)
		for _, n := range data {
			m[n] = n
		}
		// 查找
		for _, n := range data {
			if _, ok := m[n]; ok {
				continue
			}
		}
	}
}
func BenchmarkArray(b *testing.B) {
	b.ReportAllocs()
	m := make([]int, n+2)
	for i := 0; i < b.N; i++ {
		// 创建header
		m = m[0:0]
		for _, n := range data {
			m = append(m, n)
		}
		// 查找
		for _, n := range data {
			for _, nn := range data {
				if nn == n {
					break
				}
			}
		}
	}
}
