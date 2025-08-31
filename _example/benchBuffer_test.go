package eudore_test

import (
	"bytes"
	"testing"
)

func init() {
	for i := range strs {
		for n := 0; n < 4; n++ {
			strs[i] = strs[i] + strs[i]
		}
	}
}

func BenchmarkBytesAppend(b *testing.B) {
	buf := make([]byte, 64)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf = buf[:0]
		for i := 0; i < 100; i++ {
			for _, s := range strs {
				buf = append(buf, s...)
			}
		}
	}
}

func BenchmarkBytesBuferr(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 64))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		for i := 0; i < 100; i++ {
			for _, s := range strs {
				buf.WriteString(s)
			}
		}
	}
}

func BenchmarkBytesBuferr2(b *testing.B) {
	buf := &Buffer{make([]byte, 64)}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.buf = buf.buf[:0]
		for i := 0; i < 100; i++ {
			for _, s := range strs {
				buf.WriteString(s)
			}
		}
	}
}

type Buffer struct {
	buf []byte
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	return copy(b.buf[b.grow(len(p)):], p), nil
}

func (b *Buffer) WriteString(s string) (n int, err error) {
	return copy(b.buf[b.grow(len(s)):], s), nil
}

func (b *Buffer) grow(n int) int {
	l := len(b.buf)
	if l+n <= cap(b.buf) {
		b.buf = b.buf[:l+n]
		return l
	}

	c := l + n
	if c < 2*l {
		c = 2 * l
	}
	buf2 := append([]byte(nil), make([]byte, c)...)
	copy(buf2, b.buf)
	b.buf = buf2[:l+n]
	return l
}

var strs = []string{
	"&{Context:context.Background.WithCancel",
	"CancelFunc:0x4f7de0",
	"Logger:0xc00013c5a0",
	"Config:0xc000080730",
	"Database:<nil>",
	"Client:0xc000153da0",
	"Server:0xc000080780",
	"Ro",
	"uter:0xc0001421c0",
	"GetWarp:0x782160",
	"HandlerFuncs:[github.com/eudore/eudore.(*App).serveContext-fm]",
	"ContextPool:0xc0001650b0",
	"CancelError:<nil>",
	"cancelMutex:{state:0",
	"sema:0}",
	"Values:[bind",
	"0x79fac0",
	"render",
	"0x7a0580",
	"templdate",
	"0xc000158060",
	"handler-extender",
	"0xc000160cc0",
	"func-creator",
	"0xc0001628c0]}&{RouterCore:0xc0000f5480",
	"HandlerExtender:0xc0000b9800",
	"Middlewares:0xc00008da90",
	"GroupParams:route=",
	"Logger:0xc000154270",
	"LoggerKind:all",
	"Meta:0xc00013c640}",
}
