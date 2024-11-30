package eudore

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
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
	"time"
	"unsafe"
)

type clientHookCookie struct {
	next   http.RoundTripper
	cookie http.CookieJar
}

// The NewClientHookCookie function creates [ClientHook] and
// uses [http.CookieJar] to manage the cookies sent.
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

// The NewClientHookTimeout function creates [ClientHook] to implement
// request timeout, using [context.WithTimeout].
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

func (b *cancelTimerBody) Read(p []byte) (n int, err error) {
	n, err = b.rc.Read(p)
	if err == nil {
		return n, nil
	}
	if errors.Is(err, io.EOF) {
		return n, err
	}
	if time.Now().After(b.deadline) {
		err = errors.New(err.Error() +
			" (Client.Timeout or context cancellation while reading body)")
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

// NewClientHookRedirect function creates [ClientHook] to implement
// response redirection processing.
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
		case StatusMovedPermanently, StatusFound, StatusSeeOther:
		case StatusPermanentRedirect, StatusTemporaryRedirect:
			if notGetBody(req) {
				return resp, nil
			}
		default:
			return resp, nil
		}

		closeBody(req.Body)
		loc := resp.Header.Get(HeaderLocation)
		if loc == "" {
			return resp, nil
		}
		req, err = newRequest(req, loc, resp.StatusCode)
		if err != nil {
			resp.Body.Close()
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
		resp.Body.Close()
	}
}

func notGetBody(r *http.Request) bool {
	if r.Body == nil || r.Body == http.NoBody {
		return false
	}
	return r.ContentLength != 0 && r.GetBody == nil
}

func closeBody(b io.ReadCloser) {
	if b != nil {
		b.Close()
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
		if r.GetBody != nil {
			r.Body, err = r.GetBody()
			if err != nil {
				return nil, err
			}
			r.ContentLength = req.ContentLength
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

// The NewClientHookRetry function creates [ClientHook] to implement
// multiple request retries.
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
	for i := 0; i < hook.num; i++ {
		resp, err := hook.next.RoundTrip(req)
		if hook.condition(resp, err) {
			return resp, err
		}

		// reset body
		if !isNetError(err) {
			if notGetBody(req) {
				return resp, err
			}
			closeBody(req.Body)
			if resp != nil {
				resp.Body.Close()
			}
			if req.GetBody != nil {
				req.Body, err = req.GetBody()
				if err != nil {
					return resp, err
				}
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

// NewClientHookLogger function creates [ClientHook] to implement
// request logger output.
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
	next               http.RoundTripper
	username, password string
}

// NewClientRetryDigest function creates [ClientHook] to implement digest
// authentication.
//
// When [StatusUnauthorized], calculate the digest and re-initiate the request.
func NewClientHookDigest(username, password string) ClientHook {
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
	resp, err := hook.next.RoundTrip(req)
	if err != nil || resp.StatusCode != StatusUnauthorized {
		return resp, err
	}

	if notGetBody(req) {
		return resp, ErrClientBodyNotGetBody
	}
	dig := newclientDigest(resp.Header.Get(HeaderWWWAuthenticate))
	if dig == nil || dig.invalid() {
		return resp, err
	}

	resp.Body.Close()
	if req.GetBody != nil {
		closeBody(req.Body)
		req.Body, err = req.GetBody()
		if err != nil {
			return resp, err
		}
		if dig.Qop == httpDigestQopAuthInt {
			dig.Body, _ = req.GetBody()
		}
	}

	dig.Nc = "00000001"
	dig.Username = hook.username
	dig.Password = hook.password
	dig.Method = req.Method
	dig.URI = req.URL.Path
	req.Header.Set(HeaderAuthorization, dig.Encode())
	return hook.next.RoundTrip(req)
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
		ha1 = dig.digestHash(fmt.Sprintf("%s:%s:%s",
			dig.Username, dig.Realm, dig.Password,
		))
	case "SHA-256", "SHA-256-SESS":
		dig.Hash = sha256.New()
		ha1 = dig.digestHash(fmt.Sprintf("%s:%s:%s",
			dig.Username, dig.Realm, dig.Password,
		))
	}
	if strings.HasSuffix(dig.Algorithm, "-SESS") {
		ha1 = dig.digestHash(fmt.Sprintf("%s:%s:%s",
			ha1, dig.Nonce, dig.Cnonce,
		))
	}

	switch dig.Qop {
	case httpDigestQopAuth, "":
		ha2 = dig.digestHash(fmt.Sprintf("%s:%s", dig.Method, dig.URI))
	case httpDigestQopAuthInt:
		if dig.Body != nil {
			dig.Hash.Reset()
			_, _ = io.Copy(dig.Hash, dig.Body)
			ha2 = hex.EncodeToString(dig.Hash.Sum(nil))
			dig.Body.Close()
		}
		ha2 = dig.digestHash(fmt.Sprintf("%s:%s:%s", dig.Method, dig.URI, ha2))
	}

	switch dig.Qop {
	case httpDigestQopAuth, httpDigestQopAuthInt:
		dig.Response = dig.digestHash(fmt.Sprintf("%s:%s:00000001:%s:%s:%s",
			ha1, dig.Nonce, dig.Cnonce, dig.Qop, ha2,
		))
	case "":
		dig.Response = dig.digestHash(fmt.Sprintf("%s:%s:%s",
			ha1, dig.Nonce, ha2,
		))
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
	_, _ = io.WriteString(dig.Hash, s)
	return hex.EncodeToString(dig.Hash.Sum(nil))
}
