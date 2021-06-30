package policy

import (
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/eudore/eudore"
)

// NewConditions 方法使用json对象创建多个条件
func NewConditions(data map[string]interface{}) []Condition {
	conds := make([]Condition, 0, len(data))
	for key, val := range data {
		fn, ok := newConditionFuncs[key]
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

// conditionOr 定义Or条件。
type conditionOr struct {
	Conditions []Condition `json:"conditions,omitempty"`
}

// newConditionOr 方法创建一个or条件。
func newConditionOr(i interface{}) Condition {
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

// conditionAnd 定义And条件。
type conditionAnd struct {
	Conditions []Condition `json:"conditions,omitempty"`
}

// newConditionAnd 方法创建一个and条件。
func newConditionAnd(i interface{}) Condition {
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

// conditionSourceIP 定义ip检查条件。
type conditionSourceIP struct {
	SourceIP []*net.IPNet `json:"sourceip,omitempty"`
}

// newConditionSourceIP 方法创建一个ip匹配条件。
func newConditionsourceIP(i interface{}) Condition {
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
	return &conditionSourceIP{SourceIP: ipnets}
}

// Name 方法返回条件名称。
func (cond conditionSourceIP) Name() string {
	return "sourceip"
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

// MarshalJSON 方法实现ip条件的序列化。
func (cond conditionSourceIP) MarshalJSON() ([]byte, error) {
	data := make([]string, 0, len(cond.SourceIP))
	for _, i := range cond.SourceIP {
		data = append(data, i.IP.String())
	}
	return json.Marshal(data)
}

// conditionTime 定义当前时间现在条件。
type conditionTime struct {
	Before time.Time `json:"Before"`
	After  time.Time `json:"after"`
}

// newConditionTime 方法创建一个时间条件。
func newConditionTime(i interface{}) Condition {
	cond := &conditionTime{}
	data, ok := i.(map[string]interface{})
	if ok {
		cond.After = timeParse(eudore.GetString(data["after"]))
		cond.Before = timeParse(eudore.GetString(data["before"]))
		if cond.Before.Equal(time.Time{}) {
			cond.Before = time.Date(9999, 12, 31, 0, 0, 0, 0, time.Now().Location())
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
	return current.Before(cond.Before) && current.After(cond.After)
}

// conditionMethod 定义请求方法条件。
type conditionMethod struct {
	Methods []string `json:"methods"`
}

// newConditionMethod 创建一个请求方法条件。
func newConditionMethod(i interface{}) Condition {
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

type conditionParams map[string][]string

// newConditionParams 创建一个请求方法条件。
func newConditionParams(i interface{}) Condition {
	data, ok := i.(map[string]interface{})
	if !ok {
		return nil
	}
	cond := make(conditionParams, len(data))
	for k, v := range data {
		cond[k] = eudore.GetStrings(v)
	}
	return cond
}

// Name 方法返回条件名称。
func (cond conditionParams) Name() string {
	return "params"
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
