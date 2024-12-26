package term

type OutputFormat int

const (
	HTMLWindow  OutputFormat = iota // Open in browser
	HTMLPage                        // Print HTML page
	HTMLContent                     // Print HTML content
	Raw                             // Print raw text, useful for debugging
	Custom                          // Print nothing, user is expected to call the HTML function
)

type TermOption func(*Term)

// Detach from the stdout/stderr of the current process.
func Detach() func(t *Term) {
	return func(t *Term) {
		t.attachOutput = false
	}
}

// Format sets the format of the terminal output.
// The default is FormatRaw.
func Format(format OutputFormat) func(t *Term) {
	return func(t *Term) {
		t.format = format
	}
}

// BindPort will start a web server to serve the terminal output on the specified port.
func BindPort(port int) func(t *Term) {
	return func(t *Term) {
		t.format = Custom
		t.port = port
		t.cacheOutput = true
	}
}
