package df

import (
	"cmp"
	"iter"

	"github.com/discoverkl/goterm/term"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type chartConfig struct {
	// general configs which need to be initialized
	width  int
	height int

	// general options
	name   string
	xLabel string
	yLabel string

	// for gonum plot
	ratio float64
	plotX iter.Seq[float64]
	lines []*LineData
}

type LineData struct {
	Name string
	X    []float64
	Y    []float64
	Fn   func(float64) float64
}

type ChartOption func(*chartConfig)

func Name(name string) ChartOption {
	return func(c *chartConfig) {
		c.name = name
	}
}

func XName(name string) ChartOption {
	return func(c *chartConfig) {
		c.xLabel = name
	}
}

func YName(name string) ChartOption {
	return func(c *chartConfig) {
		c.yLabel = name
	}
}

func LineFn(name string, fn func(float64) float64) ChartOption {
	return func(c *chartConfig) {
		c.lines = append(c.lines, &LineData{Name: name, Fn: fn})
	}
}

func LineXY(name string, x, y []float64) ChartOption {
	return func(c *chartConfig) {
		c.lines = append(c.lines, &LineData{Name: name, X: x, Y: y})
	}
}

func PlotX(x iter.Seq[float64]) ChartOption {
	return func(c *chartConfig) {
		c.plotX = x
	}
}

func Size(width, height int) ChartOption {
	return func(c *chartConfig) {
		c.width = width
		c.height = height
	}
}

func Ratio(ratio float64) ChartOption {
	return func(c *chartConfig) {
		c.ratio = ratio
	}
}

func (d *dataFrame) configEcharts(chart any, options ...ChartOption) *chartConfig {
	c := &chartConfig{}
	for _, option := range options {
		option(c)
	}

	name := c.name
	xname := cmp.Or(c.xLabel, d.GetColumnAt(0).Name())
	yname := c.yLabel

	switch chart := chart.(type) {
	case *charts.Bar:
		chart.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title: name,
			}),
			charts.WithXAxisOpts(opts.XAxis{
				Name: xname,
			}),
			charts.WithYAxisOpts(opts.YAxis{
				Name: yname,
			}),
		)
	case *charts.RectChart:
		chart.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title: name,
			}),
			charts.WithXAxisOpts(opts.XAxis{
				Name: xname,
			}),
			charts.WithYAxisOpts(opts.YAxis{
				Name: yname,
			}),
		)
	}
	return c
}

func (d *dataFrame) Bar(options ...ChartOption) {
	bar := charts.NewBar()
	c := d.configEcharts(&bar.RectChart, options...)

	bar.SetXAxis(d.GetColumnAt(0).AsString())
	for i := 1; i < len(d.Columns()); i++ {
		series := d.GetColumnAt(i)
		var items []opts.BarData
		for _, v := range series.Data() {
			items = append(items, opts.BarData{Value: v})
		}
		bar.AddSeries(series.Name(), items)
	}

	d.printChart(NewEChart(bar), c)
}

func (d *dataFrame) Line(options ...ChartOption) {
	line := charts.NewLine()
	c := d.configEcharts(&line.RectChart, options...)

	line.SetXAxis(d.GetColumnAt(0).AsString())
	for i := 1; i < len(d.Columns()); i++ {
		series := d.GetColumnAt(i)
		var items []opts.LineData
		for _, v := range series.Data() {
			items = append(items, opts.LineData{Value: v})
		}
		line.AddSeries(series.Name(), items)
	}

	d.printChart(NewEChart(line), c)
}

func (d *dataFrame) Pie(options ...ChartOption) {
	pie := charts.NewPie()
	c := d.configEcharts(pie, options...)

	names := d.GetColumnAt(0).AsString()
	series := d.GetColumnAt(1)
	var items []opts.PieData
	for j, v := range series.Data() {
		items = append(items, opts.PieData{Name: names[j], Value: v})
	}
	pie.AddSeries(series.Name(), items)

	d.printChart(NewEChart(pie), c)
}

func (d *dataFrame) XY(options ...ChartOption) {
	if len(d.Columns()) < 2 {
		return
	}
	x := d.GetColumnAt(0).ToFloat64()
	chartOPs := []ChartOption{XName(d.GetColumnAt(0).Name())}
	for i, name := range d.Columns() {
		if i == 0 {
			continue
		}
		y := d.GetColumnAt(i).ToFloat64()
		chartOPs = append(chartOPs, LineXY(name, x, y))
	}

	// chartOPs goes first for auto x label
	options = append(chartOPs, options...)
	c, err := NewXYChart(options...)
	if err != nil {
		return
	}
	d.printChart(c, c.conf)
}

func (d *dataFrame) printChart(chart term.BlockElement, c *chartConfig) {
	ops := []term.BlockOption{}
	if c.width != 0 || c.height != 0 {
		ops = append(ops, term.SizeOption(c.width, c.height))
	}
	term.Block(chart, ops...)
}
