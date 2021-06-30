package policy

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/eudore/eudore"
)

type contextKey struct {
	name string
}

// PolicyExpressions 定义policy-expressions Context Key。
var PolicyExpressions = &contextKey{"policy-expressions"}

// Policy 定义一个策略.
type Policy struct {
	PolicyID    int        `json:"policy_id" alias:"policy_id" gorm:"primaryKey" db:"policy_id"`
	PolicyName  string     `json:"policy_name" alias:"policy_name" gorm:"index" db:"policy_name"`
	Description string     `json:"description" alias:"description"`
	Statement   RawMessage `json:"statement" alias:"statement"`
	statement   []statement
}

// RawMessage 定义json字符串集合。
type RawMessage []byte

// statement 定义一条策略内容。
type statement struct {
	Effect       bool        `json:"effect"`
	Action       []string    `json:"action"`
	Resource     []string    `json:"resource"`
	Conditions   Condition   `json:"conditions,omitempty"`
	Data         Expressions `json:"data,omitempty"`
	treeAction   *starTree
	treeResource *starTree
}

// Condition 定义策略使用的条件。
type Condition interface {
	Name() string
	Match(ctx eudore.Context) bool
}

// Expressions 定义数据表达式集合
type Expressions []Expression

// Expression 定义数据表达式
type Expression interface {
	Expression(string, []string) (string, []interface{})
}

var newConditionFuncs = make(map[string]func(interface{}) Condition)
var newExpressionFuncs = make(map[string]func([]byte) (Expression, error))
var newValueFuncs = make(map[string]func(string) func(eudore.Context) interface{})

func init() {
	newConditionFuncs["or"] = newConditionOr
	newConditionFuncs["and"] = newConditionAnd
	newConditionFuncs["sourceip"] = newConditionsourceIP
	newConditionFuncs["time"] = newConditionTime
	newConditionFuncs["method"] = newConditionMethod
	newConditionFuncs["params"] = newConditionParams

	newExpressionFuncs["and"] = newExpressionAnd
	newExpressionFuncs["or"] = newExpressionOr
	newExpressionFuncs["value"] = newExpressionValue
	newExpressionFuncs["range"] = newExpressionRange
	newExpressionFuncs["sql"] = newExpressionSql

	newValueFuncs["param"] = newValueParam
	newValueFuncs["query"] = newValueQuery
}

// RegisterCondition 方法注册条件构造函数，默认存在or、and、sourceip、time、method、params。
func RegisterCondition(name string, fn func(interface{}) Condition) {
	newConditionFuncs[name] = fn
}

// RegisterExpression 方法注册数据表达式构造函数，默认存在and、or、value、range、sql。
func RegisterExpression(name string, fn func([]byte) (Expression, error)) {
	newExpressionFuncs[name] = fn
}

// RegisterValue 方法注册动态数据构造函数，默认存在param、query。
func RegisterValue(name string, fn func(string) func(eudore.Context) interface{}) {
	newValueFuncs[name] = fn
}

// StatementUnmarshal 方法将policy.Statement反序列化。
func (policy *Policy) StatementUnmarshal() error {
	if policy.Statement != nil {
		return json.Unmarshal([]byte(policy.Statement), &policy.statement)
	}
	return nil
}

// StatementMarshal 方法将policy.Statement序列化。
func (policy *Policy) StatementMarshal() error {
	body, err := json.Marshal(policy.statement)
	if err == nil {
		policy.Statement = body
	}
	return err
}

// Match 方法匹配请求上下文，默认返回false。
func (policy *Policy) Match(ctx eudore.Context, action, resource string) (Expressions, bool) {
	var datas Expressions
	var names []string
	for _, s := range policy.statement {
		ok := s.MatchAction(action) && s.MatchResource(resource) && s.MatchCondition(ctx)
		if ok {
			// 非数据权限执行行为
			if s.Data == nil {
				ctx.SetParam("policy", policy.PolicyName)
				return nil, s.Effect
			}
			names = append(names, policy.PolicyName)
			datas = append(datas, s.Data...)
		}
	}
	if datas != nil {
		ctx.SetParam("policy", strings.Join(names, ","))
		ctx.WithContext(context.WithValue(ctx.GetContext(), PolicyExpressions, datas))
	}
	return datas, datas != nil
}

// MatchAction 方法匹配描述的条件。
func (stat statement) MatchAction(action string) bool {
	return stat.treeAction.Match(action) != ""
}

// MatchResource 方法匹配描述的资源。
func (stat statement) MatchResource(resource string) bool {
	return stat.treeResource.Match(resource) != ""
}

// MatchCondition 方法匹配描述的条件。
func (stat statement) MatchCondition(ctx eudore.Context) bool {
	if stat.Conditions == nil {
		return true
	}
	return stat.Conditions.Match(ctx)
}

// UnmarshalJSON 方法实现json反序列化。
func (stat *statement) UnmarshalJSON(body []byte) error {
	var data struct {
		Effect     bool                   `json:"effect"`
		Action     []string               `json:"action"`
		Resource   []string               `json:"resource"`
		Conditions map[string]interface{} `json:"conditions,omitempty"`
		Data       RawMessage             `json:"data,omitempty"`
	}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	stat.Effect = data.Effect
	stat.Action = data.Action
	stat.Resource = data.Resource
	if len(stat.Action) == 0 {
		stat.Action = []string{"*"}
	}
	if len(stat.Resource) == 0 {
		stat.Resource = []string{"*"}
	}
	stat.treeAction = new(starTree)
	stat.treeResource = new(starTree)
	for _, i := range stat.Action {
		stat.treeAction.Insert(i)
	}
	for _, i := range stat.Resource {
		stat.treeResource.Insert(i)
	}

	conds := NewConditions(data.Conditions)
	if len(conds) > 0 {
		stat.Conditions = conditionAnd{conds}
	}
	if len(data.Data) > 0 {
		stat.Data, err = NewExpressions(data.Data)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateExpressions 方法生成请求上下文对应表的数据表达式。
//
// 如果是mysql等$1使用数据库占位符的数据库，index为第一个占位符的位置。
func CreateExpressions(ctx eudore.Context, tb string, fields []string, index int) (string, []interface{}) {
	val := ctx.GetContext().Value(PolicyExpressions)
	if val == nil {
		return "", nil
	}
	exprs := val.(Expressions)

	newexprs, values := exprs.Expression(tb, fields)
	sql := strings.Join(newexprs, " OR ")
	for i := range values {
		fn, ok := values[i].(func(ctx eudore.Context) interface{})
		if ok {
			values[i] = fn(ctx)
		}
	}
	if index > 0 {
		for {
			if strings.IndexByte(sql, '?') != -1 {
				sql = strings.Replace(sql, "?", fmt.Sprintf("$%d", index), 1)
				index++
			} else {
				break
			}
		}
	}
	return sql, values
}

// MarshalJSON returns m as the JSON encoding of m.
func (m RawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("policy.RawMessage: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

// Value 方法生成sql值。
func (m RawMessage) Value() (driver.Value, error) {
	return m.MarshalJSON()
}

// Scan 方法接收sql解析数据、
func (m *RawMessage) Scan(src interface{}) error {
	return m.UnmarshalJSON([]byte(eudore.GetString(src)))
}
