package ram

import (
	"github.com/eudore/eudore"
	"sync"
)

type (
	// Rbac 定义rbac对象。
	Rbac struct {
		sync.RWMutex
		RoleBinds       map[int][]int
		PermissionBinds map[int][]int
		Permissions     map[string]int
	}
)

// NewRbac 函数创建一个Rbac的ram处理者。
func NewRbac() *Rbac {
	return &Rbac{
		RoleBinds:       make(map[int][]int),
		PermissionBinds: make(map[int][]int),
		Permissions:     make(map[string]int),
	}
}

// Match 方法实现ram.Handler接口，匹配一个请求。
func (r *Rbac) Match(id int, name string, ctx eudore.Context) (bool, bool) {
	r.RLock()
	defer r.RUnlock()
	permid, ok := r.Permissions[name]
	if !ok {
		return false, false
	}
	for _, roles := range r.RoleBinds[id] {
		for _, perm := range r.PermissionBinds[roles] {
			if perm == permid {
				ctx.SetParam(eudore.ParamRAM, "rbac")
				return true, true
			}
		}
	}
	return false, false
}

// AddPermission 方法增加一个权限。
func (r *Rbac) AddPermission(id int, perm string) {
	r.Lock()
	defer r.Unlock()
	r.Permissions[perm] = id
}

// DeletePermission 方法删除指定权限。
func (r *Rbac) DeletePermission(perm string) {
	r.Lock()
	defer r.Unlock()
	delete(r.Permissions, perm)
}

// BindRole 方法用户绑定一个角色。
func (r *Rbac) BindRole(userid int, roleid ...int) {
	r.Lock()
	defer r.Unlock()
	r.RoleBinds[userid] = append(r.RoleBinds[userid], roleid...)
}

// UnnindRole 方法解除用户绑定的一个角色
func (r *Rbac) UnnindRole(userid int, roleid ...int) {
	r.Lock()
	defer r.Unlock()
	for _, id := range roleid {
		r.RoleBinds[userid] = removeSlice(r.RoleBinds[userid], id)
	}
}

// BindPermissions 方法角色绑定一个权限id
func (r *Rbac) BindPermissions(roleid int, permid ...int) {
	r.Lock()
	defer r.Unlock()
	r.PermissionBinds[roleid] = append(r.PermissionBinds[roleid], permid...)
}

// UnbindPermissions 方法解除角色绑定的一个权限
func (r *Rbac) UnbindPermissions(roleid int, permid ...int) {
	r.Lock()
	defer r.Unlock()
	for _, id := range permid {
		r.PermissionBinds[roleid] = removeSlice(r.PermissionBinds[roleid], id)
	}
}

func removeSlice(s []int, n int) []int {
	for i, val := range s {
		if val == n {
			s[i] = s[len(s)-1]
			s = s[:len(s)-1]
			return s
		}
	}
	return s
}
