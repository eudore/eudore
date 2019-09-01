package test

/*
goos: linux
goarch: amd64
BenchmarkMvc-2        	 2000000	       894 ns/op	      64 B/op	       3 allocs/op
BenchmarkRestful-2     	10000000	       170 ns/op	       0 B/op	       0 allocs/op
BenchmarkBeegoMvc-2   	 2000000	       886 ns/op	      16 B/op	       1 allocs/op
PASS
ok  	command-line-arguments	7.238s
*/

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/eudore/eudore"
	// "github.com/astaxie/beego"
	// "github.com/kataras/iris"
	// "github.com/kataras/iris/mvc"
)

type Base1Controller struct {
	eudore.ControllerBase
}

func (*Base1Controller) Any()        {}
func (*Base1Controller) Get()        {}
func (*Base1Controller) GetIndex()   {}
func (*Base1Controller) GetContent() {}

func BenchmarkMvc(b *testing.B) {
	app := eudore.NewCore()
	app.AddController(new(Base1Controller))
	// app.Listen(":8084")
	// app.Run()

	req, _ := eudore.NewRequestReaderTest("GET", "/base1/", nil)
	resp := eudore.NewResponseWriterTest()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.EudoreHTTP(context.Background(), resp, req)
	}
}

func BenchmarkRestful(b *testing.B) {
	app := eudore.NewCore()
	base1 := app.Group("/base1")
	base1.AnyFunc("/*", eudore.HandlerEmpty)
	base1.GetFunc("/*", eudore.HandlerEmpty)
	base1.GetFunc("/index", eudore.HandlerEmpty)
	base1.GetFunc("/content", eudore.HandlerEmpty)

	req, _ := eudore.NewRequestReaderTest("GET", "/base1/", nil)
	resp := eudore.NewResponseWriterTest()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.EudoreHTTP(context.Background(), resp, req)
	}
}

// iris一样panic，且无法获取依赖。

/*type Base2Controller struct {}
func (*Base2Controller) Any() {}
func (*Base2Controller) Get() {}
func (*Base2Controller) GetIndex() {}
func (*Base2Controller) GetContent() {}

func BenchmarkIrisMvc(b *testing.B) {
	app := iris.New()
	mvc.New(app).Handle(new(BaseController))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest("GET", "/base2/", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
	}
}*/

/*
type Base3Controller struct {
	beego.Controller
}
func (*Base3Controller) Any() {}
func (*Base3Controller) Get() {}
func (*Base3Controller) GetIndex() {}
func (*Base3Controller) GetContent() {}

func BenchmarkBeego(b *testing.B) {
	app := beego.Router("/", &Base3Controller{})

		r := httptest.NewRequest("GET", "/base2/indexx", nil)
		w := httptest.NewRecorder()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Handlers.ServeHTTP(w, r)
	}
}
*/
