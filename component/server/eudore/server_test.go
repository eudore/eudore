package eudore

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/pprof"
	"testing"

	"github.com/eudore/eudore/component/server/eudore"
)

func TestStart(t *testing.T) {
	srv := eudore.NewServer(nil)
	ln, err := net.Listen("tcp", ":8084")
	if err != nil {
		panic(err)
	}
	srv.AddListener(ln)
	// startCPUProfile()
	// time.AfterFunc(30* time.Second,stopCPUProfile)
	srv.Start()
}

func TestHttp(*testing.T) {
	srv := http.Server{
		Addr: ":8088",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("start eudore server, this default page."))
		}),
	}
	// startCPUProfile()
	// time.AfterFunc(3* time.Second,stopCPUProfile)
	srv.ListenAndServe()
}

var cpuProfile = "/tmp/cpu.out"

func startCPUProfile() {
	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can not create cpu profile output file: %s",
				err)
			return
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Can not start cpu profile: %s", err)
			f.Close()
			return
		}
	}
}

func stopCPUProfile() {
	if cpuProfile != "" {
		pprof.StopCPUProfile() // 把记录的概要信息写到已指定的文件
	}
}
