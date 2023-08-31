package eudore_test

import (
	"bytes"
	"testing"
	"unsafe"
)

func BenchmarkAppendBuff(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 2048))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		for _, s := range strs {
			buf.WriteString(s)
		}
	}
}

func BenchmarkAppendCopy(b *testing.B) {
	en := &encoder{make([]byte, 2048)}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		en.data = en.data[0:0]
		for _, s := range strs {
			en.WriteString(s)
		}
	}
}

type encoder struct {
	data []byte
}

func (en *encoder) WriteString(s string) {
	b := *(*[]byte)(unsafe.Pointer(&struct {
		string
		Cap int
	}{s, len(s)}))
	en.data = append(en.data, b...)
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
