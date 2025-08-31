package middleware

import (
	"encoding/json"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/eudore/eudore"
)

type securityPolicys struct {
	effect  sync.Map
	data    sync.Map
	policys orderData[policy]
	members orderData[member]
}

type policy struct {
	Name      string      `json:"name"`
	Statement []statement `json:"statement"`
}

type statement struct {
	Effect     bool             `json:"effect"`
	Action     []string         `json:"action"`
	Resource   []string         `json:"resource,omitempty"`
	Conditions *conditionAnd    `json:"conditions,omitempty"`
	Data       map[string][]any `json:"data,omitempty"`
	actions    *radixNode[struct{}, *struct{}]
	resources  *radixNode[struct{}, *struct{}]
}

type condition interface {
	Match(ctx eudore.Context) bool
}

type member struct {
	User   string   `json:"user"`
	Policy []string `json:"policy,omitempty"`
	Data   []string `json:"data,omitempty"`
}

type forbiddenMessage struct {
	Time       string `json:"time" protobuf:"1,name=time" yaml:"time"`
	Host       string `json:"host" protobuf:"2,name=host" yaml:"host"`
	Method     string `json:"method" protobuf:"3,name=method" yaml:"method"`
	Path       string `json:"path" protobuf:"4,name=path" yaml:"path"`
	Route      string `json:"route" protobuf:"5,name=route" yaml:"route"`
	Action     string `json:"action,omitempty" protobuf:"13,name=action" yaml:"action"`
	Resource   string `json:"resource,omitempty" protobuf:"14,name=resource" yaml:"resource"`
	Policy     string `json:"policy,omitempty" protobuf:"15,name=policy" yaml:"policy"`
	Userid     string `json:"userid,omitempty" protobuf:"16,name=userid" yaml:"userid"`
	Username   string `json:"username,omitempty" protobuf:"17,name=username" yaml:"username"`
	Status     int    `json:"status" protobuf:"6,name=status" yaml:"status"`
	XRequestID string `json:"xRequestId,omitempty" protobuf:"8,name=xRequestId" yaml:"xRequestId,omitempty"`
	XTraceID   string `json:"xTraceId,omitempty" protobuf:"9,name=xTraceId" yaml:"xTraceId,omitempty"`
	Message    any    `json:"message,omitempty" protobuf:"11,name=message" yaml:"message,omitempty"`
}

// The NewSecurityPolicysFunc function creates middleware to implement PBAC request authentication.
//
// data passes the [Policy] and [Member] data used during initialization.
//
// options: [NewOptionRouter] [NewOptionSecurityPolicysChan].
func NewSecurityPolicysFunc(data []string, options ...Option) Middleware {
	pbac := &securityPolicys{
		policys: &orderMap[policy]{mapping: make(map[string]*policy)},
		members: &orderMap[member]{mapping: make(map[string]*member)},
	}
	for _, msg := range data {
		err := pbac.addData(msg)
		if err != nil {
			panic(err)
		}
	}
	applyOption(pbac, options)

	return func(ctx eudore.Context) {
		userid := ctx.GetParam(eudore.ParamUserid)
		if pbac.Allow(ctx, userid) {
			pbac.LoadData(ctx, userid, nil)
			return
		}

		msg := forbiddenMessage{
			Time:       time.Now().Format(eudore.DefaultLoggerFormatterFormatTime),
			Host:       ctx.Host(),
			Method:     ctx.Method(),
			Path:       ctx.Path(),
			Route:      ctx.GetParam(eudore.ParamRoute),
			Action:     ctx.GetParam(eudore.ParamAction),
			Resource:   ctx.GetParam(eudore.ParamResource),
			Policy:     ctx.GetParam(eudore.ParamPolicy),
			Userid:     userid,
			Username:   ctx.GetParam(eudore.ParamUsername),
			Status:     eudore.StatusForbidden,
			Message:    "Forbidden",
			XRequestID: ctx.Response().Header().Get(eudore.HeaderXRequestID),
			XTraceID:   ctx.Response().Header().Get(eudore.HeaderXTraceID),
		}
		if userid == "" {
			msg.Status = eudore.StatusUnauthorized
			msg.Message = "Unauthorized"
		}
		ctx.WriteStatus(msg.Status)
		_ = ctx.Render(msg)
		ctx.End()
	}
}

// NewOptionSecurityPolicysChan creates [Option] to receive Policy data and
// synchronize user policy data to the middleware.
func NewOptionSecurityPolicysChan(ch chan string) Option {
	return func(data any) {
		pbac, ok := data.(*securityPolicys)
		if ok {
			go func(ch chan string) {
				for msg := range ch {
					_ = pbac.addData(msg)
				}
			}(ch)
		}
	}
}

// The NewResourceFunc function creates middleware to set [eudore.ParamResource].
//
// When matching a specified prefix, sets the subsequent path to Resource.
//
// Used by [NewSecurityPolicysFunc] to match requests.
func NewResourceFunc(data ...string) Middleware {
	if len(data) == 1 {
		prefix := data[0]
		return func(ctx eudore.Context) {
			ctx.SetParam(eudore.ParamResource, strings.TrimPrefix(ctx.Path(), prefix))
		}
	}

	radix := &radixNode[byte, []byte]{}
	for i := range data {
		path := strings.Trim(data[i], "*")
		radix.insert(path+"*", []byte(path))
	}
	return func(ctx eudore.Context) {
		prefix := radix.lookNode(ctx.Path())
		if prefix != nil {
			ctx.SetParam(eudore.ParamResource, ctx.Path()[len(prefix):])
		}
	}
}

func (pbac *securityPolicys) Allow(ctx eudore.Context, userid string) bool {
	action := ctx.GetParam(eudore.ParamAction)
	resource := ctx.GetParam(eudore.ParamResource)
	value, ok := pbac.effect.Load(userid)
	if ok {
		for _, p := range value.([]*policy) {
			for _, stmt := range p.Statement {
				if stmt.Data == nil &&
					(stmt.actions == nil || stmt.actions.lookNode(action) != nil) &&
					(stmt.resources == nil || stmt.resources.lookNode(resource) != nil) &&
					(stmt.Conditions == nil || stmt.Conditions.Match(ctx)) {
					ctx.SetParam(eudore.ParamPolicy, p.Name)
					return stmt.Effect
				}
			}
		}
	}

	if userid != "" {
		return pbac.Allow(ctx, "")
	}
	return false
}

func (pbac *securityPolicys) LoadData(ctx eudore.Context, userid string, data map[string][]any) {
	action := ctx.GetParam(eudore.ParamAction)
	resource := ctx.GetParam(eudore.ParamResource)
	value, ok := pbac.data.Load(userid)
	if ok {
		for _, p := range value.([]*policy) {
			for _, stmt := range p.Statement {
				if stmt.Data != nil &&
					(stmt.actions == nil || stmt.actions.lookNode(action) != nil) &&
					(stmt.resources == nil || stmt.resources.lookNode(resource) != nil) &&
					(stmt.Conditions == nil || stmt.Conditions.Match(ctx)) {
					if data == nil {
						data = make(map[string][]any)
					}
					for key, vals := range stmt.Data {
						data[key] = append(data[key], vals...)
					}
				}
			}
		}
	}

	if userid != "" {
		pbac.LoadData(ctx, "", data)
		return
	}
	for key, vals := range data {
		ctx.SetValue(eudore.NewContextKey(key), vals)
	}
}

func (pbac *securityPolicys) addData(data string) error {
	p := &policy{}
	err := json.Unmarshal([]byte(data), p)
	if err == nil && p.Name != "" {
		for i, stmt := range p.Statement {
			if len(stmt.Action) != 0 {
				radix := &radixNode[struct{}, *struct{}]{}
				for _, action := range stmt.Action {
					radix.insert(action, &valueStruct)
				}
				p.Statement[i].actions = radix
			}
			if len(stmt.Resource) != 0 {
				radix := &radixNode[struct{}, *struct{}]{}
				for _, resource := range stmt.Resource {
					radix.insert(resource, &valueStruct)
				}
				p.Statement[i].resources = radix
			}
		}

		pbac.policys.Set(p)
		pbac.updatePolicy(p.Name)
		return nil
	}

	m := &member{}
	err = json.Unmarshal([]byte(data), m)
	if err == nil && m.User != "" {
		pbac.members.Set(m)
		pbac.updateMember(m)
	}
	return err
}

func (pbac *securityPolicys) updatePolicy(name string) {
	for _, m := range pbac.members.Slices() {
		for _, policy := range m.Policy {
			if name == policy {
				pbac.updateMember(m)
				break
			}
		}
		for _, policy := range m.Data {
			if name == policy {
				pbac.updateMember(m)
				break
			}
		}
	}
}

func (pbac *securityPolicys) updateMember(m *member) {
	user := m.User
	if m.User == DefaultPolicyGuestUser {
		user = ""
	}

	policys := make([]*policy, 0, len(m.Policy))
	for _, name := range m.Policy {
		p := pbac.policys.Get(name)
		if p != nil {
			policys = append(policys, p)
		}
	}
	if len(policys) > 0 {
		pbac.effect.Store(user, policys)
	} else {
		pbac.effect.Delete(user)
	}

	policys = make([]*policy, 0, len(m.Data))
	for _, name := range m.Data {
		p := pbac.policys.Get(name)
		if p != nil && policyIsData(p.Statement) {
			policys = append(policys, p)
		}
	}
	if len(policys) > 0 {
		pbac.data.Store(user, policys)
	} else {
		pbac.data.Delete(user)
	}
}

func policyIsData(stmts []statement) bool {
	for _, stmt := range stmts {
		if stmt.Data != nil {
			return true
		}
	}
	return false
}

func (pbac *securityPolicys) Inject(_ eudore.Controller, router eudore.Router) error {
	router.HeadFunc("/policys", eudore.HandlerEmpty)
	router.GetFunc("/policys Action=middleware:pbac:GetPolicys", pbac.GetPolicys)
	router.GetFunc("/policys/:name Action=middleware:pbac:GetPolicysByName", pbac.GetPolicysByName)
	router.PutFunc("/policys/:name Action=middleware:pbac:PutPolicysByName", pbac.PutPolicysByName)
	router.DeleteFunc("/policys/:name Action=middleware:pbac:DeletePolicysByName", pbac.DeletePolicysByName)
	router.GetFunc("/members Action=middleware:pbac:GetMembers", pbac.GetMembers)
	router.GetFunc("/members/:name Action=middleware:pbac:GutMembersByName", pbac.GutMembersByName)
	router.PostFunc("/members/:name Action=middleware:pbac:PostMembersByName", pbac.PostMembersByName)
	router.PutFunc("/members/:name Action=middleware:pbac:PutMembersByName", pbac.PutMembersByName)
	router.DeleteFunc("/members/:name Action=middleware:pbac:DeleteMembersByName", pbac.DeleteMembersByName)
	return nil
}

func (pbac *securityPolicys) GetPolicys(ctx eudore.Context) {
	_ = ctx.Render(pbac.policys.Slices())
}

func (pbac *securityPolicys) GetPolicysByName(ctx eudore.Context) {
	p := pbac.policys.Get(ctx.GetParam("name"))
	if p == nil {
		ctx.Fatalf("policy not found %s", ctx.GetParam("name"))
		return
	}
	_ = ctx.Render(p)
}

func (pbac *securityPolicys) PutPolicysByName(ctx eudore.Context) {
	p := &policy{}
	err := json.NewDecoder(ctx).Decode(p)
	if err != nil {
		ctx.Fatal(err)
		return
	}

	p.Name = ctx.GetParam("name")
	pbac.policys.Set(p)
	_ = ctx.Render(p)
}

func (pbac *securityPolicys) DeletePolicysByName(ctx eudore.Context) {
	p := pbac.policys.Get(ctx.GetParam("name"))
	if p != nil {
		pbac.policys.Delete(p.Name)
		pbac.updatePolicy(p.Name)
	}
	ctx.WriteHeader(eudore.StatusNoContent)
}

func (pbac *securityPolicys) GetMembers(ctx eudore.Context) {
	_ = ctx.Render(pbac.members.Slices())
}

func (pbac *securityPolicys) GutMembersByName(ctx eudore.Context) {
	m := pbac.members.Get(ctx.GetParam("name"))
	if m == nil {
		ctx.Fatalf("member not found %s", ctx.GetParam("name"))
		return
	}
	m.User = ctx.GetParam("name")
	_ = ctx.Render(m)
}

func (pbac *securityPolicys) PostMembersByName(ctx eudore.Context) {
	m := &member{}
	err := ctx.Bind(m)
	if err != nil {
		ctx.Fatal(err)
		return
	}
	m.User = ctx.GetParam("name")
	o := pbac.members.Get(m.User)
	if o != nil {
		for _, p := range m.Policy {
			if sliceIndex(o.Policy, p) == -1 {
				o.Policy = append(o.Policy, p)
			}
		}
		for _, p := range m.Data {
			if sliceIndex(o.Data, p) == -1 {
				o.Data = append(o.Data, p)
			}
		}
		m.Policy = o.Policy
		m.Data = o.Data
	}
	pbac.members.Set(m)
	pbac.updateMember(m)
	_ = ctx.Render(m)
}

func sliceIndex[T comparable](vals []T, val T) int {
	for i := range vals {
		if val == vals[i] {
			return i
		}
	}
	return -1
}

func (pbac *securityPolicys) PutMembersByName(ctx eudore.Context) {
	m := &member{}
	err := ctx.Bind(m)
	if err != nil {
		ctx.Fatal(err)
		return
	}
	m.User = ctx.GetParam("name")
	pbac.members.Set(m)
	pbac.updateMember(m)
	_ = ctx.Render(m)
}

func (pbac *securityPolicys) DeleteMembersByName(ctx eudore.Context) {
	name := ctx.GetParam("name")
	m := pbac.members.Get(name)
	if m != nil {
		pbac.members.Delete(name)
		pbac.updateMember(&member{User: name})
	}
	ctx.WriteHeader(eudore.StatusNoContent)
}

type orderData[T any] interface {
	Get(key string) *T
	Set(val *T)
	Delete(val string)
	Slices() []*T
}

type orderMap[T any] struct {
	sync.RWMutex
	mapping map[string]*T
	slices  []*T
}

// The getStructName function gets the first string field of the struct.
func getStructName[T any](data *T) string {
	return *(*string)(unsafe.Pointer(data))
}

func (m *orderMap[T]) Get(key string) *T {
	m.RLock()
	defer m.RUnlock()
	return m.mapping[key]
}

func (m *orderMap[T]) Set(val *T) {
	name := getStructName(val)
	m.Lock()
	defer m.Unlock()
	if _, ok := m.mapping[name]; ok {
		for i := range m.slices {
			if name == getStructName(m.slices[i]) {
				m.slices[i] = val
				break
			}
		}
	} else {
		m.slices = append(m.slices, val)
	}
	m.mapping[name] = val
}

func (m *orderMap[T]) Delete(key string) {
	m.Lock()
	defer m.Unlock()
	val, ok := m.mapping[key]
	if ok {
		for i, v := range m.slices {
			if v == val {
				m.slices = m.slices[:i+copy(m.slices[i:], m.slices[i+1:])]
				break
			}
		}
		delete(m.mapping, key)
	}
}

func (m *orderMap[T]) Slices() []*T {
	m.RLock()
	defer m.RUnlock()
	return append([]*T(nil), m.slices...)
}
