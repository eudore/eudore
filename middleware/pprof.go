package middleware

import (
	"bytes"
	"fmt"
	"regexp"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/eudore/eudore"
)

// The NewPProfFunc function handles pprof requests.
//
// The route registration path must end with /*.
// Use Params '*' to get the processing function name.
//
//go:noinline
func NewPProfFunc() Middleware {
	return func(ctx eudore.Context) {
		name := ctx.GetParam("*")
		handler, ok := DefaultPProfHandlers[name]
		if ok {
			handler.ServeHTTP(ctx.Response(), ctx.Request())
			return
		}
		switch name {
		case "goroutine":
			handlerPProfGoroutine(ctx)
		default:
			handlerPPorfIndex(ctx)
		}
	}
}

type profile struct {
	Name  string `json:"name"`
	Href  string `json:"href"`
	Desc  string `json:"desc"`
	Count int    `json:"count"`
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

// The handlerPPorfIndex function processes the pprof index page,
// returns the index message, and response formats: format=text/json/html.
func handlerPPorfIndex(ctx eudore.Context) {
	runtimeprofiles := pprof.Profiles()
	profiles := make([]profile, 0, len(runtimeprofiles)+3)
	for _, p := range runtimeprofiles {
		profiles = append(profiles, profile{
			Name:  p.Name(),
			Href:  p.Name() + "?debug=1",
			Desc:  profileDescriptions[p.Name()],
			Count: p.Count(),
		})
	}
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

	switch getRequestFormat(ctx) {
	case QueryFormatJSON:
		ctx.SetHeader(headerContentType, eudore.MimeApplicationJSONCharsetUtf8)
		_ = eudore.HandlerDataRenderJSON(ctx, profiles)
	case QueryFormatHTML:
		ctx.SetHeader(headerContentType, eudore.MimeTextHTMLCharsetUtf8)
		_ = pprofIndexTemplate.ExecuteTemplate(ctx, "index-html", profiles)
	default:
		ctx.SetHeader(headerContentType, eudore.MimeTextPlainCharsetUtf8)
		_ = pprofIndexTemplate.ExecuteTemplate(ctx, "index-text", profiles)
	}
}

var pprofIndexTemplate, _ = template.New("index").Parse(`
{{define "index-html"}}
<html>
<head>
<title>eudore pprof</title>
<script>console.log('godoc=https://golang.org Linked godoc address');</script>
</head>
<body>
Types of profiles available:
<table>
<thead><td>Count</td><td>Profile</td><td>Descriptions</td></thead>
{{range .}}
	<tr><td>{{.Count}}</td><td><a href={{.Href}}>{{.Name}}</a></td><td>{{.Desc}}</td></tr>
{{end}}
</table>
<a href="goroutine?debug=2">full goroutine stack dump</a>
</body>
</html>
{{end}}

{{define "index-text" -}}
Types of profiles available:
Count	Profile		Descriptions
{{range . -}}
{{.Count}}	{{.Name}}	{{.Desc}}
{{end}}
{{end}}
`)

// The handlerPProfGoroutine function processes pprof Goroutine data
// and responds to three formats: format=text/json/html.
func handlerPProfGoroutine(ctx eudore.Context) {
	p := pprof.Lookup("goroutine")
	debug := eudore.GetAnyByString[int](ctx.GetQuery("debug"))
	if debug == 0 {
		name := "attachment; filename=\"goroutine\""
		ctx.SetHeader(eudore.HeaderContentDisposition, name)
		ctx.SetHeader(headerContentType, eudore.MimeApplicationOctetStream)
		_ = p.WriteTo(ctx, 0)
		return
	}

	format := ctx.GetQuery("format")
	if format == QueryFormatText {
		ctx.SetHeader(headerContentType, eudore.MimeTextPlainCharsetUtf8)
		_ = p.WriteTo(ctx, debug)
		return
	}

	var buf bytes.Buffer
	_ = p.WriteTo(&buf, debug)
	var data any
	if debug == 1 {
		data = newGoroutineDebug1(buf.String())
	} else {
		data = newGoroutineDebug2(buf.String())
	}

	if format == QueryFormatJSON {
		ctx.SetHeader(headerContentType, eudore.MimeApplicationJSONCharsetUtf8)
		_ = eudore.HandlerDataRenderJSON(ctx, data)
	} else {
		godoc := eudore.GetAnyByString(ctx.GetQuery("godoc"), ctx.GetParam("godoc"), eudore.DefaultGodocServer)
		godoc = strings.TrimSuffix(godoc, "/")
		tpl, _ := template.New("goroutine").Funcs(template.FuncMap{
			"getPackage": getGodocPackage(godoc),
			"getSource":  getGodocSource(godoc),
		}).Parse(pprofGoroutineTemplate)
		ctx.SetHeader(headerContentType, eudore.MimeTextHTMLCharsetUtf8)
		err := tpl.Execute(ctx, &goroutineData{
			Data:  data,
			Debug: debug,
		})
		ctx.Error(err)
	}
}

type goroutineData struct {
	Data  any
	Debug int
	Godoc string
}

type goroutineDebug1Block struct {
	Args  []string              `json:"args"`
	Lines []goroutineDebug1Line `json:"lines"`
}

type goroutineDebug1Line struct {
	Pointer string `json:"pointer"`
	Func    string `json:"func"`
	Pos     string `json:"pos"`
	Space   string `json:"space"`
	File    string `json:"file"`
	Line    string `json:"line"`
}

type goroutineDebug2Block struct {
	Number string                `json:"number"`
	State  string                `json:"state"`
	Lines  []goroutineDebug2Line `json:"lines"`
}

type goroutineDebug2Line struct {
	Func    string `json:"func"`
	Args    string `json:"args"`
	File    string `json:"file"`
	Line    string `json:"line"`
	Pos     string `json:"pos"`
	Created int    `json:"created,omitempty"`
}

func newGoroutineDebug1(str string) []goroutineDebug1Block {
	reg := regexp.MustCompile(`#\t0x(\S+)\t(\S+)\+0x(\S+)(\s+)(\S+):(\d+)`)
	routines := strings.Split(str[strings.IndexByte(str, '\n')+1:], "\n\n")
	blocks := make([]goroutineDebug1Block, 0, len(routines))
	for i := range routines {
		if routines[i] == "" {
			continue
		}

		arg, _, _ := strings.Cut(routines[i], "\n")
		var block goroutineDebug1Block
		block.Args = strings.Split(arg, " ")
		matchs := reg.FindAllStringSubmatch(routines[i], -1)
		for _, m := range matchs {
			block.Lines = append(block.Lines, goroutineDebug1Line{Pointer: m[1], Func: m[2], Pos: m[3], Space: m[4], File: m[5], Line: m[6]})
		}
		blocks = append(blocks, block)
	}
	return blocks
}

func newGoroutineDebug2(str string) []goroutineDebug2Block {
	reghead := regexp.MustCompile(`goroutine (\d+) \[(.*)\]`)
	regline := regexp.MustCompile(`\n(\S+)\((.*)\)\n\t(\S+):(\d+)( \+0x\S+)?|\n(?:created by )(\S+)(?: in goroutine (\d+))?\n\t(\S+):(\d+) \+0x(\S+)`)
	routines := strings.Split(str, "\n\n")
	blocks := make([]goroutineDebug2Block, 0, len(routines))
	for i := range routines {
		head := reghead.FindStringSubmatch(routines[i])
		block := goroutineDebug2Block{Number: head[1], State: head[2]}
		matchs := regline.FindAllStringSubmatch(routines[i], -1)
		for _, m := range matchs {
			if m[1] != "" {
				block.Lines = append(block.Lines, goroutineDebug2Line{Func: m[1], Args: m[2], File: m[3], Line: m[4], Pos: strings.TrimPrefix(m[5], " +0x")})
			} else {
				id, _ := strconv.Atoi(m[7])
				block.Lines = append(block.Lines, goroutineDebug2Line{Func: m[6], File: m[8], Line: m[9], Pos: m[10], Created: id})
			}
		}
		blocks = append(blocks, block)
	}
	return blocks
}

var pprofGoroutineTemplate = `
<pre>
{{- if eq .Debug 1 }}
goroutine profile: total {{len .Data}}
{{- range $index, $elem := .Data }}
{{ $elem.Args }}
{{- range $index, $elem := $elem.Lines }}
#	0x{{$elem.Pointer}}	{{getPackage $elem.Func}}+0x{{$elem.Pos}}{{$elem.Space}}{{getSource $elem.File $elem.Line}}
{{- end }}
{{ end }}
{{- else }}
{{- range $index, $elem := .Data }}
goroutine {{$elem.Number}} [{{$elem.State}}]:
{{- range $index, $elem := $elem.Lines }}
{{- if ne $elem.Args "" }}
{{getPackage $elem.Func}}({{$elem.Args}})
	{{getSource $elem.File $elem.Line}}{{if $elem.Pos}} +0x{{$elem.Pos}}{{end}}
{{- else }}
created by {{getPackage $elem.Func}}{{if ne $elem.Created 0 }} in goroutine {{$elem.Created}}{{end}}
	{{getSource $elem.File $elem.Line}} +0x{{$elem.Pos}}
{{- end }}
{{- end }}
{{ end }}
{{- end }}
</pre>
`

func getGodocPackage(godoc string) func(string) string {
	return func(pkg string) string {
		if pkg == "main.main" {
			return pkg
		}

		pos := strings.LastIndexByte(pkg, '/')
		if pos == -1 {
			pos = 0
		}
		pos = strings.IndexByte(pkg[pos:], '.') + pos
		fn := pkg[pos+1:]
		pkg = pkg[:pos]

		strs := strings.Split(fn, ".")
		obj := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(strs[0], "("), "*"), ")")
		if obj == "" || obj[0] < 'A' || 'Z' < obj[0] {
			// github.com/eudore/eudore.(*contextBase).Next
			return fmt.Sprintf("<a href='%s/pkg/%s' target='_Blank'>%s</a>.%s", godoc, pkg, pkg, fn)
		}
		if len(strs) == 2 && 0x40 < strs[1][0] && strs[1][0] < 0x5b {
			// github.com/eudore/eudore.(*App).Run
			return fmt.Sprintf("<a href='%s/pkg/%s#%s' target='_Blank'>%s.%s</a>", godoc, pkg, obj+"."+strs[1], pkg, fn)
		}

		pos = strings.IndexByte(fn, '.')
		if pos == -1 {
			// github.com/eudore/eudore/middleware.PProfGoroutine
			return fmt.Sprintf("<a href='%s/pkg/%s#%s' target='_Blank'>%s.%s</a>", godoc, pkg, obj, pkg, fn)
		}
		// github.com/eudore/eudore.(*App).serveContext
		return fmt.Sprintf("<a href='%s/pkg/%s#%s' target='_Blank'>%s.%s</a>%s", godoc, pkg, obj, pkg, fn[:pos], fn[pos:])
	}
}

func getGodocSource(godoc string) func(string, string) string {
	return func(file, line string) string {
		pos := strings.Index(file, "/src/")
		if pos != -1 {
			return fmt.Sprintf("<a href='%s%s#L%s' target='_Blank'>%s</a>:%s", godoc, file[pos:], line, file, line)
		}
		return file + ":" + line
	}
}
