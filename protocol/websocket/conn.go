package websocket

import (
	// "io"
	"net"
	// "fmt"
)

type (
	Conn struct {
		net.Conn
		frame Frame
	}
)

func NewConn(con net.Conn) *Conn {
	return &Conn{Conn: con}
}

func (con *Conn) Read(msg []byte) (n int, err error) {
	for con.frame.pos == int(con.frame.Header.Length) {
		_, err = con.ReadFrame()
		if err != nil {
			return 0, err
		}
	}
	n, err = con.frame.Read(msg)
	return n, err
}

func (con *Conn) ReadFrame() (f Frame, err error) {
	f, err = ReadFrame(con.Conn)
	con.frame = f
	return f, err
}

func (con *Conn) Write(msg []byte) (n int, err error) {
	f := Frame{
		Header: Header{
			Fin:    true,
			OpCode: FrameTypeBinary,
			Length: int64(len(msg)),
		},
		Payload: msg,
	}
	return len(msg), con.WriteFrame(f)
}

func (con *Conn) WriteString(str string) (n int, err error) {
	f := Frame{
		Header: Header{
			Fin:    true,
			OpCode: FrameTypeText,
			Length: int64(len(msg)),
		},
		Payload: strToBytes(str),
	}
	return len(msg), con.WriteFrame(f)
}

func (con *Conn) WriteFrame(f Frame) error {
	return WriteFrame(con.Conn, f)
}
