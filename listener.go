package eudore

import (
	"net"
)

type (
	GlobalListens struct {
		Names		[]string
		Listeners	[]net.Listener
	}
)
