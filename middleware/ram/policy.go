package ram

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/eudore/eudore"
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
		Effect     bool
		Action     []string
		Resource   []string
		Conditions *Conditions `json:"conditions,omitempty"`
	}
	// ConditionS 定义PBAC使用的条件对象。
	Conditions struct {
		Conditions []Condition
	}
	Condition interface {
		Name() string
		Match(ctx eudore.Context) bool
	}
	NewConditionFunc func(interface{}) Condition
	ConditionOr      struct {
		Conditions []Condition
	}
	ConditionAnd struct {
		Conditions []Condition
	}
	ConditionSourceIp struct {
		SourceIp []*net.IPNet
	}
	ConditionTime struct {
		Befor time.Time `json:"befor"`
		After time.Time `json:"after"`
	}
	ConditionMethod struct {
		Methods []string
	}
	ConditionRequest struct {
		Header []*RequestHeader
	}
	RequestHeader struct {
		Name   string
		Values []string
	}
	ConditionUserAgent struct {
		UserAgent []string
	}
	ConditionBrowser struct {
		Browsers []*UserBrowser
	}
	UserBrowser struct {
		Name string
		Min  int8
		Max  int8
	}
)

var conditionnews = make(map[string]NewConditionFunc)

func init() {
	conditionnews["or"] = NewConditionOr
	conditionnews["and"] = NewConditionAnd
	conditionnews["sourceip"] = NewConditionSourceIp
	conditionnews["time"] = NewConditionTime
	conditionnews["useragent"] = NewConditionUserAgent
	conditionnews["method"] = NewConditionMethod
	conditionnews["browser"] = NewConditionBrowser
}

func NewPolicyStringJSON(str string) *Policy {
	policy := Policy{}
	err := json.Unmarshal([]byte(str), &policy)
	if err != nil {
		panic(err)
	}

	return &policy
}

func (stat *Statement) MatchAction(action string) bool {
	for _, i := range stat.Action {
		if MatchStar(action, i) {
			return true
		}
	}
	return false
}

func (stat *Statement) MatchResource(resource string) bool {
	for _, i := range stat.Resource {
		if MatchStar(resource, i) {
			return true
		}
	}
	return false
}

func (stat *Statement) MatchCondition(ctx eudore.Context) bool {
	if stat.Conditions == nil {
		return true
	}
	return stat.Conditions.Match(ctx)
}

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

func (conds *Conditions) Match(ctx eudore.Context) bool {
	for _, i := range conds.Conditions {
		if !i.Match(ctx) {
			return false
		}
	}
	return true
}

func (conds *Conditions) UnmarshalJSON(body []byte) error {
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err
	}
	conds.Conditions = NewConditions(data)
	return nil
}
func (conds *Conditions) MarshalJSON() ([]byte, error) {
	data := make(map[string]Condition)
	for _, cond := range conds.Conditions {
		data[cond.Name()] = cond
	}
	return json.Marshal(data)
}

func NewConditionOr(i interface{}) Condition {
	data, ok := i.(map[string]interface{})
	if !ok {
		return nil
	}
	return &ConditionOr{Conditions: NewConditions(data)}
}

func (cond *ConditionOr) Name() string {
	return "or"
}
func (cond *ConditionOr) Match(ctx eudore.Context) bool {
	for _, i := range cond.Conditions {
		if i.Match(ctx) {
			return true
		}
	}
	return false
}

func NewConditionAnd(i interface{}) Condition {
	data, ok := i.(map[string]interface{})
	if !ok {
		return nil
	}
	return &ConditionAnd{Conditions: NewConditions(data)}
}

func (cond *ConditionAnd) Name() string {
	return "and"
}
func (cond *ConditionAnd) Match(ctx eudore.Context) bool {
	for _, i := range cond.Conditions {
		if !i.Match(ctx) {
			return false
		}
	}
	return true
}

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

func (cond *ConditionSourceIp) Name() string {
	return "sourceip"
}
func (cond *ConditionSourceIp) Match(ctx eudore.Context) bool {
	for _, i := range cond.SourceIp {
		if i.Contains(net.ParseIP(ctx.RealIP())) {
			return true
		}
	}
	return false
}

func (cond *ConditionSourceIp) MarshalJSON() ([]byte, error) {
	data := make([]string, 0, len(cond.SourceIp))
	for _, i := range cond.SourceIp {
		data = append(data, i.IP.String())
	}
	return json.Marshal(data)
}

func NewConditionTime(i interface{}) Condition {
	cond := &ConditionTime{}
	data, ok := i.(map[string]interface{})
	if ok {
		cond.After = TimeParse(eudore.GetString(data["after"]))
		cond.Befor = TimeParse(eudore.GetString(data["befor"]))
		if cond.Befor.Unix() == 0 {
			cond.Befor = time.Unix(9223372036854775807, 9223372036854775807)
		}
	}
	return cond
}

func (cond *ConditionTime) Name() string {
	return "time"
}
func (cond *ConditionTime) Match(ctx eudore.Context) bool {
	current := time.Now()
	return current.Before(cond.Befor) && current.After(cond.After)
}

func NewConditionMethod(i interface{}) Condition {
	return &ConditionMethod{Methods: GetArrayString(i)}
}

func (cond *ConditionMethod) Name() string {
	return "method"
}

func (cond *ConditionMethod) Match(ctx eudore.Context) bool {
	method := ctx.Method()
	for _, i := range cond.Methods {
		if i == method {
			return true
		}
	}
	return false
}

func (cond *ConditionMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(cond.Methods)
}

func NewConditionUserAgent(i interface{}) Condition {
	return &ConditionUserAgent{UserAgent: GetArrayString(i)}
}

func (cond *ConditionUserAgent) Name() string {
	return "useragent"
}

func (cond *ConditionUserAgent) Match(ctx eudore.Context) bool {
	return false
}

func (cond *ConditionUserAgent) MarshalJSON() ([]byte, error) {
	return json.Marshal(cond.UserAgent)
}

func NewConditionBrowser(i interface{}) Condition {
	cond := &ConditionBrowser{}
	for _, name := range GetArrayString(i) {
		cond.Browsers = append(cond.Browsers, NewUserBrowser(name))
	}
	return cond
}

func (cond *ConditionBrowser) Name() string {
	return "browser"
}

func (cond *ConditionBrowser) Match(ctx eudore.Context) bool {
	// method := ctx.Method()
	// for _, i := range cond.Browsers {
	// 	if i == method {
	// 		return true
	// 	}
	// }
	// return false

	return true
}

func (cond *ConditionBrowser) MarshalJSON() ([]byte, error) {
	return json.Marshal(cond.Browsers)
}

func NewUserBrowser(name string) *UserBrowser {
	pos := strings.LastIndexByte(name, '/')
	if pos == -1 {
		return &UserBrowser{Name: name, Min: 0, Max: 0x7f}
	}
	// version := name[pos+1:]
	name = name[:pos]
	var min, max int8 = 0, 0x7f

	// TODO: not use verison
	return &UserBrowser{Name: name, Min: min, Max: max}
}

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
	return time.Unix(0, 0)
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
