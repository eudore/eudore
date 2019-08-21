package rbac

import (
	"github.com/eudore/eudore"
)

type (
	// Role 定义一个rbac角色。
	Role struct {
		Name  string
		Binds []string
	}
	// Rbac 定义rbac对象。
	Rbac struct {
		Binds map[int][]*Role
		Roles map[string]*Role
	}
)

// NewRbac 函数创建一个Rbac的ram处理者。
func NewRbac() *Rbac {
	return &Rbac{
		Binds: make(map[int][]*Role),
		Roles: make(map[string]*Role),
	}
}

// NewRole 方法创建一个角色。
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
}

// RamHandle 方法实现ram.RamHandler接口。
func (r *Rbac) RamHandle(id int, name string, ctx eudore.Context) (bool, bool) {
	ctx.SetParam(eudore.ParamRam, "rbac")
	for _, i := range r.Binds[id] {
		if ok := i.Match(name); ok {
			return true, true
		}
	}
	return false, false
}

// Match 方法实现角色匹配一个行为。
func (r *Role) Match(name string) bool {
	for _, i := range r.Binds {
		if i == name {
			return true
		}
	}
	return false
}
