package df

import (
	"bytes"
	"cmp"
	"fmt"
	"image/color"
	"iter"
	"log"
	"math"

	"github.com/discoverkl/goterm/df/vs"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// Default height for gonum plot, in pixels
const DefaultPlotWidth = 0
const DefaultPlotHeight = 480
const DefaultPlotRatio = 16.0 / 9.0

// assuming 96 DPI
const Inch640px = 640 / 96
const Inch480px = 480 / 96

type XYChart struct {
	gp   *plot.Plot
	conf *chartConfig
}

func defaultConfig() *chartConfig {
	return &chartConfig{
		ratio:  DefaultPlotRatio,
		width:  DefaultPlotWidth,
		height: DefaultPlotHeight,
	}
}

func newXYChart(p *plot.Plot) *XYChart {
	return &XYChart{gp: p, conf: defaultConfig()}
}

func NewGonumPlot(p *plot.Plot) *XYChart {
	return newXYChart(p)
}

func NewXY(name string, xx []float64, yy []float64, options ...ChartOption) (*XYChart, error) {
	return create(name, nil, xx, yy, options...)
}

func NewXYFn(name string, fn func(float64) float64, options ...ChartOption) (*XYChart, error) {
	return create(name, fn, nil, nil, options...)
}

func NewXYChart(options ...ChartOption) (*XYChart, error) {
	return create("", nil, nil, nil, options...)
}

func create(name string, fn func(float64) float64, xx []float64, yy []float64, options ...ChartOption) (*XYChart, error) {
	var err error
	// Create a new plot
	p := plot.New()
	c := newXYChart(p)

	// Apply options
	for _, option := range options {
		option(c.conf)
	}

	p.Title.Text = cmp.Or(c.conf.name, name)
	p.Title.TextStyle.Font.Size = vg.Points(16)
	p.Title.Padding = vg.Points(10)
	p.X.Label.Text = cmp.Or(c.conf.xLabel, "X")
	p.Y.Label.Text = cmp.Or(c.conf.yLabel, "Y")

	p.Title.TextStyle.Color = textColor
	p.BackgroundColor = color.Transparent
	p.X.Color = axisLineColor
	p.Y.Color = axisLineColor
	p.X.Label.TextStyle.Color = axisLabelColor
	p.Y.Label.TextStyle.Color = axisLabelColor
	p.X.Tick.LineStyle.Color = axisTickColor
	p.Y.Tick.LineStyle.Color = axisTickColor
	p.X.Tick.Label.Color = textColor
	p.Y.Tick.Label.Color = textColor
	p.Legend.TextStyle.Color = legendTextColor

	// Disable automatic padding to center Y-axis
	p.X.Padding = 0
	p.Y.Padding = 0

	// Merge lines from arguments and options
	linesConfig := []*LineData{}
	if fn != nil || xx != nil {
		linesConfig = append(linesConfig, &LineData{Name: name, Fn: fn, X: xx, Y: yy})
	}
	linesConfig = append(linesConfig, c.conf.lines...)

	// Parse lines to sequences
	seqs := []iter.Seq2[float64, float64]{}
	for _, line := range linesConfig {
		var points iter.Seq2[float64, float64]
		if line.Fn != nil {
			points = getPoints(c, line.Fn)
		} else {
			points = getPoints2(line.X, line.Y)
		}
		seqs = append(seqs, points)
	}

	// Create series
	series := []plotter.XYer{}
	for _, seq := range seqs {
		pts := []plotter.XY{}
		for x, y := range seq {
			if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
				continue
			}
			pts = append(pts, plotter.XY{X: x, Y: y})
		}
		xys := plotter.XYs(pts)
		series = append(series, xys)
	}

	// Set ranges for axes
	c.adjustXYRange(series...)

	// Draw the function
	for i, xys := range series {
		line, err := plotter.NewLine(xys)
		if err != nil {
			return nil, err
		}
		line.Color = getColor(i)
		p.Add(line)
		p.Legend.Add(cmp.Or(linesConfig[i].Name, fmt.Sprintf("Line %d", i)), line)
	}

	// Add zero lines
	err = c.drawZeroLines()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *XYChart) HTML() string {
	p := c.gp
	var buf bytes.Buffer

	buf.WriteString(`<div style="padding: 16px; box-sizing: border-box">`)
	wt, err := p.WriterTo(Inch480px*vg.Length(c.conf.ratio)*vg.Inch, Inch480px*vg.Inch, "svg")
	if err != nil {
		log.Printf("print plot failed: %v", err)
		return ""
	}
	wt.WriteTo(&buf)
	buf.WriteString(`</div>`)
	return buf.String()
}

func (c *XYChart) adjustXYRange(data ...plotter.XYer) {
	p := c.gp
	var xMin, xMax, yMin, yMax float64
	for _, xys := range data {
		for i := 0; i < xys.Len(); i++ {
			x, y := xys.XY(i)
			xMin = min(xMin, x)
			xMax = max(xMax, x)
			yMin = min(yMin, y)
			yMax = max(yMax, y)
		}
	}
	// p.X.Min = min(0, xMin)
	p.X.Max = max(1, xMax)
	// p.Y.Min = min(0, yMin)
	p.Y.Max = max(1, yMax)
}

func (c *XYChart) drawZeroLines() error {
	p := c.gp
	var zeroLine *plotter.Line
	var err error

	// Add a vertical line at X=0 to emphasize the middle Y-axis
	if p.X.Min < 0 {
		zeroLine, err = plotter.NewLine(plotter.XYs{{X: 0, Y: p.Y.Min}, {X: 0, Y: p.Y.Max}})
		if err != nil {
			return err
		}
		zeroLine.LineStyle.Width = vg.Points(0.5)
		zeroLine.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(5)} // Dashed line
		zeroLine.Color = axisLineColor
		p.Add(zeroLine)
	}

	// Add a horizontal line at Y=0 to emphasize the middle X-axis
	if p.Y.Min < 0 {
		zeroLine, err = plotter.NewLine(plotter.XYs{{X: p.X.Min, Y: 0}, {X: p.X.Max, Y: 0}})
		if err != nil {
			return err
		}
		zeroLine.LineStyle.Width = vg.Points(0.5)
		zeroLine.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(5)} // Dashed line
		zeroLine.Color = axisLineColor
		p.Add(zeroLine)
	}
	return nil
}

func getPoints(c *XYChart, fn func(float64) float64) iter.Seq2[float64, float64] {
	return func(yield func(float64, float64) bool) {
		plotX := c.conf.plotX
		if plotX == nil {
			plotX = vs.X()
		}

		for x := range plotX {
			y := fn(x)
			if !yield(x, y) {
				return
			}
		}
	}
}

func getPoints2(xx []float64, yy []float64) iter.Seq2[float64, float64] {
	return func(yield func(float64, float64) bool) {
		for i := 0; i < len(xx); i++ {
			x, y := xx[i], yy[i]
			if !yield(x, y) {
				return
			}
		}
	}
}

func getPalette() []color.Color {
	return []color.Color{
		color.RGBA{R: 0x54, G: 0x70, B: 0xc6, A: 0xff},
		color.RGBA{R: 0x91, G: 0xcc, B: 0x75, A: 0xff},
		color.RGBA{R: 0xfa, G: 0xc8, B: 0x58, A: 0xff},
		color.RGBA{R: 0xee, G: 0x66, B: 0x66, A: 0xff},
		color.RGBA{R: 0x73, G: 0xc0, B: 0xde, A: 0xff},
		color.RGBA{R: 0x3b, G: 0xa2, B: 0x72, A: 0xff},
		color.RGBA{R: 0xfc, G: 0x84, B: 0x52, A: 0xff},
		color.RGBA{R: 0x9a, G: 0x60, B: 0xb4, A: 0xff},
		color.RGBA{R: 0xea, G: 0x7c, B: 0xcc, A: 0xff},
	}
}

func getColor(i int) color.Color {
	return palette[i%len(palette)]
}

var palette = getPalette()
var textColor = color.RGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xff}
var legendTextColor = color.RGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xff}
var axisLineColor = color.RGBA{R: 0x6e, G: 0x70, B: 0x79, A: 0xff}
var axisTickColor = color.RGBA{R: 0x6e, G: 0x70, B: 0x79, A: 0xff}
var axisLabelColor = color.RGBA{R: 0x6e, G: 0x70, B: 0x79, A: 0xff}

func init() {
	// set colors
	mock := false
	if mock {
		textColor = color.RGBA{R: 0xff, A: 0xff}
		legendTextColor = textColor
		axisLineColor = textColor
		axisTickColor = textColor
		axisLabelColor = textColor
	}
}
