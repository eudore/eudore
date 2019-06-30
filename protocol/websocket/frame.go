package websocket

import (
	"io"
	// "fmt"
	"encoding/binary"
	"reflect"
	"unsafe"
)

const (
	FrameTypeContinuation FrameType = 0x0
	FrameTypeText         FrameType = 0x1
	FrameTypeBinary       FrameType = 0x2
	FrameTypeClose        FrameType = 0x8
	FrameTypePing         FrameType = 0x9
	FrameTypePong         FrameType = 0xa
)

type (
	FrameType int8
	Frame     struct {
		Header
		Payload []byte
		pos     int
	}
	Header struct {
		Fin    bool
		Rsv    byte
		OpCode FrameType
		Masked bool
		Mask   [4]byte
		Length int64
	}
)

func NewFrame(op FrameType) Frame {
	return Frame{
		Header: Header{
			OpCode: op,
		},
	}
}

func ReadHeader(r io.Reader) (h Header, err error) {
	// Make slice of bytes with capacity 12 that could hold any header.
	//
	// The maximum header size is 14, but due to the 2 hop reads,
	// after first hop that reads first 2 constant bytes, we could reuse 2 bytes.
	// So 14 - 2 = 12.
	//
	// We use unsafe to stick bts to stack and avoid allocations.
	//
	// Using stack based slice is safe here, cause golang docs for io.Reader
	// says that "Implementations must not retain p".
	// See https://golang.org/pkg/io/#Reader
	var b [MaxHeaderSize - 2]byte
	bp := uintptr(unsafe.Pointer(&b))
	bh := &reflect.SliceHeader{Data: bp, Len: 2, Cap: MaxHeaderSize - 2}
	bts := *(*[]byte)(unsafe.Pointer(bh))

	// Prepare to hold first 2 bytes to choose size of next read.
	_, err = io.ReadFull(r, bts)
	if err != nil {
		return
	}

	h.Fin = bts[0]&bit0 != 0
	h.Rsv = (bts[0] & 0x70) >> 4
	h.OpCode = FrameType(bts[0] & 0x0f)

	var extra int

	if bts[1]&bit0 != 0 {
		h.Masked = true
		extra += 4
	}

	length := bts[1] & 0x7f
	switch {
	case length < 126:
		h.Length = int64(length)

	case length == 126:
		extra += 2

	case length == 127:
		extra += 8

	default:
		err = ErrHeaderLengthUnexpected
		return
	}

	if extra == 0 {
		return
	}

	// Increase len of bts to extra bytes need to read.
	// Overwrite frist 2 bytes read before.
	bts = bts[:extra]
	_, err = io.ReadFull(r, bts)
	if err != nil {
		return
	}

	switch {
	case length == 126:
		h.Length = int64(binary.BigEndian.Uint16(bts[:2]))
		bts = bts[2:]

	case length == 127:
		if bts[0]&0x80 != 0 {
			err = ErrHeaderLengthMSB
			return
		}
		h.Length = int64(binary.BigEndian.Uint64(bts[:8]))
		bts = bts[8:]
	}

	if h.Masked {
		copy(h.Mask[:], bts)
	}

	return
}

func HeaderSize(h Header) (n int) {
	switch {
	case h.Length < 126:
		n = 2
	case h.Length <= len16:
		n = 4
	case h.Length <= len64:
		n = 10
	default:
		return -1
	}
	if h.Masked {
		n += len(h.Mask)
	}
	return n
}

// WriteHeader writes header binary representation into w.
func WriteHeader(w io.Writer, h Header) error {
	// Make slice of bytes with capacity 14 that could hold any header.
	//
	// We use unsafe to stick bts to stack and avoid allocations.
	//
	// Using stack based slice is safe here, cause golang docs for io.Writer
	// says that "Implementations must not retain p".
	// See https://golang.org/pkg/io/#Writer
	var b [MaxHeaderSize]byte
	bp := uintptr(unsafe.Pointer(&b))
	bh := &reflect.SliceHeader{
		Data: bp,
		Len:  MaxHeaderSize,
		Cap:  MaxHeaderSize,
	}
	bts := *(*[]byte)(unsafe.Pointer(bh))
	_ = bts[MaxHeaderSize-1] // bounds check hint to compiler.

	if h.Fin {
		bts[0] |= bit0
	}
	bts[0] |= h.Rsv << 4
	bts[0] |= byte(h.OpCode)

	var n int
	switch {
	case h.Length <= len7:
		bts[1] = byte(h.Length)
		n = 2

	case h.Length <= len16:
		bts[1] = 126
		binary.BigEndian.PutUint16(bts[2:4], uint16(h.Length))
		n = 4

	case h.Length <= len64:
		bts[1] = 127
		binary.BigEndian.PutUint64(bts[2:10], uint64(h.Length))
		n = 10

	default:
		return ErrHeaderLengthUnexpected
	}

	if h.Masked {
		bts[1] |= bit0
		n += copy(bts[n:], h.Mask[:])
	}

	fmt.Println("writes", bts[:n])
	_, err := w.Write(bts[:n])

	return err
}

func ReadFrame(r io.Reader) (f Frame, err error) {
	f.Header, err = ReadHeader(r)
	if err != nil {
		return
	}

	if f.Header.Length > 0 {
		// int(f.Header.Length) is safe here cause we have
		// checked it for overflow above in ReadHeader.
		f.Payload = make([]byte, int(f.Header.Length))
		_, err = io.ReadFull(r, f.Payload)
	}

	return
}

// WriteFrame writes frame binary representation into w.
func WriteFrame(w io.Writer, f Frame) error {
	err := WriteHeader(w, f.Header)
	if err != nil {
		return err
	}
	_, err = w.Write(f.Payload)
	return err
}

func (frame *Frame) Read(msg []byte) (n int, err error) {
	n = int(frame.Header.Length)
	if n == frame.pos {
		return 0, io.EOF
	}
	if len(msg) < n {
		n = len(msg)
	}
	// fmt.Println(n)
	if frame.Header.Masked {
		for i := 0; i < n; i++ {
			msg[i] = frame.Payload[frame.pos] ^ frame.Header.Mask[frame.pos%4]
			frame.pos++
		}
	} else {
		copy(msg, frame.Payload[frame.pos:frame.pos+n])
		frame.pos += n
	}
	return n, nil
}

func (frame *Frame) Write(msg []byte) (n int, err error) {
	if len(frame.Payload) == 0 {
		frame.Payload = msg
	} else {
		frame.Payload = append(frame.Payload, msg...)
	}
	return len(msg), nil
}
