package ram

import (
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/eudore/eudore"
)

type (
	// Policy 定义一个策略。
	Policy struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Version     string      `json:"version"`
		Statement   []Statement `json:"statement"`
	}
	// Statement 定义一条策略内容。
	Statement struct {
		Effect     bool      `json:"effect"`
		Action     []string  `json:"action"`
		Resource   []string  `json:"resource"`
		Conditions Condition `json:"conditions,omitempty"`
	}
	// Condition 定义策略使用的条件。
	Condition interface {
		Name() string
		Match(ctx eudore.Context) bool
	}
	// conditionOr 定义Or条件。
	conditionOr struct {
		Conditions []Condition `json:"conditions,omitempty"`
	}
	// conditionAnd 定义And条件。
	conditionAnd struct {
		Conditions []Condition `json:"conditions,omitempty"`
	}
	// conditionSourceIp 定义ip检查条件。
	conditionSourceIp struct {
		SourceIp []*net.IPNet `json:"sourceip,omitempty"`
	}
	// conditionTime 定义当前时间现在条件。
	conditionTime struct {
		Befor time.Time `json:"befor"`
		After time.Time `json:"after"`
	}
	// conditionMethod 定义请求方法条件。
	conditionMethod struct {
		Methods []string `json:"methods"`
	}

	/*
		ConditionBrowser扩展: https://github.com/eudore/website/blob/master/framework/rambrowser.go
	*/
)

var conditionnews = make(map[string]func(interface{}) Condition)

func init() {
	conditionnews["or"] = NewConditionOr
	conditionnews["and"] = NewConditionAnd
	conditionnews["sourceip"] = NewConditionSourceIp
	conditionnews["time"] = NewConditionTime
	conditionnews["method"] = NewConditionMethod
}

// RegisterCondition 方法支持一个条件构造函数，默认存在or、and、sourceip、time、method。
func RegisterCondition(name string, cond func(interface{}) Condition) {
	conditionnews[name] = cond
}

// ParsePolicyString 方法使用json字符串创建一个策略对象。
func ParsePolicyString(str string) (*Policy, error) {
	policy := &Policy{}
	err := json.Unmarshal([]byte(str), policy)
	return policy, err
}

// MatchAction 方法匹配描述的条件。
func (stat Statement) MatchAction(action string) bool {
	for _, i := range stat.Action {
		if matchStar(action, i) {
			return true
		}
	}
	return false
}

// MatchResource 方法匹配描述的资源。
func (stat Statement) MatchResource(resource string) bool {
	for _, i := range stat.Resource {
		if matchStar(resource, i) {
			return true
		}
	}
	return false
}

// MatchCondition 方法匹配描述的条件。
func (stat Statement) MatchCondition(ctx eudore.Context) bool {
	if stat.Conditions == nil {
		return true
	}
	return stat.Conditions.Match(ctx)
}

// UnmarshalJSON 方法实现json反序列化。
func (stat *Statement) UnmarshalJSON(body []byte) error {
	var data struct {
		Effect     bool                   `json:"effect"`
		Action     []string               `json:"action"`
		Resource   []string               `json:"resource"`
		Conditions map[string]interface{} `json:"conditions,omitempty"`
	}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err
	}
	stat.Effect = data.Effect
	stat.Action = data.Action
	stat.Resource = data.Resource
	conds := NewConditions(data.Conditions)
	if len(conds) > 0 {
		stat.Conditions = conditionAnd{conds}
	}
	return nil
}

// NewConditions 方法使用json对象创建多个条件
func NewConditions(data map[string]interface{}) []Condition {
	conds := make([]Condition, 0, len(data))
	for key, val := range data {
		fn, ok := conditionnews[key]
		if !ok {
			continue
		}

		cond := fn(val)
		if cond != nil {
			conds = append(conds, cond)
		}
	}
	return conds
}

// NewConditionOr 方法创建一个or条件。
func NewConditionOr(i interface{}) Condition {
	data, ok := i.(map[string]interface{})
	if !ok {
		return nil
	}
	return &conditionOr{Conditions: NewConditions(data)}
}

// Name 方法返回条件名称。
func (cond conditionOr) Name() string {
	return "or"
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

// MarshalJSON 方法实现json序列化。
func (cond conditionOr) MarshalJSON() ([]byte, error) {
	data := make(map[string]Condition)
	for _, cond := range cond.Conditions {
		data[cond.Name()] = cond
	}
	return json.Marshal(data)
}

// NewConditionAnd 方法创建一个and条件。
func NewConditionAnd(i interface{}) Condition {
	data, ok := i.(map[string]interface{})
	if !ok {
		return nil
	}
	return conditionAnd{Conditions: NewConditions(data)}
}

// Name 方法返回条件名称。
func (cond conditionAnd) Name() string {
	return "and"
}

// Match 方法匹配and条件。
func (cond conditionAnd) Match(ctx eudore.Context) bool {
	for _, i := range cond.Conditions {
		if !i.Match(ctx) {
			return false
		}
	}
	return true
}

// MarshalJSON 方法实现json序列化。
func (cond conditionAnd) MarshalJSON() ([]byte, error) {
	data := make(map[string]Condition)
	for _, cond := range cond.Conditions {
		data[cond.Name()] = cond
	}
	return json.Marshal(data)
}

// NewConditionSourceIp 方法创建一个ip匹配条件。
func NewConditionSourceIp(i interface{}) Condition {
	var ipnets []*net.IPNet
	for _, i := range eudore.GetStrings(i) {
		if strings.IndexByte(i, '/') == -1 {
			i += "/32"
		}
		_, ipnet, err := net.ParseCIDR(i)
		if err == nil {
			ipnets = append(ipnets, ipnet)
		}
	}
	return &conditionSourceIp{SourceIp: ipnets}
}

// Name 方法返回条件名称。
func (cond conditionSourceIp) Name() string {
	return "sourceip"
}

// Match 方法匹配ip段。
func (cond conditionSourceIp) Match(ctx eudore.Context) bool {
	for _, i := range cond.SourceIp {
		if i.Contains(net.ParseIP(ctx.RealIP())) {
			return true
		}
	}
	return false
}

// MarshalJSON 方法实现ip条件的序列化。
func (cond conditionSourceIp) MarshalJSON() ([]byte, error) {
	data := make([]string, 0, len(cond.SourceIp))
	for _, i := range cond.SourceIp {
		data = append(data, i.IP.String())
	}
	return json.Marshal(data)
}

// NewConditionTime 方法创建一个时间条件。
func NewConditionTime(i interface{}) Condition {
	cond := &conditionTime{}
	data, ok := i.(map[string]interface{})
	if ok {
		cond.After = timeParse(eudore.GetString(data["after"]))
		cond.Befor = timeParse(eudore.GetString(data["befor"]))
		if cond.Befor.Equal(time.Time{}) {
			cond.Befor = time.Date(9999, 12, 31, 0, 0, 0, 0, time.Now().Location())
		}
	}
	return cond
}

// Name 方法返回条件名称。
func (cond conditionTime) Name() string {
	return "time"
}

// Match 方法匹配当前时间范围。
func (cond conditionTime) Match(ctx eudore.Context) bool {
	current := time.Now()
	return current.Before(cond.Befor) && current.After(cond.After)
}

// NewConditionMethod 创建一个请求方法条件。
func NewConditionMethod(i interface{}) Condition {
	return &conditionMethod{Methods: eudore.GetStrings(i)}
}

// Name 方法返回条件名称。
func (cond conditionMethod) Name() string {
	return "method"
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

// MarshalJSON 方法实现请求方法的序列化。
func (cond conditionMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(cond.Methods)
}

// timeformats 定义允许使用的时间格式。
var timeformats = []string{
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
	"2006-1-02",
	"2006-01-02",
	"15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05.999999999Z07:00",
}

// timeParse 方法通过解析内置支持的时间格式。
func timeParse(str string) time.Time {
	for _, f := range timeformats {
		t, err := time.Parse(f, str)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}

// matchStar 模式匹配对象，允许使用带'*'的模式。
func matchStar(obj, patten string) bool {
	ps := strings.Split(patten, "*")
	if len(ps) < 2 {
		return patten == obj
	}
	if !strings.HasPrefix(obj, ps[0]) {
		return false
	}
	for _, i := range ps {
		if i == "" {
			continue
		}
		pos := strings.Index(obj, i)
		if pos == -1 {
			return false
		}
		obj = obj[pos+len(i):]
	}
	return true
}
