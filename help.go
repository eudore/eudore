package eudore

import (
	"fmt"
	"strings"
	"strconv"
	"math/rand"
	"encoding/json"
)

func arrayclean(names []string) (n []string){
	for _, name := range names {
		if name != "" {
			n = append(n, name)
		}
	}
	return
}

// Each string strs handle element, if return is null, then delete this a elem.
func eachstring(strs []string, fn func(string) string) (s []string){
	for _, i := range strs {
		i = fn(i)
		if i != "" {
			s = append(s, i)
		}
	}
	return
}

// Use sep to split str into two strings.
func split2byte(str string,b byte) (string, string) {
	pos := strings.IndexByte(str, b)
	if pos == -1 {
		return "", ""
	}
	return str[:pos], str[pos + 1:]
}

/*
func split2(str string,sep string) (string, string) {
	pos := strings.IndexByte(str, sep[0])
	if pos == -1 {
		return "", ""
	}
	return str[:pos], str[pos + len(sep):]
}
*/

// Env to Arg
func env2arg(str string) string {
	k, v := split2byte(str, '=')
	k = strings.ToLower(strings.Replace(k, "_", ".", -1))[4:]
	return fmt.Sprintf("--%s=%s", k, v)
}

func GetRandomString() string {
	const letters = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	result := make([]byte, 16)
	for i := range result {
		result[i] = letters[rand.Intn(61)]
	}
	return string(result)
}

// 模式匹配对象，运行试验带'*'的模式
func MatchStar(obj, patten string) bool {
	ps := strings.Split(patten,"*")
	if len(ps) == 0 {
		return patten == obj
	}
	if !strings.HasPrefix(obj, ps[0]) {
		return false
	}
	for _,i := range ps {
		if i == "" {
			continue
		}
		pos := strings.Index(obj, i)
		if pos == -1 {
			return false
		}
		obj = obj[pos + len(i):]
	}
	return true
}

// test function, json formatted output args.
//
// 测试函数，json格式化输出args。
func Json(args ...interface{}) {
	indent, err := json.MarshalIndent(&args, "", "\t")
	fmt.Println(string(indent), err)
}




func GetInt(i interface{}) int {
	return GetDefultInt(i, 0)
}

func GetDefultInt(i interface{},n int) int {
	if v, ok := i.(int); ok {
		return v
	}	
	if v, err := strconv.Atoi(GetDefaultString(i, "")); err == nil {
		return v
	}
	return n
}

func GetString(i interface{}) string {
	return GetDefaultString(i, "")
}

func GetDefaultString(i interface{},str string) string {
	if v, ok := i.(string); ok {
		return v
	}
	if v, ok := i.(fmt.Stringer); ok {
		return v.String()
	}
	return str
}