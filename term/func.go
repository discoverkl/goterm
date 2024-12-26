package term

import (
	"iter"
	"os"
)

var sysStdout = os.Stdout
var sysStderr = os.Stderr

var term = NewTerm()

// Open opens the terminal. This function should be called at the beginning of the program.
func Open(options ...TermOption) {
	if term.closed {
		term = NewTerm()
	}
	term.Open(options...)
}

// Close closes the terminal. This function should be called at the end of the program.
func Close() {
	term.Close()
}

// HTML returns a sequence of strings for the HTML content.
// If page is true, the HTML content is a full page. Otherwise, it is a fragment.
// One should only call this function when the format option is set to Custom.
func HTML(page bool) iter.Seq[string] {
	return term.HTML(page)
}
