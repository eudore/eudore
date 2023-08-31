package eudore

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// ClientOption 定义创建客户端请求时额外选项。
type ClientOption struct {
	Context       context.Context
	Timeout       time.Duration
	Body          io.Reader
	ClientBody    ClientBody
	Values        url.Values
	Header        http.Header
	Headers       []string
	Cookies       []string
	RequestHooks  []func(*http.Request)
	ResponseHooks []func(*http.Response) error
	Retrys        []ClientRetry
	// Trace saves ClientTrace data, and enables httptrace when it is not empty.
	// Trace 保存ClientTrace数据，非空时启用httptrace。
	Trace *ClientTrace
}

// ClientRetry 定义客户端请求重试行为。
type ClientRetry struct {
	Max       int
	Condition func(int, *http.Response, error) bool
}

// ClientTrace 定义http客户端请求追踪记录的数据。
type ClientTrace struct {
	sync.Mutex            `alias:"mutex" json:"-" xml:"-" yaml:"-"`
	HTTPStart             time.Time            `alias:"http-start" json:"http-start" xml:"http-start" yaml:"http-start"`
	HTTPDone              time.Time            `alias:"http-done" json:"http-done" xml:"http-done" yaml:"http-done"`
	HTTPDuration          time.Duration        `alias:"http-duration" json:"http-duration" xml:"http-duration" yaml:"http-duration"`
	DNSStart              time.Time            `alias:"dns-start,omitempty" json:"dns-start,omitempty" xml:"dns-start,omitempty" yaml:"dns-start,omitempty"`
	DNSDone               time.Time            `alias:"dns-done,omitempty" json:"dns-done,omitempty" xml:"dns-done,omitempty" yaml:"dns-done,omitempty"`
	DNSDuration           time.Duration        `alias:"dns-duration,omitempty" json:"dns-duration,omitempty" xml:"dns-duration,omitempty" yaml:"dns-duration,omitempty"`
	DNSHost               string               `alias:"dns-host,omitempty" json:"dns-host,omitempty" xml:"dns-host,omitempty" yaml:"dns-host,omitempty"`
	DNSAddrs              []net.IPAddr         `alias:"dns-addrs,omitempty" json:"dns-addrs,omitempty" xml:"dns-addrs,omitempty" yaml:"dns-addrs,omitempty"`
	Connect               []ClientTraceConnect `alias:"connect" json:"connect" xml:"connect" yaml:"connect"`
	GetConn               time.Time            `alias:"get-conn" json:"get-conn" xml:"get-conn" yaml:"get-conn"`
	GetConnHostPort       string               `alias:"get-conn-host-port" json:"get-conn-host-port" xml:"get-conn-host-port" yaml:"get-conn-host-port"`
	GotConn               time.Time            `alias:"got-conn" json:"got-conn" xml:"got-conn" yaml:"got-conn"`
	GotFirstResponseByte  time.Time            `alias:"got-first-response-byte" json:"got-first-response-byte" xml:"got-first-response-byte" yaml:"got-first-response-byte"`
	TLSHandshakeStart     time.Time            `alias:"tls-handshake-start,omitempty" json:"tls-handshake-start,omitempty" xml:"tls-handshake-start,omitempty" yaml:"tls-handshake-start,omitempty"`
	TLSHandshakeDone      time.Time            `alias:"tls-handshake-done,omitempty" json:"tls-handshake-done,omitempty" xml:"tls-handshake-done,omitempty" yaml:"tls-handshake-done,omitempty"`
	TLSHandshakeDuration  time.Duration        `alias:"tls-handshake-duration,omitempty" json:"tls-handshake-duration,omitempty" xml:"tls-handshake-duration,omitempty" yaml:"tls-handshake-duration,omitempty"`
	TLSHandshakeError     error                `alias:"tls-handshake-error,omitempty" json:"tls-handshake-error,omitempty" xml:"tls-handshake-error,omitempty" yaml:"tls-handshake-error,omitempty"`
	TLSHandshakeIssuer    string               `alias:"tls-handshake-issuer,omitempty" json:"tls-handshake-issuer,omitempty" xml:"tls-handshake-issuer,omitempty" yaml:"tls-handshake-issuer,omitempty"`
	TLSHandshakeSubject   string               `alias:"tls-handshake-subject,omitempty" json:"tls-handshake-subject,omitempty" xml:"tls-handshake-subject,omitempty" yaml:"tls-handshake-subject,omitempty"`
	TLSHandshakeNotBefore time.Time            `alias:"tls-handshake-not-before,omitempty" json:"tls-handshake-not-before,omitempty" xml:"tls-handshake-not-before,omitempty" yaml:"tls-handshake-not-before,omitempty"`
	TLSHandshakeNotAfter  time.Time            `alias:"tls-handshake-not-after,omitempty" json:"tls-handshake-not-after,omitempty" xml:"tls-handshake-not-after,omitempty" yaml:"tls-handshake-not-after,omitempty"`
	TLSHandshakeDigest    string               `alias:"tls-handshake-digest,omitempty" json:"tls-handshake-digest,omitempty" xml:"tls-handshake-digest,omitempty" yaml:"tls-handshake-digest,omitempty"`
	WroteHeaders          http.Header          `alias:"wrote-headers,omitempty" json:"wrote-headers,omitempty" xml:"wrote-headers,omitempty" yaml:"wrote-headers,omitempty"`
}

// ClientTraceConnect 定义Trace连接信息，一个请求可能出现多连接。
type ClientTraceConnect struct {
	Network  string        `alias:"network" json:"network" xml:"network" yaml:"network"`
	Address  string        `alias:"address" json:"address" xml:"address" yaml:"address"`
	Start    time.Time     `alias:"start" json:"start" xml:"start" yaml:"start"`
	Done     time.Time     `alias:"done,omitempty" json:"done,omitempty" xml:"done,omitempty" yaml:"done,omitempty"`
	Duration time.Duration `alias:"duration,omitempty" json:"duration,omitempty" xml:"duration,omitempty" yaml:"duration,omitempty"`
	Error    error         `alias:"error,omitempty" json:"error,omitempty" xml:"error,omitempty" yaml:"error,omitempty"`
}

// NewClientOption 函数使用options创建ClientOption。
func NewClientOption(ctx context.Context, options []any) *ClientOption {
	co := &ClientOption{}
	return co.appendOptions(ctx, options)
}

func appendValues(dst, src map[string][]string) map[string][]string {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = make(map[string][]string)
	}
	for key, vals := range src {
		dst[key] = append(dst[key], vals...)
	}
	return dst
}

func (co *ClientOption) clone() *ClientOption {
	o := &ClientOption{}
	*o = *co
	if o.Values != nil {
		o.Values = appendValues(make(url.Values, len(o.Values)), o.Values)
	}
	if o.Header != nil {
		o.Header = o.Header.Clone()
	}
	return o
}

//nolint:cyclop,gocyclo
func (co *ClientOption) appendOptions(ctx context.Context, options []any) *ClientOption {
	if ctx != nil {
		co.Context = ctx
	}
	for i := range options {
		switch o := options[i].(type) {
		case context.Context:
			co.Context = o
		case ClientBody:
			co.ClientBody = o
			co.Body = o
		case io.Reader:
			co.Body = o
		case url.Values:
			co.Values = appendValues(co.Values, o)
		case http.Header:
			co.Header = appendValues(co.Header, o)
		case Cookie:
			co.Cookies = clearCap(append(co.Cookies, o.String()))
		case *http.Cookie:
			co.Cookies = clearCap(append(co.Cookies, Cookie{Name: o.Name, Value: o.Value}.String()))
		case time.Duration:
			co.Timeout = o
		case func(*http.Request):
			co.RequestHooks = clearCap(append(co.RequestHooks, o))
		case func(*http.Response) error:
			co.ResponseHooks = clearCap(append(co.ResponseHooks, o)) //nolint:bodyclose
		case ClientRetry:
			co.Retrys = clearCap(append(co.Retrys, o))
		case *ClientTrace:
			co.Trace = o
		case *ClientOption:
			co.append(o)
		}
	}

	return co
}

func (co *ClientOption) append(o *ClientOption) {
	if o == nil {
		return
	}
	co.Context = GetAnyDefault(o.Context, co.Context)
	co.Timeout = GetAnyDefault(o.Timeout, co.Timeout)
	co.Body = GetAnyDefault(o.Body, co.Body)
	co.ClientBody = GetAnyDefault(o.ClientBody, co.ClientBody)

	co.Values = appendValues(co.Values, o.Values)
	co.Header = appendValues(co.Header, o.Header)
	co.Headers = clearCap(append(co.Headers, o.Headers...))
	co.Cookies = clearCap(append(co.Cookies, o.Cookies...))
	co.RequestHooks = clearCap(append(co.RequestHooks, o.RequestHooks...))
	co.ResponseHooks = clearCap(append(co.ResponseHooks, o.ResponseHooks...)) //nolint:bodyclose
	co.Retrys = clearCap(append(co.Retrys, o.Retrys...))
	if co.Trace == nil && o.Trace != nil {
		co.Trace = &ClientTrace{}
	}
}

func (co *ClientOption) apply(req *http.Request) {
	if co.Values != nil {
		v, err := url.ParseQuery(req.URL.RawQuery)
		if err == nil {
			v = appendValues(v, co.Values)
			req.URL.RawQuery = v.Encode()
		}
	}
	if co.Header != nil {
		req.Header = appendValues(req.Header, co.Header)
	}
	if co.Headers != nil {
		for i := 0; i < len(co.Headers); i += 2 {
			req.Header.Set(co.Headers[i], co.Headers[i+1])
		}
	}
	if co.Cookies != nil {
		s := strings.Join(co.Cookies, "; ")
		if c := req.Header.Get(HeaderCookie); c != "" {
			req.Header.Set(HeaderCookie, c+"; "+s)
		} else {
			req.Header.Set(HeaderCookie, s)
		}
	}

	if co.ClientBody != nil {
		req.ContentLength = -1
		req.Header.Set(HeaderContentType, co.ClientBody.GetContentType())
		req.GetBody = co.ClientBody.GetBody
	}
	for _, hook := range co.RequestHooks {
		hook(req)
	}
}

var clientLoggerRequestIDKeys = [...]string{
	HeaderXRequestID, strings.ToLower(HeaderXRequestID),
	HeaderXTraceID, strings.ToLower(HeaderXTraceID),
}

func (co *ClientOption) release(req *http.Request, resp *http.Response, err error) error {
	trace := co.Trace
	if trace != nil {
		trace.HTTPDone = time.Now()
		trace.HTTPDuration = trace.HTTPDone.Sub(trace.HTTPStart)
	}
	for _, hook := range co.ResponseHooks {
		if err != nil {
			break
		}
		err = hook(resp)
	}

	level := LoggerDebug
	if err != nil {
		level = LoggerError
	}
	if level >= DefaultClinetLoggerLevel {
		log := NewLoggerWithContext(co.Context)
		if log != DefaultLoggerNull && level >= log.GetLevel() {
			keys := []string{"method", "scheme", "host", "path", "query", "request-header"}
			vals := []any{req.Method, req.URL.Scheme, req.Host, req.URL.Path, req.URL.RawQuery, req.Header}
			if resp != nil {
				keys = append(keys, "proto", "status", "status-code", "response-header")
				vals = append(vals, resp.Proto, resp.StatusCode, resp.Status, resp.Header)
				// append id
				for i := 0; i < 4; i += 2 {
					val := resp.Header.Get(clientLoggerRequestIDKeys[i])
					if val != "" {
						log = log.WithField(clientLoggerRequestIDKeys[i+1], val)
					}
				}
			}
			if trace != nil {
				keys = append(keys, "trace")
				vals = append(vals, trace)
				// lock trace to loggerformat
				trace.Lock()
				defer trace.Unlock()
			}

			if err != nil {
				log.WithFields(keys, vals).Error(err.Error())
			} else {
				log.WithFields(keys, vals).Debug()
			}
		}
	}

	return err
}

// NewClientOptionBasicauth 函数设置请求Basic Auth权限。
func NewClientOptionBasicauth(username, password string) *ClientOption {
	auth := username + ":" + password
	return NewClientOptionHeader(HeaderAuthorization, "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
}

// NewClientOptionBearer 函数设置请求Bearer认证。
func NewClientOptionBearer(bearer string) *ClientOption {
	return NewClientOptionHeader(HeaderAuthorization, "Bearer "+bearer)
}

// NewClientOptionUserAgent 函数设置请求UA。
func NewClientOptionUserAgent(ua string) *ClientOption {
	return NewClientOptionHeader(HeaderUserAgent, ua)
}

// NewClientOptionHost 函数设置请求Host。
func NewClientOptionHost(host string) func(*http.Request) {
	return func(req *http.Request) { req.Host = host }
}

// NewClientOptionHeader 函数设置请求Header。
func NewClientOptionHeader(key, val string) *ClientOption {
	if val == "" {
		return nil
	}
	return &ClientOption{
		Headers: []string{key, val},
	}
}

// NewClientTraceWithContext 函数将ClientTrace初始化并绑定到context.Context。
//
//nolint:funlen
func NewClientTraceWithContext(ctx context.Context, trace *ClientTrace) context.Context {
	trace.HTTPStart = time.Now()
	ctx = context.WithValue(ctx, ContextKeyClientTrace, trace)
	ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			trace.DNSStart = time.Now()
			trace.DNSHost = info.Host
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			trace.DNSDone = time.Now()
			trace.DNSDuration = trace.DNSDone.Sub(trace.DNSStart)
			trace.DNSAddrs = info.Addrs
		},
		ConnectStart: func(network, addr string) {
			trace.Lock()
			defer trace.Unlock()
			trace.Connect = append(trace.Connect, ClientTraceConnect{
				Start:   time.Now(),
				Network: network,
				Address: addr,
			})
		},
		ConnectDone: func(network, addr string, err error) {
			trace.Lock()
			defer trace.Unlock()
			for i := range trace.Connect {
				if trace.Connect[i].Network == network && trace.Connect[i].Address == addr {
					trace.Connect[i].Done = time.Now()
					trace.Connect[i].Duration = trace.Connect[i].Done.Sub(trace.Connect[i].Start)
					trace.Connect[i].Error = err
					return
				}
			}
		},
		GetConn: func(hostPort string) {
			trace.GetConn = time.Now()
			trace.GetConnHostPort = hostPort
		},
		GotConn:              func(httptrace.GotConnInfo) { trace.GotConn = time.Now() },
		GotFirstResponseByte: func() { trace.GotFirstResponseByte = time.Now() },
		TLSHandshakeStart:    func() { trace.TLSHandshakeStart = time.Now() },
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			trace.Lock()
			defer trace.Unlock()
			trace.TLSHandshakeDone = time.Now()
			trace.TLSHandshakeDuration = trace.TLSHandshakeDone.Sub(trace.TLSHandshakeStart)
			trace.TLSHandshakeError = err

			if state.PeerCertificates != nil {
				cert := state.PeerCertificates[0]
				trace.TLSHandshakeIssuer = cert.Issuer.String()
				trace.TLSHandshakeSubject = cert.Subject.String()
				trace.TLSHandshakeNotBefore = cert.NotBefore
				trace.TLSHandshakeNotAfter = cert.NotAfter
				h := sha256.New()
				h.Write(cert.Raw)
				trace.TLSHandshakeDigest = hex.EncodeToString(h.Sum(nil))
			}
		},
		WroteHeaderField: func(key string, value []string) {
			if trace.WroteHeaders == nil {
				trace.WroteHeaders = make(http.Header)
			}
			trace.WroteHeaders[key] = value
		},
	})
	return ctx
}

// NewClientRetryNetwork 函数创建一个网络重试配置。
//
// 在err不为空或DefaultClinetRetryStatus指定状态码时，重试请求。
func NewClientRetryNetwork(max int) ClientRetry {
	return ClientRetry{
		Max: max,
		Condition: func(attempt int, resp *http.Response, _ error) bool {
			retry := resp == nil || DefaultClinetRetryStatus[resp.StatusCode]
			if retry {
				time.Sleep(time.Second * time.Duration((attempt + 1)))
			}
			return retry
		},
	}
}

// NewClientRetryDigest 函数创建一个摘要认证配置，在401时重新发起请求。
func NewClientRetryDigest(username, password string) ClientRetry {
	return ClientRetry{
		Max: 1,
		Condition: func(_ int, resp *http.Response, _ error) bool {
			if resp == nil || resp.StatusCode != StatusUnauthorized {
				return false
			}

			dig := newclientDigest(resp.Header.Get(HeaderWWWAuthenticate))
			if dig == nil || dig.invalid() {
				return false
			}

			req := resp.Request
			if dig.Qop == httpDigestQopAuthInt && req.Body != nil {
				// check GetBody in dotry
				dig.Body, _ = req.GetBody()
			}

			dig.Nc = "00000001"
			dig.Username = username
			dig.Password = password
			dig.Method = req.Method
			dig.URI = req.URL.Path
			req.Header.Set(HeaderAuthorization, dig.Encode())
			return true
		},
	}
}

var (
	digestKeys = [...]string{
		"username", "uri",
		"realm", "algorithm", "nonce", "qop",
		"nc", "cnonce", "response", "opaque",
	}
	httpDigestQopAuth    = "auth"
	httpDigestQopAuthInt = "auth-int"
)

type clientDigest struct {
	Hash     hash.Hash
	Body     io.ReadCloser
	Password string
	Method   string
	Username string
	URI      string

	Realm     string
	Algorithm string
	Nonce     string
	Qop       string

	Nc       string
	Cnonce   string
	Response string
	Opaque   string
}

func newclientDigest(req string) *clientDigest {
	if !strings.HasPrefix(req, "Digest ") {
		return nil
	}
	req = req[7:]

	dig := &clientDigest{}
	for _, s := range splitDigestString(req) {
		k, v, ok := strings.Cut(s, "=")
		if !ok {
			return nil
		}
		if len(v) > 2 && v[0] == '"' && v[len(v)-1] == '"' {
			v = v[1 : len(v)-1]
		}

		switch k {
		case "realm":
			dig.Realm = v
		case "algorithm":
			dig.Algorithm = strings.ToUpper(v)
		case "nonce":
			dig.Nonce = v
		case "qop":
			dig.Qop = strings.TrimSpace(strings.SplitN(v, ",", 2)[0])
		case "opaque":
			dig.Opaque = v
		default:
			return nil
		}
	}

	return dig
}

func splitDigestString(str string) []string {
	var pos int
	var char bool
	var strs []string
	for i, b := range str {
		switch b {
		case ',':
			if char {
				continue
			}
			strs = append(strs, strings.TrimSpace(str[pos:i]))
			pos = i + 1
		case '"':
			char = !char
		}
	}
	strs = append(strs, strings.TrimSpace(str[pos:]))
	return strs
}

func (dig *clientDigest) invalid() bool {
	switch dig.Algorithm {
	case "MD5", "MD5-SESS", "SHA-256", "SHA-256-SESS":
	default:
		return true
	}
	switch dig.Qop {
	case "", httpDigestQopAuth, httpDigestQopAuthInt:
	default:
		return true
	}
	return false
}

func (dig *clientDigest) Encode() string {
	dig.Cnonce = GetStringRandom(40)
	var ha1, ha2 string
	switch dig.Algorithm {
	case "MD5", "MD5-SESS":
		dig.Hash = md5.New()
		ha1 = dig.digestHash(fmt.Sprintf("%s:%s:%s", dig.Username, dig.Realm, dig.Password))
	case "SHA-256", "SHA-256-SESS":
		dig.Hash = sha256.New()
		ha1 = dig.digestHash(fmt.Sprintf("%s:%s:%s", dig.Username, dig.Realm, dig.Password))
	}
	if strings.HasSuffix(dig.Algorithm, "-SESS") {
		ha1 = dig.digestHash(fmt.Sprintf("%s:%s:%s", ha1, dig.Nonce, dig.Cnonce))
	}

	switch dig.Qop {
	case httpDigestQopAuth, "":
		ha2 = dig.digestHash(fmt.Sprintf("%s:%s", dig.Method, dig.URI))
	case httpDigestQopAuthInt:
		if dig.Body != nil {
			dig.Hash.Reset()
			io.Copy(dig.Hash, dig.Body)
			ha2 = hex.EncodeToString(dig.Hash.Sum(nil))
			dig.Body.Close()
		}
		ha2 = dig.digestHash(fmt.Sprintf("%s:%s:%s", dig.Method, dig.URI, ha2))
	}

	switch dig.Qop {
	case httpDigestQopAuth, httpDigestQopAuthInt:
		dig.Response = dig.digestHash(fmt.Sprintf("%s:%s:00000001:%s:%s:%s", ha1, dig.Nonce, dig.Cnonce, dig.Qop, ha2))
	case "":
		dig.Response = dig.digestHash(fmt.Sprintf("%s:%s:%s", ha1, dig.Nonce, ha2))
	}

	buf := bytes.NewBufferString("Digest ")
	data := *(*[14]string)(unsafe.Pointer(dig))
	for i, s := range data[4:] {
		if s != "" {
			switch i {
			case 3, 5, 6:
				fmt.Fprintf(buf, "%s=%s, ", digestKeys[i], s)
			default:
				fmt.Fprintf(buf, "%s=\"%s\", ", digestKeys[i], s)
			}
		}
	}
	buf.Truncate(buf.Len() - 2)
	return buf.String()
}

func (dig *clientDigest) digestHash(s string) string {
	dig.Hash.Reset()
	io.WriteString(dig.Hash, s)
	return hex.EncodeToString(dig.Hash.Sum(nil))
}
