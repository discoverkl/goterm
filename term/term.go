// The term package provides a way to capture stdout and stderr and display the output in a browser.
// The package also provides a way to display charts in the browser.
// If you want to display charts, you can use the DataFrame type from the df package.
// Or you can leverage the Chart interface to display any chart in the go-echarts library.
//
// The Escape* functions are used to wrap the content in HTML tags so that it can be displayed in a browser.
package term

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

const (
	// HtmlTag is a special tag used to wrap HTML content in the buffer.
	// None html content will be wrapped in <pre> tag.
	HtmlTag       = "==========76ADCBF0-980B-4C05-951F-63340F35E9C=========="
	MaxBuffersize = 1024 * 1024 * 1024 // 1GB
)

// threadSafeWriter wraps io.Writer with a mutex for thread-safe writing
type threadSafeWriter struct {
	w  io.Writer
	mu *sync.Mutex
}

func NewThreadSafeWriter(w io.Writer) *threadSafeWriter {
	return &threadSafeWriter{
		w:  w,
		mu: &sync.Mutex{},
	}
}

func (t *threadSafeWriter) Write(p []byte) (n int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.w.Write(p)
}

// Term captures stdout and stderr and provides methods to display the output in a browser.
type Term struct {
	// Buffer to store the output
	buf *Buffer

	// Cache to store the output for reuse in the web server
	cache   bytes.Buffer
	cacheMu sync.Mutex

	// Pipes for attaching to stdout and stderr
	stdoutWriter *os.File
	stderrWriter *os.File
	// oldStdout    *os.File
	// oldStderr    *os.File

	// WaitGroups for channel writers and readers
	chWriterWg sync.WaitGroup
	chReaderWg sync.WaitGroup

	// Internal logger which writes to stderr
	logger *log.Logger
	opened bool
	closed bool

	// Options
	format       OutputFormat
	port         int
	attachOutput bool
	cacheOutput  bool
}

func (t *Term) Open(options ...TermOption) {
	if t.opened {
		panic("terminal is already opened")
	}
	t.opened = true

	// Apply options
	for _, option := range options {
		option(t)
	}

	// Save the original stdout and stderr
	// t.oldStdout = os.Stdout
	// t.oldStderr = os.Stderr

	// Create pipes for stdout and stderr
	stdoutReader, stdoutWriter, _ := os.Pipe()
	stderrReader, stderrWriter, _ := os.Pipe()
	t.stdoutWriter = stdoutWriter
	t.stderrWriter = stderrWriter

	// var err error
	// err = syscall.SetNonblock(int(stdoutWriter.Fd()), true)
	// if err != nil {
	// 	log.Println(fmt.Errorf("set none block failed: %w", err))
	// }

	// Redirect stdout and stderr to the pipes
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	// Set logger output to the buffer
	log.SetOutput(os.Stderr)

	// Start goroutines to copy the pipe contents to the buffer and original stdout/stderr
	t.chWriterWg.Add(1)
	go func() {
		defer t.chWriterWg.Done()

		defer stdoutReader.Close()
		var err error
		if t.format == Raw {
			_, err = io.Copy(io.MultiWriter(t.buf, sysStdout), stdoutReader)
		} else {
			_, err = io.Copy(t.buf, stdoutReader)
		}
		if err != nil {
			log.Printf("stdout copy error: %v", err)
		}
	}()

	t.chWriterWg.Add(1)
	go func() {
		defer t.chWriterWg.Done()

		defer stderrReader.Close()
		var err error
		if t.format == Raw {
			_, err = io.Copy(io.MultiWriter(t.buf, sysStderr), stderrReader)
		} else {
			_, err = io.Copy(t.buf, stderrReader)
		}
		if err != nil {
			log.Printf("stderr copy error: %v", err)
		}
	}()

	// Start a goroutine to read the buffer
	t.chReaderWg.Add(1)
	go func() {
		defer t.chReaderWg.Done()

		switch t.format {
		case HTMLWindow:
			t.serveHtmlContent(true, true, 0)
		case HTMLPage:
			for html := range t.internalHTML(true) {
				printToStdout(html)
			}
		case HTMLContent:
			for html := range t.internalHTML(false) {
				printToStdout(html)
			}
		case Raw:
			for range t.internalHTML(false) {
				// read and discard the output
			}
		case Custom:
			if t.port > 0 {
				// start a web server to serve the terminal output
				t.serveHtmlContent(false, false, t.port)
			} else {
				// do nothing here, assuming the user will call HTML() to get the content
			}
		default:
			panic("unknown output format")
		}
	}()
}

// Close stops capturing stdout and stderr and restores the original stdout and stderr.
func (t *Term) Close() {
	// Restore stdout and stderr
	os.Stdout = sysStdout
	os.Stderr = sysStderr
	log.SetOutput(sysStderr)

	// Close writers to stop the goroutines
	t.stdoutWriter.Close()
	t.stderrWriter.Close()

	// Wait for channel writers
	t.chWriterWg.Wait()

	// Close the channel
	t.buf.Close()

	// Wait for channel readers, including the web server and the iterator which the HTML() method returns
	t.chReaderWg.Wait()

	t.closed = true
}

// HTML returns a sequence of strings that represent the terminal output in HTML format.
// If fullPage is true, the output will be wrapped in a full HTML page with styles.
// Otherwise, the output will be some HTML content that can be embedded in a page.
func (t *Term) HTML(fullPage bool) iter.Seq[string] {
	if t.format != Custom {
		panic("format must be CustomFormat when calling HTML()")
	}
	return t.internalHTML(fullPage)
}

func (t *Term) internalHTML(fullPage bool) iter.Seq[string] {
	return func(yield func(s string) bool) {
		t.chReaderWg.Add(1)
		defer t.chReaderWg.Done()

		// Write html page prefix
		if fullPage {
			if !yield(t.getHtmlPagePrefix()) {
				return
			}
		}

		var sc *bufio.Scanner
		inHtml := false
		isFirstTextLine := true

		// convert text line to html
		var convertLine = func(line string) bool {
			// If the line is a tag line, discard it and toggle inHtml
			if strings.HasSuffix(line, HtmlTag) {
				if !inHtml && !isFirstTextLine {
					if !yield("</pre>\n") {
						return false
					}
				}
				inHtml = !inHtml
				isFirstTextLine = true
				return true // always skip the tag line
			}

			// If the line is html content, yield it directly and return
			if inHtml {
				return yield(line + "\n")
			}

			// Otherwise, wrap the line in a pre tag
			if isFirstTextLine {
				isFirstTextLine = false
				if !yield("<pre class=\"goterm\">\n") {
					return false
				}
			}
			if !yield(line + "\n") {
				return false
			}
			return true
		}

		// Read cached output first
		if t.cacheOutput {
			t.cacheMu.Lock()
			var savedStr = t.cache.String()
			t.cacheMu.Unlock()

			oldBuf := bytes.NewBufferString(savedStr)
			sc = bufio.NewScanner(oldBuf)
			sc.Buffer(nil, MaxBuffersize)
			for sc.Scan() {
				line := sc.Text()
				if !convertLine(line) {
					return
				}
			}
		}

		// Read the buffer line by line
		sc = bufio.NewScanner(t.buf)
		sc.Buffer(nil, MaxBuffersize)
		for sc.Scan() {
			line := sc.Text()

			// Update the cache
			if t.cacheOutput {
				t.cacheMu.Lock()
				t.cache.WriteString(line + "\n")
				t.cacheMu.Unlock()
			}

			if !convertLine(line) {
				return
			}
		}

		// Reaching the end of the buffer, close the pre tag if needed
		if !inHtml && !isFirstTextLine {
			if !yield("</pre>\n") {
				return
			}
		}

		// Write html page suffix
		if fullPage {
			if !yield(t.getHtmlPageSuffix()) {
				return
			}
		}
	}
}

func (t *Term) getHtmlPagePrefix() string {
	var buf bytes.Buffer

	// write html head
	buf.WriteString("<!DOCTYPE html>\n")
	buf.WriteString("<html>\n")
	buf.WriteString("<head>\n")
	buf.WriteString("<title>Term</title>\n")
	buf.WriteString("</head>\n")
	buf.WriteString("<body>\n")

	// write css style
	buf.WriteString("<style>\n")
	buf.WriteString(BodyStyle)
	buf.WriteString(IframeStyle)
	buf.WriteString(BlockStyle)
	buf.WriteString(TextStyle)
	buf.WriteString("</style>\n")

	// write script
	buf.WriteString(ScrollScript)
	return buf.String()
}

func (t *Term) getHtmlPageSuffix() string {
	var buf bytes.Buffer
	buf.WriteString("</body>\n")
	buf.WriteString("</html>\n")
	return buf.String()
}

func (t *Term) serveHtmlContent(local bool, serveOnce bool, port int) error {
	var err error

	// This WaitGroup is used only when serveOnce is true, otherwise the server will run indefinitely
	var doneCh = make(chan any)
	var doneOnce sync.Once

	// Serve the HTML content
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// The Close() method will wait for this WaitGroup to finish
		t.chReaderWg.Add(1)
		defer t.chReaderWg.Done()

		// Get a Flusher to flush the response
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		// Set the Content-Type header so that the browser can render the HTML content immediately
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		for html := range t.internalHTML(true) {
			// If client has disconnected, stop iterating and return
			if r.Context().Err() != nil {
				return
			}

			// Flush some html content to the client
			fmt.Fprint(w, html)
			flusher.Flush()
		}

		// One-time server will close the connection after serving the HTML content
		if serveOnce {
			doneOnce.Do(func() {
				close(doneCh)
			})
		}
	})

	// Get host based on the local flag
	host := "localhost"
	if !local {
		host = "0.0.0.0"
	}

	// Get port based on the given port or random port
	var listener net.Listener
	if port <= 0 {
		// Listen on a random port
		addr := fmt.Sprintf("%s:0", host)
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			return err
		}

		// Extract port from listener's address
		port = listener.Addr().(*net.TCPAddr).Port
	} else {
		// Listen on the given port
		addr := fmt.Sprintf("%s:%d", host, port)
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			return err
		}
	}

	// Create an HTTP server
	server := &http.Server{}

	// Start the HTTP server in a separate goroutine so that we can close it later using server.Shutdown()
	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			t.logger.Printf("HTTP server ListenAndServe failed: %v", err)
		}
	}()

	// Construct the URL based on the host and port
	url := fmt.Sprintf("http://localhost:%d", port)
	if port == 80 {
		// remove the port if it is 80
		url = "http://localhost"
	}

	// Open or print the URL based on the local flag
	if local {
		// Open the URL in the default browser
		err = openInBrower(url)
		if err != nil {
			return fmt.Errorf("openURL: %w", err)
		}
	} else {
		// Print the URL to the console
		t.logger.Printf("Serving HTML content at: %s", url)
	}

	if serveOnce {
		// Keep the program running until the HTML content is served
		<-doneCh
		server.Shutdown(context.Background())
		return nil
	}

	// Hanging here so that the Close() method can wait for the server to finish
	select {}
}

// NewTerm creates a new Term and copies stdout and stderr to a internal buffer.
// The output can be displayed in a browser when you use the Open method with the default HTMLWindow format.
// See the Format options for other ways to display the output.
func NewTerm() *Term {
	term := &Term{
		buf:    NewBuffer(),
		logger: log.New(sysStderr, "", log.LstdFlags),
	}
	return term
}

// printToStdout uses var declaration to make it possible to override this function in tests.
var printToStdout = func(s string) {
	fmt.Fprint(sysStdout, s)
}

// openInBrower opens the given URL in the default browser.
// It uses var declaration to make it possible to override this function in tests.
// TODO: test this function on different platforms
var openInBrower = func(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // Linux, FreeBSD, OpenBSD, NetBSD
		cmd = "xdg-open"
	}

	if runtime.GOOS == "windows" {
		// On Windows, we need to add an empty string to prevent issues with URLs starting with a quote
		args = append(args, "", url)
	} else {
		args = append(args, url)
	}

	return exec.Command(cmd, args...).Start()
}

// escapeHtml wraps the given HTML content in a special html tag.
// Remember to add a newline after the tag to make it valid.
func escapeHtml(html string) string {
	return fmt.Sprintf(`%s
%s
%s`, HtmlTag, html, HtmlTag)
}

// PrintHtml prints the given HTML content to the terminal.
func PrintHtml(html string) {
	s := escapeHtml(html)
	fmt.Println(s)
}
