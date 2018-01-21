package regexer

import (
	"io"
	"os"
	"regexp"
	"testing"
)

func Test_Simple(t *testing.T) {
	// Create a new Regexer using the compiled regexp
	r := NewRegexer(regexp.MustCompile(`(\w+)`))

	// Write our sample data to the regexer
	r.Write([]byte("Two Words"))

	// Read the two matches (relies on the regexer having a buffered channel
	t.Log("Match 1:", BytesToString(<-r.C))
	t.Log("Match 2:", BytesToString(<-r.C))

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
			t.Log(BytesToString(m))
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
