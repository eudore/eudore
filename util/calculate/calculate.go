// From: github.com/khalily/caculator
package calculate

import (
	"container/list"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	LEVEL1 = iota
	LEVEL2
	LEVEL3
)

var po = make(map[string]int)

func init() {
	po["+"] = LEVEL1
	po["-"] = LEVEL1
	po["*"] = LEVEL2
	po["/"] = LEVEL2
}

func Result(expr string) (int, error) {
	expr = strings.Replace(expr, " ", "", -1)
	exps, err := ParseExp(expr)
	if err != nil {
		return 0, err
	}
	return Caculate(Pre2stuf(exps)), nil
}

func ParseExp(s string) (exps []string, err error) {
	re, err := regexp.Compile("[0-9]+|[+*/\\-\\(\\)]{1}")
	if err != nil {
		fmt.Println("regexp compile error:", err)
		return
	}
	for _, exp := range re.FindAll([]byte(s), -1) {
		exps = append(exps, string(exp))
	}
	return
}

func isPop(list *list.List, s string) (op []string, ok bool) {
	switch string(s) {
	case "(":
		ok = false
		return
	case ")":
		ok = true
		cur := list.Back()
		for {
			prev := cur.Prev()
			if curValue, ok2 := cur.Value.(string); ok2 {
				if string(curValue) == "(" {
					list.Remove(cur)
					return
				}
				op = append(op, curValue)
				list.Remove(cur)
				cur = prev
			}
		}
	default:
		for cur := list.Back(); cur != nil; {
			prev := cur.Prev()
			if curValue, ok2 := cur.Value.(string); ok2 {
				if level, ok3 := po[curValue]; ok3 && level >= po[s] {
					ok = true
					op = append(op, curValue)
					// fmt.Println(op)
					list.Remove(cur)
				} else if curValue == "(" {
					// fmt.Println(curValue, op)
					if len(op) != 0 {
						ok = true
					} else {
						ok = false
					}
					return
				}
			}
			cur = prev
		}
	}
	return
}

func isOperate(s string) bool {
	re, _ := regexp.Compile("[+*/\\-\\(\\)]{1}")
	ok := re.MatchString(s)
	// fmt.Println(ok, s)
	return ok
}

func Pre2stuf(exps []string) (exps2 []string) {
	list1 := list.New()
	list2 := list.New()

	for _, exp := range exps {
		if isOperate(exp) {
			if op, ok := isPop(list1, exp); ok {
				for _, s := range op {
					list2.PushBack(s)
				}
			}
			if exp == ")" {
				continue
			}
			list1.PushBack(exp)
		} else {
			list2.PushBack(exp)
		}
	}

	for cur := list1.Back(); cur != nil; cur = cur.Prev() {
		// fmt.Print(cur.Value)
		list2.PushBack(cur.Value)
	}

	for cur := list2.Front(); cur != nil; cur = cur.Next() {
		if curValue, ok := cur.Value.(string); ok {
			exps2 = append(exps2, curValue)
		}
	}
	return
}

func Caculate(exps []string) int {
	list1 := list.New()

	for _, s := range exps {
		if isOperate(s) {
			back := list1.Back()
			prev := back.Prev()
			backVal, _ := back.Value.(int)
			prevVal, _ := prev.Value.(int)
			var res int
			switch s {
			case "+":
				res = prevVal + backVal
			case "-":
				res = prevVal - backVal
			case "*":
				res = prevVal * backVal
			case "/":
				res = prevVal / backVal
			}
			list1.Remove(back)
			list1.Remove(prev)
			list1.PushBack(res)
		} else {
			v, _ := strconv.Atoi(s)
			list1.PushBack(v)
		}
	}
	if list1.Len() != 1 {
		fmt.Println("caculate error")
		os.Exit(1)
	}
	res, _ := list1.Back().Value.(int)
	return res
}
