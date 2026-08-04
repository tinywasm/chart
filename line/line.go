package line

import (
	"github.com/tinywasm/color"
	"github.com/tinywasm/pdf"
)

type Chart struct {
	canvas pdf.Canvas
	title  string
	width  float64
	height float64
	series []lineSeries
}

type lineSeries struct {
	name  string
	data  []float64
	color color.Color
	width float64
}

func New(c pdf.Canvas) *Chart {
	return &Chart{
		canvas: c,
	}
}

func (c *Chart) Title(t string) *Chart {
	c.title = t
	return c
}

func (c *Chart) Height(h float64) *Chart {
	c.height = h
	return c
}

func (c *Chart) Width(w float64) *Chart {
	c.width = w
	return c
}

func (c *Chart) AddSeries(name string, data []float64, col color.Color) *Chart {
	c.series = append(c.series, lineSeries{
		name:  name,
		data:  data,
		color: col,
		width: 0.5,
	})
	return c
}

func (c *Chart) Draw() {
	if c.width == 0 {
		w, _ := c.canvas.GetPageSize()
		l, _, r, _ := c.canvas.GetMargins()
		c.width = w - l - r
	}
	if c.height == 0 {
		c.height = 100
	}

	x := c.canvas.GetX()
	y := c.canvas.GetY()

	// Title
	if c.title != "" {
		c.canvas.SetDrawingFont("B", 12)
		c.canvas.CellFormat(c.width, 10, c.title, "", 1, "C", false, 0, "")
		y = c.canvas.GetY() + 5
	}

	// Calculate Max Y
	maxVal := 0.0
	maxPoints := 0
	for _, s := range c.series {
		if len(s.data) > maxPoints {
			maxPoints = len(s.data)
		}
		for _, v := range s.data {
			if v > maxVal {
				maxVal = v
			}
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}
	if maxPoints < 2 {
		return // Not enough data to draw lines
	}

	plotH := c.height - 20
	scaleY := plotH / maxVal
	stepX := (c.width - 20) / float64(maxPoints-1)

	// Draw Axes
	c.canvas.SetDrawColor(color.Color("#000000"))
	c.canvas.SetLineWidth(0.2)
	c.canvas.Line(x, y, x, y+c.height) // Y
	c.canvas.Line(x, y+c.height, x+c.width, y+c.height) // X

	// Draw Series
	for _, s := range c.series {
		c.canvas.SetDrawColor(s.color)
		c.canvas.SetLineWidth(s.width)
		c.canvas.SetFillColor(s.color)

		for i := 0; i < len(s.data)-1; i++ {
			x1 := x + 10 + float64(i)*stepX
			y1 := y + c.height - (s.data[i] * scaleY)
			x2 := x + 10 + float64(i+1)*stepX
			y2 := y + c.height - (s.data[i+1] * scaleY)

			c.canvas.Line(x1, y1, x2, y2)
			// Dot
			c.canvas.Circle(x1, y1, s.width*2, "F")
		}
		// Last dot
		if len(s.data) > 0 {
			i := len(s.data) - 1
			x1 := x + 10 + float64(i)*stepX
			y1 := y + c.height - (s.data[i] * scaleY)
			c.canvas.Circle(x1, y1, s.width*2, "F")
		}
	}

	c.canvas.SetY(y + c.height + 20)
}
