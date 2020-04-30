package gorilla

import (
	"github.com/eudore/eudore"
	"github.com/gorilla/websocket"
	"io"
)

// StreamWebsocket 定义gorilla websocket实现eudore.Stream。
type StreamWebsocket struct {
	*websocket.Conn
	streamid  string
	readType  int
	writeType int
	reader    io.Reader
}

// NewExtendFuncStream 函数转换eudore.Stream处理函数，使用gorilla库处理websocket请求。
//
// 默认streamID为"streamid"或"UID"取参数的值。
func NewExtendFuncStream(fn func(eudore.Stream)) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		conn, err := websocket.Upgrade(ctx.Response(), ctx.Request(), nil, 2048, 2048)
		if err != nil {
			ctx.Fatal(err)
			return
		}
		var id string
		for _, i := range []string{"streamid", eudore.ParamUID} {
			id := ctx.GetParam(i)
			if id != "" {
				break
			}
		}
		fn(NewStreamWebsocket(conn, id))
	}
}

// NewStreamWebsocket 函数创建一个StreamWebsocket，默认初始读写类型均为1，即websocket.TextMessage。
func NewStreamWebsocket(conn *websocket.Conn, id string) eudore.Stream {
	return &StreamWebsocket{
		Conn:      conn,
		streamid:  id,
		readType:  1,
		writeType: 1,
	}
}

// StreamID 方法返回流id。
func (stream *StreamWebsocket) StreamID() string {
	return stream.streamid
}

// Read 方法实现io.Reader.
func (stream *StreamWebsocket) Read(b []byte) (int, error) {
	if stream.reader == nil {
		msgType, reader, err := stream.NextReader()
		if err != nil {
			return 0, err
		}
		stream.readType = msgType
		stream.reader = reader
	}
	n, err := stream.reader.Read(b)
	if err == io.EOF {
		stream.reader = nil
		err = nil
	}
	return n, err
}

// Write 方法实现io.Writer。
func (stream *StreamWebsocket) Write(b []byte) (n int, err error) {
	w, err := stream.NextWriter(stream.writeType)
	if err != nil {
		return 0, err
	}
	n, err = w.Write(b)
	if err == nil {
		w.Close()
	}
	return
}

// SendMsg 方法默认使用json序列化对象并发送。
func (stream *StreamWebsocket) SendMsg(m interface{}) error {
	return stream.WriteJSON(m)
}

// RecvMsg 方法使用json解码下一帧数据。
func (stream *StreamWebsocket) RecvMsg(m interface{}) error {
	return stream.ReadJSON(m)
}

// GetType 方法返回读类型，即请求桢类型。
func (stream *StreamWebsocket) GetType() int {
	return stream.readType
}

// SetType 方法设置写入类型，即设置响应桢类型。
func (stream *StreamWebsocket) SetType(t int) {
	stream.writeType = t
}
