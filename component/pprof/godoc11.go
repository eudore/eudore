// +build go1.11

package pprof

import (
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"golang.org/x/tools/godoc"
	"golang.org/x/tools/godoc/static"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/gatefs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

// Godoc 定义godoc server对象。
type Godoc struct {
	*godoc.Presentation
	fs    vfs.NameSpace
	route string
}

// NewGodoc 函数创建一个godoc处理对象，需要go1.11版本和GOROOT环境变量存在。
func NewGodoc(route string) http.Handler {
	goroot, gopath := os.Getenv("GOROOT"), os.Getenv("GOPATH")
	if goroot == "" {
		return nil
	}

	vfs.GOROOT = goroot
	fsGate := make(chan bool, 20)
	fs := vfs.NameSpace{}
	fs.Bind("/", gatefs.New(vfs.OS(goroot), fsGate), "/", vfs.BindReplace)
	for _, path := range strings.Split(gopath, ":") {
		fs.Bind("/src", gatefs.New(vfs.OS(path), fsGate), "/src", vfs.BindAfter)
	}
	fs.Bind("/lib/godoc", mapfs.New(static.Files), "/", vfs.BindReplace)

	corpus := godoc.NewCorpus(fs)
	corpus.Verbose = false
	corpus.MaxResults = 10000
	corpus.IndexEnabled = false
	corpus.IndexFiles = ""
	corpus.IndexDirectory = func(dir string) bool {
		return dir != "/pkg" && !strings.HasPrefix(dir, "/pkg/")
	}
	corpus.IndexThrottle = 0.75
	corpus.IndexInterval = 0
	go func() {
		err := corpus.Init()
		if err != nil {
			log.Fatal(err)
		}
		corpus.RunIndexer()
	}()
	corpus.InitVersionInfo()

	gd := &Godoc{
		route:        route,
		fs:           fs,
		Presentation: godoc.NewPresentation(corpus),
	}
	gd.Presentation.ShowTimestamps = false
	gd.Presentation.ShowPlayground = false
	gd.Presentation.DeclLinks = true

	gd.readTemplates()

	return gd
}

func (doc *Godoc) readTemplates() {
	doc.CallGraphHTML = doc.readTemplate("callgraph.html")
	doc.DirlistHTML = doc.readTemplate("dirlist.html")
	doc.ErrorHTML = doc.readTemplate("error.html")
	doc.ExampleHTML = doc.readTemplate("example.html")
	doc.GodocHTML = doc.readTemplate("godoc.html")
	doc.ImplementsHTML = doc.readTemplate("implements.html")
	doc.MethodSetHTML = doc.readTemplate("methodset.html")
	doc.PackageHTML = doc.readTemplate("package.html")
	doc.PackageRootHTML = doc.readTemplate("packageroot.html")
	doc.SearchHTML = doc.readTemplate("search.html")
	doc.SearchDocHTML = doc.readTemplate("searchdoc.html")
	doc.SearchCodeHTML = doc.readTemplate("searchcode.html")
	doc.SearchTxtHTML = doc.readTemplate("searchtxt.html")
}

func (doc *Godoc) readTemplate(name string) *template.Template {
	path := "lib/godoc/" + name

	// use underlying file system fs to read the template file
	// (cannot use template ParseFile functions directly)
	data, err := vfs.ReadFile(doc.fs, path)
	if err != nil {
		log.Fatal("readTemplate: ", err)
	}

	body := string(data)
	body = strings.Replace(body, `<script src="/lib/godoc/`, `<script src="`+doc.route+"/lib/godoc/", -1)
	body = strings.Replace(body, `<link type="text/css" rel="stylesheet" href="/lib/godoc/`, `<link type="text/css" rel="stylesheet" href="`+doc.route+"/lib/godoc/", -1)
	if name == "godoc.html" {
		body += "<script> var route='" + doc.route +
			`';
var dom = document.getElementById(location.hash.slice(1));
if(dom!=null && location.search==''){
	var newnode = document.createElement("span");
	newnode.className="selection";
	newnode.innerText=dom.nextSibling.nodeValue;
	dom.parentNode.insertBefore(newnode,dom.nextElementSibling);
	dom.parentNode.removeChild(dom.nextSibling);
}
for(var i of document.getElementsByTagName('a')){
	if(i.href.indexOf(location.origin)==0 && i.href.indexOf(route)==-1){
		i.href = location.origin+route+i.href.slice(location.origin.length);
	}
}
</script>`
	}

	// be explicit with errors (for app engine use)
	t, err := template.New(name).Funcs(doc.FuncMap()).Parse(body)
	if err != nil {
		log.Fatal("readTemplate: ", err)
	}
	return t
}
