package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/eudore/eudore"
)

type conditionAnd struct {
	Conditions []condition
	Names      []string
}

// Match conditionAnd.
func (cond *conditionAnd) Match(ctx eudore.Context) bool {
	for _, i := range cond.Conditions {
		if !i.Match(ctx) {
			return false
		}
	}
	return true
}

func (cond *conditionAnd) UnmarshalJSON(body []byte) error {
	return cond.unmarshalJSON("and", body)
}

func (cond *conditionAnd) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	for i := range cond.Names {
		if buf.Len() > 1 {
			buf.WriteByte(',')
		}
		data, err := json.Marshal(cond.Conditions[i])
		if err == nil {
			buf.WriteByte('"')
			buf.WriteString(cond.Names[i])
			buf.WriteByte('"')
			buf.WriteByte(':')
			buf.Write(data)
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (cond *conditionAnd) unmarshalJSON(t string, body []byte) error {
	var data map[string]json.RawMessage
	err := json.Unmarshal(body, &data)
	if err != nil {
		return fmt.Errorf(ErrPolicyConditionsUnmarshalError, t, err)
	}
	names, err := getMapKeys(len(data), body)
	if err != nil {
		return err
	}

	cond.Names = make([]string, 0, len(data))
	cond.Conditions = make([]condition, 0, len(data))
	for _, name := range names {
		fn, ok := DefaultPolicyConditions[name]
		if !ok {
			return fmt.Errorf(ErrPolicyConditionsUnmarshalError, t, fmt.Errorf("undefined condition %s", name))
		}

		val := fn()
		err := json.Unmarshal(data[name], val)
		if err != nil {
			return fmt.Errorf(ErrPolicyConditionsParseError, name, err)
		}
		cond.Names = append(cond.Names, name)
		cond.Conditions = append(cond.Conditions, val.(condition))
	}
	return nil
}

func getMapKeys(size int, body []byte) ([]string, error) {
	depth := 0
	dec := json.NewDecoder(bytes.NewReader(body))
	keys := make([]string, 0, size)
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}

		b, ok := t.(json.Delim)
		if ok {
			switch b {
			case '{', '[':
				depth++
				continue
			case '}', ']':
				depth--
				continue
			}
		}

		if depth == 1 {
			str, ok := t.(string)
			if !ok {
				return nil, fmt.Errorf("invalid character '%v' looking for beginning of value", t)
			}
			keys = append(keys, str)
		}
	}
	return keys, nil
}

type conditionOr struct {
	Conditions []condition
	Names      []string
}

func (cond *conditionOr) Match(ctx eudore.Context) bool {
	for _, i := range cond.Conditions {
		if i.Match(ctx) {
			return true
		}
	}
	return len(cond.Conditions) != 0
}

func (cond *conditionOr) UnmarshalJSON(body []byte) error {
	and := &conditionAnd{}
	err := and.unmarshalJSON("or", body)
	if err != nil {
		return err
	}
	cond.Conditions = and.Conditions
	cond.Names = and.Names
	return nil
}

func (cond *conditionOr) MarshalJSON() ([]byte, error) {
	return json.Marshal(&conditionAnd{cond.Conditions, cond.Names})
}

type conditionSourceIP struct {
	SourceIP []*net.IPNet `json:"sourceip,omitempty"`
}

func (cond *conditionSourceIP) Match(ctx eudore.Context) bool {
	for _, ip := range cond.SourceIP {
		if ip.Contains(net.ParseIP(ctx.RealIP())) {
			return true
		}
	}
	return false
}

func (cond *conditionSourceIP) UnmarshalJSON(body []byte) error {
	var strs []string
	err := json.Unmarshal(body, &strs)
	if err != nil {
		return fmt.Errorf(ErrPolicyConditionsUnmarshalError, "sourceip", err)
	}

	ipnets := make([]*net.IPNet, 0, len(strs))
	for _, i := range strs {
		if strings.IndexByte(i, '/') == -1 {
			i += "/32"
		}
		_, ipnet, err := net.ParseCIDR(i)
		if err != nil {
			return fmt.Errorf(ErrPolicyConditionParseError, "sourceip", "cidr", err)
		}
		ipnets = append(ipnets, ipnet)
	}
	cond.SourceIP = ipnets
	return nil
}

func (cond *conditionSourceIP) MarshalJSON() ([]byte, error) {
	nets := make([]string, len(cond.SourceIP))
	for i, ip := range cond.SourceIP {
		nets[i] = ip.String()
	}
	return json.Marshal(nets)
}

type conditionDate struct {
	After  time.Time `json:"after"`
	Before time.Time `json:"before"`
}

// Match method matches the current Date range.
func (cond *conditionDate) Match(eudore.Context) bool {
	now := time.Now()
	return (cond.After.IsZero() || now.After(cond.After)) &&
		(cond.Before.IsZero() || now.Before(cond.Before))
}

const durationDay = 24 * time.Hour

func (cond *conditionDate) UnmarshalJSON(body []byte) error {
	var date struct {
		After  string `json:"after"`
		Before string `json:"before"`
	}
	err := json.Unmarshal(body, &date)
	if err != nil {
		return fmt.Errorf(ErrPolicyConditionsUnmarshalError, "date", err)
	}
	if len(date.After) == 10 {
		date.After += " 00:00:00"
	}
	if len(date.Before) == 10 {
		date.Before += " 00:00:00"
		defer func() {
			cond.Before = cond.Before.Add(durationDay - 1)
		}()
	}

	cond.After, err = time.ParseInLocation("2006-01-02 15:04:05", date.After, eudore.DefaultValueTimeLocation)
	if err != nil && date.After != "" {
		return fmt.Errorf(ErrPolicyConditionParseError, "date", "after", err)
	}
	cond.Before, err = time.ParseInLocation("2006-01-02 15:04:05", date.Before, eudore.DefaultValueTimeLocation)
	if err != nil && date.Before != "" {
		return fmt.Errorf(ErrPolicyConditionParseError, "date", "before", err)
	}
	return nil
}

func (cond *conditionDate) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	if !cond.After.IsZero() {
		_, offset := cond.After.Zone()
		t := (cond.After.UnixNano() + int64(offset)*int64(time.Second)) % int64(durationDay)
		buf.WriteString(`"after":"`)
		if t == 0 {
			buf.WriteString(cond.After.Format("2006-01-02"))
		} else {
			buf.WriteString(cond.After.Format("2006-01-02 15:04:05"))
		}
		buf.WriteByte('"')
	}
	if !cond.After.IsZero() && !cond.Before.IsZero() {
		buf.WriteByte(',')
	}
	if !cond.Before.IsZero() {
		_, offset := cond.Before.Zone()
		t := (cond.Before.UnixNano() + int64(offset)*int64(time.Second) + 1) % int64(durationDay)
		buf.WriteString(`"before":"`)
		if t == 0 {
			buf.WriteString(cond.Before.Format("2006-01-02"))
		} else {
			buf.WriteString(cond.Before.Format("2006-01-02 15:04:05"))
		}
		buf.WriteByte('"')
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

type conditionTime struct {
	After  int64 `json:"after"`
	Before int64 `json:"before"`
}

// Match method matches the current time range.
func (cond *conditionTime) Match(eudore.Context) bool {
	now := time.Now()
	curr := int64(now.Hour()*int(time.Hour) +
		now.Minute()*int(time.Minute) +
		now.Second()*int(time.Second))
	return cond.After < curr && curr < cond.Before
}

func (cond *conditionTime) UnmarshalJSON(body []byte) error {
	var date struct {
		After  string `json:"after"`
		Before string `json:"before"`
	}
	err := json.Unmarshal(body, &date)
	if err != nil {
		return fmt.Errorf(ErrPolicyConditionsUnmarshalError, "time", err)
	}

	cond.After = 0
	if date.After != "" {
		after, err := time.Parse("15:04:05", date.After)
		if err != nil {
			return fmt.Errorf(ErrPolicyConditionParseError, "time", "after", err)
		}
		cond.After = after.UnixNano() - after.Truncate(durationDay).UnixNano()
	}
	cond.Before = int64(durationDay)
	if date.Before != "" {
		before, err := time.Parse("15:04:05", date.Before)
		if err != nil {
			return fmt.Errorf(ErrPolicyConditionParseError, "time", "before", err)
		}
		cond.Before = int64(before.Hour()*int(time.Hour) +
			before.Minute()*int(time.Minute) +
			before.Second()*int(time.Second))
	}
	if cond.After > cond.Before {
		cond.Before += int64(durationDay)
	}
	return nil
}

func (cond *conditionTime) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	if cond.After != 0 {
		fmt.Fprintf(buf, `"after":"%02d:%02d:%02d"`, cond.After/int64(time.Hour), cond.After/int64(time.Minute)%60, cond.After/int64(time.Second)%60)
	}
	if cond.After != 0 && cond.Before != int64(durationDay) {
		buf.WriteByte(',')
	}
	if cond.Before != int64(24*time.Hour) {
		fmt.Fprintf(buf, `"before":"%02d:%02d:%02d"`, cond.Before/int64(time.Hour), cond.Before/int64(time.Minute)%60, cond.Before/int64(time.Second)%60)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

type conditionMethod []string

func (cond conditionMethod) Match(ctx eudore.Context) bool {
	method := ctx.Method()
	for _, i := range cond {
		if i == method {
			return true
		}
	}
	return false
}

type conditionPath []string

func (cond conditionPath) Match(ctx eudore.Context) bool {
	path := ctx.Path()
	for _, i := range cond {
		if i == path {
			return true
		}
	}
	return false
}

type conditionParams map[string][]string

func (cond conditionParams) Match(ctx eudore.Context) bool {
	params := ctx.Params()
	for key, vals := range cond {
		if stringSliceNotIn(vals, params.Get(key)) {
			return false
		}
	}
	return true
}

func stringSliceNotIn(strs []string, str string) bool {
	for _, i := range strs {
		if i == str {
			return false
		}
	}
	return true
}

type conditionRate struct {
	Speed int64
	Max   int64
	rate  rateBucket
}

func (cond *conditionRate) Match(eudore.Context) bool {
	_, _, ok := cond.rate.Allow()
	return ok
}

func (cond *conditionRate) UnmarshalJSON(body []byte) error {
	var rate struct {
		Speed int64 `json:"speed"`
		Max   int64 `json:"max"`
	}
	err := json.Unmarshal(body, &rate)
	if err != nil {
		return err
	}

	cond.Speed = eudore.GetAnyDefault(rate.Speed, 1)
	cond.Max = eudore.GetAnyDefault(rate.Max, 1)
	cond.rate = rateBucket{
		speed: int64(time.Second) / cond.Speed,
		max:   int64(time.Second) / cond.Speed * cond.Max,
	}
	return nil
}

func (cond *conditionRate) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, `{"speed":%d,"max":%d}`, cond.Speed, cond.Max), nil
}

type conditionVersion struct {
	Name    string
	Version []conditionVersionValue
}
type conditionVersionValue struct {
	Name string
	Min  []int
	Max  []int
}

type conditionVersionString struct {
	Name    string                        `json:"name"`
	Version []conditionVersionValueString `json:"version"`
}

type conditionVersionValueString struct {
	Name string `json:"name"`
	Min  string `json:"min,omitempty"`
	Max  string `json:"max,omitempty"`
}

func (cond *conditionVersion) Match(ctx eudore.Context) bool {
	val := ctx.GetParam(cond.Name)
	if val == "" {
		return false
	}

	for _, v := range cond.Version {
		if strings.HasPrefix(val, v.Name) {
			list := condconditionVerisonParse(val[len(v.Name)+1:])
			for i := range list {
				if i < len(v.Min) && v.Min[i] > list[i] {
					return false
				}
				if i < len(v.Max) && v.Max[i] < list[i] {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (cond *conditionVersion) UnmarshalJSON(body []byte) error {
	var data conditionVersionString
	err := json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	cond.Name = data.Name
	for _, v := range data.Version {
		cond.Version = append(cond.Version, conditionVersionValue{
			v.Name,
			condconditionVerisonParse(v.Min),
			condconditionVerisonParse(v.Max),
		})
	}
	return nil
}

func (cond *conditionVersion) MarshalJSON() ([]byte, error) {
	data := &conditionVersionString{Name: cond.Name}
	for _, v := range cond.Version {
		data.Version = append(data.Version, conditionVersionValueString{
			v.Name,
			condconditionVerisonString(v.Min),
			condconditionVerisonString(v.Max),
		})
	}
	return json.Marshal(data)
}

func condconditionVerisonParse(str string) []int {
	if str == "" {
		return nil
	}

	list := make([]int, 1, 4)
	for _, s := range str {
		if 0x2F < s && s < 0x3A {
			list[len(list)-1] = list[len(list)-1]*10 + int(s-0x30)
		} else {
			list = append(list, 0)
		}
	}
	return list
}

func condconditionVerisonString(list []int) string {
	buf := make([]byte, 0, 32)
	for i := range list {
		if i != 0 {
			buf = append(buf, '.')
		}
		buf = strconv.AppendInt(buf, int64(list[i]), 10)
	}
	return string(buf)
}
