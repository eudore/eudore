package eudore

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ClientOption defines add options when creating client request.
type ClientOption struct {
	Context       context.Context
	Body          io.Reader
	Host          string
	URL           *url.URL
	Values        url.Values
	Headers       []string
	Hooks         []ClientHook
	RequestHooks  []func(*http.Request)
	ResponseHooks []func(*http.Response) error
	// Trace saves ClientTrace data, and enables httptrace when it is not empty.
	Trace *ClientTrace
}

// ClientTrace defines the data for client request tracking reords.
type ClientTrace struct {
	sync.Mutex            `json:"-" yaml:"-"`
	HTTPStart             time.Time            `json:"httpStart" yaml:"httpStart"`
	HTTPDone              time.Time            `json:"httpDone" yaml:"httpDone"`
	HTTPDuration          time.Duration        `json:"httpDuration" yaml:"httpDuration"`
	DNSStart              time.Time            `json:"dnsStart,omitempty" yaml:"dnsStart,omitempty"`
	DNSDone               time.Time            `json:"dnsDone,omitempty" yaml:"dnsDone,omitempty"`
	DNSDuration           time.Duration        `json:"dnsDuration,omitempty" yaml:"dnsDuration,omitempty"`
	DNSHost               string               `json:"dnsHost,omitempty" yaml:"dnsHost,omitempty"`
	DNSAddrs              []net.IPAddr         `json:"dnsAddrs,omitempty" yaml:"dnsAddrs,omitempty"`
	DNSError              error                `json:"dnsError,omitempty" yaml:"dnsError,omitempty"`
	Connect               []ClientTraceConnect `json:"onnect" yaml:"onnect"`
	GetConn               time.Time            `json:"getConn" yaml:"getConn"`
	GetConnHostPort       string               `json:"getConnHostPort" yaml:"getConnHostPort"`
	GotConn               time.Time            `json:"gotConn" yaml:"gotConn"`
	GotConnLocalAddr      net.Addr             `json:"gotConnLocalAddr" yaml:"gotConnLocalAddr"`
	GotConnRemoteAddr     net.Addr             `json:"gotConnRemoteAddr" yaml:"gotConnRemoteAddr"`
	GotFirstResponseByte  time.Time            `json:"gotFirstResponseByte" yaml:"gotFirstResponseByte"`
	TLSHandshakeStart     time.Time            `json:"tlsHandshakeStart,omitempty" yaml:"tlsHandshakeStart,omitempty"`
	TLSHandshakeDone      time.Time            `json:"tlsHandshakeDone,omitempty" yaml:"tlsHandshakeDone,omitempty"`
	TLSHandshakeDuration  time.Duration        `json:"tlsHandshakeDuration,omitempty" yaml:"tlsHandshakeDuration,omitempty"`
	TLSHandshakeError     error                `json:"tlsHandshakeError,omitempty" yaml:"tlsHandshakeError,omitempty"`
	TLSHandshakeIssuer    string               `json:"tlsHandshakeIssuer,omitempty" yaml:"tlsHandshakeIssuer,omitempty"`
	TLSHandshakeSubject   string               `json:"tlsHandshakeSubject,omitempty" yaml:"tlsHandshakeSubject,omitempty"`
	TLSHandshakeNotBefore time.Time            `json:"tlsHandshakeNotBefore,omitempty" yaml:"tlsHandshakeNotBefore,omitempty"`
	TLSHandshakeNotAfter  time.Time            `json:"tlsHandshakeNotAfter,omitempty" yaml:"tlsHandshakeNotAfter,omitempty"`
	TLSHandshakeDigest    string               `json:"tlsHandshakeDigest,omitempty" yaml:"tlsHandshakeDigest,omitempty"`
	WroteHeaders          http.Header          `json:"wroteHeaders,omitempty" yaml:"wroteHeaders,omitempty"`
}

// ClientTraceConnect defines Trace onnection information.
// One request may have multiple onnections.
type ClientTraceConnect struct {
	Network  string        `json:"network" yaml:"network"`
	Address  string        `json:"address" yaml:"address"`
	Start    time.Time     `json:"start" yaml:"start"`
	Done     time.Time     `json:"done,omitempty" yaml:"done,omitempty"`
	Duration time.Duration `json:"duration,omitempty" yaml:"duration,omitempty"`
	Error    error         `json:"error,omitempty" yaml:"error,omitempty"`
}

// The NewClientOption function creates [ClientOption] using any options.
//
// If you add an unsupported type, it will panic [ErrClientOptionInvalidType].
func NewClientOption(options []any) *ClientOption {
	o := &ClientOption{}
	return o.appendOptions(options)
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

func (o *ClientOption) clone() *ClientOption {
	n := &ClientOption{}
	n.append(o)
	return n
}

//nolint:cyclop,gocyclo
func (o *ClientOption) appendOptions(options []any) *ClientOption {
	for i := range options {
		switch v := options[i].(type) {
		case context.Context:
			o.Context = v
		case io.Reader:
			o.Body = v
		case url.Values:
			o.Values = appendValues(o.Values, v)
		case http.Header:
			for k, vals := range v {
				for _, v := range vals {
					o.Headers = append(o.Headers, k, v)
				}
			}
		case *Cookie:
			o.appendCookie(v.String())
		case *CookieSet:
			o.appendCookie(Cookie{Name: v.Name, Value: v.Value}.String())
		case ClientHook:
			o.Hooks = append(o.Hooks, v)
		case func(*http.Request):
			o.RequestHooks = sliceClearAppend(o.RequestHooks, v)
		case func(*http.Response) error:
			o.ResponseHooks = sliceClearAppend(o.ResponseHooks, v) //nolint:bodyclose
		case *ClientTrace:
			o.Trace = v
		case *ClientOption:
			o.append(v)
		case []any:
			o.appendOptions(v)
		default:
			panic(fmt.Errorf(ErrClientOptionInvalidType, v))
		}
	}

	return o
}

func (o *ClientOption) appendCookie(val string) {
	for i := 0; i < len(o.Headers); i += 2 {
		if o.Headers[i] == HeaderCookie {
			o.Headers[i+1] = o.Headers[i+1] + ";" + val
			return
		}
	}
	o.Headers = append(o.Headers, HeaderCookie, val)
}

func (o *ClientOption) append(a *ClientOption) {
	if a == nil {
		return
	}
	o.Context = GetAnyDefault(a.Context, o.Context)
	o.Host = GetAnyDefault(a.Host, o.Host)
	o.Values = appendValues(o.Values, a.Values)
	o.Headers = sliceClearAppend(o.Headers, a.Headers...)
	o.Hooks = sliceClearAppend(o.Hooks, a.Hooks...)
	o.RequestHooks = sliceClearAppend(o.RequestHooks, a.RequestHooks...)
	o.ResponseHooks = sliceClearAppend(o.ResponseHooks, a.ResponseHooks...) //nolint:bodyclose
	if o.URL != nil && a.URL != nil {
		o.URL = o.URL.ResolveReference(a.URL)
	} else {
		o.URL = GetAnyDefault(a.URL, o.URL)
	}
	if o.Trace == nil && a.Trace != nil {
		o.Trace = &ClientTrace{}
	}
}

func (o *ClientOption) parsePath(path string) string {
	u, err := url.Parse(path)
	if err != nil {
		return path
	}
	if o.URL != nil {
		u = o.URL.ResolveReference(u)
	}

	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" {
		_, ok := o.Context.Value(ContextKeyServer).(interface {
			ServeConn(onn net.Conn)
		})
		if ok {
			u.Host = DefaultClientInternalHost
		} else {
			u.Host = DefaultClientHost
		}
	}

	if o.Values != nil {
		if u.RawQuery == "" {
			u.RawQuery = o.Values.Encode()
		} else {
			u.RawQuery = o.Values.Encode() + "&" + u.RawQuery
		}
	}
	return u.String()
}

func (o *ClientOption) apply(req *http.Request) {
	if o.Host != "" {
		req.Host = o.Host
	}
	// if nil make in send
	req.Header = make(http.Header, len(o.Headers)/2)
	for i := 0; i < len(o.Headers); i += 2 {
		req.Header.Set(o.Headers[i], o.Headers[i+1])
	}

	body, ok := req.Body.(ClientBody)
	if ok {
		req.Header.Set(HeaderContentType, body.GetContentType())
		req.GetBody = body.GetBody
	}
	for _, hook := range o.RequestHooks {
		hook(req)
	}
}

func (o *ClientOption) release(resp *http.Response) error {
	trace := o.Trace
	if trace != nil {
		trace.Lock()
		if trace.HTTPDuration == 0 {
			trace.HTTPDone = time.Now()
			trace.HTTPDuration = trace.HTTPDone.Sub(trace.HTTPStart)
		}
		trace.Unlock()
	}
	for _, hook := range o.ResponseHooks {
		err := hook(resp)
		if err != nil {
			if DefaultClientOptionLoggerError {
				req := resp.Request
				log := NewLoggerWithContext(req.Context())
				log = log.WithFields(clientLoggerKeys2[:5], []any{
					req.Host, req.Method, req.URL.Path,
					resp.Proto, resp.StatusCode,
				})
				log = loggerValues(log, "x-response-id",
					resp.Header.Get(HeaderXRequestID),
					resp.Header.Get(HeaderXTraceID),
				)
				log.Error(err)
			}
			return err
		}
	}

	return nil
}

// The NewClientOptionHost function creates [ClientOption] setting request Host.
func NewClientOptionHost(host string) *ClientOption {
	return &ClientOption{Host: strings.TrimSuffix(host, ":")}
}

// NewClientOptionURL function creates [ClientOption] request base url.
func NewClientOptionURL(host string) *ClientOption {
	if host == "" {
		return nil
	}
	u, err := url.Parse(host)
	if err != nil {
		return nil
	}
	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	return &ClientOption{URL: u}
}

// The NewClientQuery function creates [ClientOption] set the request uri param.
func NewClientQuery(key, val string) url.Values {
	if val == "" {
		return nil
	}
	return url.Values{key: {val}}
}

// The NewClientHeader function creates [ClientOption] set the request Header.
func NewClientHeader(key, val string) http.Header {
	if val == "" {
		return nil
	}
	return http.Header{key: []string{val}}
}

// NewClientOptionBasicauth function creates [ClientOption] set BasicAuth users.
func NewClientOptionBasicauth(username, password string) http.Header {
	auth := username + ":" + password
	return NewClientHeader(
		HeaderAuthorization,
		"Basic "+base64.StdEncoding.EncodeToString([]byte(auth)),
	)
}

// The NewClientOptionBearer function creates [ClientOption] set
// request Bearer [HeaderAuthorization].
func NewClientOptionBearer(bearer string) http.Header {
	return NewClientHeader(HeaderAuthorization, "Bearer "+bearer)
}

// The NewClientOptionUserAgent function creates [ClientOption] set
// the request [HeaderUserAgent].
func NewClientOptionUserAgent(ua string) http.Header {
	return NewClientHeader(HeaderUserAgent, ua)
}

// The NewClientTraceWithContext function initializes and binds [ClientTrace] to
// [ontext.Context].
//
//nolint:funlen
func NewClientTraceWithContext(ctx context.Context, trace *ClientTrace,
) context.Context {
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
			trace.DNSError = info.Err
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
				if trace.Connect[i].Network == network &&
					trace.Connect[i].Address == addr {
					trace.Connect[i].Done = time.Now()
					trace.Connect[i].Duration = trace.Connect[i].Done.Sub(
						trace.Connect[i].Start,
					)
					trace.Connect[i].Error = err
					return
				}
			}
		},
		GetConn: func(hostPort string) {
			trace.GetConn = time.Now()
			trace.GetConnHostPort = hostPort
		},
		GotConn: func(info httptrace.GotConnInfo) {
			trace.GotConn = time.Now()
			trace.GotConnLocalAddr = info.Conn.LocalAddr()
			trace.GotConnRemoteAddr = info.Conn.RemoteAddr()
		},
		GotFirstResponseByte: func() { trace.GotFirstResponseByte = time.Now() },
		TLSHandshakeStart:    func() { trace.TLSHandshakeStart = time.Now() },
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			trace.Lock()
			defer trace.Unlock()
			trace.TLSHandshakeDone = time.Now()
			trace.TLSHandshakeDuration = trace.TLSHandshakeDone.Sub(
				trace.TLSHandshakeStart,
			)
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
