package eudore_test

/*
goos: linux
goarch: amd64
BenchmarkMiddlewareBlackTree-2        	 1000000	      1212 ns/op	       0 B/op	       0 allocs/op
BenchmarkMiddlewareBlackArray-2       	 1000000	      1956 ns/op	       0 B/op	       0 allocs/op
BenchmarkMiddlewareBlackIp2intbit-2   	 1000000	      1654 ns/op	     320 B/op	       5 allocs/op
BenchmarkMiddlewareBlackNetParse-2    	 1000000	      1989 ns/op	     360 B/op	      20 allocs/op
PASS
ok  	command-line-arguments	6.919s
*/

import (
	"bytes"
	"fmt"
	"net"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

var ips []string = []string{
	"10.0.0.0/4", "127.0.0.1/8", "192.168.1.0/24", "192.168.75.0/24", "192.168.100.0/24",
}

var requests []uint64 = []uint64{
	725415979, 2727437335, 889276411, 4005535794, 3864288534, 3906172701, 282878927, 1284469666, 730935782, 3371086418,
	1506312450, 1351422527, 1427742110, 1787801507, 2252116061, 229145224, 2463885032, 977944943, 3785363053, 3752670878,
	1109101831, 523139815, 2692892509, 822628332, 1521829731, 1137604504, 3946127316, 3492727158, 3701842868, 1345785201,
	2479587981, 1525387624, 2335875430, 2742578379, 842531784, 4164034788, 4067025409, 3579565778, 1135250289, 2272239320,
	2221887036, 47163049, 756685807, 3064055796, 2298095091, 3099116819, 4070972416, 1014033, 3023215026, 555430525,
	3702021454, 2340802113, 2507760403, 510831888, 3073321492, 4221140315, 1198583294, 1495418697, 827583711, 813333453,
	2746343126, 3755199452, 1697814659, 365059279, 3478405321, 2147566177, 281339662, 2742376600, 2293307920, 2061663865,
	913999062, 542572186, 4225265321, 633066366, 2063795404, 522841846, 195572401, 124532676, 2456662794, 3902204181,
	2491401143, 4233234751, 69766498, 388520887, 1017105985, 62871287, 3328355052, 1705168586, 2260082173, 3340006743,
	2211140888, 1906467873, 1247205260, 1492905294, 1014862918, 2587182986, 1040587870, 3570772999, 3084952258, 2425691705,
}

var requeststrs []string = []string{
	"43.60.248.43", "162.145.100.23", "53.1.71.251", "238.191.160.50", "230.84.93.22", "232.211.119.29", "16.220.99.207", "76.143.115.162", "43.145.49.230", "200.238.178.82",
	"89.200.129.2", "80.141.18.63", "85.25.157.158", "106.143.175.163", "134.60.144.93", "13.168.122.136", "146.219.230.232", "58.74.65.111", "225.160.14.109", "223.173.54.158",
	"66.27.141.7", "31.46.122.231", "160.130.71.93", "49.8.79.236", "90.181.71.99", "67.206.119.152", "235.53.31.212", "208.46.201.118", "220.165.163.180", "80.55.13.113",
	"147.203.130.141", "90.235.145.104", "139.58.161.102", "163.120.108.203", "50.56.3.200", "248.50.32.228", "242.105.226.1", "213.91.214.210", "67.170.139.113", "135.111.158.216",
	"132.111.78.60", "2.207.166.169", "45.26.27.239", "182.161.199.244", "136.250.37.243", "184.184.197.19", "242.166.28.0", "0.15.121.17", "180.50.153.178", "33.27.50.125",
	"220.168.93.78", "139.133.206.65", "149.121.99.19", "30.114.173.16", "183.47.42.20", "251.153.125.91", "71.112.237.254", "89.34.71.73", "49.83.236.223", "48.122.123.205",
	"163.177.222.214", "223.211.203.220", "101.50.152.131", "21.194.92.207", "207.84.64.201", "128.1.66.97", "16.196.231.14", "163.117.88.152", "136.177.26.16", "122.226.126.121",
	"54.122.132.214", "32.86.254.154", "251.216.110.169", "37.187.211.126", "123.3.4.204", "31.41.238.246", "11.168.50.177", "7.108.55.196", "146.109.179.10", "232.150.233.21",
	"148.127.195.183", "252.82.9.63", "4.40.141.98", "23.40.91.183", "60.159.206.65", "3.191.86.247", "198.98.170.236", "101.162.206.202", "134.182.29.253", "199.20.117.87",
	"131.203.85.24", "113.162.100.33", "74.86.215.140", "88.251.237.78", "60.125.148.70", "154.53.71.138", "62.6.28.94", "212.213.172.7", "183.224.162.194", "144.149.30.57",
}

/*
func TestMiddlewareBlackResult(t *testing.T) {
	tree := new(middleware.BlackNode)
	array := new(BlackNodeArray)
	for _, ip := range ips {
		tree.Insert(ip)
		array.Insert(ip)
	}
	for _, ip := range requests {
		if tree.Look(ip) != array.Look(ip) {
			t.Logf("tree: %t array: %t result not equal %d %s", tree.Look(ip), array.Look(ip), ip, int2ip(ip))
		}
	}
}

func BenchmarkMiddlewareBlackTree(b *testing.B) {
	node := new(middleware.BlackNode)
	for _, ip := range ips {
		node.Insert(ip)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, ip := range requests {
			node.Look(ip)
		}
	}
}
*/

func BenchmarkMiddlewareBlackArray(b *testing.B) {
	node := new(BlackNodeArray)
	b.ReportAllocs()
	for _, ip := range ips {
		node.Insert(ip)
	}
	for i := 0; i < b.N; i++ {
		for _, ip := range requests {
			node.Look(ip)
		}
	}
}

func TestMiddlewareBlackParseip(t *testing.T) {
	for _, ip := range ips {
		ip1, bit1 := ip2intbit(ip)
		ip2, bit2 := ip2netintbit(ip)
		if ip1 != ip2 || bit1 != bit2 {
			t.Log("ip parse error", ip, ip1, ip2, bit1, bit2)
		}
	}
}

func BenchmarkMiddlewareBlackIp2intbit(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, ip := range ips {
			ip2intbit(ip)
		}
	}
}

func BenchmarkMiddlewareBlackNetParse(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, ip := range ips {
			ip2netintbit(ip)
		}
	}
}

// BlackNodeArray 定义数组遍历实现ip解析
type BlackNodeArray struct {
	Data  []uint64
	Mask  []uint
	Count []uint64
}

// Insert 方法给黑名单节点新增一个ip或ip段。
func (node *BlackNodeArray) Insert(ip string) {
	iip, bit := ip2intbit(ip)
	node.Data = append(node.Data, iip>>(32-bit))
	node.Mask = append(node.Mask, 32-bit)
	node.Count = append(node.Count, 0)
}

// Look 方法匹配ip是否在黑名单节点，命中则节点计数加一。
func (node *BlackNodeArray) Look(ip uint64) bool {
	for i := range node.Data {
		if node.Data[i] == (ip >> node.Mask[i]) {
			node.Count[i]++
			return true
		}
	}
	return false
}

// BlackNodeArrayNet 定义基于net库实现ip遍历匹配，支持ipv6.
type BlackNodeArrayNet struct {
	Data  []net.IP
	Mask  []net.IPMask
	Count []uint64
}

// Insert 方法给黑名单节点新增一个ip或ip段。
func (node *BlackNodeArrayNet) Insert(ip string) {
	_, ipnet, _ := net.ParseCIDR(ip)
	node.Data = append(node.Data, ipnet.IP)
	node.Mask = append(node.Mask, ipnet.Mask)
	node.Count = append(node.Count, 0)
}

// Look 方法匹配ip是否在黑名单节点，命中则节点计数加一。
func (node *BlackNodeArrayNet) Look(ip string) bool {
	netip := net.ParseIP(ip)
	for i := range node.Data {
		if node.Data[i].Equal(netip.Mask(node.Mask[i])) {
			node.Count[i]++
			return true
		}
	}
	return false
}

func ip2netintbit(ip string) (uint64, uint) {
	ipaddr, ipnet, _ := net.ParseCIDR(ip)
	length := len(ipaddr)
	bit, _ := ipnet.Mask.Size()
	var sum uint64
	sum += uint64(ipaddr[length-4]) << 24
	sum += uint64(ipaddr[length-3]) << 16
	sum += uint64(ipaddr[length-2]) << 8
	sum += uint64(ipaddr[length-1])
	return sum, uint(bit)
}

func ip2intbit(ip string) (uint64, uint) {
	bit := 32
	pos := strings.Index(ip, "/")
	if pos != -1 {
		bit, _ = strconv.Atoi(ip[pos+1:])
		ip = ip[:pos]
	}
	return ip2int(ip), uint(bit)
}

func ip2int(ip string) uint64 {
	bits := strings.Split(ip, ".")
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum uint64
	sum += uint64(b0) << 24
	sum += uint64(b1) << 16
	sum += uint64(b2) << 8
	sum += uint64(b3)
	return sum
}

func int2ip(ip uint64) string {
	var bytes [4]uint64
	bytes[0] = ip & 0xFF
	bytes[1] = (ip >> 8) & 0xFF
	bytes[2] = (ip >> 16) & 0xFF
	bytes[3] = (ip >> 24) & 0xFF
	return fmt.Sprintf("%d.%d.%d.%d", bytes[3], bytes[2], bytes[1], bytes[0])
}

func BenchmarkMiddlewareRewrite(b *testing.B) {
	rewritedata := map[string]string{
		"/js/*":                    "/public/js/$0",
		"/api/v1/users/*/orders/*": "/api/v3/user/$0/order/$1",
		"/d/*":                     "/d/$0-$0",
		"/api/v1/*":                "/api/v3/$0",
		"/api/v2/*":                "/api/v3/$0",
		"/help/history*":           "/api/v3/history",
		"/help/history":            "/api/v3/history",
		"/help/*":                  "$0",
	}

	app := eudore.NewApp(eudore.NewLoggerInit())
	app.AddMiddleware("global", middleware.NewRewriteFunc(rewritedata))
	app.AnyFunc("/*", eudore.HandlerEmpty)
	paths := []string{"/", "/js/", "/js/index.js", "/api/v1/user", "/api/v1/user/new", "/api/v1/users/v3/orders/8920", "/api/v1/users/orders", "/api/v2", "/api/v2/user", "/d/3", "/help/history", "/help/historyv2"}
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			r.URL.Path = path
			app.ServeHTTP(w, r)
		}
	}
}
func BenchmarkMiddlewareRewriteWithZero(b *testing.B) {
	app := eudore.NewApp(eudore.NewLoggerInit())
	app.AnyFunc("/*", eudore.HandlerEmpty)
	paths := []string{"/", "/js/", "/js/index.js", "/api/v1/user", "/api/v1/user/new", "/api/v1/users/v3/orders/8920", "/api/v1/users/orders", "/api/v2", "/api/v2/user", "/d/3", "/help/history", "/help/historyv2"}
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			r.URL.Path = path
			app.ServeHTTP(w, r)
		}
	}
}

func BenchmarkMiddlewareRewriteWithRouter(b *testing.B) {
	routerdata := map[string]interface{}{
		"/js/*0":                     newRewriteFunc("/public/js/$0"),
		"/api/v1/users/:0/orders/*1": newRewriteFunc("/api/v3/user/$0/order/$1"),
		"/d/*0":                      newRewriteFunc("/d/$0-$0"),
		"/api/v1/*0":                 newRewriteFunc("/api/v3/$0"),
		"/api/v2/*0":                 newRewriteFunc("/api/v3/$0"),
		"/help/history*0":            newRewriteFunc("/api/v3/history"),
		"/help/history":              newRewriteFunc("/api/v3/history"),
		"/help/*0":                   newRewriteFunc("$0"),
	}
	app := eudore.NewApp(eudore.NewLoggerInit())
	app.AddMiddleware("global", middleware.NewRouterFunc(routerdata))
	app.AnyFunc("/*", eudore.HandlerEmpty)
	paths := []string{"/", "/js/", "/js/index.js", "/api/v1/user", "/api/v1/user/new", "/api/v1/users/v3/orders/8920", "/api/v1/users/orders", "/api/v2", "/api/v2/user", "/d/3", "/help/history", "/help/historyv2"}
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			r.URL.Path = path
			app.ServeHTTP(w, r)
		}
	}
}

func newRewriteFunc(path string) eudore.HandlerFunc {
	paths := strings.Split(path, "$")
	Index := make([]string, 1, len(paths)*2-1)
	Data := make([]string, 1, len(paths)*2-1)
	Index[0] = ""
	Data[0] = paths[0]
	for _, path := range paths[1:] {
		Index = append(Index, path[0:1])
		Data = append(Data, "")
		if path[1:] != "" {
			Index = append(Index, "")
			Data = append(Data, path[1:])
		}
	}
	return func(ctx eudore.Context) {
		buffer := bytes.NewBuffer(nil)
		for i := range Index {
			if Index[i] == "" {
				buffer.WriteString(Data[i])
			} else {
				buffer.WriteString(ctx.GetParam(Index[i]))
			}
		}
		ctx.Request().URL.Path = buffer.String()
	}
}
