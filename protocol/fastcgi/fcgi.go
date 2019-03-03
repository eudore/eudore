// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fcgi implements the FastCGI protocol.
//
// See https://fast-cgi.github.io/ for an unofficial mirror of the
// original documentation.
//
// Currently only the responder role is supported.
package fastcgi

// This file defines the raw protocol and some utilities used by the child and
// the host.

import (
	"os"
	"net"
	"errors"
	"context"
	"encoding/binary"
	"github.com/eudore/eudore/protocol"
)

// recType is a record type, as defined by
// https://web.archive.org/web/20150420080736/http://www.fastcgi.com/drupal/node/6?q=node/22#S8
type recType uint8

const (
	typeBeginRequest    recType = 1
	typeAbortRequest    recType = 2
	typeEndRequest      recType = 3
	typeParams          recType = 4
	typeStdin           recType = 5
	typeStdout          recType = 6
	typeStderr          recType = 7
	typeData            recType = 8
	typeGetValues       recType = 9
	typeGetValuesResult recType = 10
	typeUnknownType     recType = 11
)

// keep the connection between web-server and responder open after request
const flagKeepConn = 1

const (
	maxWrite = 65535 // maximum record body
	maxPad   = 255
)

const (
	roleResponder = iota + 1 // only Responders are implemented.
	roleAuthorizer
	roleFilter
)

const (
	statusRequestComplete = iota
	statusCantMultiplex
	statusOverloaded
	statusUnknownRole
)

type Fastcgi struct{}

type (
	Handler = protocol.Handler
	Header = protocol.Header
)

func (f *Fastcgi) EudoreConn(ctx context.Context, rw net.Conn, h Handler) {
	newChild(ctx, rw, h).serve()
}



type header struct {
	Version       uint8
	Type          recType
	Id            uint16
	ContentLength uint16
	PaddingLength uint8
	Reserved      uint8
}

type beginRequest struct {
	role     uint16
	flags    uint8
	reserved [5]uint8
}

func (br *beginRequest) read(content []byte) error {
	if len(content) != 8 {
		return errors.New("fcgi: invalid begin request record")
	}
	br.role = binary.BigEndian.Uint16(content)
	br.flags = content[2]
	return nil
}

// for padding so we don't have to allocate all the time
// not synchronized because we don't care what the contents are
var pad [maxPad]byte

func (h *header) init(recType recType, reqId uint16, contentLength int) {
	h.Version = 1
	h.Type = recType
	h.Id = reqId
	h.ContentLength = uint16(contentLength)
	h.PaddingLength = uint8(-contentLength & 7)
}


// Serve accepts incoming FastCGI connections on the listener l, creating a new
// goroutine for each. The goroutine reads requests and then calls handler
// to reply to them.
// If l is nil, Serve accepts connections from os.Stdin.
// If handler is nil, http.DefaultServeMux is used.
func Serve(l net.Listener, handler Handler) error {
	if l == nil {
		var err error
		l, err = net.FileListener(os.Stdin)
		if err != nil {
			return err
		}
		defer l.Close()
	}
	for {
		rw, err := l.Accept()
		if err != nil {
			return err
		}
		c := newChild(nil, rw, handler)
		go c.serve()
	}
}
