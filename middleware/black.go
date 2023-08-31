package middleware

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eudore/eudore"
)

// Black 定义黑名单中间件后台。
type black struct {
	White *BlackNode
	Black *BlackNode
}

// newBlack 函数创建一个黑名单后台。
func newBlack() *black {
	return &black{
		White: new(BlackNode),
		Black: new(BlackNode),
	}
}

// NewBlackFunc 函数创建一个黑名单处理函数，传入map[string]bool类型参数记录初始化使用的黑/白名单，白名单值为true/黑名单值为false。
func NewBlackFunc(data map[string]bool, router eudore.Router) eudore.HandlerFunc {
	b := newBlack()
	for k, v := range data {
		if v {
			b.InsertWhite(k)
		} else {
			b.InsertBlack(k)
		}
	}
	if router != nil {
		b.InjectRoutes(router)
	}
	return b.HandleHTTP
}

// InjectRoutes 方法将黑名单后台管理功能注入到路由器中。
func (b *black) InjectRoutes(router eudore.Router) {
	router.AnyFunc("/black/ui", HandlerAdmin)
	router.GetFunc("/black/data", b.data)
	router.PutFunc("/black/white/:ip", b.putIP)
	router.PutFunc("/black/black/:ip black=black", b.putIP)
	router.DeleteFunc("/black/white/:ip", b.DeleteAllIP)
	router.DeleteFunc("/black/black/:ip black=black", b.DeleteAllIP)
}

func (b *black) data(ctx eudore.Context) any {
	ctx.SetHeader(eudore.HeaderXEudoreAdmin, "black")
	return map[string]any{
		"white": b.White.List(nil, 0, 32),
		"black": b.Black.List(nil, 0, 32),
	}
}

func (b *black) putIP(ctx eudore.Context) {
	ip := fmt.Sprintf("%s/%s", ctx.GetParam("ip"), eudore.GetAnyByString(ctx.GetQuery("mask"), "32"))
	ctx.Infof("%s insert %s ip: %s", ctx.RealIP(), eudore.GetAnyByString(ctx.GetQuery("black"), "white"), ip)
	if ctx.GetParam("black") != "" {
		b.InsertBlack(ip)
	} else {
		b.InsertWhite(ip)
	}
}

func (b *black) DeleteAllIP(ctx eudore.Context) {
	ip := fmt.Sprintf("%s/%s", ctx.GetParam("ip"), eudore.GetAnyByString(ctx.GetQuery("mask"), "32"))
	ctx.Infof("%s DeleteAll %s ip: %s", ctx.RealIP(), eudore.GetAnyByString(ctx.GetQuery("black"), "white"), ip)
	if ctx.GetParam("black") != "" {
		b.DeleteAllBlack(ip)
	} else {
		b.DeleteAllWhite(ip)
	}
}

// HandleHTTP 方法定义黑名单后台的中间件处理函数。
func (b *black) HandleHTTP(ctx eudore.Context) {
	ip := ip2int(ctx.RealIP())
	if b.White.Look(ip) {
		return
	}
	if b.Black.Look(ip) {
		ctx.WriteHeader(eudore.StatusForbidden)
		ctx.WriteString("black list deny your ip " + ctx.RealIP())
		ctx.End()
	}
}

// InsertWhite 方法新增一个白名单ip或ip段。
func (b *black) InsertWhite(ip string) {
	b.White.Insert(ip)
}

// InsertBlack 方法新增一个黑名单ip或ip段。
func (b *black) InsertBlack(ip string) {
	b.Black.Insert(ip)
}

// DeleteAllWhite 方法删除一个白名单ip或ip段。
func (b *black) DeleteAllWhite(ip string) {
	b.White.DeleteAll(ip)
}

// DeleteAllBlack 方法删除一个黑名单ip或ip段。
func (b *black) DeleteAllBlack(ip string) {
	b.Black.DeleteAll(ip)
}

// BlackNode 定义黑名单存储树节点。
type BlackNode struct {
	Childrens [2]*BlackNode
	Data      bool
	Count     uint64
}

// BlackInfo 定义黑名单规则信息。
type BlackInfo struct {
	Addr  string `alias:"addr" json:"addr"`
	Mask  uint64 `alias:"mask" json:"mask"`
	Count uint64 `alias:"count" json:"count"`
}

// Insert 方法给黑名单节点新增一个ip或ip段。
func (node *BlackNode) Insert(ipstr string) {
	ip, bit := ip2intbit(ipstr)
	for num := uint(31); bit > 0; bit-- {
		i := ip >> num & 0x01
		if node.Childrens[i] == nil {
			node.Childrens[i] = new(BlackNode)
		}
		node = node.Childrens[i]
		num--
	}
	node.Data = true
}

// DeleteAll 方法给黑名单节点删除一个ip或ip段包全部子网段。
func (node *BlackNode) DeleteAll(ipstr string) {
	var lastnode *BlackNode
	var lastindex uint64
	rootnode := node
	ip, bit := ip2intbit(ipstr)
	for num := uint(31); bit > 0; bit-- {
		i := ip >> num & 0x01
		if node.Childrens[i] == nil {
			return
		}
		if node.Data || node.Childrens[1^i] != nil {
			lastnode = node
			lastindex = i
		}
		node = node.Childrens[i]
		num--
	}
	if lastnode != nil {
		lastnode.Childrens[lastindex] = nil
	} else {
		*rootnode = BlackNode{}
	}
}

// Look 方法匹配ip是否在黑名单节点，命中则节点计数加一。
func (node *BlackNode) Look(ip uint64) bool {
	for num := uint(32); num > 0; num-- {
		if node.Data {
			node.Count++
			return true
		}
		i := ip >> (num - 1) & 0x01
		if node.Childrens[i] == nil {
			return false
		}
		node = node.Childrens[i]
	}
	node.Count++
	return true
}

// List 方法递归获取全部黑名单规则信息。
func (node *BlackNode) List(data []BlackInfo, prefix, bit uint64) []BlackInfo {
	if node.Data {
		data = append(data, BlackInfo{
			Addr:  int2ip(prefix << bit),
			Mask:  32 - bit,
			Count: node.Count,
		})
	}
	for i, child := range node.Childrens {
		if child != nil {
			data = child.List(data, prefix<<1|uint64(i), bit-1)
		}
	}
	return data
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
	if len(bits) != 4 {
		return DefaultBlackInvalidAddress
	}
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
