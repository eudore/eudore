package eudore

import (
	"strings"
	"net/http"
)

type (
	Header = http.Header
	Cookie = http.Cookie
	CookieRead struct {
		Name  string
		Value string
	}
	// From net/http.Cookie
/*	CookieWrite struct {
		Name  string
		Value string

		Path       string    // optional
		Domain     string    // optional
		Expires    time.Time // optional
		RawExpires string    // for reading cookies only

		// MaxAge=0 means no 'Max-Age' attribute specified.
		// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
		// MaxAge>0 means Max-Age attribute present and given in seconds
		MaxAge   int
		Secure   bool
		HttpOnly bool
		SameSite SameSite // Go 1.11
		Raw      string
		Unparsed []string // Raw text of unparsed attribute-value pairs
	}*/
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
