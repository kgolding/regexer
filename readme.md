# Regexer

Simple go package to return regexp matches on a channel using an io.Writer interface.

* Handles buffering of incoming data, and destruction of unmatched data
* io.Writer interface
* Returns [][]byte or []string using BytesToString helper

## Usage

Create a regexer using a compiled regexp

    r := NewRegexer(regexp.MustCompile(`(\w+)`))

Get results by reading the C channel, relying on the channel closing to exit the loop, and using the BytesToString helper to convert the results to string from the default []byte

	go func() {
		for m := range r.C {
			t.Log(BytesToString(m))
		}
	}()

Send data to the regxer using the io.Writer Write method

	r.Write([]byte("This shouldn't match"))

Or for example from a file

	f, err := os.Open("words.txt")
	if err != nil {
		t.Fatal("Unable to open words.txt:", err)
	}
	defer f.Close()
	_, err = io.Copy(r, f)

See the regexer_test.go file for full examples
