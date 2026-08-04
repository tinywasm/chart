package chart_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tinywasm/chart/bar"
	"github.com/tinywasm/chart/line"
	"github.com/tinywasm/chart/pie"
	"github.com/tinywasm/font"
	"github.com/tinywasm/pdf"
)

func getFontDir(t *testing.T) string {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/tinywasm/pdf")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to find pdf module directory: %v", err)
	}
	dir := strings.TrimSpace(string(out))
	return filepath.Join(dir, "fpdf", "fonts")
}

func TestCharts(t *testing.T) {
	const out = "test_charts.pdf"
	_ = os.Remove(out)
	defer os.Remove(out) // clean up after test execution

	fontDir := getFontDir(t)

	d := font.Declare("Roboto", fontDir)
	tf, err := pdf.LoadDeclared(d)
	if err != nil {
		t.Fatalf("loading typeface: %v", err)
	}

	doc := pdf.NewDocument(tf)
	doc.AddPage()

	doc.SetDrawingFont("B", 16)
	doc.CellFormat(0, 8, "Chart Examples", "", 1, "L", false, 0, "")
	doc.SetY(doc.GetY() + 5)

	// Bar Chart
	doc.SetDrawingFont("B", 12)
	doc.CellFormat(0, 6, "Bar Chart", "", 1, "L", false, 0, "")
	doc.SetY(doc.GetY() + 2)

	bar.New(doc).
		Title("Monthly Sales").
		Height(100).
		AddBar(120, "Jan", "#3264C8").
		AddBar(140, "Feb", "#C86432").
		AddBar(110, "Mar", "#32C864").
		Draw()

	doc.SetY(doc.GetY() + 10)

	// Line Chart
	doc.SetDrawingFont("B", 12)
	doc.CellFormat(0, 6, "Line Chart", "", 1, "L", false, 0, "")
	doc.SetY(doc.GetY() + 2)

	line.New(doc).
		Title("Growth Trends").
		Height(100).
		AddSeries("Revenue", []float64{10, 15, 13, 17, 20, 25, 22}, "#0000FF").
		Draw()

	doc.SetY(doc.GetY() + 10)

	// Pie Chart
	doc.SetDrawingFont("B", 12)
	doc.CellFormat(0, 6, "Pie Chart", "", 1, "L", false, 0, "")
	doc.SetY(doc.GetY() + 2)

	pie.New(doc).
		Title("Market Share").
		Height(120).
		AddSlice("A", 40, "#FF0000").
		AddSlice("B", 30, "#00FF00").
		AddSlice("C", 30, "#0000FF").
		Draw()

	writeErr := doc.WritePdf(out)
	if writeErr != nil {
		t.Fatalf("WritePdf failed: %v", writeErr)
	}

	st, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat %s: %v", out, err)
	}
	if st.Size() == 0 {
		t.Fatalf("PDF %s is empty", out)
	}
}
