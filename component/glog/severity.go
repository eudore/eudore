package glog

import (
	"strings"
	"strconv"
	"sync/atomic"
)

type Severity int32 // sync/atomic int32

const SeverityChar = "DIWEF"


const (
	debugLog Severity = iota
	infoLog
	warningLog
	errorLog
	fatalLog
	numSeverity = 5
)

var SeverityName = []string{
	debugLog:   "DEBUG",
	infoLog:    "INFO",
	warningLog: "WARNING",
	errorLog:   "ERROR",
	fatalLog:   "FATAL",
}

// get returns the value of the Severity.
func (s *Severity) get() Severity {
	return Severity(atomic.LoadInt32((*int32)(s)))
}

// set sets the value of the Severity.
func (s *Severity) set(val Severity) {
	atomic.StoreInt32((*int32)(s), int32(val))
}

// String is part of the flag.Value interface.
func (s *Severity) String() string {
	return strconv.FormatInt(int64(*s), 10)
}

// Get is part of the flag.Value interface.
func (s *Severity) Get() interface{} {
	return *s
}

func SeverityByName(s string) (Severity, bool) {
	s = strings.ToUpper(s)
	for i, name := range SeverityName {
		if name == s {
			return Severity(i), true
		}
	}
	return 0, false
}
