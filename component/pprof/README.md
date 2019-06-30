# Pprof

适配`net/http/pprof`库的功能。

```golang
package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/pprof"
)

func main() {
	app := eudore.NewCore()
	pprof.Inject(app.Group("/eudore/debug"))
	
	app.Listen(":8088")
	app.Run()
}
```

然后阅览器访问地址`http://127.0.0.1:8088/eudore/debug/pprof/`，就可以显示pprof，路径前缀可改。