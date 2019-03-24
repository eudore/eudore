package fastcgi

import (
	"sync"
	"io"
	"bytes"
	"encoding/binary"
)

// for padding so we don't have to allocate all the time
// not synchronized because we don't care what the contents are
var pad [maxPad]byte


// conn sends records over rwc
type conn struct {
	mutex sync.Mutex
	rwc   io.ReadWriteCloser

	// to avoid allocations
	buf bytes.Buffer
	h   cgiHeader
}

func newConn(rwc io.ReadWriteCloser) *conn {
	return &conn{rwc: rwc}
}

func (c *conn) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.rwc.Close()
}

// writeRecord writes and sends a single record.
func (c *conn) writeRecord(recType recType, reqId uint16, b []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.buf.Reset()
	c.h.init(recType, reqId, len(b))
	if err := binary.Write(&c.buf, binary.BigEndian, c.h); err != nil {
		return err
	}
	if _, err := c.buf.Write(b); err != nil {
		return err
	}
	if _, err := c.buf.Write(pad[:c.h.PaddingLength]); err != nil {
		return err
	}
	_, err := c.rwc.Write(c.buf.Bytes())
	return err
}

func (c *conn) writeEndRequest(reqId uint16, appStatus int, protocolStatus uint8) error {
	b := make([]byte, 8)
	binary.BigEndian.PutUint32(b, uint32(appStatus))
	b[4] = protocolStatus
	return c.writeRecord(typeEndRequest, reqId, b)
}

func (c *conn) writePairs(recType recType, reqId uint16, pairs map[string]string) error {
	w := newWriter(c, recType, reqId)
	b := make([]byte, 8)
	for k, v := range pairs {
		n := encodeSize(b, uint32(len(k)))
		n += encodeSize(b[n:], uint32(len(v)))
		if _, err := w.Write(b[:n]); err != nil {
			return err
		}
		if _, err := w.WriteString(k); err != nil {
			return err
		}
		if _, err := w.WriteString(v); err != nil {
			return err
		}
	}
	w.Close()
	return nil
}

func encodeSize(b []byte, size uint32) int {
	if size > 127 {
		size |= 1 << 31
		binary.BigEndian.PutUint32(b, size)
		return 4
	}
	b[0] = byte(size)
	return 1
}
