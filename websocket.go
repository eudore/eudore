package eudore
// source: github.com/gobwas/ws

import (
	"io"
	"net"
	"hash"
	"sync"
	"bufio"
	"unsafe"
	"reflect"
	"crypto/sha1"
	"encoding/base64"
)

const (
	// RFC6455: The value of this header field MUST be a nonce consisting of a
	// randomly selected 16-byte value that has been base64-encoded (see
	// Section 4 of [RFC4648]).  The nonce MUST be selected randomly for each
	// connection.
	nonceKeySize = 16
	nonceSize    = 24 // base64.StdEncoding.EncodedLen(nonceKeySize)

	// RFC6455: The value of this header field is constructed by concatenating
	// /key/, defined above in step 4 in Section 4.2.2, with the string
	// "258EAFA5- E914-47DA-95CA-C5AB0DC85B11", taking the SHA-1 hash of this
	// concatenated value to obtain a 20-byte value and base64- encoding (see
	// Section 4 of [RFC4648]) this 20-byte hash.
	acceptSize = 28 // base64.StdEncoding.EncodedLen(sha1.Size)
)


var (
	ErrHandshakeBadProtocol = NewError(StatusHTTPVersionNotSupported, "handshake error: bad HTTP protocol version")
	ErrHandshakeBadMethod = NewError(StatusMethodNotAllowed, "handshake error: bad HTTP request method")
	ErrHandshakeBadHost = NewError(StatusBadRequest, "handshake error: bad Host heade")
	ErrHandshakeBadUpgrade = NewError(StatusBadRequest, "handshake error: bad Upgrade header")
	ErrHandshakeBadConnection = NewError(StatusBadRequest, "handshake error: bad Connection header")
	ErrHandshakeBadSecAccept = NewError(StatusBadRequest, "handshake error: bad Sec-Websocket-Accept header")
	ErrHandshakeBadSecKey = NewError(StatusBadRequest, "handshake error: bad Sec-Websocket-Key header")
	ErrHandshakeBadSecVersion = NewError(StatusBadRequest, "handshake error: bad Sec-Websocket-Version header")
	ErrHandshakeUpgradeRequired = NewError(StatusUpgradeRequired, "handshake error: bad Sec-Websocket-Version header")
)


func UpgradeHttp(ctx Context) (net.Conn, error) {
	conn, err := ctx.Response().Hijack()
	if err != nil {

	}

	rw := bufio.NewWriter(conn)
	var nonce string
	if ctx.Method() != MethodGet {
		err = ErrHandshakeBadMethod
	} else if ctx.Request().Proto() != "HTTP/1.1" {
		err = ErrHandshakeBadProtocol
	} else if ctx.Host() == "" {
		err = ErrHandshakeBadHost
	} else if ctx.GetHeader("Upgrade") != "websocket" {
		err = ErrHandshakeBadUpgrade
	} else if ctx.GetHeader("Connection") != "Upgrade" {
		err = ErrHandshakeBadConnection
	} else if v := ctx.GetHeader("Sec-Websocket-Version"); v != "13" {
		if v != "" {
			err = ErrHandshakeUpgradeRequired
		} else {
			err = ErrHandshakeBadSecVersion
		}
	} else if nonce = ctx.GetHeader("Sec-Websocket-Key"); nonce == "" {
		err = ErrHandshakeBadSecKey
	}

	if err == nil {
		httpWriteResponseUpgrade(rw, []byte(nonce))
		err = rw.Flush()
	}else {
		var code int = 500
		if err2, ok := err.(*ErrorHttp); ok {
			code = err2.Code()
		}
		ctx.WriteHeader(code)
		ctx.WriteString(err.Error())
		err = rw.Flush()
	}
	return conn, err
}

func httpWriteResponseUpgrade(bw *bufio.Writer, nonce []byte) {
	const textHeadUpgrade = "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n"
	bw.WriteString(textHeadUpgrade)
	bw.WriteString("Sec-Websocket-Accept: ")
	writeAccept(bw, nonce)
	bw.WriteString("\r\n\r\n")
}


func writeAccept(w io.Writer, nonce []byte) (int, error) {
	var b [acceptSize]byte
	bp := uintptr(unsafe.Pointer(&b))
	bts := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: bp,
		Len:  acceptSize,
		Cap:  acceptSize,
	}))

	initAcceptFromNonce(bts, nonce)

	return w.Write(bts)
}

// initAcceptFromNonce fills given slice with accept bytes generated from given
// nonce bytes. Given buffer should be exactly acceptSize bytes.
func initAcceptFromNonce(dst, nonce []byte) {
	if len(dst) != acceptSize {
		panic("accept buffer is invalid")
	}
	if len(nonce) != nonceSize {
		panic("nonce is invalid")
	}

	sha := acquireSha1()
	defer releaseSha1(sha)

	sha.Write(nonce)
	sha.Write(webSocketMagic)

	var sb [sha1.Size]byte
	sh := uintptr(unsafe.Pointer(&sb))
	sum := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: sh,
		Len:  0,
		Cap:  sha1.Size,
	}))
	sum = sha.Sum(sum)

	base64.StdEncoding.Encode(dst, sum)
}



var webSocketMagic = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

var sha1Pool sync.Pool

// nonce helps to put nonce bytes on the stack and then retrieve stack-backed
// slice with unsafe.
type nonce [nonceSize]byte

func (n *nonce) bytes() []byte {
	h := uintptr(unsafe.Pointer(n))
	b := &reflect.SliceHeader{Data: h, Len: nonceSize, Cap: nonceSize}
	return *(*[]byte)(unsafe.Pointer(b))
}
func acquireSha1() hash.Hash {
	if h := sha1Pool.Get(); h != nil {
		return h.(hash.Hash)
	}
	return sha1.New()
}

func releaseSha1(h hash.Hash) {
	h.Reset()
	sha1Pool.Put(h)
}

