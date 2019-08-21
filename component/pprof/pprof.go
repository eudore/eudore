package pprof

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/eudore/eudore"
)

// InjectRoutes 函数实现注入pprof路由。
func RoutesInject(r eudore.RouterMethod) {
	r = r.Group("/pprof")
	r.AnyFunc("/*name", Index)
	r.AnyFunc("/cmdline", Cmdline)
	r.AnyFunc("/profile", Profile)
	r.AnyFunc("/symbol", Symbol)
	r.AnyFunc("/trace", Trace)
}

// Cmdline responds with the running program's
// command line, with arguments separated by NUL bytes.
// The package initialization registers it as /debug/pprof/cmdline.
func Cmdline(ctx eudore.Context) {
	ctx.SetHeader("X-Content-Type-Options", "nosniff")
	ctx.SetHeader("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(ctx, strings.Join(os.Args, "\x00"))
}

func sleep(ctx eudore.Context, d time.Duration) {
	select {
	case <-time.After(d):
	case <-ctx.Context().Done():
	}
}

func serveError(ctx eudore.Context, status int, txt string) {
	ctx.SetHeader("Content-Type", "text/plain; charset=utf-8")
	ctx.SetHeader("X-Go-Pprof", "1")
	ctx.Response().Header().Del("Content-Disposition")
	ctx.WriteHeader(status)
	fmt.Fprintln(ctx, txt)
}

// Profile responds with the pprof-formatted cpu profile.
// Profiling lasts for duration specified in seconds GET parameter, or for 30 seconds if not specified.
// The package initialization registers it as /debug/pprof/profile.
func Profile(ctx eudore.Context) {
	ctx.SetHeader("X-Content-Type-Options", "nosniff")
	sec, err := strconv.ParseInt(ctx.GetQuery("seconds"), 10, 64)
	if sec <= 0 || err != nil {
		sec = 30
	}

	// Set Content Type assuming StartCPUProfile will work,
	// because if it does it starts writing.
	ctx.SetHeader("Content-Type", "application/octet-stream")
	ctx.SetHeader("Content-Disposition", `attachment; filename="profile"`)
	if err := pprof.StartCPUProfile(ctx); err != nil {
		// StartCPUProfile failed, so no writes yet.
		serveError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("Could not enable CPU profiling: %s", err))
		return
	}
	sleep(ctx, time.Duration(sec)*time.Second)
	pprof.StopCPUProfile()
}

// Trace responds with the execution trace in binary form.
// Tracing lasts for duration specified in seconds GET parameter, or for 1 second if not specified.
// The package initialization registers it as /debug/pprof/trace.
func Trace(ctx eudore.Context) {
	ctx.SetHeader("X-Content-Type-Options", "nosniff")
	sec, err := strconv.ParseFloat(ctx.GetQuery("seconds"), 64)
	if sec <= 0 || err != nil {
		sec = 1
	}

	// Set Content Type assuming trace.Start will work,
	// because if it does it starts writing.
	ctx.SetHeader("Content-Type", "application/octet-stream")
	ctx.SetHeader("Content-Disposition", `attachment; filename="trace"`)
	if err := trace.Start(ctx); err != nil {
		// trace.Start failed, so no writes yet.
		serveError(ctx, eudore.StatusInternalServerError,
			fmt.Sprintf("Could not enable tracing: %s", err))
		return
	}
	sleep(ctx, time.Duration(sec*float64(time.Second)))
	trace.Stop()
}

// Symbol looks up the program counters listed in the request,
// responding with a table mapping program counters to function names.
// The package initialization registers it as /debug/pprof/symbol.
func Symbol(ctx eudore.Context) {
	ctx.SetHeader("X-Content-Type-Options", "nosniff")
	ctx.SetHeader("Content-Type", "text/plain; charset=utf-8")

	// We have to read the whole POST body before
	// writing any output. Buffer the output here.
	var buf bytes.Buffer

	// We don't know how many symbols we have, but we
	// do have symbol information. Pprof only cares whether
	// this number is 0 (no symbols available) or > 0.
	fmt.Fprintf(&buf, "num_symbols: 1\n")

	var b *bufio.Reader
	if ctx.Method() == "POST" {
		b = bufio.NewReader(ctx)
	} else {
		b = bufio.NewReader(strings.NewReader(ctx.Request().RequestURI()[len(ctx.Method()):]))
	}

	for {
		word, err := b.ReadSlice('+')
		if err == nil {
			word = word[0 : len(word)-1] // trim +
		}
		pc, _ := strconv.ParseUint(string(word), 0, 64)
		if pc != 0 {
			f := runtime.FuncForPC(uintptr(pc))
			if f != nil {
				fmt.Fprintf(&buf, "%#x %s\n", pc, f.Name())
			}
		}

		// Wait until here to check for err; the last
		// symbol will have an err because it doesn't end in +.
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(&buf, "reading request: %v\n", err)
			}
			break
		}
	}

	ctx.Write(buf.Bytes())
}

// Handler returns an HTTP handler that serves the named profile.
func Handler(name string) eudore.HandlerFunc {
	return handler(name).Handle
}

type handler string

func (name handler) Handle(ctx eudore.Context) {
	ctx.SetHeader("X-Content-Type-Options", "nosniff")
	p := pprof.Lookup(string(name))
	if p == nil {
		serveError(ctx, http.StatusNotFound, "Unknown profile")
		return
	}
	gc, _ := strconv.Atoi(ctx.GetQuery("gc"))
	if name == "heap" && gc > 0 {
		runtime.GC()
	}
	debug, _ := strconv.Atoi(ctx.GetQuery("debug"))
	if debug != 0 {
		ctx.SetHeader("Content-Type", "text/plain; charset=utf-8")
	} else {
		ctx.SetHeader("Content-Type", "application/octet-stream")
		ctx.SetHeader("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	}
	p.WriteTo(ctx, debug)
}

var profileDescriptions = map[string]string{
	"allocs":       "A sampling of all past memory allocations",
	"block":        "Stack traces that led to blocking on synchronization primitives",
	"cmdline":      "The command line invocation of the current program",
	"goroutine":    "Stack traces of all current goroutines",
	"heap":         "A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.",
	"mutex":        "Stack traces of holders of contended mutexes",
	"profile":      "CPU profile. You can specify the duration in the seconds GET parameter. After you get the profile file, use the go tool pprof command to investigate the profile.",
	"threadcreate": "Stack traces that led to the creation of new OS threads",
	"trace":        "A trace of execution of the current program. You can specify the duration in the seconds GET parameter. After you get the trace file, use the go tool trace command to investigate the trace.",
}

// Index responds with the pprof-formatted profile named by the request.
// For example, "/debug/pprof/heap" serves the "heap" profile.
// Index responds to a request for "/debug/pprof/" with an HTML page
// listing the available profiles.
func Index(ctx eudore.Context) {
	if name := ctx.GetParam("name"); name != "" {
		handler(name).Handle(ctx)
		return
	}

	type profile struct {
		Name  string
		Href  string
		Desc  string
		Count int
	}
	var profiles []profile
	for _, p := range pprof.Profiles() {
		profiles = append(profiles, profile{
			Name:  p.Name(),
			Href:  p.Name() + "?debug=1",
			Desc:  profileDescriptions[p.Name()],
			Count: p.Count(),
		})
	}

	// Adding other profiles exposed from within this package
	for _, p := range []string{"cmdline", "profile", "trace"} {
		profiles = append(profiles, profile{
			Name: p,
			Href: p,
			Desc: profileDescriptions[p],
		})
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	if err := indexTmpl.Execute(ctx, profiles); err != nil {
		log.Print(err)
	}
}

var indexTmpl = template.Must(template.New("index").Parse(`<html>
<head>
<title>/debug/pprof/</title>
<style>
.profile-name{
	display:inline-block;
	width:6rem;
}
</style>
</head>
<body>
/debug/pprof/<br>
<br>
Types of profiles available:
<table>
<thead><td>Count</td><td>Profile</td></thead>
{{range .}}
	<tr>
	<td>{{.Count}}</td><td><a href={{.Href}}>{{.Name}}</a></td>
	</tr>
{{end}}
</table>
<a href="goroutine?debug=2">full goroutine stack dump</a>
<br/>
<p>
Profile Descriptions:
<ul>
{{range .}}
<li><div class=profile-name>{{.Name}}:</div> {{.Desc}}</li>
{{end}}
</ul>
</p>
</body>
</html>
`))
