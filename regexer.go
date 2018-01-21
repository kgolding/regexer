/*
	io.Writer that output matches based on the given regex

*/
package regexer

import (
	"regexp"
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
		C:     make(chan [][]byte, 5),
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
	for _, m := range matches {
		ms := make([][]byte, len(m))
		for i := 0; i < len(m)-1; i += 2 {
			ms[i] = r.rxBuf[m[i]:m[i+1]]
		}
		r.C <- ms
		if lastByteUsed < m[1] {
			lastByteUsed = m[1]
		}
	}
	r.rxBuf = r.rxBuf[lastByteUsed:]
	// Purge old data
	if len(r.rxBuf) > MAX_BUFFER_SIZE {
		r.rxBuf = r.rxBuf[len(r.rxBuf)-MAX_BUFFER_SIZE:]
	}
	return len(b), nil
}
