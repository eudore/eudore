package ram

import (
	"github.com/eudore/eudore"
)

type (
	// Rbac 定义rbac对象。
	Rbac struct {
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

// BindRole 方法用户绑定一个角色。
func (r *Rbac) BindRole(userid, roleid int) {
	r.RoleBinds[userid] = append(r.RoleBinds[userid], roleid)
}

// BindPermissions 方法角色绑定一个权限id
func (r *Rbac) BindPermissions(roleid, permid int) {
	r.PermissionBinds[roleid] = append(r.PermissionBinds[roleid], permid)

}

// AddPermission 方法增加一个权限。
func (r *Rbac) AddPermission(id int, permid string) {
	r.Permissions[permid] = id
}

// Match 方法实现ram.Handler接口，匹配一个请求。
func (r *Rbac) Match(id int, name string, ctx eudore.Context) (bool, bool) {
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
