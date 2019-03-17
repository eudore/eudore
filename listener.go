package eudore

import (
	"os"
	"net"
	"sync"
	"strings"
)

type (
	Listeninfo struct {
		Using bool
		Addr string
		Listener net.Listener `json:"-"`
	}
	globalListener struct {
		mu			sync.Locker
		Listeners	[]*Listeninfo
	}
)

var (
	GlobalListener *globalListener
)

func init() {
	addrs := os.Getenv(EUDORE_GRACEFUL_ADDRS)
	GlobalListener = &globalListener{}
	for i, addr := range strings.Split(addrs, ",") {
		file := os.NewFile(uintptr(i + 3), "")
		ln, err := net.FileListener(file)
		if err == nil {
			GlobalListener.Listeners = append(GlobalListener.Listeners, &Listeninfo{
				Using:		false,
				Addr:		addr,
				Listener:	ln,
			})
		}
	}
}

func (gl *globalListener) Listen(addr string) (net.Listener, error) {
	for _, v := range gl.Listeners {
		if addr == v.Addr {
			v.Using = true
			return v.Listener, nil
		}
	}
	ln, err := newListener(addr)
	if err == nil {
		GlobalListener.Listeners = append(GlobalListener.Listeners, &Listeninfo{
			Using:		true,
			Addr:		addr,
			Listener:	ln,
		})
	}
	return ln, err
}

func newListener(addr string) (net.Listener, error) {
	if strings.HasPrefix(addr, "unix://") {
		return net.Listen("unix", addr[7:])
	}
	if strings.HasPrefix(addr, "tcp://") {
		return net.Listen("tcp", addr[6:])
	}
	return net.Listen("tcp", addr)
}

func (gl *globalListener) AllListener() ([]string, []*os.File) {
	var addrs = make([]string, 0, len(gl.Listeners))
	var files = make([]*os.File, 0, len(gl.Listeners))
	for _, g := range gl.Listeners {
		if g.Using {
			fd, err := g.Listener.(*net.TCPListener).File()
			if err == nil {
				addrs = append(addrs, g.Addr)
				files = append(files, fd)
			}
		}
	}
	return addrs, files
}