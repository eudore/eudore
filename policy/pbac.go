package policy

import (
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
	UserID      int       `json:"user_id" alias:"user_id"`
	PolicyID    int       `json:"policy_id" alias:"policy_id"`
	Index       int       `json:"index" alias:"index" description:"越大优先级越高，如果小于0移除授权"`
	Description string    `json:"description" alias:"description"`
	Expiration  time.Time `json:"expiration" alias:"expiration"`
	Policy      *Policy   `json:"-" alias:"-"`
}

// NewPolicys 函数创建默认策略访问控制器
func NewPolicys() *Policys {
	policys := &Policys{
		Signaturer: NewSignaturerJwt([]byte("eudore")),
		ActionFunc: func(ctx eudore.Context) string {
			return ctx.GetParam(eudore.ParamAction)
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
	ctx.SetParam(eudore.ParamResource, resource)

	// 获取用户信息。
	userid, err := ctl.GetUserFunc(ctx)
	if err != nil {
		ctl.ForbendFunc(ctx, action, resource, err.Error())
		return
	}
	ctx.SetParam(eudore.ParamUserid, fmt.Sprint(userid))

	var now = time.Now()
	var datas map[string][]interface{}
	var names []string

	// 遍历用户的全部授权的全部stmt
matchPolicys:
	for _, m := range ctl.GetMember(userid) {
		p := m.Policy
		if !m.Expiration.IsZero() && m.Expiration.Before(now) {
			continue
		}

		for _, s := range p.Statement {
			if s.MatchAction(action) && s.MatchResource(resource) && s.MatchCondition(ctx) {
				// 非数据权限执行行为
				if s.Data == nil {
					ctx.SetParam(eudore.ParamPolicy, p.PolicyName)
					if !s.Effect {
						ctl.ForbendFunc(ctx, action, resource, "")
					}
					return
				}
				names = append(names, p.PolicyName)
				if datas == nil {
					datas = make(map[string][]interface{})
				}
				for key, val := range s.data {
					datas[key] = append(datas[key], val...)
				}
			}
		}
	}

	if userid != 0 {
		userid = 0
		goto matchPolicys
	}
	// 数据权限
	if datas != nil {
		ctx.SetParam(eudore.ParamPolicy, strings.Join(names, ","))
		for key, val := range datas {
			ctx.SetValue(eudore.NewContextKey("policy-"+key), val)
		}
		return
	}
	ctl.ForbendFunc(ctx, action, resource, fmt.Sprintf("User %s not match policys.", ctx.GetParam(eudore.ParamUserid)))
}

// GetMember 方法获取指定userid绑定的Member。
func (ctl *Policys) GetMember(userid int) []*Member {
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
	Time       string `json:"time"`
	Host       string `json:"host"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Route      string `json:"route"`
	Action     string `json:"action"`
	Resource   string `json:"resource"`
	Status     int    `json:"status"`
	Error      string `json:"error,omitempty"`
	Message    string `json:"message"`
	XRequestID string `json:"x-request-id,omitempty"`
	XTraceID   string `json:"x-trace-id,omitempty"`
}

func (ctl *Policys) handleForbidden(ctx eudore.Context, action, resource, err string) {
	msg := forbiddenMessage{
		Time:       time.Now().Format(eudore.DefaultLoggerTimeFormat),
		Host:       ctx.Host(),
		Method:     ctx.Method(),
		Path:       ctx.Path(),
		Route:      ctx.GetParam(eudore.ParamRoute),
		Action:     action,
		Resource:   resource,
		Status:     403,
		Error:      err,
		Message:    "forbidden",
		XRequestID: ctx.Response().Header().Get(eudore.HeaderXRequestID),
		XTraceID:   ctx.Response().Header().Get(eudore.HeaderXTraceID),
	}
	if ctx.GetParam(eudore.ParamUserid) == "0" {
		msg.Status = 401
		msg.Message = "unauthorized"
	}
	ctx.WriteHeader(msg.Status)
	ctx.Render(msg)
	ctx.End()
}

const stringBearer = "Bearer "

// SignatureUser 定义默认的用户信息，也可以组合该对象使用自定义签名对象。
type SignatureUser struct {
	// 唯一必要的属性，指定请求的userid
	UserID int `json:"userid" alias:"userid"`
	// 如果非空，则为base64([]Statement)
	Policy     string `json:"policy,omitempty" alias:"policy"`
	Expiration int64  `json:"expiration" alias:"expiration"`
}

// NewBearer 默认的Bearer签名方法。
func (ctl *Policys) NewBearer(userid int, policy string, expires int64) string {
	return stringBearer + ctl.Signaturer.Signed(&SignatureUser{
		UserID:     userid,
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
			return 0, fmt.Errorf("bearer policy base64 decode error: %s", err.Error())
		}
		user.Policy = string(body)

		var statements []Statement
		err = json.Unmarshal(body, &statements)
		if err != nil {
			return 0, fmt.Errorf("bearer policy statements unmarshal json error: %s", err.Error())
		}
		action := ctx.GetParam(eudore.ParamAction)
		resource := ctx.GetParam(eudore.ParamResource)
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
func (ctl *Policys) AddPolicy(body string) error {
	policy, err := NewPolicy(body)
	if err != nil {
		return err
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
		ctl.Policys.Store(member.PolicyID, policy)
	}
	member.Policy = policy.(*Policy)

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
