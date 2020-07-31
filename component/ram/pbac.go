package ram

import (
	"strconv"
	"strings"
	"sync"

	"github.com/eudore/eudore"
)

type (
	// Pbac 定义PBAC鉴权对象。
	Pbac struct {
		sync.RWMutex
		PolicyBinds  map[int][]int
		PolicyIndexs map[int][]int
		Policys      map[int]*Policy
		GetResource  func(eudore.Context) string `json:"-" alice:"-"`
	}
)

// NewPbac 函数创建一个*Pbac对象
func NewPbac() *Pbac {
	return &Pbac{
		PolicyBinds:  make(map[int][]int),
		PolicyIndexs: make(map[int][]int),
		Policys:      make(map[int]*Policy),
		GetResource:  getResource,
	}
}

// Match 方法实现ram.Handler接口，匹配一个请求。
func (p *Pbac) Match(id int, action string, ctx eudore.Context) (bool, bool) {
	p.RLock()
	defer p.RUnlock()
	resource := p.GetResource(ctx)
	bs, ok := p.PolicyBinds[id]
	if ok {
		for _, b := range bs {
			ps, ok := p.Policys[b]
			if !ok {
				continue
			}
			for _, s := range ps.Statement {
				if s.MatchAction(action) && s.MatchResource(resource) && s.MatchCondition(ctx) {
					ctx.SetParam(eudore.ParamRAM, "pbac")
					ctx.SetParam("policy", ps.Name)
					return s.Effect, true
				}
			}
		}
	}
	return false, false
}

func getResource(ctx eudore.Context) string {
	path := ctx.Path()
	// 移除无效的前缀
	prefix := ctx.GetParam("prefix")
	if prefix != "" {
		path = path[len(prefix):]
	}
	ctx.SetParam("resource", path)
	return path
}

// BindPolicy 方法给一个用户id绑定一个策略id
func (p *Pbac) BindPolicy(id, index, policyid int) {
	p.Lock()
	defer p.Unlock()
	binds := append(p.PolicyBinds[id], policyid)
	indexs := append(p.PolicyIndexs[id], index)
	for i := len(binds) - 1; i > 0; i-- {
		if indexs[i] > indexs[i-1] {
			binds[i], binds[i-1] = binds[i-1], binds[i]
			indexs[i], indexs[i-1] = indexs[i-1], indexs[i]
		}
	}
	p.PolicyBinds[id] = binds
	p.PolicyIndexs[id] = indexs
}

// BindPolicyString 方法给一个用户id绑定多个策略字符串id，使用逗号未分隔符
//
// p.BindPolicyString(0, "0,1,2")
func (p *Pbac) BindPolicyString(id, index int, str string) {
	strs := strings.Split(str, ",")
	for _, i := range strs {
		n, err := strconv.Atoi(i)
		if err == nil {
			p.BindPolicy(id, index, n)
		}
	}
}

// AddPolicy 给PBAC添加一个策略。
func (p *Pbac) AddPolicy(id int, policy *Policy) {
	p.Lock()
	defer p.Unlock()
	p.Policys[id] = policy
}

// AddPolicyStringJson 方法给PBAC添加一个策略，策略类型是JSON字符串。
func (p *Pbac) AddPolicyStringJson(id int, str string) error {
	policy, err := ParsePolicyString(str)
	if err == nil {
		p.AddPolicy(id, policy)
	}
	return err
}

// DeletePolicy 方法删除一个指定id的测试。
func (p *Pbac) DeletePolicy(id int) {
	p.Lock()
	defer p.Unlock()
	delete(p.Policys, id)
}
