package eudore_test

import (
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/eudore/eudore"
	. "github.com/eudore/eudore/middleware"
)

func TestMiddlewareBlack(*testing.T) {
	data := map[string]bool{
		"0.0.0.0/0":      false,
		"127.0.0.1":      true,
		"127.0.0.2":      true,
		"127.0.0.3":      true,
		"192.168.1.0/24": true,
		"::/0":           false,
		"::1":            true,
		"::2":            true,
		"::3":            true,
		"::ff01":         true,
	}

	app := NewApp()
	admin := app.Group("/eudore/debug")
	admin.AddMiddleware(NewLoggerLevelFunc(func(Context) int { return 4 }))
	app.GetFunc("/v", NewBlackListFunc(data, NewOptionRouter(admin)))

	req := func(ip string, status int) {
		app.GetRequest("/v", http.Header{HeaderXRealIP: {ip}}, NewClientCheckStatus(status))
	}
	req("127.0.0.1", 200)
	req("127.0.0.2", 200)
	req("127.0.0.3", 200)
	req("127.0.0.4", 403)
	req("::1", 200)
	req("::2", 200)
	req("::3", 200)
	req("::4", 403)
	req("::ff01", 200)
	req("::FF", 403)

	app.SetValue(ContextKeyClient, app.NewClient(
		NewClientOptionURL("/eudore/debug/black"),
	))
	app.GetRequest("/eudore/debug/black", NewClientCheckStatus(200))
	app.PutRequest("allow/10.127.87.0?mask=32", NewClientCheckStatus(200))
	app.PutRequest("allow/10:127:87::?mask=128", NewClientCheckStatus(200))
	app.PutRequest("deny/10.127.87.0?mask=24", NewClientCheckStatus(200))
	app.PutRequest("deny/10:127:87::?mask=24", NewClientCheckStatus(200))
	app.DeleteRequest("allow/10.127.87.0", NewClientCheckStatus(200))
	app.DeleteRequest("allow/10:127:87::", NewClientCheckStatus(200))
	app.DeleteRequest("deny/10.127.87.0?mask=24", NewClientCheckStatus(200))
	app.DeleteRequest("deny/10:127:87::?mask=24", NewClientCheckStatus(200))

	app.PutRequest("allow/10.127.87.2000", NewClientCheckStatus(500))
	app.DeleteRequest("allow/10.127.87.2000", NewClientCheckStatus(500))

	func() {
		defer func() {
			recover()
		}()
		NewBlackListFunc(map[string]bool{
			"10.127.87.2000": true,
		})
	}()

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareBlackData(*testing.T) {
	type SubnetList interface {
		Insert(string)
		Delete(string)
		Look4(uint32) bool
		Look6([2]uint64) bool
		List() any
	}

	fn := func(list SubnetList, cidrs, look []string) {
		for _, ip := range cidrs {
			list.Insert(ip)
		}

		for _, ip := range look {
			pos := strings.IndexByte(ip, '/')
			if pos != -1 {
				ip = ip[:pos]
			}
			if strings.IndexByte(ip, '.') != -1 {
				list.Look4(mip2int4(ip))
			} else {
				list.Look6(mip2int6(ip))
			}
		}

		for _, ip := range look {
			list.Delete(ip)
		}
	}

	cidrs := [][]string{
		{
			"127.0.0.1",
			"127.0.0.1",
			"127.0.0.2",
			"127.0.0.3",
			"127.1.0.1",
			"172.16.0.1/32",
			"172.16.0.0/16",
			"172.16.255.0/20",
			"172.16.0.0/24",
		},
		{
			"127.1.0.4",
			"127.255.0.0",
			"::1",
			"::2",
		},
		{
			"127::0:0:1",
			"127::0:0:1",
			"127::0:0:2",
			"127::0:0:3",
			"127::1:0:1",
			"172::16:0:1/128",
			"172::16:0:0/112",
			"172::16:255:0/124",
			"172::16:0:0/120",
		},
		{
			"127::1:0:4",
			"127::255:0:0",
			"127.0.0.1",
		},
	}

	fn(SubnetListV4().(SubnetList), cidrs[0], append(cidrs[1], cidrs[0]...))
	fn(SubnetListV6().(SubnetList), cidrs[2], append(cidrs[3], cidrs[2]...))
}

func mip2int4(ip string) uint32 {
	var fields [4]uint32
	var val, pos int
	for i := 0; i < len(ip); i++ {
		if ip[i] >= '0' && ip[i] <= '9' {
			val = val*10 + int(ip[i]) - '0'
		} else if ip[i] == '.' {
			fields[pos] = uint32(val)
			pos++
			val = 0
		}
	}
	fields[3] = uint32(val)

	return fields[0]<<24 | fields[1]<<16 | fields[2]<<8 | fields[3]
}

func mip2int6(ip string) [2]uint64 {
	ipv6, _ := netip.ParseAddr(ip)
	bytes := ipv6.As16()

	return [2]uint64{
		uint64(bytes[0])<<56 | uint64(bytes[1])<<48 | uint64(bytes[2])<<40 | uint64(bytes[3])<<32 |
			uint64(bytes[4])<<24 | uint64(bytes[5])<<16 | uint64(bytes[6])<<8 | uint64(bytes[7]),
		uint64(bytes[8])<<56 | uint64(bytes[9])<<48 | uint64(bytes[10])<<40 | uint64(bytes[11])<<32 |
			uint64(bytes[12])<<24 | uint64(bytes[13])<<16 | uint64(bytes[14])<<8 | uint64(bytes[15]),
	}
}

func TestMiddlewareBreaker(*testing.T) {
	app := NewApp()
	app.AddMiddleware("global", NewLoggerLevelFunc(func(Context) int { return 4 }))
	app.AddMiddleware(NewCircuitBreakerFunc(
		NewOptionCircuitBreakerConfig(3, 3, time.Millisecond*10, time.Millisecond*100),
		NewOptionRouter(app.Group("/eudore/debug")), // 创建熔断器并注入管理路由
	))
	app.GetFunc("/skip", NewCircuitBreakerFunc(
		NewOptionKeyFunc(func(Context) string { return "" }),
	))
	app.AnyFunc("/*", func(ctx Context) {
		if ctx.Request().URL.RawQuery != "" {
			ctx.Fatal("test err")
			return
		}
		ctx.WriteString("route: " + ctx.GetParam("route"))
	})

	app.GetRequest("/skip")
	// 错误请求
	for i := 0; i < 4; i++ {
		app.GetRequest("/faile?i=" + strconv.Itoa(i))
	}
	time.Sleep(time.Millisecond * 100)
	// 除非熔断后访问
	app.GetRequest("/halfopen")
	for i := 0; i < 4; i++ {
		app.GetRequest("/halfopen")
		time.Sleep(time.Millisecond * 15)
	}

	app.GetRequest("/eudore/debug/breaker", http.Header{HeaderAccept: {MimeApplicationJSON}})
	app.GetRequest("/eudore/debug/breaker/L3NraXA", NewClientCheckStatus(200))
	app.GetRequest("/eudore/debug/breaker/100", NewClientCheckStatus(500))
	app.PutRequest("/eudore/debug/breaker/100/state/3", NewClientCheckStatus(500))
	app.PutRequest("/eudore/debug/breaker/L3NraXA/state/3", NewClientCheckStatus(500))
	app.PutRequest("/eudore/debug/breaker/L3NraXA/state/0", NewClientCheckStatus(200))

	time.Sleep(time.Microsecond * 100)
	app.CancelFunc()
	app.Run()
}

func TestMiddlewareCacheData(*testing.T) {
	app := NewApp()
	app.AddMiddleware(
		NewCacheFunc(time.Millisecond*5, NewOptionCacheCleanup(app.Context, time.Millisecond*20)),
		NewBodyLimitFunc(4<<20),
	)
	app.AnyFunc("/sf", func(ctx Context) {
		ctx.Redirect(301, "/")
		if ctx.Method() == "GET" {
			w := ctx.Response()
			w.Push("/", nil)
			w.Flush()
			w.Hijack()
			b, ok := w.(interface{ Body() []byte })
			if ok {
				b.Body()
			}
		}
		ctx.Info(ctx.Response().Status(), ctx.Response().Size())
	})
	app.AnyFunc("/*", func(ctx Context) {
		time.Sleep(time.Millisecond * 2)
		ctx.WriteString("hello eudore")
	})

	app.GetRequest("/sf")
	wg := sync.WaitGroup{}
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 4; i++ {
				app.GetRequest("/?c=" + fmt.Sprint(i))
				app.GetRequest("/?c=" + fmt.Sprint(i))
				time.Sleep(time.Millisecond * 10)
				app.GetRequest("/?c=" + fmt.Sprint(i))
			}
			wg.Done()
		}()
	}
	wg.Wait()

	app.GetRequest("/sf", NewClientHeader(HeaderAccept, MimeApplicationJSON))
	app.PostRequest("/sf")
	app.GetRequest("/hello", NewClientBodyJSON(map[string]any{
		"body": "json",
	}))
	time.Sleep(time.Millisecond * 220)
	app.GetRequest("/s")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRateRequest(*testing.T) {
	app := NewApp()
	app.AnyFunc("/*", NewRateRequestFunc(1, 3, NewOptionRateState()))
	app.AnyFunc("/skip", NewRateRequestFunc(1, 3,
		NewOptionKeyFunc(func(ctx Context) string {
			if ctx.Path() == "/skip" {
				return ""
			}
			return ctx.RealIP()
		}),
	))
	app.AnyFunc("/clear", NewRateRequestFunc(20*100, 1,
		NewOptionRateCleanup(app.Context, time.Millisecond, 0),
	))
	app.AnyFunc("/reset", NewRateRequestFunc(20*100, 1))

	go func() {
		for i := 0; i < 8; i++ {
			app.GetRequest("/")
		}
	}()
	app.GetRequest("/clear")
	app.GetRequest("/skip")
	app.GetRequest("/reset")
	time.Sleep(time.Millisecond * 100)
	app.GetRequest("/reset")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRateSpeed(*testing.T) {
	app := NewApp()
	app.AddMiddleware("global", NewLoggerLevelFunc(func(Context) int { return 4 }))
	app.AddMiddleware(NewRateSpeedFunc(16<<10, 64<<10,
		NewOptionRateCleanup(app.Context, time.Millisecond, 10),
		NewOptionRateCleanup(app.Context, time.Millisecond*10, 0),
		NewOptionKeyFunc(func(ctx Context) string {
			if ctx.Path() == "/skip" {
				return ""
			}
			return ctx.RealIP()
		}),
	))
	app.GetFunc("/skip", HandlerEmpty)
	app.PostFunc("/read", func(ctx Context) {
		ctx.Body()
	})
	app.AnyFunc("/write", func(ctx Context) {
		ctx.Write([]byte("rate speed 16B"))
		ctx.WriteString("rate speed 16B")
	})
	app.AnyFunc("/deadline",
		func(ctx Context) {
			c, _ := context.WithTimeout(ctx.Context(), time.Microsecond)
			ctx.SetContext(c)
		},
		NewRateSpeedFunc(1024, 128),
		func(ctx Context) {
			ctx.Body()
		},
	)
	app.AnyFunc("/wait",
		NewRateSpeedFunc(1024, 256),
		func(ctx Context) {
			ctx.Body()
		},
	)
	app.AnyFunc("/cannel",
		func(ctx Context) {
			c, cannel := context.WithCancel(ctx.Context())
			go cannel()
			ctx.SetContext(c)
		},
		NewRateSpeedFunc(1024, 128),
		func(ctx Context) {
			ctx.Body()
			ctx.Write([]byte("body"))
			ctx.WriteString("body")
		},
	)

	app.GetRequest("/skip")
	app.PostRequest("/read", strings.NewReader("read body"))
	app.PostRequest("/read", strings.NewReader("read body"))
	app.PostRequest("/read", strings.NewReader("read body"))
	app.PostRequest("/deadline", strings.NewReader("wait body"))
	app.PostRequest("/wait", strings.NewReader("wait body"))
	app.PostRequest("/cannel", strings.NewReader("wait body"))
	app.PostRequest("/write")
	time.Sleep(time.Second / 10)

	app.CancelFunc()
	app.Run()
}
