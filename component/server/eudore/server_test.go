package eudore

import (
	"testing"
	
	// "time"
	"os"
	"fmt"
	"runtime/pprof"
	"context"
	"net/http"
	"github.com/eudore/eudore/component/server/eudore"
	"github.com/eudore/eudore/protocol"
)

func TestStart(t *testing.T) {
	srv := eudore.Server{}
	srv.Set("", &eudore.HttpConfig{
		Addr:	":8088",
	})
	srv.Set("", protocol.HandlerFunc(func(ctx context.Context, w protocol.ResponseWriter, r protocol.RequestReader) {
		w.Write([]byte("start eudore server, this default page."))
	}))
	// startCPUProfile()
	// time.AfterFunc(30* time.Second,stopCPUProfile)
	t.Log(srv.Start())
}




func TestHttp(t *testing.T) {
	srv := http.Server{
		Addr:	":8088",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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