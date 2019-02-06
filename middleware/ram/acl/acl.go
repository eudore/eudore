package acl

import (
	"eudore"
	"eudore/middleware/ram"
)

type Acl struct {
	Data map[int]map[string]bool
}

var (
	empty 				=	struct{}{}
)


func NewAcl() *Acl {
	return &Acl{
		Data:			make(map[int]map[string]bool),
	}
}

func (a *Acl) Handle(ctx eudore.Context) {
	ram.DefaultHandle(ctx, a)
}

func (a *Acl) RamHandle(id int, perm string, ctx eudore.Context) (bool, bool) {
	ctx.SetParam(eudore.ParamRam, "acl")
	ps, ok := a.Data[id]
	if ok {
		isgrant, ok := ps[perm]
		if ok {
			return isgrant, true
		}
	}
	return false, false
}

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

func (a *Acl) AddAllowPermission(id int, perms []string) {
	a.AddPermission(id, perms, true)
}

func (a *Acl) AddDenyPermission(id int, perms []string) {
	a.AddPermission(id, perms, false)
}



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
