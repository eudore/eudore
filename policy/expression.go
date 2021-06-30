package policy

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eudore/eudore"
)

/*
[
    {kind: and       name:   data:[]},
    {kind: value     name:   value: []},
    {kind: range     name:   Min:    Max},
    {kind: time      name:   after:  before:},
    {kind: string    name:   value:  expr: prefix like}
]
*/

type expressionBase struct {
	Kind  string `json:"kind"`
	Table string `json:"table"`
	Name  string `json:"name"`
}

// CheckTable 方法检查表达式是否符合表和字段。
func (expr expressionBase) CheckTable(tb string, fields []string) bool {
	if expr.Table != "" && expr.Table != tb {
		return true
	}
	for _, i := range fields {
		if i == expr.Name {
			return false
		}
	}
	return true
}

// NewExpressions 函数使用json生成多表达式。
func NewExpressions(body []byte) (Expressions, error) {
	var data []RawMessage
	err := json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	var exprs Expressions
	for i := range data {
		var base expressionBase
		err := json.Unmarshal(data[i], &base)
		if err != nil {
			return nil, err
		}

		fn, ok := newExpressionFuncs[base.Kind]
		if !ok {
			continue
		}
		expr, err := fn(data[i])
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}
	return exprs, nil
}

// Expression 方法Expressions生成多表达式合集。
func (exprs Expressions) Expression(tb string, fields []string) ([]string, []interface{}) {
	var strs []string
	var vals []interface{}
	for i := range exprs {
		str, val := exprs[i].Expression(tb, fields)
		if str != "" {
			strs = append(strs, str)
			vals = append(vals, val...)
		}
	}
	return strs, vals
}

type expressionAnd struct {
	Data Expressions `json:"data"`
}

func newExpressionAnd(body []byte) (Expression, error) {
	var exprdata struct {
		Data RawMessage `json:"data"`
	}
	json.Unmarshal(body, &exprdata)
	exprs, err := NewExpressions(exprdata.Data)
	return expressionAnd{exprs}, err
}

func (exprs expressionAnd) Expression(tb string, fields []string) (string, []interface{}) {
	strs, vals := exprs.Data.Expression(tb, fields)
	return "(" + strings.Join(strs, " AND ") + ")", vals
}

type expressionOr struct {
	Data Expressions `json:"data"`
}

func newExpressionOr(body []byte) (Expression, error) {
	var exprdata struct {
		Data RawMessage `json:"data"`
	}
	json.Unmarshal(body, &exprdata)
	exprs, err := NewExpressions(exprdata.Data)
	return expressionOr{exprs}, err
}

func (exprs expressionOr) Expression(tb string, fields []string) (string, []interface{}) {
	strs, vals := exprs.Data.Expression(tb, fields)
	return "(" + strings.Join(strs, " OR ") + ")", vals
}

type expressionValue struct {
	expressionBase
	Not    bool          `json:"not,omitempty"`
	Value  []interface{} `json:"value"`
	format string
}

func newExpressionValue(body []byte) (Expression, error) {
	var expr expressionValue
	err := json.Unmarshal(body, &expr)
	if err != nil {
		return nil, err
	}
	expr.Value = getDataFunc(expr.Value)
	switch len(expr.Value) {
	case 0:
		return nil, fmt.Errorf("Expression Value")
	case 1:
		expr.format = fmt.Sprintf("%s = ?", expr.Name)
	default:
		expr.format = fmt.Sprintf("%s in (%s)", expr.Name, strings.Repeat(",?", len(expr.Value))[1:])
	}
	if expr.Not {
		expr.format = "NOT " + expr.format
	}
	return &expr, nil
}

func (expr expressionValue) Expression(tb string, fields []string) (string, []interface{}) {
	if expr.CheckTable(tb, fields) {
		return "", nil
	}
	return tb + "." + expr.format, expr.Value
}

type expressionRange struct {
	expressionBase
	Not    bool        `json:"not,omitempty"`
	Min    interface{} `json:"min"`
	Max    interface{} `json:"max"`
	values []interface{}
	format string
}

func newExpressionRange(body []byte) (Expression, error) {
	var expr expressionRange
	err := json.Unmarshal(body, &expr)
	if err != nil {
		return nil, err
	}

	if expr.Min == nil && expr.Max == nil {
		return nil, fmt.Errorf("ExpressionRange min or max must not nil")
	}
	if expr.Min == nil {
		expr.values = getDataFunc([]interface{}{expr.Max})
		expr.format = fmt.Sprintf("%s <= ?", expr.Name)
	} else if expr.Max == nil {
		expr.values = getDataFunc([]interface{}{expr.Min})
		expr.format = fmt.Sprintf("%s >= ?", expr.Name)
	} else {
		expr.values = getDataFunc([]interface{}{expr.Min, expr.Max})
		expr.format = fmt.Sprintf("%s BETWEEN ? AND ?", expr.Name)
	}
	if expr.Not {
		expr.format = "NOT " + expr.format
	}
	return &expr, nil
}

func (expr expressionRange) Expression(tb string, fields []string) (string, []interface{}) {
	if expr.CheckTable(tb, fields) {
		return "", nil
	}
	return tb + "." + expr.format, expr.values
}

type expressionSql struct {
	expressionBase
	Sql    string        `json:"sql"`
	Value  []interface{} `json:"value"`
	format string
}

func newExpressionSql(body []byte) (Expression, error) {
	var expr expressionSql
	err := json.Unmarshal(body, &expr)
	if err != nil {
		return nil, err
	}

	expr.Value = getDataFunc(expr.Value)
	return &expr, nil
}

func (expr expressionSql) Expression(tb string, fields []string) (string, []interface{}) {
	if expr.CheckTable(tb, fields) {
		return "", nil
	}
	return strings.Replace(expr.format, "{{table}}", tb, -1), expr.Value
}

func getDataFunc(data []interface{}) []interface{} {
	for i := range data {
		name, ok := data[i].(string)
		if ok && strings.HasPrefix(name, "value:") {
			pos := strings.IndexByte(name[6:], ':')
			if pos != -1 {
				fn, ok := newValueFuncs[name[6:pos+6]]
				if ok {
					data[i] = fn(name[pos+7:])
				}
			}
		}
	}
	return data
}

func newValueParam(str string) func(eudore.Context) interface{} {
	return func(ctx eudore.Context) interface{} {
		val := ctx.GetParam(str)
		if val == "" {
			val = "0"
		}
		return val
	}
}

func newValueQuery(str string) func(eudore.Context) interface{} {
	return func(ctx eudore.Context) interface{} {
		val := ctx.GetQuery(str)
		if val == "" {
			val = "0"
		}
		return val
	}
}
