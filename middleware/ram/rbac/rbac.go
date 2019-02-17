package rbac

import (
	"github.com/eudore/eudore"
)

type (
	Role struct {
		Name 	string
		Binds 	[]string
	}
	Rbac struct {
		Binds	map[int][]*Role
		Roles	map[string]*Role
	}
)


func NewRbac() *Rbac {
	return &Rbac{
		Binds:	make(map[int][]*Role),
		Roles:	make(map[string]*Role),
	}
}

func (r *Rbac) NewRole(name string, perms []string) {
	r.Roles[name] = &Role{
		Name:	name,
		Binds:	perms,
	}
}

func (r *Rbac) AddRoles(name string,role *Role) {
	r.Roles[name] = role
}

func (r *Rbac) BindRoles(id int, rolesname []string) {
	roles := make([]*Role, len(rolesname))
	for i, name := range rolesname {
		if role, ok := r.Roles[name]; ok {
			roles[i] = role
		}
	}
	r.Binds[id] = roles
}

func (r *Rbac) RamHandle(id int, name string, ctx eudore.Context) (bool, bool) {
	ctx.SetParam(eudore.ParamRam, "rbac")
	for _, i := range r.Binds[id] {
		if ok := i.Match(name); ok {
			return true, true
		}
	}
	return false, false
}

func (r *Role) Match(name string) bool {
	for _, i := range r.Binds {
		if i == name {
			return true
		}
	}
	return false
}