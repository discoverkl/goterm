package term

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestOpenInCustomFormat(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Normal text should be wrapped in a pre tag.
		{"", preText("")},
		{"\n", preText("\n")},
		{"hi", preText("hi")},
		{"a\nb", preText("a\nb")},

		// Even if the text is HTML, it should be wrapped in a pre tag.
		{"<span>hi</span>", preText("<span>hi</span>")},

		// Escaped HTML should be output as is.
		{escapeHtml("<span>hi</span>"), "<span>hi</span>\n"},
		{escapeHtml("<image src=\"hi\" />"), "<image src=\"hi\" />\n"},

		// The HTML tag must be followed by a newline to be valid.
		// Print* functions always add a newline.
		{escapeHtml("<span>hi</span>") + "\nhi", "<span>hi</span>\n" + preText("hi")},
	}

	for _, test := range tests {
		Open(Format(Custom))
		fmt.Println(test.input)
		Close()

		a := slices.Collect(HTML(false))
		got := strings.Join(a, "")

		if got != test.want {
			t.Errorf("got %q, want %q", got, test.want)
		}
	}
}

func TestOpenMultipleTimes(t *testing.T) {
	// Open should panic if the terminal is already open.
	Open(Format(Custom))
	defer Close()
	assertPanic(t, func() {
		Open(Format(Custom))
	})
}

func TestOpenBigBlock(t *testing.T) {
	var overhead int
	for _, size := range []int{1, 2, 10, 1024, 10 * 1024 * 1024} { // max 10MB (the actual limit is 1GB)
		Open(Format(Custom))
		time.Sleep(time.Millisecond * 100)
		var bigText = strings.Repeat("a", size)
		fmt.Println(bigText)
		Close()

		var count int
		for s := range HTML(false) {
			count += len(s)
		}

		// The overhead should be constant for different input sizes.
		if overhead == 0 {
			overhead = count - len(bigText)
		} else {
			if overhead != count-len(bigText) {
				t.Errorf("Overhead changed from %d to %d when input size it %d", overhead, count-len(bigText), len(bigText))
			}
		}
		log.Printf("Input size: %d, output size: %d", len(bigText), count)
	}
}

func TestClose(t *testing.T) {
	// Close should panic if the terminal is already closed.
	Open(Format(Custom))
	Close()
	assertPanic(t, Close)
}

func TestHTML(t *testing.T) {
	// Mock the openInBrowser function to avoid opening the browser but still curl the url.
	holdOpen := openInBrower
	openInBrower = mockOpenInBrowser
	defer func() { openInBrower = holdOpen }()

	// Mock the printToStdout function to avoid a lot of print to the terminal.
	holdPrint := printToStdout
	printToStdout = func(string) {}
	defer func() { printToStdout = holdPrint }()

	const text = "<span>hi</span>"

	// Only Custom format should allow to call the HTML() method.
	for _, format := range []OutputFormat{
		HTMLWindow,
		HTMLPage,
		HTMLContent,
		Raw,
		Custom,
	} {
		Open(Format(format))
		PrintHtml(text)
		Close()

		var foo = func() {
			a := slices.Collect(HTML(false))
			got := strings.Join(a, "")
			if got != text+"\n" {
				t.Errorf("got %q, want %q", got, text)
			}
		}
		if format == Custom {
			foo()
		} else {
			assertPanic(t, foo)
		}

	}
}

func mockOpenInBrowser(url string) error {
	// get the url using http.Get
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// Utility function to test if a function panics
func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

func preText(s string) string {
	return fmt.Sprintf("<pre class=\"goterm\">\n%s\n</pre>\n", s)
}
