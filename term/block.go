package term

import (
	"encoding/base64"
	"fmt"
	"html"
	"image/color"
	"strings"
)

// BlockElement represents a horizontally centered or scrollable block of HTML content.
// There are two types of block elements: one can have auto size (such as an img), and
// the other can not (such as an iframe).
// For elements which can not have auto size, the box element should have a 100% width
// and auto overflow-x. So that it can take the responsibility for horizontal scrolling from the row element.
type BlockElement interface {
	HTML() string // Returns the HTML content of the block
}

type BlockWithOption interface {
	BlockElement
	Options() []BlockOption
}

type BlockOption func(*blockConfig)

type blockConfig struct {
	width      int
	height     int
	background color.Color
	color      color.Color
	opacity    *float64
}

func SizeOption(width, height int) BlockOption {
	return func(c *blockConfig) {
		c.width = width
		c.height = height
	}
}

func BackgroundOption(c color.Color) BlockOption {
	return func(conf *blockConfig) {
		conf.background = c
	}
}

func ColorOption(c color.Color) BlockOption {
	return func(conf *blockConfig) {
		conf.color = c
	}
}

func OpacityOption(x float64) BlockOption {
	return func(conf *blockConfig) {
		conf.opacity = &x
	}
}

func Block(e BlockElement, ops ...BlockOption) {
	BlockSize(e, 0, 0, ops...)
}

func BlockSize(e BlockElement, width, height int, ops ...BlockOption) {
	if block, ok := e.(BlockWithOption); ok {
		// Apply default options for BlockWithOption elements.
		// log.Println(true)
		ops = append(block.Options(), ops...)
	}
	PrintBlockSize(e.HTML(), width, height, ops...)
}

func PrintBlock(html string, ops ...BlockOption) {
	PrintBlockSize(html, 0, 0, ops...)
}

// PrintBlockSize supports HTML Page, Iframe, and other HTML elements.
func PrintBlockSize(html string, width, height int, ops ...BlockOption) {
	var conf blockConfig
	for _, op := range ops {
		op(&conf)
	}

	if width == 0 {
		width = conf.width
	}

	if height == 0 {
		height = conf.height
	}

	// config row style
	var row string
	if conf.background != nil {
		row += fmt.Sprintf("background-color: %s;", colorToCSS(conf.background))
	}

	// config box style
	var css string
	if width > 0 {
		css += fmt.Sprintf("width: %dpx;", width)
	}
	if height > 0 {
		css += fmt.Sprintf("height: %dpx;", height)
	}
	if conf.color != nil {
		css += fmt.Sprintf("color: %s;", colorToCSS(conf.color))
	}
	if conf.opacity != nil {
		css += fmt.Sprintf("opacity: %.2f;", *conf.opacity)
	}

	// prompt page html to iframe
	if strings.HasSuffix(strings.TrimSpace(html), "</html>") {
		html = EscapeIframe(html, "")
	}

	// goterm-box: iframe content should default to 100% width and auto overflow-x
	// TODO: try a more robust way to detect if the content is an iframe
	if strings.HasPrefix(html, "<iframe") {
		css = "width: 100%;" + css
		css += "overflow-x: auto;"
	}
	html = fmt.Sprintf("<div class='goterm-row' style='%s'><div style='%s' class='goterm-box'>%s</div></div>", row, css, html)
	html = strings.ReplaceAll(html, " style=''", "")
	PrintHtml(html)
}

type Image string

func (url Image) HTML() string {
	return fmt.Sprintf(`<img src="%s">`, url)
}

func ImageData(mime string, data []byte) BlockElement {
	encoded := base64.StdEncoding.EncodeToString(data)
	url := fmt.Sprintf("data:%s;base64,%s", mime, encoded)
	return Image(url)
}

// EscapeIframe wraps the given HTML content in an iframe tag and escapes it for srcdoc attribute.
// If the pageHtml starts with "http", it will be used as the source url of the iframe.
func EscapeIframe(pageHtml string, klass string) string {
	var attr, value = "src", pageHtml
	if !strings.HasPrefix(pageHtml, "http") {
		attr, value = "srcdoc", escapeForSrcdoc(pageHtml)
	}
	return fmt.Sprintf(`<iframe class="%s" %s="%s"></iframe>`, klass, attr, value)
}

// escapeForSrcdoc escapes the given content for the srcdoc attribute of an iframe.
func escapeForSrcdoc(content string) string {
	// First, escape HTML special characters
	escapedContent := html.EscapeString(content)

	// Now, escape double quotes for srcdoc attribute
	escapedContent = strings.ReplaceAll(escapedContent, `"`, `&#34;`)

	return escapedContent
}

func colorToCSS(c color.Color) string {
	r, g, b, a := c.RGBA()
	return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", uint8(r>>8), uint8(g>>8), uint8(b>>8), float64(a)/65535.0)
}
