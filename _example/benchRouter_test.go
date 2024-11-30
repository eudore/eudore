package eudore_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/eudore/eudore"
)

type Route struct {
	Method string
	Path   string
}

var (
	static = []*Route{
		{"GET", "/"},
		{"GET", "/cmd.html"},
		{"GET", "/code.html"},
		{"GET", "/contrib.html"},
		{"GET", "/contribute.html"},
		{"GET", "/debugging_with_gdb.html"},
		{"GET", "/docs.html"},
		{"GET", "/effective_go.html"},
		{"GET", "/files.log"},
		{"GET", "/gccgo_contribute.html"},
		{"GET", "/gccgo_install.html"},
		{"GET", "/go-logo-black.png"},
		{"GET", "/go-logo-blue.png"},
		{"GET", "/go-logo-white.png"},
		{"GET", "/go1.1.html"},
		{"GET", "/go1.2.html"},
		{"GET", "/go1.html"},
		{"GET", "/go1compat.html"},
		{"GET", "/go_faq.html"},
		{"GET", "/go_mem.html"},
		{"GET", "/go_spec.html"},
		{"GET", "/help.html"},
		{"GET", "/ie.css"},
		{"GET", "/install-source.html"},
		{"GET", "/install.html"},
		{"GET", "/logo-153x55.png"},
		{"GET", "/Makefile"},
		{"GET", "/root.html"},
		{"GET", "/share.png"},
		{"GET", "/sieve.gif"},
		{"GET", "/tos.html"},
		{"GET", "/articles/"},
		{"GET", "/articles/go_command.html"},
		{"GET", "/articles/index.html"},
		{"GET", "/articles/wiki/"},
		{"GET", "/articles/wiki/edit.html"},
		{"GET", "/articles/wiki/final-noclosure.go"},
		{"GET", "/articles/wiki/final-noerror.go"},
		{"GET", "/articles/wiki/final-parsetemplate.go"},
		{"GET", "/articles/wiki/final-template.go"},
		{"GET", "/articles/wiki/final.go"},
		{"GET", "/articles/wiki/get.go"},
		{"GET", "/articles/wiki/http-sample.go"},
		{"GET", "/articles/wiki/index.html"},
		{"GET", "/articles/wiki/Makefile"},
		{"GET", "/articles/wiki/notemplate.go"},
		{"GET", "/articles/wiki/part1-noerror.go"},
		{"GET", "/articles/wiki/part1.go"},
		{"GET", "/articles/wiki/part2.go"},
		{"GET", "/articles/wiki/part3-errorhandling.go"},
		{"GET", "/articles/wiki/part3.go"},
		{"GET", "/articles/wiki/test.bash"},
		{"GET", "/articles/wiki/test_edit.good"},
		{"GET", "/articles/wiki/test_Test.txt.good"},
		{"GET", "/articles/wiki/test_view.good"},
		{"GET", "/articles/wiki/view.html"},
		{"GET", "/codewalk/"},
		{"GET", "/codewalk/codewalk.css"},
		{"GET", "/codewalk/codewalk.js"},
		{"GET", "/codewalk/codewalk.xml"},
		{"GET", "/codewalk/functions.xml"},
		{"GET", "/codewalk/markov.go"},
		{"GET", "/codewalk/markov.xml"},
		{"GET", "/codewalk/pig.go"},
		{"GET", "/codewalk/popout.png"},
		{"GET", "/codewalk/run"},
		{"GET", "/codewalk/sharemem.xml"},
		{"GET", "/codewalk/urlpoll.go"},
		{"GET", "/devel/"},
		{"GET", "/devel/release.html"},
		{"GET", "/devel/weekly.html"},
		{"GET", "/gopher/"},
		{"GET", "/gopher/appenginegopher.jpg"},
		{"GET", "/gopher/appenginegophercolor.jpg"},
		{"GET", "/gopher/appenginelogo.gif"},
		{"GET", "/gopher/bumper.png"},
		{"GET", "/gopher/bumper192x108.png"},
		{"GET", "/gopher/bumper320x180.png"},
		{"GET", "/gopher/bumper480x270.png"},
		{"GET", "/gopher/bumper640x360.png"},
		{"GET", "/gopher/doc.png"},
		{"GET", "/gopher/frontpage.png"},
		{"GET", "/gopher/gopherbw.png"},
		{"GET", "/gopher/gophercolor.png"},
		{"GET", "/gopher/gophercolor16x16.png"},
		{"GET", "/gopher/help.png"},
		{"GET", "/gopher/pkg.png"},
		{"GET", "/gopher/project.png"},
		{"GET", "/gopher/ref.png"},
		{"GET", "/gopher/run.png"},
		{"GET", "/gopher/talks.png"},
		{"GET", "/gopher/pencil/"},
		{"GET", "/gopher/pencil/gopherhat.jpg"},
		{"GET", "/gopher/pencil/gopherhelmet.jpg"},
		{"GET", "/gopher/pencil/gophermega.jpg"},
		{"GET", "/gopher/pencil/gopherrunning.jpg"},
		{"GET", "/gopher/pencil/gopherswim.jpg"},
		{"GET", "/gopher/pencil/gopherswrench.jpg"},
		{"GET", "/play/"},
		{"GET", "/play/fib.go"},
		{"GET", "/play/hello.go"},
		{"GET", "/play/life.go"},
		{"GET", "/play/peano.go"},
		{"GET", "/play/pi.go"},
		{"GET", "/play/sieve.go"},
		{"GET", "/play/solitaire.go"},
		{"GET", "/play/tree.go"},
		{"GET", "/progs/"},
		{"GET", "/progs/cgo1.go"},
		{"GET", "/progs/cgo2.go"},
		{"GET", "/progs/cgo3.go"},
		{"GET", "/progs/cgo4.go"},
		{"GET", "/progs/defer.go"},
		{"GET", "/progs/defer.out"},
		{"GET", "/progs/defer2.go"},
		{"GET", "/progs/defer2.out"},
		{"GET", "/progs/eff_bytesize.go"},
		{"GET", "/progs/eff_bytesize.out"},
		{"GET", "/progs/eff_qr.go"},
		{"GET", "/progs/eff_sequence.go"},
		{"GET", "/progs/eff_sequence.out"},
		{"GET", "/progs/eff_unused1.go"},
		{"GET", "/progs/eff_unused2.go"},
		{"GET", "/progs/error.go"},
		{"GET", "/progs/error2.go"},
		{"GET", "/progs/error3.go"},
		{"GET", "/progs/error4.go"},
		{"GET", "/progs/go1.go"},
		{"GET", "/progs/gobs1.go"},
		{"GET", "/progs/gobs2.go"},
		{"GET", "/progs/image_draw.go"},
		{"GET", "/progs/image_package1.go"},
		{"GET", "/progs/image_package1.out"},
		{"GET", "/progs/image_package2.go"},
		{"GET", "/progs/image_package2.out"},
		{"GET", "/progs/image_package3.go"},
		{"GET", "/progs/image_package3.out"},
		{"GET", "/progs/image_package4.go"},
		{"GET", "/progs/image_package4.out"},
		{"GET", "/progs/image_package5.go"},
		{"GET", "/progs/image_package5.out"},
		{"GET", "/progs/image_package6.go"},
		{"GET", "/progs/image_package6.out"},
		{"GET", "/progs/interface.go"},
		{"GET", "/progs/interface2.go"},
		{"GET", "/progs/interface2.out"},
		{"GET", "/progs/json1.go"},
		{"GET", "/progs/json2.go"},
		{"GET", "/progs/json2.out"},
		{"GET", "/progs/json3.go"},
		{"GET", "/progs/json4.go"},
		{"GET", "/progs/json5.go"},
		{"GET", "/progs/run"},
		{"GET", "/progs/slices.go"},
		{"GET", "/progs/timeout1.go"},
		{"GET", "/progs/timeout2.go"},
		{"GET", "/progs/update.bash"},
	}

	githubAPI = []*Route{
		// OAuth Authorizations
		{"GET", "/authorizations"},
		{"GET", "/authorizations/:id"},
		{"POST", "/authorizations"},
		//{"PUT", "/authorizations/clients/:client_id"},
		//{"PATCH", "/authorizations/:id"},
		{"DELETE", "/authorizations/:id"},
		{"GET", "/applications/:client_id/tokens/:access_token"},
		{"DELETE", "/applications/:client_id/tokens"},
		{"DELETE", "/applications/:client_id/tokens/:access_token"},

		// Activity
		{"GET", "/events"},
		{"GET", "/repos/:owner/:repo/events"},
		{"GET", "/networks/:owner/:repo/events"},
		{"GET", "/orgs/:org/events"},
		{"GET", "/users/:user/received_events"},
		{"GET", "/users/:user/received_events/public"},
		{"GET", "/users/:user/events"},
		{"GET", "/users/:user/events/public"},
		{"GET", "/users/:user/events/orgs/:org"},
		{"GET", "/feeds"},
		{"GET", "/notifications"},
		{"GET", "/repos/:owner/:repo/notifications"},
		{"PUT", "/notifications"},
		{"PUT", "/repos/:owner/:repo/notifications"},
		{"GET", "/notifications/threads/:id"},
		//{"PATCH", "/notifications/threads/:id"},
		{"GET", "/notifications/threads/:id/subscription"},
		{"PUT", "/notifications/threads/:id/subscription"},
		{"DELETE", "/notifications/threads/:id/subscription"},
		{"GET", "/repos/:owner/:repo/stargazers"},
		{"GET", "/users/:user/starred"},
		{"GET", "/user/starred"},
		{"GET", "/user/starred/:owner/:repo"},
		{"PUT", "/user/starred/:owner/:repo"},
		{"DELETE", "/user/starred/:owner/:repo"},
		{"GET", "/repos/:owner/:repo/subscribers"},
		{"GET", "/users/:user/subscriptions"},
		{"GET", "/user/subscriptions"},
		{"GET", "/repos/:owner/:repo/subscription"},
		{"PUT", "/repos/:owner/:repo/subscription"},
		{"DELETE", "/repos/:owner/:repo/subscription"},
		{"GET", "/user/subscriptions/:owner/:repo"},
		{"PUT", "/user/subscriptions/:owner/:repo"},
		{"DELETE", "/user/subscriptions/:owner/:repo"},

		// Gists
		{"GET", "/users/:user/gists"},
		{"GET", "/gists"},
		//{"GET", "/gists/public"},
		//{"GET", "/gists/starred"},
		{"GET", "/gists/:id"},
		{"POST", "/gists"},
		//{"PATCH", "/gists/:id"},
		{"PUT", "/gists/:id/star"},
		{"DELETE", "/gists/:id/star"},
		{"GET", "/gists/:id/star"},
		{"POST", "/gists/:id/forks"},
		{"DELETE", "/gists/:id"},

		// Git Data
		{"GET", "/repos/:owner/:repo/git/blobs/:sha"},
		{"POST", "/repos/:owner/:repo/git/blobs"},
		{"GET", "/repos/:owner/:repo/git/commits/:sha"},
		{"POST", "/repos/:owner/:repo/git/commits"},
		//{"GET", "/repos/:owner/:repo/git/refs/*ref"},
		{"GET", "/repos/:owner/:repo/git/refs"},
		{"POST", "/repos/:owner/:repo/git/refs"},
		//{"PATCH", "/repos/:owner/:repo/git/refs/*ref"},
		//{"DELETE", "/repos/:owner/:repo/git/refs/*ref"},
		{"GET", "/repos/:owner/:repo/git/tags/:sha"},
		{"POST", "/repos/:owner/:repo/git/tags"},
		{"GET", "/repos/:owner/:repo/git/trees/:sha"},
		{"POST", "/repos/:owner/:repo/git/trees"},

		// Issues
		{"GET", "/issues"},
		{"GET", "/user/issues"},
		{"GET", "/orgs/:org/issues"},
		{"GET", "/repos/:owner/:repo/issues"},
		{"GET", "/repos/:owner/:repo/issues/:number"},
		{"POST", "/repos/:owner/:repo/issues"},
		//{"PATCH", "/repos/:owner/:repo/issues/:number"},
		{"GET", "/repos/:owner/:repo/assignees"},
		{"GET", "/repos/:owner/:repo/assignees/:assignee"},
		{"GET", "/repos/:owner/:repo/issues/:number/comments"},
		//{"GET", "/repos/:owner/:repo/issues/comments"},
		//{"GET", "/repos/:owner/:repo/issues/comments/:id"},
		{"POST", "/repos/:owner/:repo/issues/:number/comments"},
		//{"PATCH", "/repos/:owner/:repo/issues/comments/:id"},
		//{"DELETE", "/repos/:owner/:repo/issues/comments/:id"},
		{"GET", "/repos/:owner/:repo/issues/:number/events"},
		//{"GET", "/repos/:owner/:repo/issues/events"},
		//{"GET", "/repos/:owner/:repo/issues/events/:id"},
		{"GET", "/repos/:owner/:repo/labels"},
		{"GET", "/repos/:owner/:repo/labels/:name"},
		{"POST", "/repos/:owner/:repo/labels"},
		//{"PATCH", "/repos/:owner/:repo/labels/:name"},
		{"DELETE", "/repos/:owner/:repo/labels/:name"},
		{"GET", "/repos/:owner/:repo/issues/:number/labels"},
		{"POST", "/repos/:owner/:repo/issues/:number/labels"},
		{"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name"},
		{"PUT", "/repos/:owner/:repo/issues/:number/labels"},
		{"DELETE", "/repos/:owner/:repo/issues/:number/labels"},
		{"GET", "/repos/:owner/:repo/milestones/:number/labels"},
		{"GET", "/repos/:owner/:repo/milestones"},
		{"GET", "/repos/:owner/:repo/milestones/:number"},
		{"POST", "/repos/:owner/:repo/milestones"},
		//{"PATCH", "/repos/:owner/:repo/milestones/:number"},
		{"DELETE", "/repos/:owner/:repo/milestones/:number"},

		// Miscellaneous
		{"GET", "/emojis"},
		{"GET", "/gitignore/templates"},
		{"GET", "/gitignore/templates/:name"},
		{"POST", "/markdown"},
		{"POST", "/markdown/raw"},
		{"GET", "/meta"},
		{"GET", "/rate_limit"},

		// Organizations
		{"GET", "/users/:user/orgs"},
		{"GET", "/user/orgs"},
		{"GET", "/orgs/:org"},
		//{"PATCH", "/orgs/:org"},
		{"GET", "/orgs/:org/members"},
		{"GET", "/orgs/:org/members/:user"},
		{"DELETE", "/orgs/:org/members/:user"},
		{"GET", "/orgs/:org/public_members"},
		{"GET", "/orgs/:org/public_members/:user"},
		{"PUT", "/orgs/:org/public_members/:user"},
		{"DELETE", "/orgs/:org/public_members/:user"},
		{"GET", "/orgs/:org/teams"},
		{"GET", "/teams/:id"},
		{"POST", "/orgs/:org/teams"},
		//{"PATCH", "/teams/:id"},
		{"DELETE", "/teams/:id"},
		{"GET", "/teams/:id/members"},
		{"GET", "/teams/:id/members/:user"},
		{"PUT", "/teams/:id/members/:user"},
		{"DELETE", "/teams/:id/members/:user"},
		{"GET", "/teams/:id/repos"},
		{"GET", "/teams/:id/repos/:owner/:repo"},
		{"PUT", "/teams/:id/repos/:owner/:repo"},
		{"DELETE", "/teams/:id/repos/:owner/:repo"},
		{"GET", "/user/teams"},

		// Pull Requests
		{"GET", "/repos/:owner/:repo/pulls"},
		{"GET", "/repos/:owner/:repo/pulls/:number"},
		{"POST", "/repos/:owner/:repo/pulls"},
		//{"PATCH", "/repos/:owner/:repo/pulls/:number"},
		{"GET", "/repos/:owner/:repo/pulls/:number/commits"},
		{"GET", "/repos/:owner/:repo/pulls/:number/files"},
		{"GET", "/repos/:owner/:repo/pulls/:number/merge"},
		{"PUT", "/repos/:owner/:repo/pulls/:number/merge"},
		{"GET", "/repos/:owner/:repo/pulls/:number/comments"},
		//{"GET", "/repos/:owner/:repo/pulls/comments"},
		//{"GET", "/repos/:owner/:repo/pulls/comments/:number"},
		{"PUT", "/repos/:owner/:repo/pulls/:number/comments"},
		//{"PATCH", "/repos/:owner/:repo/pulls/comments/:number"},
		//{"DELETE", "/repos/:owner/:repo/pulls/comments/:number"},

		// Repositories
		{"GET", "/user/repos"},
		{"GET", "/users/:user/repos"},
		{"GET", "/orgs/:org/repos"},
		{"GET", "/repositories"},
		{"POST", "/user/repos"},
		{"POST", "/orgs/:org/repos"},
		{"GET", "/repos/:owner/:repo"},
		//{"PATCH", "/repos/:owner/:repo"},
		{"GET", "/repos/:owner/:repo/contributors"},
		{"GET", "/repos/:owner/:repo/languages"},
		{"GET", "/repos/:owner/:repo/teams"},
		{"GET", "/repos/:owner/:repo/tags"},
		{"GET", "/repos/:owner/:repo/branches"},
		{"GET", "/repos/:owner/:repo/branches/:branch"},
		{"DELETE", "/repos/:owner/:repo"},
		{"GET", "/repos/:owner/:repo/collaborators"},
		{"GET", "/repos/:owner/:repo/collaborators/:user"},
		{"PUT", "/repos/:owner/:repo/collaborators/:user"},
		{"DELETE", "/repos/:owner/:repo/collaborators/:user"},
		{"GET", "/repos/:owner/:repo/comments"},
		{"GET", "/repos/:owner/:repo/commits/:sha/comments"},
		{"POST", "/repos/:owner/:repo/commits/:sha/comments"},
		{"GET", "/repos/:owner/:repo/comments/:id"},
		//{"PATCH", "/repos/:owner/:repo/comments/:id"},
		{"DELETE", "/repos/:owner/:repo/comments/:id"},
		{"GET", "/repos/:owner/:repo/commits"},
		{"GET", "/repos/:owner/:repo/commits/:sha"},
		{"GET", "/repos/:owner/:repo/readme"},
		//{"GET", "/repos/:owner/:repo/contents/*path"},
		//{"PUT", "/repos/:owner/:repo/contents/*path"},
		//{"DELETE", "/repos/:owner/:repo/contents/*path"},
		//{"GET", "/repos/:owner/:repo/:archive_format/:ref"},
		{"GET", "/repos/:owner/:repo/keys"},
		{"GET", "/repos/:owner/:repo/keys/:id"},
		{"POST", "/repos/:owner/:repo/keys"},
		//{"PATCH", "/repos/:owner/:repo/keys/:id"},
		{"DELETE", "/repos/:owner/:repo/keys/:id"},
		{"GET", "/repos/:owner/:repo/downloads"},
		{"GET", "/repos/:owner/:repo/downloads/:id"},
		{"DELETE", "/repos/:owner/:repo/downloads/:id"},
		{"GET", "/repos/:owner/:repo/forks"},
		{"POST", "/repos/:owner/:repo/forks"},
		{"GET", "/repos/:owner/:repo/hooks"},
		{"GET", "/repos/:owner/:repo/hooks/:id"},
		{"POST", "/repos/:owner/:repo/hooks"},
		//{"PATCH", "/repos/:owner/:repo/hooks/:id"},
		{"POST", "/repos/:owner/:repo/hooks/:id/tests"},
		{"DELETE", "/repos/:owner/:repo/hooks/:id"},
		{"POST", "/repos/:owner/:repo/merges"},
		{"GET", "/repos/:owner/:repo/releases"},
		{"GET", "/repos/:owner/:repo/releases/:id"},
		{"POST", "/repos/:owner/:repo/releases"},
		//{"PATCH", "/repos/:owner/:repo/releases/:id"},
		{"DELETE", "/repos/:owner/:repo/releases/:id"},
		{"GET", "/repos/:owner/:repo/releases/:id/assets"},
		{"GET", "/repos/:owner/:repo/stats/contributors"},
		{"GET", "/repos/:owner/:repo/stats/commit_activity"},
		{"GET", "/repos/:owner/:repo/stats/code_frequency"},
		{"GET", "/repos/:owner/:repo/stats/participation"},
		{"GET", "/repos/:owner/:repo/stats/punch_card"},
		{"GET", "/repos/:owner/:repo/statuses/:ref"},
		{"POST", "/repos/:owner/:repo/statuses/:ref"},

		// Search
		{"GET", "/search/repositories"},
		{"GET", "/search/code"},
		{"GET", "/search/issues"},
		{"GET", "/search/users"},
		{"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword"},
		{"GET", "/legacy/repos/search/:keyword"},
		{"GET", "/legacy/user/search/:keyword"},
		{"GET", "/legacy/user/email/:email"},

		// Users
		{"GET", "/users/:user"},
		{"GET", "/user"},
		//{"PATCH", "/user"},
		{"GET", "/users"},
		{"GET", "/user/emails"},
		{"POST", "/user/emails"},
		{"DELETE", "/user/emails"},
		{"GET", "/users/:user/followers"},
		{"GET", "/user/followers"},
		{"GET", "/users/:user/following"},
		{"GET", "/user/following"},
		{"GET", "/user/following/:user"},
		{"GET", "/users/:user/following/:target_user"},
		{"PUT", "/user/following/:user"},
		{"DELETE", "/user/following/:user"},
		{"GET", "/users/:user/keys"},
		{"GET", "/user/keys"},
		{"GET", "/user/keys/:id"},
		{"POST", "/user/keys"},
		//{"PATCH", "/user/keys/:id"},
		{"DELETE", "/user/keys/:id"},
	}

	gplusAPI = []*Route{
		// People
		{"GET", "/people/:userId"},
		{"GET", "/people"},
		{"GET", "/activities/:activityId/people/:collection"},
		{"GET", "/people/:userId/people/:collection"},
		{"GET", "/people/:userId/openIdConnect"},

		// Activities
		{"GET", "/people/:userId/activities/:collection"},
		{"GET", "/activities/:activityId"},
		{"GET", "/activities"},

		// Comments
		{"GET", "/activities/:activityId/comments"},
		{"GET", "/comments/:commentId"},

		// Moments
		{"POST", "/people/:userId/moments/:collection"},
		{"GET", "/people/:userId/moments/:collection"},
		{"DELETE", "/moments/:id"},
	}

	parseAPI = []*Route{
		// Objects
		{"POST", "/1/classes/:className"},
		{"GET", "/1/classes/:className/:objectId"},
		{"PUT", "/1/classes/:className/:objectId"},
		{"GET", "/1/classes/:className"},
		{"DELETE", "/1/classes/:className/:objectId"},

		// Users
		{"POST", "/1/users"},
		{"GET", "/1/login"},
		{"GET", "/1/users/:objectId"},
		{"PUT", "/1/users/:objectId"},
		{"GET", "/1/users"},
		{"DELETE", "/1/users/:objectId"},
		{"POST", "/1/requestPasswordReset"},

		// Roles
		{"POST", "/1/roles"},
		{"GET", "/1/roles/:objectId"},
		{"PUT", "/1/roles/:objectId"},
		{"GET", "/1/roles"},
		{"DELETE", "/1/roles/:objectId"},

		// Files
		{"POST", "/1/files/:fileName"},

		// Analytics
		{"POST", "/1/events/:eventName"},

		// Push Notifications
		{"POST", "/1/push"},

		// Installations
		{"POST", "/1/installations"},
		{"GET", "/1/installations/:objectId"},
		{"PUT", "/1/installations/:objectId"},
		{"GET", "/1/installations"},
		{"DELETE", "/1/installations/:objectId"},

		// Cloud Functions
		{"POST", "/1/functions"},
	}

	apis = [][]*Route{githubAPI, gplusAPI, parseAPI}
)

func benchmarkRoutes(b *testing.B, router http.Handler, routes []*Route) {
	r := httptest.NewRequest("GET", "/", nil)
	u := r.URL
	w := httptest.NewRecorder()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			r.Method = route.Method
			u.Path = route.Path
			r.URL.Path = route.Path
			router.ServeHTTP(w, r)
		}
	}
}

// eudore.App Router
func loadEuodreRoutes(routes []*Route) http.Handler {
	app := eudore.NewApp()
	router := app.Group(" loggerkind=~all")
	num, _ := strconv.Atoi(os.Getenv("NUM"))
	for i := 0; i < num; i++ {
		router.AddMiddleware(func(eudore.Context) {})
	}
	for _, r := range routes {
		router.AddHandler(r.Method, r.Path, eudore.HandlerFuncs{eudoreHandler(r.Method, r.Path)})
	}
	router.AddHandler("404", "", func(ctx eudore.Context) {
		fmt.Println(404, ctx.Path())
	})
	return app
}

var (
	eudoreCtx = os.Getenv("EUDORECTX") == ""
	ok        = []byte("OK")
)

func eudoreHandler(method, path string) eudore.HandlerFunc {
	if eudoreCtx {
		return func(ctx eudore.Context) {
			ctx.Write(ok)
		}
	}
	return eudore.NewHandlerHTTPFunc2(func(w http.ResponseWriter, r *http.Request) {
		w.Write(ok)
	})
}

func BenchmarkEudoreStatic(b *testing.B) {
	benchmarkRoutes(b, loadEuodreRoutes(static), static)
}

func BenchmarkEudoreGitHubAPI(b *testing.B) {
	benchmarkRoutes(b, loadEuodreRoutes(githubAPI), githubAPI)
}

func BenchmarkEudoreGplusAPI(b *testing.B) {
	benchmarkRoutes(b, loadEuodreRoutes(gplusAPI), gplusAPI)
}

func BenchmarkEudoreParseAPI(b *testing.B) {
	benchmarkRoutes(b, loadEuodreRoutes(parseAPI), parseAPI)
}

// eudore router
func loadErouterRoutes(routes []*Route) http.Handler {
	router := NewRouterCoreMux2()
	for _, r := range routes {
		router.HandleFunc(r.Method, r.Path, erouterHandler(r.Method, r.Path))
	}
	return router
}

func erouterHandler(method, path string) Handler {
	return func(w http.ResponseWriter, req *http.Request, p *eudore.Params) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func BenchmarkErouterStatic(b *testing.B) {
	benchmarkRoutes(b, loadErouterRoutes(static), static)
}

func BenchmarkErouterGitHubAPI(b *testing.B) {
	benchmarkRoutes(b, loadErouterRoutes(githubAPI), githubAPI)
}

func BenchmarkErouterGplusAPI(b *testing.B) {
	benchmarkRoutes(b, loadErouterRoutes(gplusAPI), gplusAPI)
}

func BenchmarkErouterParseAPI(b *testing.B) {
	benchmarkRoutes(b, loadErouterRoutes(parseAPI), parseAPI)
}

// Handler 是Erouter处理一个请求的方法，在http.HandlerFunc基础上增加了Parmas。
type Handler func(http.ResponseWriter, *http.Request, *eudore.Params)

// routerCore interface, performing routing, middleware registration, and processing http requests.
type routerCore interface {
	HandleFunc(string, string, Handler)
	ServeHTTP(http.ResponseWriter, *http.Request)
}

var (
	// Page404 是404返回的body
	Page404 = []byte("404 page not found\n")
	// Page405 是405返回的body
	Page405 = []byte("405 method not allowed\n")
	// RouterAllMethod 是默认Any的全部方法
)

// 数组参数的复用池。
var paramArrayPool = sync.Pool{
	New: func() interface{} {
		return &eudore.Params{eudore.ParamRoute, ""}
	},
}

// 默认405处理，返回405状态码和允许的方法
func defaultRouter405Func(w http.ResponseWriter, req *http.Request, param *eudore.Params) {
	w.Header().Add("Allow", strings.Join(eudore.DefaultRouterAllMethod, ", "))
	w.WriteHeader(405)
	w.Write(Page405)
}

// 默认404处理，返回404状态码
func defaultRouter404Func(w http.ResponseWriter, req *http.Request, param *eudore.Params) {
	w.WriteHeader(404)
	w.Write(Page404)
}

// routerCoreMux is implemented based on the radix tree to realize registration and matching of all routers.
type routerCoreMux struct {
	root       *nodeMux
	anyMethods []string
	allMethods []string
	params404  eudore.Params
	params405  eudore.Params
	handler404 Handler
	handler405 Handler
}

type nodeMux struct {
	path   string
	name   string
	route  string
	Cchild []*nodeMux
	Pchild []*nodeMux
	Wchild *nodeMux
	// handlers
	handlers   []nodeMuxHandler
	anyParams  eudore.Params
	anyHandler Handler
}

type nodeMuxHandler struct {
	method string
	params eudore.Params
	funcs  Handler
}

func NewRouterCoreMux2() routerCore {
	return &routerCoreMux{
		root:       &nodeMux{},
		anyMethods: append([]string{}, eudore.DefaultRouterAnyMethod...),
		allMethods: append([]string{}, eudore.DefaultRouterAllMethod...),
		handler404: defaultRouter404Func,
		handler405: defaultRouter405Func,
	}
}

// HandleFunc method register a new route to the router
//
// The router matches the handlers available to the current path from
// the middleware tree and adds them to the front of the handler.
func (mux *routerCoreMux) HandleFunc(method string, path string, handler Handler) {
	switch method {
	case "NOTFOUND", "404":
		mux.params404 = eudore.NewParamsRoute(path)[2:]
		mux.handler404 = handler
	case "METHODNOTALLOWED", "405":
		mux.params405 = eudore.NewParamsRoute(path)[2:]
		mux.handler405 = handler
	case eudore.MethodAny:
		mux.insertRoute(method, path, handler)
	default:
		for _, m := range mux.anyMethods {
			if method == m {
				mux.insertRoute(method, path, handler)
				return
			}
		}
	}
}

// 实现http.Handler接口，进行路由匹配并处理http请求。
func (r *routerCoreMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p := paramArrayPool.Get().(*eudore.Params)
	*p = (*p)[0:2]
	hs := r.Match(req.Method, req.URL.Path, p)
	hs(w, req, p)
	paramArrayPool.Put(p)
}

// Match a request, if the method does not allow direct return to node405, no match returns node404.
func (mux *routerCoreMux) Match(method, path string, params *eudore.Params) Handler {
	node := mux.root.lookNode(path, params)
	// 404
	if node == nil {
		*params = params.Add(mux.params404...)
		return mux.handler404
	}

	params.Set(eudore.ParamRoute, node.route)
	for _, h := range node.handlers {
		if h.method == method {
			*params = params.Add(h.params...)
			return h.funcs
		}
	}
	// any method
	if node.anyHandler != nil {
		for _, m := range mux.anyMethods {
			if m == method {
				*params = params.Add(node.anyParams...)
				return node.anyHandler
			}
		}
	}
	// 405
	*params = params.Add("allow", strings.Join(mux.getAllows(node), ", ")).Add(mux.params405...)
	return mux.handler405
}

func (mux *routerCoreMux) getAllows(node *nodeMux) []string {
	if node.handlers == nil {
		return mux.anyMethods
	}

	methods := make([]string, len(node.handlers))
	for i, h := range node.handlers {
		methods[i] = h.method
	}
	return methods
}

// Add a new route Node.
func (mux *routerCoreMux) insertRoute(method, path string, val Handler) {
	node := mux.root
	params := eudore.NewParamsRoute(path)

	// create a node
	for _, route := range getSplitPath(params.Get(eudore.ParamRoute)) {
		next := &nodeMux{path: route}
		switch route[0] {
		case ':', '*':
			if len(route) > 1 {
				route = route[1:]
			}
			next.name = route
		}
		node = node.insertNode(next)
		if route[0] == '*' {
			break
		}
	}
	node.route = params.Get(eudore.ParamRoute)
	node.setHandler(method, params[2:], val)
}

func (node *nodeMux) setHandler(method string, params eudore.Params, handler Handler) {
	if method == eudore.MethodAny {
		node.anyParams = params
		node.anyHandler = handler
		return
	}

	for i, h := range node.handlers {
		if h.method == method {
			node.handlers[i].params = params
			node.handlers[i].funcs = handler
			return
		}
	}

	node.handlers = append(node.handlers, nodeMuxHandler{
		method, params, handler,
	})
}

// insertNode add a child node to the node.
func (node *nodeMux) insertNode(next *nodeMux) *nodeMux {
	switch {
	case next.name == "": // const
		return node.insertNodeConst(next)
	case next.path[0] == ':': // param
		node.Pchild, next = nodeMuxSetNext(node.Pchild, next)
	case next.path[0] == '*': // wildcard
		// Copy next data and keep the original child information of node.Wchild.
		if node.Wchild != nil {
			node.Wchild.path = next.path
			node.Wchild.name = next.name
			return node.Wchild
		}
		node.Wchild = next
	}
	return next
}

func nodeMuxSetNext(nodes []*nodeMux, next *nodeMux) ([]*nodeMux, *nodeMux) {
	path := next.path
	// check if the current node exists.
	for _, node := range nodes {
		if node.path == path {
			return nodes, node
		}
	}
	return append(nodes, next), next
}

// The insertNodeConst method handles adding constant nodes.
func (node *nodeMux) insertNodeConst(next *nodeMux) *nodeMux {
	for i := range node.Cchild {
		prefix, find := getSubsetPrefix(next.path, node.Cchild[i].path)
		if find {
			// Split node path
			if len(prefix) != len(node.Cchild[i].path) {
				node.Cchild[i].path = node.Cchild[i].path[len(prefix):]
				node.Cchild[i] = &nodeMux{
					path:   prefix,
					Cchild: []*nodeMux{node.Cchild[i]},
				}
			}
			next.path = next.path[len(prefix):]

			if next.path == "" {
				return node.Cchild[i]
			}
			return node.Cchild[i].insertNode(next)
		}
	}

	node.Cchild = append(node.Cchild, next)
	// Constant node is sorted by first char.
	for i := len(node.Cchild) - 1; i > 0; i-- {
		if node.Cchild[i].path[0] < node.Cchild[i-1].path[0] {
			node.Cchild[i], node.Cchild[i-1] = node.Cchild[i-1], node.Cchild[i]
		}
	}
	return next
}

//nolint:cyclop,gocyclo
func (node *nodeMux) lookNode(key string, params *eudore.Params) *nodeMux {
	if key != "" {
		// constant Node match
		for _, child := range node.Cchild {
			if child.path[0] >= key[0] {
				length := len(child.path)
				if len(key) >= length && key[:length] == child.path {
					if n := child.lookNode(key[length:], params); n != nil {
						return n
					}
				}
				break
			}
		}

		// parameter matching, Check if there is a parameter match
		if node.Pchild != nil {
			pos := strings.IndexByte(key, '/')
			if pos == -1 {
				pos = len(key)
			}
			currentKey, nextkey := key[:pos], key[pos:]

			for _, child := range node.Pchild {
				if n := child.lookNode(nextkey, params); n != nil {
					*params = params.Add(child.name, currentKey)
					return n
				}
			}
		}
	} else if node.route != "" {
		// constant match, return data
		return node
	}

	// wildcard node
	if node.Wchild != nil {
		*params = params.Add(node.Wchild.name, key)
		return node.Wchild
	}
	// can't match, return nil
	return nil
}

func getSplitPath(path string) []string {
	var strs []string
	bytes := make([]byte, 0, 64)
	var isblock, isconst bool
	for _, b := range path {
		// block pattern
		if isblock {
			if b == '}' {
				if len(bytes) != 0 && bytes[len(bytes)-1] != '\\' {
					isblock = false
					continue
				}
				// escaping }
				bytes = bytes[:len(bytes)-1]
			}
			bytes = append(bytes, string(b)...)
			continue
		}
		switch b {
		case '/':
			// constant mode, creates a new string in non-constant mode
			if !isconst {
				isconst = true
				strs = append(strs, string(bytes))
				bytes = bytes[:0]
			}
		case ':', '*':
			// variable pattern or wildcard pattern
			isconst = false
			strs = append(strs, string(bytes))
			bytes = bytes[:0]
		case '{':
			isblock = true
			continue
		}
		bytes = append(bytes, string(b)...)
	}
	strs = append(strs, string(bytes))
	if strs[0] == "" {
		strs = strs[1:]
	}
	return strs
}

// Get the largest common prefix of the two strings,
// return the largest common prefix and have the largest common prefix.
func getSubsetPrefix(str2, str1 string) (string, bool) {
	if len(str2) < len(str1) {
		str1, str2 = str2, str1
	}

	for i := range str1 {
		if str1[i] != str2[i] {
			return str1[:i], i > 0
		}
	}
	return str1, true
}
