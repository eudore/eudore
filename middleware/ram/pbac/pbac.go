package pbac

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware/ram"
)

type (
	// Pbac 定义PBAC鉴权对象。
	Pbac struct {
		Binds   map[int][]int  `json:"-" key:"-"`
		Policys map[int]Policy `json:"-" key:"-"`
	}
	// Policy 定义一个策略。
	Policy struct {
		// Version string
		// Description string `json:"description",omitempty`
		Statement []Statement `json:"statement"`
	}
	// Statement 定义一条策略内容。
	Statement struct {
		Effect    bool
		Action    []string
		Resource  []string
		Condition Condition `json:"condition,omitempty"`
	}
	// Condition 定义PBAC使用的条件对象。
	Condition struct {
		Addr   []string
		Method []string
		Bash   bool
	}
	// Params 包含PBAC匹配需要的参数
	Params struct {
		Action   string
		Resource string
		Context  eudore.Context
	}
)

// NewEudoreParams 函数将行为和eudore请求上下文转换成PBAC条件参数。
func NewEudoreParams(action string, ctx eudore.Context) *Params {
	return &Params{
		Action:   action,
		Resource: ctx.Path(),
		Context:  ctx,
	}
}

// NewPbac 函数创建一个*Pbac对象
func NewPbac() *Pbac {
	return &Pbac{
		Binds:   make(map[int][]int),
		Policys: make(map[int]Policy),
	}
}

// Bind 方法给一个用户id绑定多个策略id
func (p *Pbac) Bind(id int, ps []int) {
	p.Binds[id] = append(p.Binds[id], ps...)
}

// BindString 方法给一个用户id绑定多个策略字符串id，使用逗号未分隔符
//
// p.BindString(0, "0,1,2")
func (p *Pbac) BindString(id int, str string) {
	strs := strings.Split(str, ",")
	var ps []int = make([]int, len(strs))
	for i, id := range strs {
		ps[i], _ = strconv.Atoi(id)
	}
	p.Binds[id] = append(p.Binds[id], ps...)
}

// AddPolicy 给PBAC添加一个策略。
func (p *Pbac) AddPolicy(id int, policy *Policy) {
	p.Policys[id] = *policy
}

// AddPolicyStringJson 给PBAC添加一个策略，策略类型是JSON字符串。
func (p *Pbac) AddPolicyStringJson(id int, str string) {
	policy := Policy{}
	if err := json.Unmarshal([]byte(str), &policy); err == nil {
		p.Policys[id] = policy
	} else {
		panic(err)
	}
}

// Handle 实现eduore请求上下文处理函数
func (p *Pbac) Handle(ctx eudore.Context) {
	ram.DefaultHandle(ctx, p)
}

// RamHandle 方法实现ram.RamHandler接口，匹配一个请求。
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

// Match 方法检查策略声明是否匹配
func (s *Statement) Match(param *Params) (bool, bool) {
	if MatchStarAny(param.Action, s.Action) &&
		MatchStarAny(param.Resource, s.Resource) &&
		s.Condition.Match(param.Context) {
		return s.Effect, true
	}
	return false, false
}

// Match 方法检查请求上下文条件是否匹配
func (c *Condition) Match(ctx eudore.Context) bool {
	if len(c.Addr) > 0 && !MatchAddrAny(ctx.RealIP(), c.Addr) {
		return false
	}
	if len(c.Method) > 0 && !MatchStringAny(ctx.Method(), c.Method) {
		return false
	}
	return true
}

// MatchStarAny 函数匹配对象是否匹配多个带星号的模式。
func MatchStarAny(obj string, patten []string) bool {
	for _, i := range patten {
		if eudore.MatchStar(obj, i) {
			return true
		}
	}
	return false
}

// MatchAddrAny 函数用来匹配ip地址是否在多个ip段中，当前未实现。
func MatchAddrAny(addr string, patten []string) bool {
	return addr == patten[0]
}

// MatchStringAny 函数匹配字符串是否在一个字符串数组中
func MatchStringAny(str string, strs []string) bool {
	for _, i := range strs {
		if str == i {
			return true
		}
	}
	return false
}
