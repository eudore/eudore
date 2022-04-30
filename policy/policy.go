package policy

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
	"unsafe"

	"github.com/eudore/eudore"
)

var (
	// ErrFormatPolcyUnmarshalError 定义策略json解析错误。
	ErrFormatPolcyUnmarshalError = "policy unmarshal json error: %v"
	// ErrFormatDataParseError 定义策略数据解析错误。
	ErrFormatDataParseError = "policy data parse %s error: %v"
	// ErrFormatConditionsUnmarshalError 定义策略条件json解析错误。
	ErrFormatConditionsUnmarshalError = "policy conditions unmarshal json %s error: %v"
	// ErrFormatConditionsParseError 定义NewConditions解析策略条件错误。
	ErrFormatConditionsParseError = "policy conditions parse %s error: %v"
	// ErrFormatConditionParseError 定义策略指定条件解析错误。
	ErrFormatConditionParseError = "policy conditions %s parse %s error: %v"

	conditionObjects = make(map[string]func() Condition)
	dataObjects      = make(map[string]func() interface{})
)

func init() {
	conditionObjects = map[string]func() Condition{
		"and":      func() Condition { return &conditionAnd{} },
		"or":       func() Condition { return &conditionOr{} },
		"sourceip": func() Condition { return &conditionSourceIP{} },
		"date":     func() Condition { return &conditionDate{} },
		"time":     func() Condition { return &conditionTime{} },
		"method":   func() Condition { return &conditionMethod{} },
		"params":   func() Condition { return &conditionParams{} },
	}
	dataObjects = map[string]func() interface{}{
		"menu": func() interface{} { return new(string) },
	}
}

// Policy 定义一个策略.
type Policy struct {
	PolicyID    int         `json:"policy_id" alias:"policy_id"`
	PolicyName  string      `json:"policy_name" alias:"policy_name"`
	Description string      `json:"description" alias:"description"`
	Statement   []Statement `json:"statement" alias:"statement"`
}

// Statement 定义一条策略内容。
type Statement struct {
	Effect       bool                         `json:"effect"`
	Action       []string                     `json:"action"`
	Resource     []string                     `json:"resource"`
	Conditions   map[string]json.RawMessage   `json:"conditions,omitempty"`
	Data         map[string][]json.RawMessage `json:"data,omitempty"`
	treeAction   *starTree
	treeResource *starTree
	conditions   Condition                `json:"-"`
	data         map[string][]interface{} `json:"-"`
}

type _statement Statement

// Condition 定义策略使用的条件。
type Condition interface {
	Match(ctx eudore.Context) bool
}

// NewPolicy 方法使用字符串创建一个策略。
func NewPolicy(body string) (*Policy, error) {
	policy := &Policy{}
	return policy, json.Unmarshal([]byte(body), policy)
}

// MatchAction 方法匹配描述的条件。
func (stmt Statement) MatchAction(action string) bool {
	return stmt.treeAction.Match(action) != ""
}

// MatchResource 方法匹配描述的资源。
func (stmt Statement) MatchResource(resource string) bool {
	return stmt.treeResource.Match(resource) != ""
}

// MatchCondition 方法匹配描述的条件。
func (stmt Statement) MatchCondition(ctx eudore.Context) bool {
	if stmt.Conditions == nil {
		return true
	}
	return stmt.conditions.Match(ctx)
}

// MatchData 方法返回匹配时的权限数据。
func (stmt Statement) MatchData() map[string][]interface{} {
	return stmt.data
}

// UnmarshalJSON 方法实现json反序列化。
func (stmt *Statement) UnmarshalJSON(body []byte) error {
	err := json.Unmarshal(body, (*_statement)(unsafe.Pointer(stmt)))
	if err != nil {
		return fmt.Errorf(ErrFormatPolcyUnmarshalError, err)
	}
	conds, err := NewConditions(stmt.Conditions)
	if err != nil {
		return err
	}
	if conds != nil {
		stmt.conditions = conditionAnd{stmt.Conditions, conds}
	}

	stmt.data, err = newDatas(stmt.Data)
	if err != nil {
		return err
	}

	if len(stmt.Action) == 0 {
		stmt.Action = []string{"*"}
	}
	if len(stmt.Resource) == 0 {
		stmt.Resource = []string{"*"}
	}
	stmt.treeAction = newStarTree(stmt.Action)
	stmt.treeResource = newStarTree(stmt.Resource)
	return nil
}

type conditionAnd struct {
	Data       map[string]json.RawMessage
	Conditions []Condition
}
type conditionOr struct {
	Data       map[string]json.RawMessage
	Conditions []Condition
}

// conditionSourceIP 定义ip检查条件。
type conditionSourceIP struct {
	SourceIP []*net.IPNet `json:"sourceip,omitempty"`
}
type conditionDate struct {
	Before time.Time `json:"Before"`
	After  time.Time `json:"after"`
}
type _conditionDate struct {
	Before string `json:"Before"`
	After  string `json:"after"`
}
type conditionTime struct {
	Before time.Time `json:"Before"`
	After  time.Time `json:"after"`
}

// conditionMethod 定义请求方法条件。
type conditionMethod struct {
	Methods []string `json:"methods"`
}
type conditionParams map[string][]string

// NewConditions 方法解析多个策略条件。
func NewConditions(data map[string]json.RawMessage) ([]Condition, error) {
	var conds []Condition
	for key, val := range data {
		fn, ok := conditionObjects[key]
		if !ok {
			continue
		}
		cond := fn()

		err := json.Unmarshal(val, cond)
		if err != nil {
			return nil, fmt.Errorf(ErrFormatConditionsParseError, key, err)
		}
		conds = append(conds, cond)
	}
	return conds, nil
}

func newDatas(body map[string][]json.RawMessage) (map[string][]interface{}, error) {
	datas := make(map[string][]interface{})
	for key, vals := range body {
		for _, val := range vals {
			fn, ok := dataObjects[key]
			if !ok {
				continue
			}
			data := fn()

			err := json.Unmarshal(val, data)
			if err != nil {
				return nil, fmt.Errorf(ErrFormatDataParseError, key, err.Error())
			}
			datas[key] = append(datas[key], data)
		}
	}
	return datas, nil
}

// Match conditionAnd
func (cond conditionAnd) Match(ctx eudore.Context) bool {
	for _, i := range cond.Conditions {
		if !i.Match(ctx) {
			return false
		}
	}
	return true
}
func (cond *conditionAnd) UnmarshalJSON(body []byte) error {
	err := json.Unmarshal(body, &cond.Data)
	if err != nil {
		return fmt.Errorf(ErrFormatConditionsUnmarshalError, "and", err)
	}
	conds, err := NewConditions(cond.Data)
	cond.Conditions = conds
	return err
}

// Match 方法匹配or条件。
func (cond conditionOr) Match(ctx eudore.Context) bool {
	for _, i := range cond.Conditions {
		if i.Match(ctx) {
			return true
		}
	}
	return false
}
func (cond *conditionOr) UnmarshalJSON(body []byte) error {
	err := json.Unmarshal(body, &cond.Data)
	if err != nil {
		return fmt.Errorf(ErrFormatConditionsUnmarshalError, "or", err)
	}
	conds, err := NewConditions(cond.Data)
	cond.Conditions = conds
	return err
}

// Match 方法匹配ip段。
func (cond conditionSourceIP) Match(ctx eudore.Context) bool {
	for _, i := range cond.SourceIP {
		if i.Contains(net.ParseIP(ctx.RealIP())) {
			return true
		}
	}
	return false
}
func (cond *conditionSourceIP) UnmarshalJSON(body []byte) error {
	var strs []string
	err := json.Unmarshal(body, &strs)
	if err != nil {
		return fmt.Errorf(ErrFormatConditionsUnmarshalError, "sourceip", err)
	}
	var ipnets []*net.IPNet
	for _, i := range strs {
		if strings.IndexByte(i, '/') == -1 {
			i += "/32"
		}
		_, ipnet, err := net.ParseCIDR(i)
		if err != nil {
			return fmt.Errorf(ErrFormatConditionParseError, "sourceip", "cidr", err)
		}
		ipnets = append(ipnets, ipnet)
	}
	cond.SourceIP = ipnets
	return nil
}

// Match 方法匹配当前时间范围。
func (cond conditionDate) Match(ctx eudore.Context) bool {
	current := time.Now()
	return current.Before(cond.Before) && current.After(cond.After)
}
func (cond *conditionDate) UnmarshalJSON(body []byte) error {
	var date _conditionDate
	err := json.Unmarshal(body, &date)
	if err != nil {
		return fmt.Errorf(ErrFormatConditionsUnmarshalError, "date", err)
	}
	cond.Before, err = time.Parse("2006-01-02", date.Before)
	if err != nil && date.Before != "" {
		return fmt.Errorf(ErrFormatConditionParseError, "date", "before", err)
	}
	cond.After, err = time.Parse("2006-01-02", date.After)
	if err != nil && date.After != "" {
		return fmt.Errorf(ErrFormatConditionParseError, "date", "after", err)
	}
	return nil
}

// Match 方法匹配当前时间范围。
func (cond conditionTime) Match(ctx eudore.Context) bool {
	current := time.Now()
	current = time.Date(0, 0, 0, current.Hour(), current.Minute(), current.Second(), 0, current.Location())
	return current.Before(cond.Before) && current.After(cond.After)
}
func (cond *conditionTime) UnmarshalJSON(body []byte) error {
	var date _conditionDate
	err := json.Unmarshal(body, &date)
	if err != nil {
		return fmt.Errorf(ErrFormatConditionsUnmarshalError, "time", err)
	}
	cond.Before, err = time.Parse("15:04:05", date.Before)
	if err != nil && date.Before != "" {
		return fmt.Errorf(ErrFormatConditionParseError, "time", "before", err)
	}
	cond.After, err = time.Parse("15:04:05", date.After)
	if err != nil && date.After != "" {
		return fmt.Errorf(ErrFormatConditionParseError, "time", "after", err)
	}
	return nil
}

// Match 方法匹配http请求方法。
func (cond conditionMethod) Match(ctx eudore.Context) bool {
	method := ctx.Method()
	for _, i := range cond.Methods {
		if i == method {
			return true
		}
	}
	return false
}
func (cond *conditionMethod) UnmarshalJSON(body []byte) error {
	return json.Unmarshal(body, &cond.Methods)
}

// Match 方法匹配http请求方法。
func (cond conditionParams) Match(ctx eudore.Context) bool {
	for key, vals := range cond {
		if stringSliceNotIn(vals, ctx.GetParam(key)) {
			return false
		}
	}
	return true
}
