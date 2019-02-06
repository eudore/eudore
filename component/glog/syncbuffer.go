package glog

import (
	"bufio"
	"os"
	"time"
	"bytes"
	"fmt"
	"runtime"
)
var MaxSize uint64 = 1024 * 1024 * 1800

// SyncBuffer joins a bufio.Writer to its underlying file, providing access to the
// file's Sync method and providing a wrapper for the Write method that provides log
// file rotation. There are conflicting methods, so the file cannot be embedded.
// l.mu is held for all its methods.

type SyncBuffer struct {
	*bufio.Writer
	file   *os.File
	// sev    Severity
	nbytes uint64 // The number of bytes written to this file
}

func (sb *SyncBuffer) Sync() error {
	return sb.file.Sync()
}

func (sb *SyncBuffer) Write(p []byte) (n int, err error) {
	if sb.nbytes+uint64(len(p)) >= MaxSize {
		if err := sb.RotateFile(time.Now()); err != nil {
			return 0, err
		}
	}
	n, err = sb.Writer.Write(p)
	sb.nbytes += uint64(n)
	return n, err
}

// rotateFile closes the SyncBuffer's file and starts a new one.
func (sb *SyncBuffer) RotateFile(now time.Time) error {
	if sb.file != nil {
		sb.Flush()
		sb.file.Close()
	}
	var err error
	sb.file, err = os.OpenFile("/tmp/access.log",os.O_CREATE|os.O_WRONLY|os.O_APPEND,0666) // create(SeverityName[sb.sev], now)
	sb.nbytes = 0
	if err != nil {
		return err
	}

	sb.Writer = bufio.NewWriterSize(sb.file, bufferSize)

	// Write header.
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Log file created at: %s\n", now.Format("2006/01/02 15:04:05"))
	fmt.Fprintf(&buf, "Binary: Built with %s %s for %s/%s\n", runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&buf, "Log line format: [IWEF]mmdd hh:mm:ss.uuuuuu threadid file:line] msg\n")
	n, err := sb.file.Write(buf.Bytes())
	sb.nbytes += uint64(n)
	return err
}


// bufferSize sizes the buffer associated with each log file. It's large
// so that log records can accumulate without the logging thread blocking
// on disk I/O. The flushDaemon will block instead.
const bufferSize = 256 * 1024
