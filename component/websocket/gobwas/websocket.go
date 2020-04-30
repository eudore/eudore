package gobwas

import (
	"encoding/json"
	"io"
	"net"

	"github.com/eudore/eudore"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// StreamWebsocket 定义gobwas websocket实现eudore.Stream。
type StreamWebsocket struct {
	net.Conn
	streamid  string
	readType  int
	writeType int
	reader    *wsutil.Reader
	writer    *wsutil.Writer
}

// NewExtendFuncStream 函数转换eudore.Stream处理函数，使用gobwas库处理websocket请求。
//
// 默认streamID为"streamid"或"UID"取参数的值。
func NewExtendFuncStream(fn func(eudore.Stream)) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		conn, _, _, err := ws.UpgradeHTTP(ctx.Request(), ctx.Response())
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
		fn(NewStreamWebsocketServer(conn, id))
	}
}

// NewStreamWebsocketServer 函数创建一个服务端StreamWebsocket，默认初始读写类型均为1，即ws.OpText。
func NewStreamWebsocketServer(conn net.Conn, id string) eudore.Stream {
	return &StreamWebsocket{
		Conn:      conn,
		streamid:  id,
		readType:  1,
		writeType: 1,
		reader:    wsutil.NewReader(conn, ws.StateServerSide),
		writer:    wsutil.NewWriter(conn, ws.StateServerSide, 1),
	}
}

// NewStreamWebsocketClient 函数创建一个客户端StreamWebsocket，默认初始读写类型均为1，即ws.OpText。
func NewStreamWebsocketClient(conn net.Conn, id string) eudore.Stream {
	return &StreamWebsocket{
		Conn:      conn,
		streamid:  id,
		readType:  1,
		writeType: 1,
		reader:    wsutil.NewReader(conn, ws.StateClientSide),
		writer:    wsutil.NewWriter(conn, ws.StateClientSide, 1),
	}
}

// StreamID 方法返回流id。
func (stream *StreamWebsocket) StreamID() string {
	return stream.streamid
}

// Read 方法实现io.Reader.
func (stream *StreamWebsocket) Read(b []byte) (int, error) {
	for {
		n, err := stream.reader.Read(b)
		if err == nil || err == io.EOF {
			return n, nil
		}

		if err == wsutil.ErrNoFrameAdvance {
			header, err := stream.reader.NextFrame()
			if err == nil {
				stream.readType = int(header.OpCode)
				continue
			}
		}
		return 0, err
	}
}

// Write 方法实现io.Writer。
func (stream *StreamWebsocket) Write(b []byte) (n int, err error) {
	n, err = stream.writer.Write(b)
	stream.writer.Flush()
	return
}

// SendMsg 方法默认使用json序列化对象并发送。
func (stream *StreamWebsocket) SendMsg(m interface{}) error {
	err := json.NewEncoder(stream.writer).Encode(m)
	if err == nil {
		stream.writer.Flush()
	}
	return err
}

// RecvMsg 方法使用json解码下一帧数据。
func (stream *StreamWebsocket) RecvMsg(m interface{}) error {
	header, err := stream.reader.NextFrame()
	if err != nil {
		return err
	}
	stream.readType = int(header.OpCode)
	return json.NewDecoder(stream.reader).Decode(m)
}

// GetType 方法返回读类型，即请求桢类型。
func (stream *StreamWebsocket) GetType() int {
	return stream.readType
}

// SetType 方法设置写入类型，即设置响应桢类型。
func (stream *StreamWebsocket) SetType(t int) {
	stream.writer.Flush()
	stream.writer.Reset(stream.Conn, ws.StateServerSide, ws.OpCode(t))
	stream.writeType = t
}
