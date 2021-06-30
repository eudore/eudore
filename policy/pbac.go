/*
Package policy 实现基于策略访问控制。

pbac 通过策略限制访问权限，每个策略拥有多条描述，按照顺序依次匹配，命中则执行effect。

pbac条件直接使用And关系,允许使用多种多样的方法限制请求，额外条件可以使用policy.RegisterCondition函数注册条件。

如果一个策略Statement的Data属性不为空，则为数据权限，在没有匹配到一个非数据权限时会通过鉴权，保存多数据权限的Data Expression，用户对指定表操作时生产对应的表数据限定sql。
*/
package policy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

// Policys 定义策略访问控制器。
type Policys struct {
	Policys sync.Map // policyid:policy
	Members sync.Map // userid:policys
	// 默认使用jwt sha1进行签名，默认值：NewSignaturerJwt([]byte("eudore")),
	Signaturer Signaturer
	// 获取请求action的函数
	// 默认返回action param
	ActionFunc func(eudore.Context) string
	// 获取请求资源的函数
	// 默认移除请求路径的前缀"/api/"为资源路径
	ResourceFunc func(eudore.Context) string
	// 获取用户id的函数
	// 默认从请求Authorization header获取jwt Bearer数据解析用户信息。
	GetUserFunc func(eudore.Context) (int, error)
	// 权限拒绝时执行的函数
	// 默认返回403和相关权限信息
	ForbendFunc func(eudore.Context, string, string, string)
}

// Signaturer 定义Policys进行用户信息签名的对象。
type Signaturer interface {
	Signed(interface{}) string
	Parse(string, interface{}) error
}

// Member 定义Policy授权对象。
type Member struct {
	UserID   int `json:"user_id" alias:"user_id" gorm:"primaryKey"`
	PolicyID int `json:"policy_id" alias:"policy_id" gorm:"primaryKey"`
	// Index 越大优先级越高，如果小于0移除授权
	Index       int       `json:"index" alias:"index"`
	Description string    `json:"description" alias:"description"`
	Expiration  time.Time `json:"expiration" alias:"expiration"`
	policy      *Policy
}

// NewPolicys 函数创建默认策略访问控制器
func NewPolicys() *Policys {
	policys := &Policys{
		Signaturer: NewSignaturerJwt([]byte("eudore")),
		ActionFunc: func(ctx eudore.Context) string {
			return ctx.GetParam("action")
		},
		ResourceFunc: func(ctx eudore.Context) string {
			return strings.TrimPrefix(ctx.Path(), "/api/")
		},
	}
	policys.ForbendFunc = policys.handleForbidden
	policys.GetUserFunc = policys.parseSignatureUser
	return policys
}

// HandleHTTP 方法实现eudore.handlerHTTP(handler.go#L49)接口，作为请求处理中间件的处理函数，实现访问控制鉴权。
//
// 请求的param action为空回跳过鉴权方法。
func (ctl *Policys) HandleHTTP(ctx eudore.Context) {
	action := ctl.ActionFunc(ctx)
	if action == "" {
		return
	}
	resource := ctl.ResourceFunc(ctx)
	ctx.SetParam("Resource", resource)

	// 获取用户信息。
	userid, err := ctl.GetUserFunc(ctx)
	if err != nil {
		ctl.ForbendFunc(ctx, action, resource, err.Error())
		return
	}
	ctx.SetParam("Userid", fmt.Sprint(userid))

	fmt.Println("pbac", userid, ctl.getMemberByUser(userid))
	var now = time.Now()
	var datas Expressions
	var names []string

	// 遍历用户的全部授权的全部stmt
matchPolicys:
	for _, m := range ctl.getMemberByUser(userid) {
		p := m.policy
		fmt.Println(p.PolicyName, p.PolicyID)
		if !m.Expiration.IsZero() && m.Expiration.Before(now) {
			fmt.Println("expiration")
			continue
		}

		for _, s := range m.policy.statement {
			ok := s.MatchAction(action) && s.MatchResource(resource) && s.MatchCondition(ctx)
			fmt.Println(p.PolicyName, s.MatchAction(action), s.MatchResource(resource), s.MatchCondition(ctx), s.Data)
			if ok {
				// 非数据权限执行行为
				if s.Data == nil {
					ctx.SetParam("Policy", m.policy.PolicyName)
					if !s.Effect {
						ctl.ForbendFunc(ctx, action, resource, "")
					}
					return
				}
				names = append(names, m.policy.PolicyName)
				datas = append(datas, s.Data...)
			}
		}
	}

	if userid != 0 {
		userid = 0
		goto matchPolicys
	}
	// 数据权限
	if datas != nil {
		ctx.SetParam("Policy", strings.Join(names, ","))
		ctx.WithContext(context.WithValue(ctx.GetContext(), PolicyExpressions, datas))
		return
	}
	ctl.ForbendFunc(ctx, action, resource, fmt.Sprintf("User %s not match policys.", ctx.GetParam("Userid")))
}

func (ctl *Policys) getMemberByUser(userid int) []*Member {
	ps, ok := ctl.Members.Load(userid)
	if ok {
		return ps.([]*Member)
	}
	return nil
}

// HandleRuntime 方法返回Policys运行时数据。
func (ctl *Policys) HandleRuntime(ctx eudore.Context) interface{} {
	var policys []Policy
	ctl.Policys.Range(func(key, val interface{}) bool {
		policys = append(policys, *val.(*Policy))
		return true
	})
	sort.Slice(policys, func(i, j int) bool {
		return policys[i].PolicyID < policys[j].PolicyID
	})

	members := make(map[int][]*Member)
	ctl.Members.Range(func(key, val interface{}) bool {
		members[key.(int)] = val.([]*Member)
		return true
	})
	return struct {
		Policys []Policy    `json:"policys"`
		Members interface{} `json:"members"`
	}{policys, members}
}

type forbiddenMessage struct {
	Status   int    `json:"status"`
	Message  string `json:"message"`
	Action   string `json:"action"`
	Resource string `json:"resource"`
	Error    string `json:"error,omitempty"`
}

func (ctl *Policys) handleForbidden(ctx eudore.Context, action, resource, err string) {
	ctx.WriteHeader(403)
	ctx.Render(forbiddenMessage{
		Status:   403,
		Message:  "forbidden",
		Action:   action,
		Resource: resource,
		Error:    err,
	})
	ctx.End()
}

const stringBearer = "Bearer "

// SignatureUser 定义默认的用户信息，也可以组合该对象使用自定义签名对象。
type SignatureUser struct {
	// 唯一必要的属性，指定请求的userid
	UserID int `json:"userid" alias:"userid"`
	// 日志显示使用
	UserName string `json:"username" alias:"username"`
	// 如果非空，则为base64([]Statement)
	Policy     string `json:"policy" alias:"policy,omitempty"`
	Expiration int64  `json:"expiration" alias:"expiration"`
}

// NewBearer 默认的Bearer签名方法。
func (ctl *Policys) NewBearer(userid int, name, policy string, expires int64) string {
	return stringBearer + ctl.Signaturer.Signed(&SignatureUser{
		UserID:     userid,
		UserName:   name,
		Policy:     base64.StdEncoding.EncodeToString([]byte(policy)),
		Expiration: expires,
	})
}

func (ctl *Policys) parseSignatureUser(ctx eudore.Context) (int, error) {
	bearer := ctx.GetHeader(eudore.HeaderAuthorization)
	if bearer == "" {
		return 0, nil
	}
	if !strings.HasPrefix(bearer, stringBearer) {
		return 0, nil
	}

	// 验证bearer
	var user SignatureUser
	err := ctl.Signaturer.Parse(bearer[7:], &user)
	if err != nil {
		return 0, fmt.Errorf("bearer parse error: %s", err.Error())
	}
	if user.Expiration < time.Now().Unix() {
		return 0, fmt.Errorf("bearer expires at %v", user.Expiration)
	}

	// 检查policy权限限定
	if user.Policy != "" {
		body, err := base64.StdEncoding.DecodeString(user.Policy)
		if err != nil {
			return 0, fmt.Errorf("parse bearer policy base64 decode error: %s", err.Error())
		}
		user.Policy = string(body)
		var statements []statement
		err = json.Unmarshal([]byte(user.Policy), &statements)
		if err != nil {
			return 0, fmt.Errorf("parse bearer policy error: %s", err.Error())
		}
		action := ctx.GetParam("action")
		resource := ctx.GetParam("resource")
		for _, s := range statements {
			if s.MatchAction(action) && s.MatchResource(resource) && s.MatchCondition(ctx) {
				if s.Effect {
					return user.UserID, nil
				}
				return 0, nil
			}
		}
		return 0, nil
	}
	return user.UserID, nil
}

// AddPolicy 方法实现添加一个策略，如果策略stmt为空则删除策略。
func (ctl *Policys) AddPolicy(policy *Policy) error {
	err := policy.StatementUnmarshal()
	if err != nil {
		return fmt.Errorf("Policys.AddPolicy add policy %d error: %s", policy.PolicyID, err.Error())
	}
	if policy.PolicyName == "" {
		policy.PolicyName = fmt.Sprintf("pbac:policy:%d", policy.PolicyID)
	}
	p, ok := ctl.Policys.Load(policy.PolicyID)
	if ok {
		*p.(*Policy) = *policy
	} else {
		ctl.Policys.Store(policy.PolicyID, policy)
	}
	return nil
}

// AddMember 方法对指定userid和policyid进行授权。
//
// Index越大匹配优先级越高，如果小于0则解除授权。
//
// 使用Expiration属性设置授权过期时间，Description属性仅记录描述。
func (ctl *Policys) AddMember(member *Member) {
	// 加载member策略
	policy, ok := ctl.Policys.Load(member.PolicyID)
	if !ok {
		policy = &Policy{
			PolicyID: member.PolicyID,
		}
		// TODO: policy tree is nil
		ctl.Policys.Store(member.PolicyID, policy)
	}
	member.policy = policy.(*Policy)

	ims, ok := ctl.Members.Load(member.UserID)
	if !ok {
		ctl.Members.Store(member.UserID, []*Member{member})
		return
	}
	ms := ims.([]*Member)
	for i := range ms {
		if ms[i].PolicyID == member.PolicyID {
			ms = append(ms[:i], ms[i+1:]...)
			break
		}
	}
	if member.Index > -1 {
		ms = append(ms, member)
		sort.Slice(ms, func(i, j int) bool {
			return ms[i].Index > ms[j].Index
		})
	}
	ctl.Members.Store(member.UserID, ms)
}

// AddPermission 方法创建一个指定roleid的策略，绑定对于的actions。
func (ctl *Policys) AddPermission(roleid int, actions ...string) {
	p, ok := ctl.Policys.Load(roleid)
	if !ok {
		p = &Policy{PolicyID: roleid}
		ctl.Policys.Store(roleid, p)
	}

	policy := p.(*Policy)
	if policy.Statement == nil {
		*policy = Policy{
			PolicyID:   roleid,
			PolicyName: fmt.Sprintf("rbac:role:%d", roleid),
			statement: []statement{
				{
					Effect:     true,
					Resource:   []string{"*"},
					treeAction: new(starTree),
					treeResource: &starTree{
						wildcard: &starTree{Name: "*"},
					},
				},
			},
		}
	}

	policy.statement[0].Action = append(policy.statement[0].Action, actions...)
	for _, i := range actions {
		policy.statement[0].treeAction.Insert(i)
	}
	policy.StatementMarshal()
}

// DeletePermission 方法删除role绑定的actions。
func (ctl *Policys) DeletePermission(roleid int, actions ...string) {
	p, ok := ctl.Policys.Load(roleid)
	if !ok {
		return
	}
	pp := p.(*Policy)
	for _, action := range actions {
		for i := range pp.statement[0].Action {
			if action == pp.statement[0].Action[i] {
				pp.statement[0].Action = append(pp.statement[0].Action[:i], pp.statement[0].Action[i+1:]...)
				break
			}
		}
	}
	tree := new(starTree)
	for _, i := range pp.statement[0].Action {
		tree.Insert(i)
	}
	pp.statement[0].treeAction = tree
	pp.StatementMarshal()
}

// AddRole 方法给指定user绑定role。
func (ctl *Policys) AddRole(userid, roleid int) {
	ctl.AddMember(&Member{
		UserID:   userid,
		PolicyID: roleid,
	})
}

// DeleteRole 方法删除指定user绑定role。
func (ctl *Policys) DeleteRole(userid, roleid int) {
	ctl.AddMember(&Member{
		UserID:   userid,
		PolicyID: roleid,
		Index:    -1,
	})
}
