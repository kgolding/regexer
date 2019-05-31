/*
	io.Writer that output matches based on the given regex

*/
package regexer

import (
	"errors"
	"regexp"
	"time"
)

// Make buffer size in memory (4K)
const MAX_BUFFER_SIZE = 1024 * 4

type Regexer struct {
	rxBuf []byte // data buffer
	regex *regexp.Regexp
	C     chan [][]byte // Returns regexp [][]byte matches, one at a time
}

// Convert a regex match [][]byte to [][]string
func BytesToString(m [][]byte) []string {
	s := make([]string, len(m))
	for i, b := range m {
		s[i] = string(b)
	}
	return s
}

// Returns a new *Regexer
func NewRegexer(regex *regexp.Regexp) *Regexer {
	return &Regexer{
		rxBuf: make([]byte, 0),
		regex: regex,
		C:     make(chan [][]byte, 10),
	}
}

// Closes the *Regexer's channel
func (r *Regexer) Close() {
	close(r.C)
}

// Write data and attempt to match, adding to the internal buffer
func (r *Regexer) Write(b []byte) (int, error) {
	r.rxBuf = append(r.rxBuf, b...)

	matches := r.regex.FindAllSubmatchIndex(r.rxBuf, -1)
	lastByteUsed := 0
	var err error
	for _, match := range matches {
		retLen := len(match)
		ret := make([][]byte, retLen/2)
		for i := 0; i < retLen-1; i += 2 {
			ret[i/2] = r.rxBuf[match[i]:match[i+1]]
		}
		select { // Do not block if chan not being emptied
		case r.C <- ret:
		case <-time.After(time.Second * 10):
			err = errors.New("match channel blocked")
		}
		if lastByteUsed < match[1] {
			lastByteUsed = match[1]
		}
	}
	r.rxBuf = r.rxBuf[lastByteUsed:]
	// Purge old data
	if len(r.rxBuf) > MAX_BUFFER_SIZE {
		r.rxBuf = r.rxBuf[len(r.rxBuf)-MAX_BUFFER_SIZE:]
	}
	return len(b), err
}
