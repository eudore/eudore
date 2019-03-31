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
	"context"
	"net/textproto"
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

type Fastcgi struct{
	h		protocol.Handler
}

func NewServer(h protocol.Handler) *Fastcgi{
	return &Fastcgi{h}
}

func (f *Fastcgi) EudoreConn(ctx context.Context, rw net.Conn) {
	newChild(ctx, rw, f.h).serve()
}


// Serve accepts incoming FastCGI connections on the listener l, creating a new
// goroutine for each. The goroutine reads requests and then calls handler
// to reply to them.
// If l is nil, Serve accepts connections from os.Stdin.
// If handler is nil, http.DefaultServeMux is used.
func Serve(l net.Listener, handler protocol.Handler) error {
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


type Header map[string][]string

func (h Header) Get(key string) string {
	return textproto.MIMEHeader(h).Get(key)
}

func (h Header) Set(key ,value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

func (h Header) Add(key ,value string) {
	textproto.MIMEHeader(h).Add(key, value)
}

func (h Header) Del(key string) {
	textproto.MIMEHeader(h).Del(key)
}

func (h Header) Range(fn func(string, string)) {
	for k, v := range h {
		for _, vv := range v {
			fn(k, vv)
		}
	}
}
