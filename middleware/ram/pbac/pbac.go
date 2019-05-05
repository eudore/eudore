package pbac

import (
	"strconv"
	"strings"
	"encoding/json"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware/ram"
)

type (
	Pbac struct {
		Binds map[int][]int
		Policys	map[int]Policy
	}

	Policy struct {
		// Version string
		// Description string `json:"description",omitempty`
		Statement	[]Statement `json:"statement"`
	}
	Statement struct {
		Effect bool
		Action []string
		Resource []string
		Condition Condition `json:"condition",omitempty`
	}
	Condition struct {
		Addr []string
		Method []string
		Bash bool
	}

	Params struct {
		Action string
		Resource string
		Context eudore.Context
	}
)

func NewEudoreParams(action string, ctx eudore.Context) *Params {
	return &Params{
		Action:		action, 
		Resource:	ctx.Path(),
		Context:	ctx,
	}
}

func NewPbac() *Pbac{
	return &Pbac{
		Binds: 		make(map[int][]int),
		Policys:	make(map[int]Policy),
	}
}



func (p *Pbac) Bind(id int, ps []int) {
	p.Binds[id] = append(p.Binds[id], ps...)
}


func (p *Pbac) BindString(id int, str string) {
	strs := strings.Split(str, ",")
	var ps []int = make([]int, len(strs))
	for i, id := range strs {
		ps[i], _ = strconv.Atoi(id) 
	}
	p.Binds[id] = append(p.Binds[id], ps...)
}


func(p *Pbac) AddPolicy(id int, policy *Policy) {
	p.Policys[id] = *policy
}

func(p *Pbac) AddPolicyStringJson(id int, str string) {
	policy := Policy{}
	if err := json.Unmarshal([]byte(str), &policy); err == nil {
		p.Policys[id] = policy
	}else {
		panic(err)
	}
}

func (p *Pbac) Handle(ctx eudore.Context) {
	ram.DefaultHandle(ctx, p)
}

func (p *Pbac) RamHandle(id int, action string, ctx eudore.Context) (bool, bool) {
	ctx.SetParam(eudore.ParamRam, "pbac")
	param := NewEudoreParams(action, ctx)
	bs, ok := p.Binds[id]
	if ok {
		for _, b := range bs {
			ps, ok := p.Policys[b]
			if !ok {
				continue
			}
			for _, s := range ps.Statement {
				b, ok := s.Match(param)
				if ok {
					return b, true
				}
			}
		}
	}
	return false, false
}

// 检查策略声明是否匹配
func (s *Statement) Match(param *Params) (bool, bool) {
	if MatchStarAny(param.Action, s.Action) && 
		MatchStarAny(param.Resource, s.Resource) &&
		s.Condition.Match(param.Context) {
		return s.Effect, true
	}
	return false, false
}

// 检查条件是否匹配
func (c *Condition) Match(ctx eudore.Context) bool {
	if len(c.Addr) > 0 && !MatchAddrAny(ctx.RemoteAddr(), c.Addr) {
		return false
	}
	if len(c.Method) > 0 && !MatchStringAny(ctx.Method(), c.Method) {
		return false
	}
	return true
}


func MatchStarAny(obj string, patten []string) bool {
	for _, i := range patten {
		if eudore.MatchStar(obj, i) {
			return true
		}
	}
	return false
}

func MatchAddrAny(addr string, patten []string) bool {
	return addr == patten[0]
}

func MatchStringAny(str string, strs []string) bool {
	for _, i := range strs {
		if str == i {
			return true
		}
	}
	return false
}
