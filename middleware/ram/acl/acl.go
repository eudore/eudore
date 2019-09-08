package acl

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware/ram"
)

// Acl 是acl权限鉴权对象
type Acl struct {
	Data map[int]map[string]bool `json:"-" key:"-"`
}

var (
	empty = struct{}{}
)

// NewAcl 函数创建一个Acl对象。
func NewAcl() *Acl {
	return &Acl{
		Data: make(map[int]map[string]bool),
	}
}

// Handle 实现eduore请求上下文处理函数
func (a *Acl) Handle(ctx eudore.Context) {
	ram.DefaultHandle(ctx, a)
}

// RamHandle 方法实现ram.RamHandler接口，匹配一个请求。
func (a *Acl) RamHandle(id int, perm string, ctx eudore.Context) (bool, bool) {
	ctx.SetParam(eudore.ParamRAM, "acl")
	ps, ok := a.Data[id]
	if ok {
		isgrant, ok := ps[perm]
		if ok {
			return isgrant, true
		}
	}
	return false, false
}

// AddPermission 方法给ACL指定用户添加多项权限
func (a *Acl) AddPermission(id int, perms []string, allow bool) {
	ps, ok := a.Data[id]
	if !ok {
		ps = make(map[string]bool)
		a.Data[id] = ps
	}
	for _, p := range perms {
		ps[p] = allow
	}
}

// AddAllowPermission 方法给指定用户id添加允许的权限
func (a *Acl) AddAllowPermission(id int, perms []string) {
	a.AddPermission(id, perms, true)
}

// AddDenyPermission 方法给指定用户id添加拒绝的权限
func (a *Acl) AddDenyPermission(id int, perms []string) {
	a.AddPermission(id, perms, false)
}

// DelPermission 放删除指定用户id的多项权限。
func (a *Acl) DelPermission(id int, perms []string) {
	ps, ok := a.Data[id]
	if ok {
		for _, p := range perms {
			delete(ps, p)
		}
		if len(ps) == 0 {
			delete(a.Data, id)
		}
	}
}
