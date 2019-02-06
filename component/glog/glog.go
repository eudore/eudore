package glog

import (
	"fmt"
	"os"
	"io"
	"time"
	"sync"
	"runtime"
	"strings"
	"eudore"
)


var timeNow	= time.Now // Stubbed out for testing.
var pid		= os.Getpid()


func init() {
	eudore.RegisterComponent("logger-glog", func(...interface{}) eudore.Component {
		return New()
	})
}

var _ eudore.Logger = &LoggingT{}

func TimeoutFlush(timeout time.Duration) {
	done := make(chan bool, 1)
	go func() {
		// Flush() // calls logging.lockAndFlushAll()
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		fmt.Fprintln(os.Stderr, "glog: Flush took longer than", timeout)
	}
}

// stacks is a wrapper for runtime.Stack that attempts to recover the data for all goroutines.
func stacks(all bool) []byte {
	// We don't know how big the traces are, so grow a few times if they don't fit. Start large, though.
	n := 10000
	if all {
		n = 100000
	}
	var trace []byte
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, all)
		if nbytes < len(trace) {
			return trace[:nbytes]
		}
		n *= 2
	}
	return trace
}

// LoggingT collects all the global state of the logging setup.
type LoggingT struct {
	eudore.ComponentName
	toStdout     bool // The -logtostderr flag.
	// alsoToStderr bool // The -alsologtostderr flag.
	level		int
	depth		int
	
	// freeList is a list of byte buffers, maintained under freeListMu.
	freeList *buffer
	// freeListMu maintains the free list. It is separate from the main mutex
	// so buffers can be grabbed and printed to without holding the main lock,
	// for better parallelization.
	freeListMu sync.Mutex

	// mu protects the remaining elements of this structure and is
	// used to synchronize logging.
	mu sync.Mutex
	// file holds writer for each of the log types.
	Out		FlushSyncWriter
	Err		func(error)
	// vmap is a cache of the V Level for each V() call site, identified by PC.
	// It is wiped whenever the vmodule flag changes state.
	// vmap map[uintptr]Level
	// filterLength stores the length of the vmodule filter chain. If greater
	// than zero, it means vmodule is enabled. It may be read safely
	// using sync.LoadInt32, but is only modified under mu.
	filterLength int32
	// traceLocation is the state of the -log_backtrace_at flag.
	traceLocation traceLocation
	// These flags are modified only under lock, although verbosity may be fetched
	// safely using atomic.LoadInt32.
	// vmodule   moduleSpec // The state of the -vmodule flag.
	// verbosity Level      // V logging level, the value of the -v flag/
}

type FlushSyncWriter interface {
	io.Writer
	Flush() error
	Sync() error
	RotateFile(time.Time) error 
}


func New() *LoggingT {
	l := &LoggingT{
		Out:	&SyncBuffer{},
		toStdout:	true,
		level:	5,
		depth:	1,
	}
	l.Out.RotateFile(timeNow())
	go l.FlushDaemon()
	return l
}

func (l *LoggingT) Version() string {
	return "eudore logger glog"
}


func (l *LoggingT) Debug(args ...interface{}){
	l.PrintDepth(0, l.depth, args...)
}

func (l *LoggingT) Info(args ...interface{}){
	l.PrintDepth(1, l.depth, args...)

}

func (l *LoggingT) Warning(args ...interface{}){
	l.PrintDepth(2, l.depth, args...)

}

func (l *LoggingT) Error(args ...interface{}){
	l.PrintDepth(3, l.depth, args...)

}

func (l *LoggingT) Fatal(args ...interface{}){
	l.PrintDepth(4, l.depth, args...)

}


// getBuffer returns a new, ready-to-use buffer.
func (l *LoggingT) GetBuffer() *buffer {
	l.freeListMu.Lock()
	b := l.freeList
	if b != nil {
		l.freeList = b.next
	}
	l.freeListMu.Unlock()
	if b == nil {
		b = new(buffer)
	} else {
		b.next = nil
		b.Reset()
	}
	return b
}

// putBuffer returns a buffer to the free list.
func (l *LoggingT) PutBuffer(b *buffer) {
	if b.Len() >= 256 {
		// Let big buffers die a natural death.
		return
	}
	l.freeListMu.Lock()
	b.next = l.freeList
	l.freeList = b
	l.freeListMu.Unlock()
}

/*
header formats a log header as defined by the C++ implementation.
It returns a buffer containing the formatted header and the user's file and line number.
The depth specifies how many stack frames above lives the source line to be identified in the log message.

Log lines have this form:
	Lmmdd hh:mm:ss.uuuuuu threadid file:line] msg...
where the fields are defined as follows:
	L                A single character, representing the log level (eg 'I' for INFO)
	mm               The month (zero padded; ie May is '05')
	dd               The day (zero padded)
	hh:mm:ss.uuuuuu  Time in hours, minutes and fractional seconds
	threadid         The space-padded thread ID as returned by GetTID()
	file             The file name
	line             The line number
	msg              The user-supplied message
*/
func (l *LoggingT) Header(s int, depth int) (*buffer, string, int) {
	_, file, line, ok := runtime.Caller(3 + depth)
	if !ok {
		file = "???"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return l.FormatHeader(s, file, line), file, line
}

// formatHeader formats a log header using the provided file name and line number.
func (l *LoggingT) FormatHeader(s int, file string, line int) *buffer {
	now := timeNow()
	if line < 0 {
		line = 0 // not a real line number, but acceptable to someDigits
	}
	// if s > fatalLog {
	// 	s = infoLog // for safety.
	// }
	buf := l.GetBuffer()

	// Avoid Fprintf, for speed. The format is so simple that we can do it quickly by hand.
	// It's worth about 3X. Fprintf is hard.
	_, month, day := now.Date()
	hour, minute, second := now.Clock()
	// Lmmdd hh:mm:ss.uuuuuu threadid file:line]
	buf.tmp[0] = SeverityChar[Severity(s)]
	fmt.Println(month,int(month),day,hour)
	buf.twoDigits(1, int(month))
	buf.twoDigits(3, day)
	buf.tmp[5] = ' '
	buf.twoDigits(6, hour)
	buf.tmp[8] = ':'
	buf.twoDigits(9, minute)
	buf.tmp[11] = ':'
	buf.twoDigits(12, second)
	buf.tmp[14] = '.'
	buf.nDigits(6, 15, now.Nanosecond()/1000, '0')
	buf.tmp[21] = ' '
	buf.nDigits(7, 22, pid, ' ')
	buf.tmp[29] = ' '
	buf.Write(buf.tmp[:30])
	buf.WriteString(file)
	buf.tmp[0] = ':'
	n := buf.someDigits(1, line)
	// buf.tmp[n+1] = ']'
	buf.tmp[n+2] = ' '
	buf.Write(buf.tmp[:n+3])
	fmt.Println(buf)
	return buf
}

func (l *LoggingT) Println(s int, args ...interface{}) {
	buf, file, line := l.Header(s, 0)
	fmt.Fprintln(buf, args...)
	l.Output(s, buf, file, line)
}

func (l *LoggingT) Print(s int, args ...interface{}) {
	l.PrintDepth(s, 1, args...)
}

func (l *LoggingT) PrintDepth(s int, depth int, args ...interface{}) {
	buf, file, line := l.Header(s, depth)
	fmt.Fprint(buf, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.Output(s, buf, file, line)
}

func (l *LoggingT) Printf(s int, format string, args ...interface{}) {
	buf, file, line := l.Header(s, 0)
	fmt.Fprintf(buf, format, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.Output(s, buf, file, line)
}

// printWithFileLine behaves like print but uses the provided file and line number.  If
// alsoLogToStderr is true, the log message always appears on standard error; it
// will also appear in the log file unless --logtostderr is set.
func (l *LoggingT) PrintWithFileLine(s int, file string, line int, args ...interface{}) {
	buf := l.FormatHeader(s, file, line)
	fmt.Fprint(buf, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.Output(s, buf, file, line)
}

// output writes the data to the log files and releases the buffer.
func (l *LoggingT) Output(s int, buf *buffer, file string, line int) {
	l.mu.Lock()
	if l.traceLocation.isSet() && l.traceLocation.match(file, line) {
		buf.Write(stacks(false))
	}
	data := buf.Bytes()
	// if !flag.Parsed() {
	// 	os.Stderr.Write([]byte("ERROR: logging before flag.Parse: "))
	// 	os.Stderr.Write(data)
	// } else
	if l.toStdout {
		os.Stdout.Write(data)
	} 
	//	else {
	// 	if alsoToStderr { // || s >= l.stderrThreshold.get() {
	// 		os.Stderr.Write(data)
	// 	}
		l.Out.Write(data)
		// if l.file[s] == nil {
		// 	if err := l.createFiles(s); err != nil {
		// 		os.Stderr.Write(data) // Make sure the message appears somewhere.
		// 		l.exit(err)
		// 	}
		// }
	//}
	l.PutBuffer(buf)
	l.mu.Unlock()
	// if stats := intStats[s]; stats != nil {
	// 	atomic.AddInt64(&stats.lines, 1)
	// 	atomic.AddInt64(&stats.bytes, int64(len(data)))
	// }
}


// logExitFunc provides a simple mechanism to override the default behavior
// of exiting on error. Used in testing and to guarantee we reach a required exit
// for fatal logs. Instead, exit could be a function rather than a method but that
// would make its use clumsier.
var logExitFunc func(error)

// exit is called if there is trouble creating or writing log files.
// It flushes the logs and exits the program; there's no point in hanging around.
// l.mu is held.
func (l *LoggingT) exit(err error) {
	fmt.Fprintf(os.Stderr, "log: exiting because of error: %s\n", err)
	// If logExitFunc is set, we do that instead of exiting.
	if l.Err != nil {
		l.Err(err)
		return
	}
	l.Flush()
	os.Exit(2)
}


const flushInterval = 1 * time.Second

// flushDaemon periodically flushes the log file buffers.
func (l *LoggingT) FlushDaemon() {
	for _ = range time.NewTicker(flushInterval).C {
		l.Flush()
	}
}

// lockAndFlushAll is like flushAll but locks l.mu first.
// flushAll flushes all the logs and attempts to "sync" their data to disk.
func (l *LoggingT) Flush() {
	l.mu.Lock()
	l.Out.Flush()
	l.Out.Sync()
	l.mu.Unlock()
}
