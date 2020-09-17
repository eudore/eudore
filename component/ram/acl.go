package ram

import (
	"github.com/eudore/eudore"
	"sync"
)

// Acl 是acl权限鉴权对象
type Acl struct {
	sync.RWMutex
	AllowBinds  map[int]map[int]struct{}
	DenyBinds   map[int]map[int]struct{}
	Permissions map[string]int
}

var (
	empty = struct{}{}
)

// NewAcl 函数创建一个Acl对象。
func NewAcl() *Acl {
	return &Acl{
		AllowBinds:  make(map[int]map[int]struct{}),
		DenyBinds:   make(map[int]map[int]struct{}),
		Permissions: make(map[string]int),
	}
}

// Name 方法返回acl name。
func (acl *Acl) Name() string {
	return "acl"
}

// Match 方法实现ram.Handler接口，匹配一个请求。
func (acl *Acl) Match(id int, perm string, ctx eudore.Context) (bool, bool) {
	acl.RLock()
	defer acl.RUnlock()
	permid, ok := acl.Permissions[perm]
	if ok {
		_, ok = acl.AllowBinds[id][permid]
		if ok {
			return true, true
		}

		_, ok = acl.DenyBinds[id][permid]
		if ok {
			return false, true
		}
	}

	return false, false
}

// AddPermission 方法增加一个权限。
func (acl *Acl) AddPermission(id int, perm string) {
	acl.Lock()
	defer acl.Unlock()
	acl.Permissions[perm] = id
}

// DeletePermission 方法删除一个权限。
func (acl *Acl) DeletePermission(perm string) {
	acl.Lock()
	defer acl.Unlock()
	delete(acl.Permissions, perm)
}

// BindPermission 方法绑定一个权限。
func (acl *Acl) BindPermission(id int, permid int, allow bool) {
	if allow {
		acl.BindAllowPermission(id, permid)
	} else {
		acl.BindDenyPermission(id, permid)
	}
}

// BindAllowPermission 方法给指定用户id添加允许的权限
func (acl *Acl) BindAllowPermission(id int, permid int) {
	acl.Lock()
	defer acl.Unlock()
	ps, ok := acl.AllowBinds[id]
	if !ok {
		ps = make(map[int]struct{})
		acl.AllowBinds[id] = ps
	}
	ps[permid] = empty
}

// BindDenyPermission 方法给指定用户id添加拒绝的权限
func (acl *Acl) BindDenyPermission(id int, permid int) {
	acl.Lock()
	defer acl.Unlock()
	ps, ok := acl.DenyBinds[id]
	if !ok {
		ps = make(map[int]struct{})
		acl.DenyBinds[id] = ps
	}
	ps[permid] = empty
}

// UnbindPermission 方法删除指定用户id的权限。
func (acl *Acl) UnbindPermission(id int, permid int) {
	acl.UnbindAllowPermission(id, permid)
	acl.UnbindDenyPermission(id, permid)
}

// UnbindAllowPermission 方法删除指定用户id的允许权限。
func (acl *Acl) UnbindAllowPermission(id int, permid int) {
	acl.Lock()
	defer acl.Unlock()
	ps, ok := acl.AllowBinds[id]
	if ok {
		delete(ps, permid)
		if len(ps) == 0 {
			delete(acl.AllowBinds, id)
		}
	}
}

// UnbindDenyPermission 方法删除指定用户id的拒绝权限。
func (acl *Acl) UnbindDenyPermission(id int, permid int) {
	acl.Lock()
	defer acl.Unlock()
	ps, ok := acl.DenyBinds[id]
	if ok {
		delete(ps, permid)
		if len(ps) == 0 {
			delete(acl.DenyBinds, id)
		}
	}
}
