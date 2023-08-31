package eudore

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func loadDefaultFuncDefine(fc FuncCreator) {
	f := fc.RegisterFunc
	// func(T) bool
	_ = f("zero", fbZero[string], fbZero[int], fbZero[uint], fbZero[float64], fbZero[bool], fcAnyZero)
	_ = f("nozero", fbNozero[string], fbNozero[int], fbNozero[uint], fbNozero[float64], fbNozero[bool], fcAnyNozero)
	_ = f("min", fbMin[int], fbMin[uint], fbMin[float64], fbStringMin) // min=1
	_ = f("max", fbMax[int], fbMax[uint], fbMax[float64], fbStringMax) // max=1
	_ = f("equal", fbEqual[string], fbEqual[int], fbEqual[uint])       // equal=string equal!=string
	_ = f("enum", fbEnum[int], fbEnum[uint], fbEnum[string])           // enum=1,2,3 enum!=1,2,3
	_ = f("len", fbStringLen, fbAnyLen)                                // len=3 len<3 len>3
	_ = f("num", fbStringNum)
	_ = f("integer", fbStringInteger)
	_ = f("domain", fbStringDomain)
	_ = f("mail", fbStringMail)
	_ = f("phone", fbStringPhone)
	_ = f("regexp", fbStringRegexp)
	_ = f("patten", fbStringpPatten)
	_ = f("prefix", fbStringFuncBool(strings.HasPrefix))   // prefix=string prefix!=string
	_ = f("suffix", fbStringFuncBool(strings.HasSuffix))   // suffix=string
	_ = f("contains", fbStringFuncBool(strings.Contains))  // contains=string
	_ = f("fold", fbStringFuncBool(strings.EqualFold))     // fold=string
	_ = f("count", fbStringFuncInt(strings.Count))         // count=number,string count<number,string count>number,string
	_ = f("compare", fbStringFuncInt(strings.Compare))     // compare=number,string
	_ = f("index", fbStringFuncInt(strings.Index))         // index=number,string
	_ = f("lastindex", fbStringFuncInt(strings.LastIndex)) // lastindex=number,string
	_ = f("after", fbTimeAfter)
	_ = f("before", fbTimeBefore)
	// func(T) T
	_ = f("default", fsDefault[string], fsDefault[int], fsDefault[uint], fsDefault[float64], fsDefault[bool], fsAnyDefault)
	_ = f("value", fsValue[string], fsValue[int], fsValue[uint], fsValue[float64], fsValue[bool], fsTimeValue)
	_ = f("add", fsAdd[int], fsAdd[uint], fsAdd[float64], fsTimeAdd)
	_ = f("now", fsStringNow, fsTimeNow)
	_ = f("len", fsStringLen)
	_ = f("md5", fsStringMd5)
	_ = f("tolower", strings.ToLower)                           // tolower
	_ = f("toupper", strings.ToUpper)                           // toupper
	_ = f("totitle", strings.Title)                             //nolint:staticcheck
	_ = f("replace", fsStringReplace)                           // replace=old,new replace=-1,old,new
	_ = f("trimspace", strings.TrimSpace)                       // trimspace
	_ = f("trim", fsStringFuncString(strings.Trim))             // trim=string
	_ = f("trimprefix", fsStringFuncString(strings.TrimPrefix)) // trim=trimprefix
	_ = f("trimsuffix", fsStringFuncString(strings.TrimSuffix)) // trim=trimsuffix
	_ = f("hide", fsStringHide)
	_ = f("hidesurname", fsStringHideSurname)
	_ = f("hidename", fsStringHideName)
	_ = f("hidemail", fsStringHideMail)
	_ = f("hidephone", fsStringHidePhone)
	// alias
	_ = f("must", fbNozero[string], fbNozero[int], fbNozero[uint], fbNozero[float64], fbNozero[bool], fcAnyNozero)
	_ = f("eq", fbEqual[string], fbEqual[int], fbEqual[uint])
}

func trimFuncOperate(str string) string {
	for _, r := range []byte{'!', '<', '>', '=', ':'} {
		if len(str) > 0 && str[0] == r {
			str = str[1:]
		}
	}
	return str
}

func fbZero[T int | uint | float64 | string | bool](i T) bool {
	var t T
	return i == t
}

func fcAnyZero(i any) bool {
	t, ok := i.(time.Time)
	if ok {
		return t.IsZero()
	}
	return i == nil || reflect.ValueOf(i).IsZero()
}

func fbNozero[T int | uint | float64 | string | bool](i T) bool {
	var t T
	return i != t
}

// fcAnyNozero 函数验证一个对象是否为零值，使用reflect.Value.IsZero函数实现。
func fcAnyNozero(i any) bool {
	return !fcAnyZero(i)
}

// fcgMin 函数生成一个验证value最小值的验证函数。
func fbMin[T int | uint | float64](s string) (func(T) bool, error) {
	val, err := GetAnyByStringWithError[T](trimFuncOperate(s))
	if err != nil {
		return nil, err
	}
	return func(num T) bool {
		return num >= val
	}, nil
}

// fbMax 函数生成一个验证value最大值的验证函数。
func fbMax[T int | uint | float64](s string) (func(T) bool, error) {
	val, err := GetAnyByStringWithError[T](trimFuncOperate(s))
	if err != nil {
		return nil, err
	}
	return func(num T) bool {
		return num <= val
	}, nil
}

// fbStringMin 函数生成一个验证string最小值的验证函数。
func fbStringMin(s string) (func(string) bool, error) {
	min, err := strconv.ParseInt(trimFuncOperate(s), 10, 32)
	if err != nil {
		return nil, err
	}
	intmin := int(min)
	return func(arg string) bool {
		num, err := strconv.Atoi(arg)
		if err != nil {
			return false
		}
		return num >= intmin
	}, nil
}

// fbStringMax 函数生成一个验证string最大值的验证函数。
func fbStringMax(s string) (func(string) bool, error) {
	max, err := strconv.ParseInt(trimFuncOperate(s), 10, 32)
	if err != nil {
		return nil, err
	}
	intmax := int(max)
	return func(arg string) bool {
		num, err := strconv.Atoi(arg)
		if err != nil {
			return false
		}
		return num <= intmax
	}, nil
}

func fbEqual[T string | int | uint](s string) (func(T) bool, error) {
	b := !strings.HasPrefix(s, "!=")
	val, err := GetAnyByStringWithError[T](trimFuncOperate(s))
	if err != nil {
		return nil, err
	}
	return func(arg T) bool {
		return val == arg == b
	}, nil
}

func fbEnum[T string | int | uint](s string) (func(arg T) bool, error) {
	b := !strings.HasPrefix(s, "!=")
	strs := strings.Split(trimFuncOperate(s), ",")
	values := make([]T, len(strs))
	for i := range strs {
		val, err := GetAnyByStringWithError[T](strs[i])
		if err != nil {
			return nil, err
		}
		values[i] = val
	}

	if len(strs) < 9 {
		return func(arg T) bool {
			for _, val := range values {
				if val == arg {
					return b
				}
			}
			return !b
		}, nil
	}

	macths := make(map[T]struct{}, len(strs))
	for i := range values {
		macths[values[i]] = struct{}{}
	}
	return func(arg T) bool {
		_, has := macths[arg]
		return has == b
	}, nil
}

func integerCompare(op byte, a, b int) bool {
	switch op {
	case '>':
		return a > b
	case '<':
		return a < b
	case '!':
		return a != b
	default:
		return a == b
	}
}

// fbStringLen 函数生一个验证字符串长度'>','<','='指定长度的验证函数。
func fbStringLen(s string) (func(s string) bool, error) {
	length, err := strconv.ParseInt(trimFuncOperate(s), 10, 32)
	if err != nil {
		return nil, err
	}

	f, l := s[0], int(length)
	return func(s string) bool {
		return integerCompare(f, len(s), l)
	}, nil
}

// fbAnyLen 函数生一个验证字符串长度'>','<','='指定长度的验证函数。
func fbAnyLen(s string) (func(i any) bool, error) {
	length, err := strconv.ParseInt(trimFuncOperate(s), 10, 32)
	if err != nil {
		return nil, err
	}
	f, l := s[0], int(length)
	return func(i any) bool {
		v := reflect.Indirect(reflect.ValueOf(i))
		switch v.Kind() {
		case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
			return integerCompare(f, v.Len(), l)
		default:
			return false
		}
	}, nil
}

// fbStringNum 函数验证一个字符串是否为float64。
func fbStringNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func fbStringInteger(s string) bool {
	for _, b := range s {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

func fbStringDomain(s string) bool {
	pos := strings.LastIndexByte(s, '.')
	first := strings.IndexByte(s, '.')
	return first > 0 && pos > 0 && pos != len(s)-1
}

func fbStringMail(s string) bool {
	pos := strings.IndexByte(s, '@')
	if pos > 0 && pos != len(s)-1 {
		return fbStringDomain(s[pos+1:])
	}
	return false
}

func fbStringPhone(s string) bool {
	s = strings.Replace(s, " ", "", 5)
	// 中国大陆移动电话 运营商/归属地/客户号码 1xx/xxxx/xxxx
	if len(s) == 11 && s[0] == '1' && fbStringInteger(s) {
		return true
	}
	// 国际电话号码格式 国际冠码/国际电话区号/电话号码 国际冠码00或+
	if len(s) > 9 && len(s) < 21 && (s[0] == '+' || strings.HasPrefix(s, "00")) &&
		fbStringInteger(strings.ReplaceAll(s[1:], "-", "")) {
		return true
	}
	// 中国大陆固定电话 长途冠码/省市区号/电话号码 长途冠码0 省市区号2位或3位 电话号码7位或8位
	if len(s) > 9 && len(s) < 13 && s[0] == '0' {
		pos := strings.IndexByte(s, '-')
		if pos == 2 || pos == 3 {
			return fbStringInteger(s[:pos]) && fbStringInteger(s[pos+1:])
		}
	}
	return false
}

// fbStringpPatten 模式匹配对象，允许使用带'*'的模式。
func fbStringpPatten(str string) (func(string) bool, error) {
	b := !strings.HasPrefix(str, "!=")
	patten := trimFuncOperate(str)
	return func(obj string) bool {
		parts := strings.Split(patten, "*")
		if len(parts) < 2 {
			return patten == obj == b
		}
		if !strings.HasPrefix(obj, parts[0]) {
			return !b
		}
		for _, i := range parts {
			if i == "" {
				continue
			}
			pos := strings.Index(obj, i)
			if pos == -1 {
				return !b
			}
			obj = obj[pos+len(i):]
		}
		return b
	}, nil
}

// fbStringRegexp 函数生成一个正则检测字符串的验证函数。
func fbStringRegexp(s string) (func(arg string) bool, error) {
	b := !strings.HasPrefix(s, "!=")
	re, err := regexp.Compile(trimFuncOperate(s))
	if err != nil {
		return nil, err
	}
	// 返回正则匹配校验函数
	return func(arg string) bool {
		return re.MatchString(arg) == b
	}, nil
}

func fbStringFuncBool(fn func(string, string) bool) func(string) (func(string) bool, error) {
	return func(s string) (func(string) bool, error) {
		b := !strings.HasPrefix(s, "!=")
		s = trimFuncOperate(s)
		return func(str string) bool {
			return fn(str, s) == b
		}, nil
	}
}

func fbStringFuncInt(fn func(string, string) int) func(string) (func(string) bool, error) {
	return func(s string) (func(string) bool, error) {
		num, str, ok := strings.Cut(trimFuncOperate(s), ",")
		if !ok || num == "" || str == "" {
			return nil, fmt.Errorf("funcCreator setstring format must 'name=num,string', current: %s", s)
		}

		n, err := GetAnyByStringWithError[int](num)
		if err != nil {
			return nil, err
		}

		f := s[0]
		return func(s string) bool {
			return integerCompare(f, fn(s, str), n)
		}, nil
	}
}

func fbTimeAfter(str string) (func(any) bool, error) {
	t, err := GetAnyByStringWithError[time.Time](trimFuncOperate(str))
	if err != nil {
		return nil, err
	}
	return func(i any) bool {
		t2, ok := i.(time.Time)
		if ok {
			return t2.After(t)
		}
		return false
	}, nil
}

func fbTimeBefore(str string) (func(any) bool, error) {
	t, err := GetAnyByStringWithError[time.Time](trimFuncOperate(str))
	if err != nil {
		return nil, err
	}
	return func(i any) bool {
		t2, ok := i.(time.Time)
		if ok {
			return t2.Before(t)
		}
		return false
	}, nil
}

func fsDefault[T string | int | uint | float64 | bool](T) T {
	var t T
	return t
}

func fsAnyDefault(i any) any {
	_, ok := i.(time.Time)
	if ok {
		return time.Time{}
	}
	return nil
}

func fsValue[T string | int | uint | float64 | bool](s string) (func(T) T, error) {
	val, err := GetAnyByStringWithError[T](trimFuncOperate(s))
	if err != nil {
		return nil, err
	}
	return func(arg T) T {
		return val
	}, nil
}

func fsTimeValue(s string) (func(i any) any, error) {
	t, err := GetAnyByStringWithError[time.Time](trimFuncOperate(s))
	if err != nil {
		return nil, err
	}
	return func(i any) any {
		return t
	}, nil
}

func fsAdd[T string | int | uint | float64](s string) (func(T) T, error) {
	val, err := GetAnyByStringWithError[T](trimFuncOperate(s))
	if err != nil {
		return nil, err
	}
	return func(arg T) T {
		return arg + val
	}, nil
}

func fsTimeAdd(s string) (func(i any) any, error) {
	d, err := time.ParseDuration(trimFuncOperate(s))
	if err != nil {
		return nil, err
	}
	return func(i any) any {
		t, ok := i.(time.Time)
		if ok {
			return t.Add(d)
		}
		return i
	}, nil
}

func fsStringNow(str string) (func(string) string, error) {
	f := trimFuncOperate(str)
	if f == time.Now().Format(f) {
		return fsValue[string](str)
	}
	return func(string) string {
		return time.Now().Format(f)
	}, nil
}

func fsTimeNow(any) any {
	return time.Now()
}

func fsStringLen(str string) string {
	return strconv.Itoa(len(str))
}

func fsStringMd5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func fsStringReplace(s string) (func(string) string, error) {
	o, n, _ := strings.Cut(trimFuncOperate(s), ",")
	num, err := GetAnyByStringWithError(o, -1)
	if err == nil {
		o, n, _ = strings.Cut(n, ",")
	}
	return func(str string) string {
		return strings.Replace(str, o, n, num)
	}, nil
}

func fsStringFuncString(fn func(string, string) string) func(string) (func(string) string, error) {
	return func(s string) (func(string) string, error) {
		s = trimFuncOperate(s)
		return func(str string) string {
			return fn(str, s)
		}, nil
	}
}

func fsStringHide(string) string {
	return "***"
}

func nameHasChinese(str string) bool {
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			return true
		}
	}
	return false
}

func fsStringHideSurname(s string) string {
	if len(s) < 3 {
		return "****"
	}
	if nameHasChinese(s) {
		return "*" + string([]rune(s)[1:])
	}
	_, n, ok := strings.Cut(s, " ")
	if ok {
		return "**** " + n
	}
	return "****" + s[len(s)-2:]
}

func fsStringHideName(s string) string {
	if len(s) < 3 {
		return "****"
	}
	if nameHasChinese(s) {
		return string([]rune(s)[0]) + "**"
	}
	sur, _, ok := strings.Cut(s, " ")
	if ok {
		return sur + " ****"
	}
	return sur[0:2] + "****"
}

func fsStringHideMail(s string) string {
	name, domain, ok := strings.Cut(s, "@")
	if ok {
		if len(name) > 8 {
			return name[:3] + "****" + "@" + domain
		}
		if len(name) > 4 {
			return name[:2] + "****" + "@" + domain
		}
		return "****@" + domain
	}
	return s
}

func fsStringHidePhone(phone string) string {
	s := strings.Replace(phone, " ", "", 5)
	// China 中国大陆移动电话
	if len(s) == 11 && s[0] == '1' && fbStringInteger(s) {
		return phone[:len(phone)-8] + "****" + phone[len(phone)-4:]
	}
	// 国际电话号码格式 国际冠码/国际电话区号/电话号码 国际冠码00或+
	if len(s) > 9 && len(s) < 21 && (s[0] == '+' || strings.HasPrefix(s, "00")) &&
		fbStringInteger(strings.ReplaceAll(s[1:], "-", "")) {
		return phone[:len(phone)-8] + "****" + phone[len(phone)-4:]
	}
	// 中国大陆固定电话 长途冠码/省市区号/电话号码 长途冠码0 省市区号2位或3位 电话号码7位或8位
	if len(s) > 9 && len(s) < 13 && s[0] == '0' {
		pos := strings.IndexByte(s, '-')
		if pos == 2 || pos == 3 {
			return phone[:len(phone)-4] + "****"
		}
	}
	return phone
}
