package df

import (
	"fmt"

	"github.com/discoverkl/goterm/term"
	"github.com/go-echarts/go-echarts/v2/render"
)

type RenderMode string

var (
	echartRenderMode = DivMode
)

// Default height for the echart iframe, in pixels. The chart only needs 530px, but the body has a padding of 8px.
const (
	eChartIframeHeight            = 550
	eChartDivHeight               = 0
	IFrameMode         RenderMode = "iframe"
	DivMode            RenderMode = "div"
)

func EChartRenderMode(mode RenderMode) {
	echartRenderMode = mode
}

type EChart struct {
	chart render.Renderer
}

func NewEChart(chart render.Renderer) *EChart {
	return &EChart{chart: chart}
}

func (c *EChart) HTML() string {
	html := string(c.chart.RenderContent())

	switch echartRenderMode {
	case IFrameMode:
		// Chart class has a minimum width of 916px to fit the echart.
		return term.EscapeIframe(html, "echart")
	case DivMode:
		// Wrap the whole page in a div will prevent auto iframe wraping in the PrintBlockSize function.
		return escapeEChartWithDiv(html)
	default:
		panic("unsupported render mode")
	}
}

func (c *EChart) Options() []term.BlockOption {
	return []term.BlockOption{
		term.SizeOption(0, eChartDefaultHeight()),
	}
}

func eChartDefaultHeight() int {
	switch echartRenderMode {
	case IFrameMode:
		return eChartIframeHeight
	case DivMode:
		return eChartDivHeight
	default:
		panic("unknown echart render mode")
	}
}

func escapeEChartWithDiv(html string) string {
	return fmt.Sprintf("<div class='echart'>%s</div>", html)
}
