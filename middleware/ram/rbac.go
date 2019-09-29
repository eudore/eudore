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

func (r *Rbac) BindRole(userid, roleid int) {
	r.RoleBinds[userid] = append(r.RoleBinds[userid], roleid)
}

func (r *Rbac) BindPermissions(roleid, permid int) {
	r.PermissionBinds[roleid] = append(r.PermissionBinds[roleid], permid)

}

// AddPermission 方法增加一个权限。
func (r *Rbac) AddPermission(id int, permid string) {
	r.Permissions[permid] = id
}

/*// NewRole 方法创建一个角色。
func (r *Rbac) NewRole(name string, perms []string) {
	r.Roles[name] = &Role{
		Name:  name,
		Binds: perms,
	}
}

// AddRoles 方法添加一个角色。
func (r *Rbac) AddRoles(name string, role *Role) {
	r.Roles[name] = role
}

// BindRoles 方法给指定用户绑定角色。
func (r *Rbac) BindRoles(id int, rolesname []string) {
	roles := make([]*Role, len(rolesname))
	for i, name := range rolesname {
		if role, ok := r.Roles[name]; ok {
			roles[i] = role
		}
	}
	r.Binds[id] = roles
}*/

// RamHandle 方法实现ram.RamHandler接口。
func (r *Rbac) RamHandle(id int, name string, ctx eudore.Context) (bool, bool) {
	ctx.SetParam(eudore.ParamRAM, "rbac")
	permid, ok := r.Permissions[name]
	if !ok {
		return false, false
	}
	for _, roles := range r.RoleBinds[id] {
		for _, perm := range r.PermissionBinds[roles] {
			if perm == permid {
				return true, true
			}
		}
	}
	return false, false
}
