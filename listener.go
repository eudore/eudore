package eudore

import (
	"os"
	"net"
	"strings"
)

type (
	Listeninfo struct {
		Using bool
		Addrs string
		Listener net.Listener
	}
	globalListener struct {
		Using		[]bool
		Addrs		[]string
		Listeners	[]net.Listener
	}
)

var (
	GlobalListener *globalListener
)

func init() {
	addrs := os.Getenv(GRACEFUL_ENVIRON_ADDRS)
	GlobalListener = &globalListener{
		Addrs:	strings.Split(addrs,","),
	}
}

func (gl *globalListener) GetListener(addr string) net.Listener {
	for i, v := range gl.Addrs {
		if addr == v {
			gl.Using[i] = true
			return gl.Listeners[i]
		}
	}
	if strings.HasPrefix(addr, "unix://") {
		ln, _ := net.Listen("unix", addr[7:])
		return ln
	}
	ln, _ := net.Listen("tcp", addr)
	return ln
}
