package middleware

import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/eudore/eudore"
)

var (
	// Export to benchmark.
	SubnetListV4 = func() any { return new(subnetListV4) }
	SubnetListV6 = func() any { return new(subnetListV6) }
	ipv4Loopback = ip2int4("127.0.0.1")
	ipv6Loopback = ip2int6("::1")
)

type black struct {
	White4 subnetList
	Black4 subnetList
	White6 subnetList
	Black6 subnetList
}

// NewBlackListFunc function creates middleware to implement IP blacklist and
// whitelist matching
//
// When the blacklist is matched but the whitelist is not matched,
// [eudore.StatusForbidden] is returned.
//
// data key is CIDR, use bool value true for whitelist, false for blacklist.
//
// options: [NewOptionRouter].
func NewBlackListFunc(data map[string]bool, options ...Option) Middleware {
	b := &black{
		White4: new(subnetListV4),
		Black4: new(subnetListV4),
		White6: new(subnetListV6),
		Black6: new(subnetListV6),
	}
	applyOption(b, options)
	for k, v := range data {
		list, err := b.find(k, v)
		if err != nil {
			panic(err)
		}
		list.Insert(k)
	}

	return func(ctx eudore.Context) {
		ip := ctx.RealIP()
		if strings.IndexByte(ip, '.') > 0 {
			v := ip2int4(ip)
			if b.White4.Look4(v) || !b.Black4.Look4(v) {
				return
			}
		} else {
			v := ip2int6(ip)
			if b.White6.Look6(v) || !b.Black6.Look6(v) {
				return
			}
		}
		writePage(ctx, eudore.StatusForbidden, DefaultPageBlack, ip)
		ctx.End()
	}
}

func (b *black) find(cidr string, allow bool) (subnetList, error) {
	if strings.LastIndexByte(cidr, '/') == -1 {
		if strings.IndexByte(cidr, '.') == -1 {
			cidr += "/128"
		} else {
			cidr += "/32"
		}
	}
	prefix, err := netip.ParsePrefix(cidr)
	if prefix.Addr().Is4() {
		if allow {
			return b.White4, nil
		}
		return b.Black4, nil
	}
	if prefix.Addr().Is6() {
		if allow {
			return b.White6, nil
		}
		return b.Black6, nil
	}

	return nil, err
}

func (b *black) data() any {
	return map[string]any{
		"white4": b.White4.List(),
		"black4": b.Black4.List(),
		"white6": b.White6.List(),
		"black6": b.Black6.List(),
	}
}

func (b *black) putIP(ctx eudore.Context) {
	ip := getAddrMask(ctx.GetParam("ip"), ctx.GetQuery("mask"))
	ctx.Infof("%s Insert %s ip: %s", ctx.RealIP(), ctx.GetParam("list"), ip)
	list, err := b.find(ip, ctx.GetParam("list") == "white")
	if err != nil {
		ctx.Fatal(err)
		return
	}
	list.Insert(ip)
}

func (b *black) deleteIP(ctx eudore.Context) {
	ip := getAddrMask(ctx.GetParam("ip"), ctx.GetQuery("mask"))
	ctx.Infof("%s Delete %s ip: %s", ctx.RealIP(), ctx.GetParam("list"), ip)
	list, err := b.find(ip, ctx.GetParam("list") == "white")
	if err != nil {
		ctx.Fatal(err)
		return
	}
	list.Delete(ip)
}

func getAddrMask(ip, mask string) string {
	if mask == "" {
		return ip
	}
	return fmt.Sprintf("%s/%s", ip, mask)
}

type subnetList interface {
	Insert(ip string)
	Delete(ip string)
	Look4(ip uint32) bool
	Look6(ip [2]uint64) bool
	List() any
}

// subnetInfo defines CIDR rule data.
type subnetInfo struct {
	Addr  string `alias:"addr" json:"addr"`
	Mask  uint32 `alias:"mask" json:"mask"`
	Count uint64 `alias:"count" json:"count"`
}

type subnetListMutex struct {
	sync.RWMutex
	subnetList
}

func (list *subnetListMutex) Insert(ip string) {
	list.Lock()
	defer list.Unlock()
	list.subnetList.Insert(ip)
}

func (list *subnetListMutex) Delete(ip string) {
	list.Lock()
	defer list.Unlock()
	list.subnetList.Delete(ip)
}

func (list *subnetListMutex) Look4(ip uint32) bool {
	list.RLock()
	defer list.RUnlock()
	return list.subnetList.Look4(ip)
}

func (list *subnetListMutex) Look6(ip [2]uint64) bool {
	list.RLock()
	defer list.RUnlock()
	return list.subnetList.Look6(ip)
}

func (list *subnetListMutex) List() any {
	list.RLock()
	defer list.RUnlock()
	return list.subnetList.List()
}

type subnetListMixin struct {
	V4 subnetList
	V6 subnetList
}

func (list subnetListMixin) Insert(ip string) {
	if strings.IndexByte(ip, '.') > 0 {
		list.V4.Insert(ip)
	} else {
		list.V6.Insert(ip)
	}
}

func (list subnetListMixin) Look(ip string) bool {
	if strings.IndexByte(ip, '.') > 0 {
		return list.V4.Look4(ip2int4(ip))
	}
	return list.V6.Look6(ip2int6(ip))
}

// subnetListV4 defines the blacklist storage tree node.
type subnetListV4 struct {
	Childrens [2]*subnetListV4
	Data      bool
	Count     uint64
	Value     any
}

// The Insert method adds a CIDR to the blacklist node.
func (node *subnetListV4) Insert(addr string) {
	if strings.Contains(addr, ".") {
		ip, mask := ip2int4mask(addr)
		for start, end := 31, 32-mask; start >= end; start-- {
			bit := ip >> start & 0x01
			if node.Childrens[bit] == nil {
				node.Childrens[bit] = new(subnetListV4)
			}
			node = node.Childrens[bit]
		}
		node.Data = true
	}
}

// The Insert method deletes the CIDR of the blacklist node.
func (node *subnetListV4) Delete(addr string) {
	var nodes []*subnetListV4
	ip, mask := ip2int4mask(addr)
	for start, end := 31, 32-mask; start >= end; start-- {
		bit := ip >> start & 0x01
		if node.Childrens[bit] == nil {
			return
		}
		nodes = append(nodes, node)
		node = node.Childrens[bit]
	}
	node.Data = false
	nodes = append(nodes, node, new(subnetListV4))

	for i := len(nodes) - 2; i >= 0; i-- {
		node := nodes[i]
		if node.Data {
			return
		}
		if node.Childrens[0] == nodes[i+1] {
			node.Childrens[0] = nil
		}
		if node.Childrens[1] == nodes[i+1] {
			node.Childrens[1] = nil
		}
		if node.Childrens[0] != nil || node.Childrens[1] != nil {
			return
		}
	}
}

func (node *subnetListV4) Look4(ip uint32) bool {
	for mask := 31; mask >= 0; mask-- {
		if node.Data {
			atomic.AddUint64(&node.Count, 1)
			return true
		}
		bit := ip >> mask & 0x01
		if node.Childrens[bit] == nil {
			return false
		}
		node = node.Childrens[bit]
	}
	atomic.AddUint64(&node.Count, 1)
	return true
}

func (node *subnetListV4) Look6(ip [2]uint64) bool {
	if ip == ipv6Loopback {
		return node.Look4(ipv4Loopback)
	}
	return false
}

// The List method recursively get all blacklist rule data.
func (node *subnetListV4) List() any {
	return node.list32([]subnetInfo{}, 0, 32)
}

func (node *subnetListV4) list32(data []subnetInfo, prefix, bit uint32,
) []subnetInfo {
	if node.Data {
		data = append(data, subnetInfo{
			Addr:  int2ip4(prefix << bit),
			Mask:  32 - bit,
			Count: node.Count,
		})
	}
	for i, child := range node.Childrens {
		if child != nil {
			data = child.list32(data, prefix<<1|uint32(i), bit-1)
		}
	}
	return data
}

// subnetListV6 defines the blacklist storage tree node.
type subnetListV6 struct {
	Childrens [2]*subnetListV6
	Data      bool
	Count     uint64
}

func (node *subnetListV6) Insert(addr string) {
	if strings.Contains(addr, ":") {
		ip, mask := ip2int6mask(addr)
		for start, end := 127, 128-mask; start >= end; start-- {
			bit := ip.right(start)[1] & 0x01
			if node.Childrens[bit] == nil {
				node.Childrens[bit] = new(subnetListV6)
			}
			node = node.Childrens[bit]
		}
		node.Data = true
	}
}

func (node *subnetListV6) Delete(addr string) {
	var nodes []*subnetListV6
	ip, mask := ip2int6mask(addr)
	for start, end := 127, 128-mask; start >= end; start-- {
		bit := ip.right(start)[1] & 0x01
		if node.Childrens[bit] == nil {
			return
		}
		nodes = append(nodes, node)
		node = node.Childrens[bit]
	}
	node.Data = false
	nodes = append(nodes, node, new(subnetListV6))

	for i := len(nodes) - 2; i >= 0; i-- {
		node := nodes[i]
		if node.Data {
			return
		}
		if node.Childrens[0] == nodes[i+1] {
			node.Childrens[0] = nil
		}
		if node.Childrens[1] == nodes[i+1] {
			node.Childrens[1] = nil
		}
		if node.Childrens[0] != nil || node.Childrens[1] != nil {
			return
		}
	}
}

func (node *subnetListV6) Look4(uint32) bool {
	return false
}

func (node *subnetListV6) Look6(addr [2]uint64) bool {
	ip := uint128(addr)
	for mask := 127; mask >= 0; mask-- {
		if node.Data {
			atomic.AddUint64(&node.Count, 1)
			return true
		}
		bit := ip.right(mask)[1] & 0x01
		if node.Childrens[bit] == nil {
			return false
		}
		node = node.Childrens[bit]
	}
	atomic.AddUint64(&node.Count, 1)
	return true
}

func (node *subnetListV6) List() any {
	return node.list128([]subnetInfo{}, uint128{}, 128)
}

func (node *subnetListV6) list128(data []subnetInfo, prefix uint128, bit int,
) []subnetInfo {
	if node.Data {
		data = append(data, subnetInfo{
			Addr:  int2ip6(prefix.left(bit)),
			Mask:  uint32(128 - bit),
			Count: node.Count,
		})
	}
	for i, child := range node.Childrens {
		if child != nil {
			data = child.list128(data,
				prefix.left(1).or(uint128{0, uint64(i)}),
				bit-1,
			)
		}
	}
	return data
}

func ip2int4mask(ip string) (uint32, int) {
	bit := 32
	if pos := strings.LastIndexByte(ip, '/'); pos != -1 {
		v, err := strconv.Atoi(ip[pos+1:])
		if err == nil && -1 < v && v < bit {
			bit = v
		}
		ip = ip[:pos]
	}
	return ip2int4(ip), bit
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
	bytes := [4]byte{
		byte(ip>>24) & 0xFF,
		byte(ip>>16) & 0xFF,
		byte(ip>>8) & 0xFF,
		byte(ip) & 0xFF,
	}
	return netip.AddrFrom4(bytes).String()
}

func ip2int6mask(ipstr string) (uint128, int) {
	bit := 128
	if pos := strings.LastIndexByte(ipstr, '/'); pos != -1 {
		v, err := strconv.Atoi(ipstr[pos+1:])
		if err == nil && -1 < v && v < bit {
			bit = v
		}
		ipstr = ipstr[:pos]
	}
	return ip2int6(ipstr), bit
}

func ip2int6(ipstr string) [2]uint64 {
	pos := strings.Index(ipstr, "::")
	if pos > 0 {
		high, pos := parsehex(ipstr[:pos])
		low, _ := parsehex(ipstr[pos+2:])
		high = high.left(128 - pos*16)
		return [2]uint64{
			high[0] + low[0],
			high[1] + low[1],
		}
	}

	ip, _ := parsehex(ipstr)
	return ip
}

func parsehex(s string) (uint128, int) {
	var ip uint128
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
			ip = ip.left(16)
			ip[1] += val
			val = 0
			pos++
		}
	}
	ip = ip.left(16)
	ip[1] += val
	pos++
	return ip, pos
}

func int2ip6(ip [2]uint64) string {
	bytes := [16]byte{}
	binary.BigEndian.PutUint64(bytes[0:8], ip[0])
	binary.BigEndian.PutUint64(bytes[8:16], ip[1])
	return netip.AddrFrom16(bytes).String()
}

// uint128 is defined as a 128-bit integer, [0] represents the upper 64 bits,
// and [1] represents the lower 64 bits.
type uint128 [2]uint64

// Bitwise OR operation |.
func (u uint128) or(v uint128) uint128 {
	return uint128{u[0] | v[0], u[1] | v[1]}
}

// Left shift operation <<.
func (u uint128) left(n int) uint128 {
	switch {
	case n < 64:
		return uint128{(u[0] << n) | (u[1] >> (64 - n)), u[1] << n}
	case n < 128:
		return uint128{u[1] << (n - 64), 0}
	default:
		return uint128{0, 0}
	}
}

// Right shift operation >>.
func (u uint128) right(n int) uint128 {
	switch {
	case n < 64:
		return uint128{u[0] >> n, (u[1] >> n) | (u[0] << (64 - n))}
	default: // max 127
		return uint128{0, u[0] >> (n - 64)}
	}
}
