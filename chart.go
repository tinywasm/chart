package chart

// Canvas represents the drawing surface that a chart needs.
// Any pdf.Document implementing these methods satisfies this interface structurally.
type Canvas interface {
	// Geometry
	GetMargins() (left, top, right, bottom float64)
	GetPageSize() (width, height float64)
	GetX() float64
	GetY() float64
	SetY(y float64)

	// Path Drawing
	MoveTo(x, y float64)
	LineTo(x, y float64)
	ArcTo(x, y, rx, ry, degRotate, degStart, degEnd float64)
	DrawPath(styleStr string)
	Line(x1, y1, x2, y2 float64)

	// Shapes
	Rect(x, y, w, h float64, styleStr string)
	Circle(x, y, r float64, styleStr string)

	// Styling
	SetDrawColor(r, g, b int)
	SetFillColor(r, g, b int)
	SetTextColor(r, g, b int)
	SetLineWidth(width float64)
	SetFont(styleStr string, size float64)

	// Text
	Text(x, y float64, txtStr string)
	CellFormat(w, h float64, txtStr, borderStr string, ln int, alignStr string, fill bool, link int, linkStr string)
	GetStringWidth(txtStr string) float64
}
