package eudore

import (
	"strings"
	"net/http"
	"net/textproto"
)

type (
	Params = textproto.MIMEHeader
	Header = textproto.MIMEHeader
	// From net/http.Cookie
	CookieWrite = http.Cookie
	CookieRead struct {
		Name  string
		Value string
	}
	// source net/http
	//
	// 来源net/http
	PushOptions = http.PushOptions
)

func ReadCookies(lines []string) []*CookieRead {
	if len(lines) == 0 {
		return []*CookieRead{}
	}
	cookies := []*CookieRead{}
	for _, line := range lines {
		parts := strings.Split(line, "; ")
		if len(parts) == 0 {
			continue
		}
		// Per-line attributes
		for i := 0; i < len(parts); i++ {
			if len(parts[i]) == 0 {
				continue
			}
			name, val := parts[i], ""
			if j := strings.Index(name, "="); j >= 0 {
				name, val = name[:j], name[j+1:]
			}
			if !isCookieNameValid(name) {
				continue
			}
			val, ok := parseCookieValue(val, true)
			if !ok {
				continue
			}
			cookies = append(cookies, &CookieRead{Name: name, Value: val})
		}
	}
	return cookies
}
