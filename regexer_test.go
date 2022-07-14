package regexer

import (
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
)

func Test_Simple(t *testing.T) {
	// Create a new Regexer using the compiled regexp
	r := NewRegexer(regexp.MustCompile(`(\w+)`))

	// Write our sample data to the regexer
	r.Write([]byte("Two Words"))

	// Read the two matches (relies on the regexer having a buffered channel
	m1 := BytesToString(<-r.C)
	if len(m1) != 2 {
		t.Errorf("expected two result got %d", len((m1)))
	}
	t.Logf("Match 1 (%d): %s", len(m1), m1)
	m2 := BytesToString(<-r.C)
	t.Logf("Match 2 (%d): %s", len(m2), m2)

	r.Close()
}

func Test_Simple2(t *testing.T) {
	// Create a new Regexer using the compiled regexp
	r := NewRegexer(regexp.MustCompile(`(\w+) (\d+)`))

	// Create a channel to signal finished
	finished := make(chan struct{}, 1)

	// Create a background process to receive matches
	go func() {
		counter := 0
		// The loop runs until the regexer is closed
		for m := range r.C {
			counter++
			t.Logf("Match: %#v", BytesToString(m))
		}
		if counter != 3 {
			t.Errorf("Expected 3 matches, got %d", counter)
		}
		close(finished)
	}()

	// Write our sample data to the regexer
	r.Write([]byte("This shouldn't match"))
	r.Write([]byte(" Match 1"))
	r.Write([]byte(" Match 2"))
	r.Write([]byte(" Match")) // Break match across writes
	r.Write([]byte(" 3"))
	r.Close()

	// Wait for background process to finish
	<-finished
}

func Test_File(t *testing.T) {
	r := NewRegexer(regexp.MustCompile(`(\w+)`))

	// Create a channel to signal finished
	finished := make(chan struct{}, 1)

	go func() { // Set up process to listen for matches
		counter := 0
		for range r.C {
			counter++
		}
		if counter != 135 {
			t.Errorf("Expected 135 matches/words, got %d", counter)
		}
		close(finished)
	}()

	// Open file
	f, err := os.Open("words.txt")
	if err != nil {
		t.Fatal("Unable to open words.txt:", err)
	}
	defer f.Close()

	_, err = io.Copy(r, f) // Copy the file into the Regexer which will send matches to the channel
	if err != nil {
		t.Error(err)
	}
	r.Close() // Close the regexer C channel

	// Wait for background process to finish
	<-finished
}

func LargeDataTest(l testing.TB, num_tests int) {
	l.Logf("Write 2,010 bytes %d times, a total of %.3f Mb", num_tests, float32(num_tests*2010)/1024/1024)
	r := NewRegexer(regexp.MustCompile(`(1)`))

	// Create a channel to signal finished
	finished := make(chan struct{}, 1)

	go func() { // Set up process to listen for matches
		bufTot := 0
		counter := 0
		bufMax := 0
		for range r.C {
			counter++
			bufLen := len(r.rxBuf)
			bufTot += bufLen
			if bufLen > bufMax {
				bufMax = bufLen
			}
			if len(r.rxBuf) > MAX_BUFFER_SIZE*10 {
				l.Errorf("Overrun rxBuf at %d bytes", len(r.rxBuf))
			}
		}
		l.Logf("Average rxBuf len was %d, max was %d after %d writes", bufTot/counter, bufMax, counter)
		if counter != num_tests {
			l.Errorf("Expected %d matches/words, got %d", num_tests, counter)
		} else {
			l.Logf("Expected %d matches/words, got %d :)", num_tests, counter)
		}
		close(finished)
	}()

	for i := 0; i < num_tests; i++ {
		_, err := r.Write([]byte(strings.Repeat("TEST TEST ", 1000))) // 1000 bytes
		if err != nil {
			l.Error(err)
		}
		_, err = r.Write([]byte{0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39}) // 0-9 - 10 bytes
		if err != nil {
			l.Error(err)
		}
		_, err = r.Write([]byte(strings.Repeat("TEST TEST ", 1000))) // 1000 bytes
		if err != nil {
			l.Error(err)
		}
	}

	r.Close() // Close the regexer C channel

	// Wait for background process to finish
	<-finished
}

func Test_LargeData(t *testing.T) {
	LargeDataTest(t, 10000)
}

func BenchmarkLargeData(b *testing.B) {
	LargeDataTest(b, b.N)
}

func Test_NullSubGroups(t *testing.T) {
	// Create a new Regexer using the compiled regexp
	r := NewRegexer(regexp.MustCompile(`(aaa)(?:(.*))(bbb)`))

	// Write our sample data to the regexer
	r.Write([]byte("aaabbb"))

	m := BytesToString(<-r.C)

	if len(m) != 4 {
		t.Errorf("expected 4 elements, got %d", len(m))
	}

	r.Close()
}
