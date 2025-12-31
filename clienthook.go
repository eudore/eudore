package eudore

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type clientHookCookie struct {
	next   http.RoundTripper
	cookie http.CookieJar
}

// NewClientHookCookie function creates [ClientHook] and uses [http.CookieJar]
// to manage the cookies sent.
func NewClientHookCookie(jar http.CookieJar) ClientHook {
	if jar == nil {
		jar, _ = cookiejar.New(nil)
	}
	return &clientHookCookie{cookie: jar}
}
func (*clientHookCookie) Name() string { return "cookie" }
func (hook *clientHookCookie) Wrap(rt http.RoundTripper) http.RoundTripper {
	return &clientHookCookie{
		next:   rt,
		cookie: hook.cookie,
	}
}

func (hook *clientHookCookie) RoundTrip(req *http.Request) (*http.Response, error) {
	cookies := hook.cookie.Cookies(req.URL)
	if cookies != nil {
		h := make(http.Header)
		for _, cooke := range cookies {
			h.Add(HeaderCookie, cooke.String())
		}
		appendValues(h, req.Header)
		req.Header, h = h, req.Header
		defer func() {
			req.Header, h = h, req.Header
		}()
	}

	resp, err := hook.next.RoundTrip(req)
	if resp != nil {
		cookies = resp.Cookies()
		if len(cookies) > 0 {
			hook.cookie.SetCookies(req.URL, cookies)
		}
	}
	return resp, err
}

type clientHookTimeout struct {
	next    http.RoundTripper
	timeout time.Duration
}

// NewClientHookTimeout function creates [ClientHook] to implement request
// timeout, using [context.WithTimeout].
func NewClientHookTimeout(timeout time.Duration) ClientHook {
	return &clientHookTimeout{timeout: timeout}
}
func (*clientHookTimeout) Name() string { return "timeout" }
func (hook *clientHookTimeout) String() string {
	return "timeout timeout=" + hook.timeout.String()
}

func (hook *clientHookTimeout) Wrap(rt http.RoundTripper) http.RoundTripper {
	return &clientHookTimeout{
		next:    rt,
		timeout: hook.timeout,
	}
}

func (hook *clientHookTimeout) RoundTrip(req *http.Request) (*http.Response, error) {
	if hook.timeout <= 0 {
		return hook.next.RoundTrip(req)
	}
	ctx, cancel := withTimeout(req.Context(), hook.timeout)
	resp, err := hook.next.RoundTrip(req.WithContext(ctx))
	if resp != nil {
		if resp.Body == nil {
			resp.Body = http.NoBody
			cancel()
		} else {
			deadline, _ := ctx.Deadline()
			resp.Body = &cancelTimerBody{
				rc:       resp.Body,
				stop:     cancel,
				deadline: deadline,
			}
		}
	}
	return resp, err
}

func withTimeout(ctx context.Context, timeout time.Duration,
) (context.Context, func()) {
	return context.WithTimeout(ctx, timeout)
}

type cancelTimerBody struct {
	rc       io.ReadCloser
	stop     func() // stops the time.Timer waiting to cancel the request
	deadline time.Time
}

func (b *cancelTimerBody) Read(p []byte) (int, error) {
	n, err := b.rc.Read(p)
	if err == nil {
		return n, nil
	}
	if errors.Is(err, io.EOF) {
		return n, err
	}
	if time.Now().After(b.deadline) {
		msg := " (Client.Timeout or context cancellation while reading body)"
		err = NewErrorWithWrapped(err, err.Error()+msg)
	}
	return n, err
}

func (b *cancelTimerBody) Close() error {
	err := b.rc.Close()
	b.stop()
	return err
}

type clientHookRedirect struct {
	next  http.RoundTripper
	check func(req *http.Request, via []*http.Request) error
}

// NewClientHookRedirect function creates [ClientHook] to implement response
// redirection processing.
//
// If the Body is not empty, the request must mean [http.Request.GetBody] to
// redirect.
func NewClientHookRedirect(fn func(req *http.Request, via []*http.Request) error) ClientHook {
	if fn == nil {
		fn = func(_ *http.Request, _ []*http.Request) error {
			return nil
		}
	}
	return &clientHookRedirect{check: fn}
}
func (*clientHookRedirect) Name() string { return "redirect" }
func (hook *clientHookRedirect) Wrap(rt http.RoundTripper) http.RoundTripper {
	return &clientHookRedirect{
		next:  rt,
		check: hook.check,
	}
}

func (hook *clientHookRedirect) RoundTrip(req *http.Request) (*http.Response, error) {
	var reqs []*http.Request
	for {
		reqs = append(reqs, req)
		resp, err := hook.next.RoundTrip(req)
		if err != nil {
			return resp, err
		}

		switch resp.StatusCode {
		case StatusMovedPermanently, StatusFound, StatusSeeOther,
			StatusPermanentRedirect, StatusTemporaryRedirect:
		default:
			return resp, nil
		}

		loc := resp.Header.Get(HeaderLocation)
		if loc == "" {
			return resp, nil
		}
		req, err = newRequest(req, loc, resp.StatusCode)
		if err != nil {
			_ = resp.Body.Close()
			return nil, &url.Error{
				Op:  strings.ToTitle(reqs[0].Method),
				URL: stripPassword(resp.Request.URL),
				Err: err,
			}
		}
		req.Response = resp

		err = hook.check(req, reqs)
		if errors.Is(err, http.ErrUseLastResponse) {
			return resp, nil
		} else if err != nil {
			return resp, &url.Error{
				Op:  strings.ToTitle(reqs[0].Method),
				URL: stripPassword(resp.Request.URL),
				Err: err,
			}
		}
		_ = resp.Body.Close()
	}
}

func newRequest(req *http.Request, loc string, code int) (*http.Request, error) {
	locu, err := url.Parse(loc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Location header %q: %w", loc, err)
	}
	u := req.URL.ResolveReference(locu)

	r := &http.Request{
		Method:  req.Method,
		URL:     u,
		Host:    u.Host,
		Header:  make(http.Header),
		GetBody: req.GetBody,
	}
	r = r.WithContext(req.Context())
	appendValues(r.Header, req.Header)

	switch code {
	case StatusMovedPermanently, StatusFound, StatusSeeOther:
		if r.Method != MethodGet && MethodGet != MethodHead {
			r.Method = MethodGet
		}
	case StatusPermanentRedirect, StatusTemporaryRedirect:
		r.Body = req.Body
		r.ContentLength = req.ContentLength
		err := resetBody(r)
		if err != nil {
			return nil, err
		}
	}
	// Host
	if req.Host != req.URL.Host && locu.Host == "" {
		r.Host = req.Host
	}
	// Add the Referer header from the most recent
	// request URL to the new one, if it's not https->http:
	if r.Header.Get(HeaderReferer) == "" &&
		(req.URL.Scheme == u.Scheme || u.Scheme == "https") {
		ref := req.URL.String()
		if req.URL.User != nil {
			auth := req.URL.User.String() + "@"
			ref = strings.Replace(ref, auth, "", 1)
		}
		req.Header.Add(HeaderReferer, ref)
	}
	return r, nil
}

func stripPassword(u *url.URL) string {
	_, passSet := u.User.Password()
	if passSet {
		return strings.Replace(u.String(), u.User.String()+"@", u.User.Username()+":***@", 1)
	}
	return u.String()
}

type clientHookRetry struct {
	next      http.RoundTripper
	num       int
	max       int
	intervals []time.Duration
	status    map[int]struct{}
}

// NewClientHookRetry function creates [ClientHook] to implement multiple
// request retries.
//
// If error is not empty or StatusCode is in status, resend the request.
//
// If the request has a body but does not support GetBody,
// it cannot be retried if it is not a network error.
func NewClientHookRetry(num int, intervals []time.Duration,
	status map[int]struct{},
) ClientHook {
	if intervals == nil {
		intervals = append([]time.Duration{}, DefaultClinetRetryInterval...)
	}
	if status == nil {
		status = mapClone(DefaultClinetRetryStatus)
	}

	return &clientHookRetry{
		num:       num,
		max:       len(intervals) / 2,
		intervals: intervals,
		status:    status,
	}
}

func (hook *clientHookRetry) Name() string { return "retry" }
func (hook *clientHookRetry) String() string {
	status := make([]string, 0, len(hook.status))
	for k := range hook.status {
		status = append(status, strconv.Itoa(k))
	}
	return fmt.Sprintf("retry num=%d status=[%s]",
		hook.num, strings.Join(status, ","),
	)
}

func (hook *clientHookRetry) Wrap(rt http.RoundTripper) http.RoundTripper {
	return &clientHookRetry{
		next:      rt,
		num:       hook.num,
		max:       hook.max,
		intervals: hook.intervals,
		status:    hook.status,
	}
}

func (hook *clientHookRetry) RoundTrip(req *http.Request) (*http.Response, error) {
	for i := 0; i <= hook.num; i++ {
		resp, err := hook.next.RoundTrip(req)
		if hook.condition(resp, err) {
			return resp, err
		}

		// reset body
		if !isNetError(err) {
			err := resetBody(req)
			if err != nil {
				return resp, err
			}
			if resp != nil {
				_ = resp.Body.Close()
			}
		}
		if i < hook.max {
			time.Sleep(hook.intervals[i] + randInterval(hook.intervals[i]))
		} else {
			time.Sleep(hook.intervals[hook.max-1])
		}
	}
	return hook.next.RoundTrip(req)
}

func (hook *clientHookRetry) condition(resp *http.Response, err error) bool {
	if resp != nil {
		_, ok := hook.status[resp.StatusCode]
		if ok {
			return false
		}
	}

	return err == nil
}

func isNetError(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr)
}

func randInterval(n time.Duration) time.Duration {
	if n <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(n)))
}

type clientHookLogger struct {
	next   http.RoundTripper
	level  LoggerLevel
	slow   time.Duration
	params []string
}

// NewClientHookLogger function creates [ClientHook] to implement request
// logger output.
//
// Output [LoggerError] when the response error is not empty;
// output [LoggerWarning] when the request time is greater than slow;
// otherwise output [LoggerInfo] or [LoggerDebug].
//
// Available params: scheme query byte-in byte-out x-response-id
// request-header response-header trace
//
// Default params: byte-in byte-out x-response-id.
func NewClientHookLogger(level LoggerLevel, slow time.Duration, params ...string) ClientHook {
	if params == nil {
		params = []string{"byte-in", "byte-out", "x-response-id"}
	}
	return &clientHookLogger{
		level:  level,
		slow:   slow,
		params: params,
	}
}
func (*clientHookLogger) Name() string { return "logger" }
func (hook *clientHookLogger) String() string {
	return "logger level=" + hook.level.String() + " slow=" + hook.slow.String()
}

func (hook *clientHookLogger) Wrap(rt http.RoundTripper) http.RoundTripper {
	return &clientHookLogger{
		next:   rt,
		level:  hook.level,
		slow:   hook.slow,
		params: hook.params,
	}
}

var (
	clientLoggerKeys1 = [...]string{"host", "method", "path", "duration"}
	clientLoggerKeys2 = [...]string{"host", "method", "path", "proto", "status", "duration"}
)

func (hook *clientHookLogger) RoundTrip(req *http.Request) (*http.Response, error) {
	now := time.Now()
	resp, err := hook.next.RoundTrip(req)
	log := NewLoggerWithContext(req.Context())
	if log.GetLevel() > hook.level {
		return resp, err
	}
	dura := time.Since(now)

	if resp == nil {
		log = log.WithFields(clientLoggerKeys1[:], []any{
			req.Host, req.Method, req.URL.Path,
			GetStringDuration(dura),
		})
	} else {
		log = log.WithFields(clientLoggerKeys2[:], []any{
			req.Host, req.Method, req.URL.Path,
			resp.Proto, resp.StatusCode, GetStringDuration(dura),
		})
	}
	log = hook.withParams(log, req, resp).WithField("time", now)

	switch {
	case err != nil:
		log.Error(err.Error())
	case dura > hook.slow && hook.level <= LoggerWarning:
		log.Warning()
	case hook.level == LoggerInfo:
		log.Info()
	case hook.level == LoggerDebug:
		log.Debug()
	}

	return resp, err
}

//nolint:cyclop,gocyclo
func (hook *clientHookLogger) withParams(log Logger, req *http.Request, resp *http.Response) Logger {
	for _, key := range hook.params {
		switch key {
		case "scheme":
			log = log.WithField(key, req.URL.Scheme)
		case "query":
			log = loggerValues(log, key, req.URL.RawQuery)
		case "byte-in":
			if req.ContentLength > 0 {
				log = log.WithField(key, req.ContentLength)
			}
		case "byte-out":
			if resp != nil && resp.ContentLength > 0 {
				log = log.WithField(key, resp.ContentLength)
			}
		case "x-response-id":
			if resp != nil {
				log = loggerValues(log, key,
					resp.Header.Get(HeaderXRequestID),
					resp.Header.Get(HeaderXTraceID),
				)
			}
		case "request-header":
			log = log.WithField(key, req.Header)
		case "response-header":
			if resp != nil {
				log = log.WithField(key, resp.Header)
			}
		case "trace":
			trace, ok := req.Context().Value(ContextKeyClientTrace).(*ClientTrace)
			if ok {
				trace.Lock()
				trace.HTTPDone = time.Now()
				trace.HTTPDuration = trace.HTTPDone.Sub(trace.HTTPStart)
				trace.Unlock()
				log = log.WithField(key, trace)
			}
		}
	}
	return log
}

func loggerValues(log Logger, key string, vals ...string) Logger {
	for _, val := range vals {
		if val != "" {
			return log.WithField(key, val)
		}
	}
	return log
}

type clientHookDigest struct {
	next         http.RoundTripper
	DomainDigest sync.Map
	username     string
	password     string
	count        int64
}

// NewClientHookDigestAuth function creates [ClientHook] to implement digest
// authentication.
//
// When [StatusUnauthorized], calculate the digest and re-initiate the request.
func NewClientHookDigestAuth(username, password string) ClientHook {
	return &clientHookDigest{
		username: username,
		password: password,
	}
}
func (*clientHookDigest) Name() string { return "" }
func (hook *clientHookDigest) String() string {
	return "digest username=" + hook.username
}

func (hook *clientHookDigest) Wrap(rt http.RoundTripper) http.RoundTripper {
	return &clientHookDigest{
		next:     rt,
		username: hook.username,
		password: hook.password,
	}
}

func (hook *clientHookDigest) RoundTrip(req *http.Request) (*http.Response, error) {
	val, ok := hook.DomainDigest.Load(req.Host)
	if ok {
		// signature
		dig := val.(*ClientDigest).Clone()
		dig.Username = hook.username
		dig.Password = hook.password
		dig.Method = req.Method
		dig.URI = req.URL.RequestURI()
		if dig.Qop == httpDigestQopAuthInt && req.GetBody != nil {
			dig.Body, _ = req.GetBody()
		} else if strings.Contains(dig.Qop, ",") {
			dig.Qop = httpDigestQopAuth
		}
		dig.Nc = fmt.Sprintf("%08x", atomic.AddInt64(&hook.count, 1)&0xffffffff)
		dig.Cnonce = GetStringRandom(8)
		dig.Response = dig.Digest()
		req.Header.Set(HeaderAuthorization, dig.Encode())
	}
	resp, err := hook.next.RoundTrip(req)
	if resp == nil || resp.StatusCode != StatusUnauthorized {
		return resp, err
	}

	auth := resp.Header.Get(HeaderWWWAuthenticate)
	dig := NewClientDigest(auth)
	if dig == nil || dig.invalid() || (ok && !strings.Contains(auth, "stale=true")) {
		return resp, err
	}

	// svae host digest
	domainw := GetAnyDefault(dig.Domain, req.Host)
	dig.Domain = ""
	for _, domain := range strings.Fields(domainw) {
		hook.DomainDigest.Store(domain, dig)
	}

	// try next
	err = resetBody(req)
	if err != nil {
		return resp, err
	}
	_ = resp.Body.Close()
	return hook.RoundTrip(req)
}

func resetBody(req *http.Request) error {
	if req.Body == nil || req.Body == http.NoBody {
		return nil
	}
	_ = req.Body.Close()
	if req.ContentLength != 0 && req.GetBody == nil {
		return ErrClientBodyNotGetBody
	}
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return err
		}
		req.Body = body
	}
	return nil
}

var (
	digestKeys = [...]string{
		"", "domain", "", "uri", "username", "",
		"realm", "algorithm", "qop", "opaque",
		"nonce", "cnonce", "nc", "response",
	}
	httpDigestQopAuth    = "auth"
	httpDigestQopAuthInt = "auth-int"
)

// ClientDigest defines the HTTP Digest data.
//
// Userhash and Authorization-Info are not supported.
//
// RFC 2069: An Extension to HTTP : Digest Access Authentication
//
// RFC 2617: HTTP Authentication: Basic and Digest Access Authentication
//
// RFC 7616: HTTP Digest Access Authentication.
type ClientDigest struct {
	Body      io.ReadCloser // Request.Body
	Domain    string        // Request.Host
	Method    string        // Request.Method
	URI       string        // Request.RequestURI
	Username  string
	Password  string
	Realm     string
	Algorithm string
	Qop       string // allow: ""/auth/auth-int/auth,auth-int
	Opaque    string
	Nonce     string // server-specified string
	Cnonce    string // client-specified string
	Nc        string // client nonce count
	Response  string // it proves that the user knows a password.
}

// NewClientDigest function parses the Digest string and creates [ClientDigest].
//
//nolint:cyclop,gocyclo
func NewClientDigest(req string) *ClientDigest {
	if !strings.HasPrefix(req, "Digest ") {
		return nil
	}

	dig := &ClientDigest{}
	for _, s := range splitDigestString(req[7:]) {
		k, v, ok := strings.Cut(s, "=")
		if !ok {
			return nil
		}
		if len(v) > 2 && v[0] == '"' && v[len(v)-1] == '"' {
			v = v[1 : len(v)-1]
		}

		switch k {
		case "domain":
			dig.Domain = v
		case "uri":
			dig.URI = v
		case "username":
			dig.Username = v
		case "realm":
			dig.Realm = v
		case "algorithm":
			dig.Algorithm = strings.ToUpper(v)
		case "qop":
			dig.Qop = v
		case "opaque":
			dig.Opaque = v
		case "nonce":
			dig.Nonce = v
		case "cnonce":
			dig.Cnonce = v
		case "nc":
			dig.Nc = v
		case "response":
			dig.Response = v
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

func (dig *ClientDigest) invalid() bool {
	switch dig.Algorithm {
	case "", "MD5", "MD5-SESS", "SHA-256", "SHA-256-SESS", "SHA-512-256", "SHA-512-256-SESS":
	default:
		return true // rfc7616 - 3.3
	}
	switch dig.Qop {
	case "", httpDigestQopAuth, httpDigestQopAuthInt:
	default:
		for _, s := range strings.Split(dig.Qop, ",") {
			s = strings.TrimSpace(s)
			if s != httpDigestQopAuth && s != httpDigestQopAuthInt {
				return true
			}
		}
	}
	return false
}

// Clone method copy new [ClientDigest].
func (dig *ClientDigest) Clone() *ClientDigest {
	d := new(ClientDigest)
	*d = *dig
	return d
}

// Encode method formats the Digest message to set the
// [HeaderAuthorization] and [HeaderWWWAuthenticate].
func (dig *ClientDigest) Encode() string {
	buf := bytes.NewBufferString("Digest ")
	data := *(*[14]string)(unsafe.Pointer(dig))
	for i, s := range data {
		if digestKeys[i] != "" && s != "" {
			switch i {
			case 7, 8, 11:
				fmt.Fprintf(buf, "%s=%s, ", digestKeys[i], s)
			default:
				fmt.Fprintf(buf, "%s=\"%s\", ", digestKeys[i], s)
			}
		}
	}
	buf.Truncate(buf.Len() - 2)
	return buf.String()
}

// Digest method calculates the current data's Digest value.
func (dig *ClientDigest) Digest() string {
	var h hash.Hash
	var ha1, ha2 string
	switch dig.Algorithm {
	case "MD5", "MD5-SESS", "":
		h = md5.New()
	case "SHA-256", "SHA-256-SESS":
		h = sha256.New()
	case "SHA-512-256", "SHA-512-256-SESS":
		h = sha512.New512_256()
	}
	ha1 = digestHash(h, fmt.Sprintf("%s:%s:%s", dig.Username, dig.Realm, dig.Password))
	// Session variant
	if strings.HasSuffix(dig.Algorithm, "-SESS") {
		ha1 = digestHash(h, fmt.Sprintf("%s:%s:%s",
			ha1, dig.Nonce, dig.Cnonce,
		))
	}

	switch dig.Qop {
	case httpDigestQopAuth, "":
		ha2 = digestHash(h, fmt.Sprintf("%s:%s", dig.Method, dig.URI))
	case httpDigestQopAuthInt:
		if dig.Body != nil {
			h.Reset()
			_, _ = io.Copy(h, dig.Body)
			ha2 = hex.EncodeToString(h.Sum(nil))
			_ = dig.Body.Close()
		}
		ha2 = digestHash(h, fmt.Sprintf("%s:%s:%s", dig.Method, dig.URI, ha2))
	}

	if dig.Qop == "" && dig.Algorithm == "" {
		// rfc2069
		return digestHash(h, fmt.Sprintf("%s:%s:%s",
			ha1, dig.Nonce, ha2,
		))
	}
	// rfc7616
	return digestHash(h, fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		ha1, dig.Nonce, dig.Nc, dig.Cnonce, dig.Qop, ha2,
	))
}

func digestHash(h hash.Hash, s string) string {
	h.Reset()
	_, _ = io.WriteString(h, s)
	return hex.EncodeToString(h.Sum(nil))
}
