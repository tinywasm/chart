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
	"github.com/tinywasm/pdf/fpdf"
)

type CanvasAdapter struct {
	*fpdf.Fpdf
	fontFamily string
}

func (a *CanvasAdapter) SetFont(styleStr string, size float64) {
	a.Fpdf.SetFont(a.fontFamily, styleStr, size)
}

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

	regData, err := os.ReadFile(filepath.Join(fontDir, "Roboto-Regular.ttf"))
	if err != nil {
		t.Fatalf("loading regular font: %v", err)
	}
	boldData, err := os.ReadFile(filepath.Join(fontDir, "Roboto-Bold.ttf"))
	if err != nil {
		t.Fatalf("loading bold font: %v", err)
	}

	pdf := fpdf.New(
		fpdf.WriteFileFunc(func(filePath string, content []byte) error {
			return os.WriteFile(filePath, content, 0644)
		}),
		fpdf.ReadFileFunc(os.ReadFile),
	)
	pdf.AddUTF8FontFromBytes("Roboto", "", regData)
	pdf.AddUTF8FontFromBytes("Roboto", "B", boldData)

	pdf.AddPage()
	pdf.SetFont("Roboto", "B", 16)
	pdf.CellFormat(0, 8, "Chart Examples", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	adapter := &CanvasAdapter{
		Fpdf:       pdf,
		fontFamily: "Roboto",
	}

	// Bar Chart
	pdf.SetFont("Roboto", "B", 12)
	pdf.CellFormat(0, 6, "Bar Chart", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	bar.New(adapter).
		Title("Monthly Sales").
		Height(100).
		AddBar(120, "Jan", "#3264C8").
		AddBar(140, "Feb", "#C86432").
		AddBar(110, "Mar", "#32C864").
		Draw()

	pdf.Ln(10)

	// Line Chart
	pdf.SetFont("Roboto", "B", 12)
	pdf.CellFormat(0, 6, "Line Chart", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	line.New(adapter).
		Title("Growth Trends").
		Height(100).
		AddSeries("Revenue", []float64{10, 15, 13, 17, 20, 25, 22}, "#0000FF").
		Draw()

	pdf.Ln(10)

	// Pie Chart
	pdf.SetFont("Roboto", "B", 12)
	pdf.CellFormat(0, 6, "Pie Chart", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	pie.New(adapter).
		Title("Market Share").
		Height(120).
		AddSlice("A", 40, "#FF0000").
		AddSlice("B", 30, "#00FF00").
		AddSlice("C", 30, "#0000FF").
		Draw()

	writeErr := pdf.OutputFileAndClose(out)
	if writeErr != nil {
		t.Fatalf("OutputFileAndClose failed: %v", writeErr)
	}

	st, err := os.Stat(out)
	if err != nil {
		t.Fatalf("stat %s: %v", out, err)
	}
	if st.Size() == 0 {
		t.Fatalf("PDF %s is empty", out)
	}
}
