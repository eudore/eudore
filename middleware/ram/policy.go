package ram

import (
	"encoding/json"
	"fmt"
	"github.com/eudore/eudore"
	"net"
	"strings"
	"time"
)

type (
	// Policy 定义一个策略。
	Policy struct {
		// Version string
		// Description string `json:"description",omitempty`
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Version     string      `json:"version"`
		Statement   []Statement `json:"statement"`
	}
	// Statement 定义一条策略内容。
	Statement struct {
		Effect     bool          `json:"effect"`
		Action     []string      `json:"action"`
		Resource   []string      `json:"resource"`
		Conditions *ConditionAnd `json:"conditions,omitempty"`
	}
	// Condition 定义策略使用的条件。
	Condition interface {
		Name() string
		Match(ctx eudore.Context) bool
	}
	// NewConditionFunc 定义条件构造函数。
	NewConditionFunc func(interface{}) Condition
	// ConditionOr 定义Or条件。
	ConditionOr struct {
		Conditions []Condition `json:"conditions,omitempty"`
	}
	// ConditionAnd 定义And条件。
	ConditionAnd struct {
		Conditions []Condition `json:"conditions,omitempty"`
	}
	// ConditionSourceIp 定义ip检查条件。
	ConditionSourceIp struct {
		SourceIp []*net.IPNet `json:"sourceip,omitempty"`
	}
	// ConditionTime 定义当前时间现在条件。
	ConditionTime struct {
		Befor time.Time `json:"befor"`
		After time.Time `json:"after"`
	}
	// ConditionMethod 定义请求方法条件。
	ConditionMethod struct {
		Methods []string `json:"methods"`
	}

	/*
		browser条件扩展: https://github.com/eudore/website/blob/master/internal/middleware/rambrowser.go
	*/
)

var conditionnews = make(map[string]NewConditionFunc)

func init() {
	conditionnews["or"] = NewConditionOr
	conditionnews["and"] = NewConditionAnd
	conditionnews["sourceip"] = NewConditionSourceIp
	conditionnews["time"] = NewConditionTime
	conditionnews["method"] = NewConditionMethod
}

// RegisterCondition 方法支持一个条件构造函数，默认存在or、and、sourceip、time、method。
func RegisterCondition(name string, cond NewConditionFunc) {
	conditionnews[name] = cond
}

// NewPolicyStringJSON 方法使用json字符串创建一个策略对象。
func NewPolicyStringJSON(str string) *Policy {
	policy := Policy{}
	err := json.Unmarshal([]byte(str), &policy)
	if err != nil {
		panic(err)
	}

	return &policy
}

// MatchAction 方法匹配描述的条件。
func (stat *Statement) MatchAction(action string) bool {
	for _, i := range stat.Action {
		if MatchStar(action, i) {
			return true
		}
	}
	return false
}

// MatchResource 方法匹配描述的资源。
func (stat *Statement) MatchResource(resource string) bool {
	for _, i := range stat.Resource {
		if MatchStar(resource, i) {
			return true
		}
	}
	return false
}

// MatchCondition 方法匹配描述的条件。
func (stat *Statement) MatchCondition(ctx eudore.Context) bool {
	if stat.Conditions == nil {
		return true
	}
	return stat.Conditions.Match(ctx)
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
	return &ConditionOr{Conditions: NewConditions(data)}
}

// Name 方法返回条件名称。
func (cond *ConditionOr) Name() string {
	return "or"
}

// Match 方法匹配or条件。
func (cond *ConditionOr) Match(ctx eudore.Context) bool {
	for _, i := range cond.Conditions {
		if i.Match(ctx) {
			return true
		}
	}
	return false
}

// UnmarshalJSON 方法实现json反序列化。
func (cond *ConditionOr) UnmarshalJSON(body []byte) error {
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err
	}
	cond.Conditions = NewConditions(data)
	return nil
}

// MarshalJSON 方法实现json序列化。
func (cond *ConditionOr) MarshalJSON() ([]byte, error) {
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
	return &ConditionAnd{Conditions: NewConditions(data)}
}

// Name 方法返回条件名称。
func (cond *ConditionAnd) Name() string {
	return "and"
}

// Match 方法匹配and条件。
func (cond *ConditionAnd) Match(ctx eudore.Context) bool {
	for _, i := range cond.Conditions {
		if !i.Match(ctx) {
			return false
		}
	}
	return true
}

// UnmarshalJSON 方法实现json反序列化。
func (cond *ConditionAnd) UnmarshalJSON(body []byte) error {
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err
	}
	cond.Conditions = NewConditions(data)
	return nil
}

// MarshalJSON 方法实现json序列化。
func (cond *ConditionAnd) MarshalJSON() ([]byte, error) {
	data := make(map[string]Condition)
	for _, cond := range cond.Conditions {
		data[cond.Name()] = cond
	}
	return json.Marshal(data)
}

// NewConditionSourceIp 方法创建一个ip匹配条件。
func NewConditionSourceIp(i interface{}) Condition {
	var ipnets []*net.IPNet
	for _, i := range GetArrayString(i) {
		if strings.IndexByte(i, '/') == -1 {
			i += "/32"
		}
		_, ipnet, err := net.ParseCIDR(i)
		if err == nil {
			ipnets = append(ipnets, ipnet)
		}
	}
	return &ConditionSourceIp{SourceIp: ipnets}
}

// Name 方法返回条件名称。
func (cond *ConditionSourceIp) Name() string {
	return "sourceip"
}

// Match 方法匹配ip段。
func (cond *ConditionSourceIp) Match(ctx eudore.Context) bool {
	for _, i := range cond.SourceIp {
		if i.Contains(net.ParseIP(ctx.RealIP())) {
			return true
		}
	}
	return false
}

// MarshalJSON 方法实现ip条件的序列化。
func (cond *ConditionSourceIp) MarshalJSON() ([]byte, error) {
	data := make([]string, 0, len(cond.SourceIp))
	for _, i := range cond.SourceIp {
		data = append(data, i.IP.String())
	}
	return json.Marshal(data)
}

// NewConditionTime 方法创建一个时间条件。
func NewConditionTime(i interface{}) Condition {
	cond := &ConditionTime{}
	data, ok := i.(map[string]interface{})
	if ok {
		cond.After = TimeParse(getString(data["after"]))
		cond.Befor = TimeParse(getString(data["befor"]))
		if cond.Befor.Unix() == 0 {
			cond.Befor = time.Unix(9223372036854775807, 9223372036854775807)
		}
	}
	return cond
}

// Name 方法返回条件名称。
func (cond *ConditionTime) Name() string {
	return "time"
}

// Match 方法匹配当前时间范围。
func (cond *ConditionTime) Match(ctx eudore.Context) bool {
	current := time.Now()
	return current.Before(cond.Befor) && current.After(cond.After)
}

// NewConditionMethod 创建一个请求方法条件。
func NewConditionMethod(i interface{}) Condition {
	return &ConditionMethod{Methods: GetArrayString(i)}
}

// Name 方法返回条件名称。
func (cond *ConditionMethod) Name() string {
	return "method"
}

// Match 方法匹配http请求方法。
func (cond *ConditionMethod) Match(ctx eudore.Context) bool {
	method := ctx.Method()
	for _, i := range cond.Methods {
		if i == method {
			return true
		}
	}
	return false
}

// MarshalJSON 方法实现请求方法的序列化。
func (cond *ConditionMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(cond.Methods)
}

// GetArrayString 方法将一个对象转换成字符串数组。
func GetArrayString(i interface{}) []string {
	strs, ok := i.([]interface{})
	if ok {
		data := make([]string, len(strs))
		for i, str := range strs {
			data[i] = fmt.Sprint(str)
		}
		return data
	}
	return nil
}

// TimeParse 方法通过解析内置支持的时间格式。
func TimeParse(str string) time.Time {
	var formats = []string{
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
	for _, f := range formats {
		t, err := time.Parse(f, str)
		if err == nil {
			return t
		}
	}
	return time.Unix(1, 0)
}

// MatchStar 模式匹配对象，允许使用带'*'的模式。
func MatchStar(obj, patten string) bool {
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

func getString(i interface{}) string {
	if i == nil {
		return ""
	}
	if v, ok := i.(string); ok && v != "" {
		return v
	}
	return ""
}
