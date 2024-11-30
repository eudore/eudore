package eudore_test

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"testing"

	"github.com/eudore/eudore/middleware"
)

type SubnetList interface {
	Insert(string)
	Delete(string)
	Look4(uint32) bool
	Look6([2]uint64) bool
}

var (
	nets4   = &IPNets{hits: make(map[string]struct{})}
	node4   = middleware.SubnetListV4().(SubnetList)
	subnet4 = []string{
		"10.0.0.0/8", "127.0.0.1/8", "192.168.1.0/24", "192.168.75.0/24", "192.168.100.0/24",
	}
	matchs4 = [1000]string{}
	matchi4 = [1000]uint32{}
	keys4   = make(map[string]int)
	nets6   = &IPNets{hits: make(map[string]struct{})}
	node6   = middleware.SubnetListV6().(SubnetList)
	subnet6 = []string{
		"0a00:0000::/28", "7f00:0001::/28", "c0a8:0100::/24", "c0a8:4b00::/24", "c0a8:6400::/24",
	}
	matchs6 = [1000]string{}
	matchi6 = [1000][2]uint64{}
	keys6   = make(map[string]int)
)

func init() {
	for i := range subnet4 {
		nets4.Insert(subnet4[i])
		node4.Insert(subnet4[i])
	}
	for i := range subnet6 {
		nets6.Insert(subnet6[i])
		node6.Insert(subnet6[i])
	}

	for i := range matchs4 {
		matchi4[i] = rand.Uint32()
		matchs4[i] = int2ip4(matchi4[i])
		keys4[matchs4[i]] = i
		matchi6[i] = [2]uint64{rand.Uint64(), rand.Uint64()}
		matchs6[i] = int2ip6(matchi6[i])
		keys6[matchs6[i]] = i
		if i < 20 {
			nets4.Insert(matchs4[i] + "/24")
			node4.Insert(matchs4[i] + "/24")
		}
		if i < 20 {
			nets6.Insert(matchs6[i] + "/24")
			node6.Insert(matchs6[i] + "/24")
		}
	}

	del4 := [10000]string{}
	for i := range del4 {
		for {
			del4[i] = int2ip4(rand.Uint32())
			_, ok := keys4[del4[i]]
			if !ok {
				break
			}
		}
	}
	for i := range del4 {
		node4.Insert(del4[i])
	}
	for i := range del4 {
		node4.Delete(del4[i])
	}
	del6 := [10000]string{}
	for i := range del6 {
		for {
			del6[i] = int2ip6([2]uint64{rand.Uint64(), rand.Uint64()})
			_, ok := keys6[del6[i]]
			if !ok {
				break
			}
		}
	}
	for i := range del6 {
		node6.Insert(del6[i])
	}
	for i := range del6 {
		node6.Delete(del6[i])
	}
}

func BenchmarkBlackCompare4(b *testing.B) {
	b.ReportAllocs()
	fn := func(ip uint32, addr string) {
		a := nets4.Look(addr)
		b := node4.Look4(ip)
		if a != b {
			panic(fmt.Sprintf("ip %s %t %t", addr, a, b))
		}
	}
	for i := 0; i < b.N; i++ {
		for i := range matchs4 {
			fn(matchi4[i], matchs4[i])
		}
		for i := 0; i < 1000; i++ {
			ip := rand.Uint32()
			fn(ip, int2ip4(ip))
		}
	}
}

func BenchmarkBlackNetsString4(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for i := range matchs4 {
			nets4.Look(matchs4[i])
		}
	}
}

func BenchmarkBlackNodeString4(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for i := range matchs4 {
			node4.Look4(ip2int4(matchs4[i]))
		}
	}
}

func BenchmarkBlackNodeInteger4(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for i := range matchi4 {
			node4.Look4(matchi4[i])
		}
	}
}

func BenchmarkBlackCompare6(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for i := range matchs6 {
			a := nets6.Look(matchs6[i])
			b := node6.Look6(matchi6[i])
			if a != b {
				panic(fmt.Sprintf("ip %s %t %t", matchs6[i], a, b))
			}
		}
	}
}

func BenchmarkBlackNetsString6(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for i := range matchs6 {
			nets6.Look(matchs6[i])
		}
	}
}

func BenchmarkBlackNodeString6(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for i := range matchs6 {
			node6.Look6(ip2int6(matchs6[i]))
		}
	}
}

func BenchmarkBlackNodeInteger6(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for i := range matchi6 {
			node6.Look6(matchi6[i])
		}
	}
}

type IPNets struct {
	nets []*net.IPNet
	hits map[string]struct{}
}

func (n *IPNets) Insert(str string) {
	if strings.Contains(str, ".") && !strings.Contains(str, "/") {
		n.hits[str] = struct{}{}
		return
	} else if strings.Contains(str, ":") && !strings.Contains(str, "/") {
		n.hits[str] = struct{}{}
		return
	}
	_, net, err := net.ParseCIDR(str)
	if err == nil {
		n.nets = append(n.nets, net)
	}
}

func (n *IPNets) Look(str string) bool {
	_, ok := n.hits[str]
	if ok {
		return true
	}
	ip := net.ParseIP(str)
	for i := range n.nets {
		if n.nets[i].Contains(ip) {
			return true
		}
	}
	return false
}

func ip2int4(ip string) uint32 {
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

func int2ip4(ip uint32) string {
	bytes := [4]uint32{
		(ip >> 24) & 0xFF,
		(ip >> 16) & 0xFF,
		(ip >> 8) & 0xFF,
		ip & 0xFF,
	}
	return fmt.Sprintf("%d.%d.%d.%d", bytes[0], bytes[1], bytes[2], bytes[3])
}

func ip2int6(ipstr string) [2]uint64 {
	pos := strings.Index(ipstr, "::")
	if pos > 0 {
		high, pos := parsehex(ipstr[:pos])
		low, _ := parsehex(ipstr[pos+2:])
		high = lShift(high, 128-uint8(pos)*16)
		return [2]uint64{
			high[0] + low[0],
			high[1] + low[1],
		}
	}

	ip, _ := parsehex(ipstr)
	return ip
}

func parsehex(s string) ([2]uint64, int) {
	var ip [2]uint64
	var val uint64
	pos := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			val = val<<4 + uint64(c-'0')
		case c >= 'a' && c <= 'f':
			val = val<<4 + uint64(c-'a'+10)
		case c >= 'A' && c <= 'F':
			val = val<<4 + uint64(c-'A'+10)
		case c == ':':
			ip = lShift(ip, 16)
			ip[1] += val
			val = 0
			pos++
		}
	}
	ip = lShift(ip, 16)
	ip[1] += val
	pos++
	return ip, pos
}

func int2ip6(ip [2]uint64) string {
	bytes := [8]uint64{
		(ip[0] >> 48) & 0xFFFF,
		(ip[0] >> 32) & 0xFFFF,
		(ip[0] >> 16) & 0xFFFF,
		ip[0] & 0xFFFF,
		(ip[1] >> 48) & 0xFFFF,
		(ip[1] >> 32) & 0xFFFF,
		(ip[1] >> 16) & 0xFFFF,
		ip[1] & 0xFFFF,
	}
	return fmt.Sprintf("%x:%x:%x:%x:%x:%x:%x:%x",
		bytes[0], bytes[1], bytes[2], bytes[3],
		bytes[4], bytes[5], bytes[6], bytes[7],
	)
}

func lShift(u [2]uint64, n uint8) [2]uint64 {
	if n >= 128 {
		// 如果移位大于等于 128 位，结果为 0
		return [2]uint64{0, 0}
	} else if n >= 64 {
		// 如果移位大于等于 64 位，则将低 64 位移动到高位
		return [2]uint64{u[1] << (n - 64), 0}
	}
	// 常规左移处理：低位左移 n 位，并将高位加上从低位移动过来的部分
	return [2]uint64{(u[0] << n) | (u[1] >> (64 - n)), u[1] << n}
}
