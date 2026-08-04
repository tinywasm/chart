package bar

import (
	"github.com/tinywasm/color"
	. "github.com/tinywasm/fmt"
	"github.com/tinywasm/pdf"
)

type Chart struct {
	canvas pdf.Canvas
	title  string
	width  float64
	height float64
	bars   []barData
}

type barData struct {
	label string
	value float64
	color color.Color
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

func (c *Chart) AddBar(val float64, label string, col ...color.Color) *Chart {
	var colVal color.Color
	if len(col) > 0 {
		colVal = col[0]
	} else {
		colVal = "#646464" // Default color
	}
	c.bars = append(c.bars, barData{
		label: label,
		value: val,
		color: colVal,
	})
	return c
}

func (c *Chart) Draw() {
	if c.width == 0 {
		// Use available width
		w, _ := c.canvas.GetPageSize()
		l, _, r, _ := c.canvas.GetMargins()
		c.width = w - l - r
	}
	if c.height == 0 {
		c.height = 100 // Default height
	}

	x := c.canvas.GetX()
	y := c.canvas.GetY()

	// Title
	if c.title != "" {
		c.canvas.SetDrawingFont("B", 12)
		c.canvas.CellFormat(c.width, 10, c.title, "", 1, "C", false, 0, "")
		y = c.canvas.GetY() + 5
	}

	// Calculate Scale
	maxVal := 0.0
	for _, b := range c.bars {
		if b.value > maxVal {
			maxVal = b.value
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	// Margins inside chart area
	margin := 20.0
	plotH := c.height - margin
	scaleY := plotH / maxVal
	barWidth := (c.width - margin) / float64(len(c.bars))

	// Draw Axes
	c.canvas.SetDrawColor(color.Color("#000000"))
	c.canvas.SetLineWidth(0.2)
	c.canvas.Line(x, y, x, y+c.height) // Y Axis
	c.canvas.Line(x, y+c.height, x+c.width, y+c.height) // X Axis

	// Draw Bars
	for i, bar := range c.bars {
		h := bar.value * scaleY
		bx := x + 10 + float64(i)*barWidth // 10 offset from Y axis
		by := y + c.height - h

		c.canvas.SetFillColor(bar.color)
		// Use simple rect
		c.canvas.Rect(bx+2, by, barWidth-4, h, "F")

		// Draw Text
		c.canvas.SetTextColor(color.Color("#000000"))
		c.canvas.SetDrawingFont("", 8)

		// Value on top
		valStr := Sprintf("%.1f", bar.value)
		wVal := c.canvas.GetStringWidth(valStr)
		c.canvas.Text(bx+(barWidth-wVal)/2, by-2, valStr)

		// Label on bottom
		wLbl := c.canvas.GetStringWidth(bar.label)
		c.canvas.Text(bx+(barWidth-wLbl)/2, y+c.height+4+4, bar.label) // +4 to descend below axis
	}

	c.canvas.SetY(y + c.height + 20) // Move below chart
}
