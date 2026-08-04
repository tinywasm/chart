package pie

import (
	"math"

	"github.com/tinywasm/chart"
	"github.com/tinywasm/color"
)

type Chart struct {
	canvas chart.Canvas
	title  string
	width  float64
	height float64
	slices []pieSlice
}

type pieSlice struct {
	label string
	value float64
	color color.Color
}

func New(c chart.Canvas) *Chart {
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

func (c *Chart) AddSlice(label string, val float64, col color.Color) *Chart {
	c.slices = append(c.slices, pieSlice{
		label: label,
		value: val,
		color: col,
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
		c.canvas.SetFont("B", 12)
		c.canvas.CellFormat(c.width, 10, c.title, "", 1, "C", false, 0, "")
		y = c.canvas.GetY() + 5
	}

	total := 0.0
	for _, s := range c.slices {
		total += s.value
	}

	cx := x + c.width/2
	cy := y + c.height/2
	radius := c.height / 2
	if c.width < c.height {
		radius = c.width / 2
	}
	radius -= 10 // Padding

	startAngle := 0.0

	c.canvas.SetLineWidth(0.2)
	c.canvas.SetDrawColor(255, 255, 255) // White borders

	for _, s := range c.slices {
		angle := (s.value / total) * 360.0
		endAngle := startAngle + angle

		r, g, b, err := s.color.RGB()
		if err != nil {
			r, g, b = 100, 100, 100
		}
		c.canvas.SetFillColor(r, g, b)

		c.canvas.MoveTo(cx, cy)
		c.canvas.ArcTo(cx, cy, radius, radius, 0, startAngle, endAngle)
		c.canvas.LineTo(cx, cy)
		c.canvas.DrawPath("F")

		// Label (Radial)
		midAngle := startAngle + angle/2
		midRad := midAngle * math.Pi / 180

		// Using standard trigonometry assuming standard orientation
		// Adjust signs if necessary after visual inspection
		tx := cx + (radius * 0.7) * math.Cos(midRad)
		ty := cy - (radius * 0.7) * math.Sin(midRad)

		c.canvas.SetTextColor(255, 255, 255)
		c.canvas.SetFont("B", 10)
		txt := s.label
		if len(txt) > 0 {
			wTxt := c.canvas.GetStringWidth(txt)
			c.canvas.Text(tx-wTxt/2, ty+3, txt)
		}

		startAngle += angle
	}

	c.canvas.SetY(y + c.height + 20)
}
