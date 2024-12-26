# Goterm

Goterm is a library which can render charts, images and other HTML based contents fairly easily from command line.

You can use `fmt.Printf` normally and all the contents will be printed on an html page in realtime. It also provides functions like `term.Block` to output varies kinds of web based contents.

It has built-in support for `go-echarts` and `gogum/plot` and can be extended to support any other web based rendering technologies.

# Basic Usage

## Chart

### DataFrame

The easist way to plot charts is to use a DataFrame object.

```go
package main

import (
	"fmt"
	"github.com/discoverkl/goterm/term"
	"github.com/discoverkl/goterm/df"
)

func main() {
	term.Open()
	defer term.Close()

	d := df.NewDataFrame()

	d.SetColumn(df.NewStringSeries("X Name", 10))
	d.SetColumn(df.NewRandomIntSeries("Series 1", 10, 300))
	d.SetColumn(df.NewRandomIntSeries("Series 2", 10, 300))
	d.SetColumn(df.NewRandomIntSeries("Series 3", 10, 300))

	fmt.Println(d)
	d.Pie(df.Name("Pie Chart"))
	d.Bar(df.Name("Bar Chart"), df.YName("Y Name"))
	d.Line(df.Name("Line Chart"), df.YName("Y Name"))
}
```

A `Series` is just a named array, you can create it using `df.NewSeries`.

```go
package main

import (
	"fmt"
	"github.com/discoverkl/goterm/term"
	"github.com/discoverkl/goterm/df"
)

func main() {
	term.Open()
	defer term.Close()

	s1 := df.NewSeries("name", []string{"A", "B", "C", "D", "E"})
	s2 := df.NewSeries("value", []float64{1, 3, 5, 6, 8})

	d := df.NewDataFrame(s1, s2)
	fmt.Println(d)

	d.Pie()
}

```

## Images

Goterm supports both online and embedded images.

### `Image`

```go
package main

import (
	"github.com/discoverkl/goterm/term"
)

func main() {
	term.Open()
	defer term.Close()

    term.Block(term.Image("https://go.dev/images/gophers/ladder.svg"))
}

```

### `ImageData`

#### SVG

```go
package main

import (
	"github.com/discoverkl/goterm/term"
	"github.com/discoverkl/goterm/df"
)

func main() {
	term.Open()
	defer term.Close()

	term.Block(df.ExampleSVG)
}
```

#### PNG

```go
package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math/cmplx"
	"github.com/discoverkl/goterm/term"
)

func main() {
	term.Open()
	defer term.Close()

	img := term.ImageData("image/png", pngdata())
	term.BlockSize(img, 0, 450)
}

func pngdata() []byte {
	const (
		xmin, ymin, xmax, ymax = -2, -2, +2, +2
		width, height          = 1024, 1024
	)

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for py := 0; py < height; py++ {
		y := float64(py)/height*(ymax-ymin) + ymin
		for px := 0; px < width; px++ {
			x := float64(px)/width*(xmax-xmin) + xmin
			z := complex(x, y)

			img.Set(px, py, mandelbrot(z))
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()

}

func mandelbrot(z complex128) color.Color {
	const iterations = 200
	const contrast = 15

	var v complex128
	for n := uint8(0); n < iterations; n++ {
		v = v*v + z
		if cmplx.Abs(v) > 2 {
			return color.Gray{255 - contrast*n}
		}
	}
	return color.Black
}

```

## Plot functions with `XYChart`

```go
package main

import (
	"math"
	"github.com/discoverkl/goterm/term"
	"github.com/discoverkl/goterm/df"
)

func main() {
	term.Open()
	defer term.Close()

	c, _ := df.NewXYFn("sin(x)", math.Sin)
	term.Block(c)
}
```

Multi-series is also supported:

```go
package main

import (
	"math"

	"github.com/discoverkl/goterm/df"
	"github.com/discoverkl/goterm/df/vs"
	"github.com/discoverkl/goterm/term"
)

func main() {
	term.Open()
	defer term.Close()

	xrange := df.PlotX(vs.Range(-4, 4, 0.04))
	s1 := df.LineFn("sin(x)", math.Sin)
	s2 := df.LineFn("cos(x)", math.Cos)
	s3 := df.LineFn("x*ln(x)", func(x float64) float64 { return x * math.Log(x) })

	c, _ := df.NewXYChart(xrange, s1, s2, s3, df.Name("Functions"))
	term.Block(c)
}
```


## General HTML: `term.PrintBlock`

You can print any HTML content as a block, even a whole web page (which will be embedded in an iframe automatically).

```go
package main

import (
	"github.com/discoverkl/goterm/term"
)

func main() {
	term.Open()
	defer term.Close()

	size := term.SizeOption(0, 1200)
	term.PrintBlock(html, size)
}

const html = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Table of Contents - The Go Programming Language</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            padding: 20px;
            max-width: 800px;
            margin: 0 auto;
        }
        h1 {
            text-align: center;
        }
        ol {
            padding-left: 20px;
        }
    </style>
</head>
<body>
    <h1>Table of Contents</h1>
    <ol>
        <li>Introduction</li>
        <li>Getting Started</li>
            <ol>
                <li>Installation</li>
                <li>Writing a Simple Program</li>
            </ol>
        <li>Program Structure</li>
            <ol>
                <li>Names</li>
                <li>Declarations</li>
                <li>Variables</li>
                <li>Assignments</li>
                <li>Type Declarations</li>
            </ol>
        <li>Basic Data Types</li>
            <ol>
                <li>Integers</li>
                <li>Floats</li>
                <li>Complex Numbers</li>
                <li>Booleans</li>
                <li>Strings</li>
            </ol>
        <li>Composite Types</li>
            <ol>
                <li>Arrays</li>
                <li>Slices</li>
                <li>Maps</li>
                <li>Structs</li>
            </ol>
        <li>Functions</li>
            <ol>
                <li>Function Declarations</li>
                <li>Recursion</li>
                <li>Multiple Return Values</li>
            </ol>
        <li>Methods</li>
        <li>Interfaces and Other Types</li>
        <li>Concurrency</li>
            <ol>
                <li>Goroutines</li>
                <li>Channels</li>
                <li>The select Statement</li>
            </ol>
        <li>Packages and the Go Tool</li>
        <li>Testing</li>
        <li>Reflection</li>
        <li>Low-Level Programming</li>
        <li>Appendix</li>
            <ol>
                <li>Language Details</li>
                <li>Package Documentation</li>
                <li>Command-Line Interface</li>
            </ol>
    </ol>
</body>
</html>`
```

# Customization

You can customize those charts freely since there are just `BlockElements`.

For example, you can use `go-echarts` or `gonum/plot` modules to plot any chart and then call `NewGoEChart` or `NewGonumPlot` to get a `BlockElement` from it.

Of course you can define your own `BlockElement` and reuse it.

## The `BlockElement` interface

```go
type BlockElement interface {
	HTML() string // Returns the HTML content of the block
}
```

### ECharts

This example shows how to display a custom `go-echarts` plot.

```go
package main

import (
	"math/rand"
	"github.com/discoverkl/goterm/term"
	"github.com/discoverkl/goterm/df"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
)

func main() {
	term.Open()
	defer term.Close()

	term.Block(df.NewEChart(bar()))
}

// generate random data for bar chart
func generateBarItems() []opts.BarData {
	items := make([]opts.BarData, 0)
	for i := 0; i < 7; i++ {
		items = append(items, opts.BarData{Value: rand.Intn(300)})
	}
	return items
}

func bar() render.Renderer {
	// create a new bar instance
	bar := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Bar chart generated by go-echarts",
		Subtitle: "It's extremely easy to use, right?",
	}))

	// Put data into instance
	bar.SetXAxis([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}).
		AddSeries("Category A", generateBarItems()).
		AddSeries("Category B", generateBarItems())

	return bar
}
```

### Gonum

This example shows how to display a custom `gonum` plot.

```go
package main

import (
	"github.com/discoverkl/goterm/term"
	"github.com/discoverkl/goterm/df"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func main() {
	term.Open()
	defer term.Close()

	groupA := plotter.Values{20, 35, 30, 35, 27}
	groupB := plotter.Values{25, 32, 34, 20, 25}
	groupC := plotter.Values{12, 28, 15, 21, 8}

	p := plot.New()

	p.Title.Text = "Bar chart"
	p.Y.Label.Text = "Heights"

	w := vg.Points(20)

	barsA, err := plotter.NewBarChart(groupA, w)
	if err != nil {
		panic(err)
	}
	barsA.LineStyle.Width = vg.Length(0)
	barsA.Color = plotutil.Color(0)
	barsA.Offset = -w

	barsB, err := plotter.NewBarChart(groupB, w)
	if err != nil {
		panic(err)
	}
	barsB.LineStyle.Width = vg.Length(0)
	barsB.Color = plotutil.Color(1)

	barsC, err := plotter.NewBarChart(groupC, w)
	if err != nil {
		panic(err)
	}
	barsC.LineStyle.Width = vg.Length(0)
	barsC.Color = plotutil.Color(2)
	barsC.Offset = w

	p.Add(barsA, barsB, barsC)
	p.Legend.Add("Group A", barsA)
	p.Legend.Add("Group B", barsB)
	p.Legend.Add("Group C", barsC)
	p.Legend.Top = true
	p.NominalX("One", "Two", "Three", "Four", "Five")

	term.Block(df.NewGonumPlot(p))
}
```

# Options

## Term Options

### Run as web server

```go
package main

import (
	"log"
	"github.com/discoverkl/goterm/term"
	"github.com/discoverkl/goterm/df"
)

func main() {
	term.Open(term.BindPort(8080))
	defer term.Close()

	log.Println("Run as web server")
	term.Block(df.ExampleEChartsPlot)
}
```

## Block Options

### Set the width and height of a `BlockElement`

```go
package main

import (
	"github.com/discoverkl/goterm/term"
	"github.com/discoverkl/goterm/df"
)

func main() {
	term.Open()
	defer term.Close()

	size := term.SizeOption(800, 600)
	term.Block(df.ExamplePNG, size)
}
```