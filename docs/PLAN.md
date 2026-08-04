---
PLAN: "feat: Colapsar chart.Canvas en pdf.Canvas"
TAG: v0.1.0
STATUS: completed
SESSION: 9144200332693953888
---

## Antes de escribir código: lee [CONSTRUCTION_HARNESS.md](CONSTRUCTION_HARNESS.md)

**Es vinculante, no orientativo.** Los principios que gobiernan este trabajo:

| # | Principio | Cómo se aplica aquí |
|---|---|---|
| 1 | Typed over `any` | `chart.Canvas` pide `(r,g,b int)` sueltos; `pdf.Canvas` pide `color.Color`. Queda el tipado. |
| 4 | One way to do each thing | Hoy hay **dos** interfaces `Canvas` con nombres de método distintos para lo mismo. Queda una. |
| 9 | Lego pieces, never forks | El contrato de dibujo lo posee `tinywasm/pdf` (ya lo publicó, con guardia de compilación). `chart` lo consume, no lo reinventa. |

Y la regla del harness que decide dónde vive:

> **A missing contract at a boundary is a defect in the library, not in the consumer.**
> If two libraries meet and there is no type to name the thing that crosses between
> them, **the type is missing upstream.**

---

## 1. Síntoma

`chart/chart.go` declara su propio `Canvas`. `pdf/canvas.go` declara otro. Divergen en
tres puntos, verificados en el código publicado (`chart` v0.0.2, `pdf` v0.1.0):

| | `chart.Canvas` | `pdf.Canvas` |
|---|---|---|
| Color | `SetFillColor(r, g, b int)` | `SetFillColor(c color.Color)` |
| Fuente | `SetFont(styleStr string, size float64)` | `SetDrawingFont(style string, size float64)` |
| Garantía de compilación | ninguna | `var _ Canvas = (*Document)(nil)` |

**Consecuencia real:** `*pdf.Document` ya no satisface `chart.Canvas`. `bar.New`,
`line.New`, `pie.New` no aceptan un documento real.

---

## 2. Causa

Dos agentes distintos implementaron cada lado sin verse. El tipado estructural de Go no
avisa entre módulos publicados por separado — compiló en los dos repos y sólo falla al
intentar usarlos juntos.

Y no se detectó porque **`chart/tests/chart_test.go` nunca probó contra `*pdf.Document`**.
Construye un `fpdf.Fpdf` crudo (`github.com/tinywasm/pdf/fpdf`, el motor interno, no la
API pública) envuelto en un `CanvasAdapter` local que cierra sobre `"Roboto"` a mano:

```go
type CanvasAdapter struct {
    *fpdf.Fpdf
    fontFamily string
}
func (a *CanvasAdapter) SetFont(styleStr string, size float64) {
    a.Fpdf.SetFont(a.fontFamily, styleStr, size)
}
```

Eso es exactamente el patrón de string de familia suelto que `pdf` v0.1.0 eliminó —
reintroducido aquí, en un test, por necesidad de tener *algo* que satisficiera
`chart.Canvas`. Es el hueco que la regla del harness sobre *consumer-shaped tests*
existe para atrapar: el test debe ejercitar el colaborador real, no un doble hecho a
medida.

---

## 3. Decisión: el contrato vive en `tinywasm/pdf`

No en `chart`, al revés de lo que decidió la primera versión de este plan («Go tiene
interfaces estructurales, que viva donde se consume»). Se revierte esa decisión con esta
evidencia encima:

- **`pdf` es upstream.** `chart` ya depende de `pdf` en su `go.mod`; `pdf` no sabe que
  `chart` existe. Por vocabulario del harness, el tipo que falta se declara upstream.
- **`pdf.Canvas` ya tiene guardia de compilación**, `var _ Canvas = (*Document)(nil)`, que
  sólo prueba algo si vive junto a la implementación que dice satisfacer.
- **`pdf.Canvas` está mejor diseñada en los dos puntos que divergen**: usa `color.Color`
  en vez de ints sueltos, y `SetDrawingFont` evita a propósito chocar con un futuro
  `SetFont` real de `Document` (comentado así en `canvas.go:126-129`).

**Sin alias.** `chart.Canvas = pdf.Canvas` dejaría dos nombres para un tipo — el mismo
error que se descartó para `pdf.Color` en `pdf/docs/PLAN.md` §5. `bar.New`, `line.New` y
`pie.New` pasan a pedir `pdf.Canvas` directamente.

---

## 4. Cambios

### 4.1 `chart/chart.go`

Borrar la interfaz `Canvas`. Si el archivo queda vacío, se elimina.

### 4.2 `bar/bar.go`, `line/line.go`, `pie/pie.go`

- Import `"github.com/tinywasm/chart"` → `"github.com/tinywasm/pdf"`.
- `canvas chart.Canvas` → `canvas pdf.Canvas`; `New(c chart.Canvas)` → `New(c pdf.Canvas)`.
- **Simplificar las llamadas de color.** Hoy cada `Draw()` descompone a mano:

  ```go
  r, g, b, err := bar.color.RGB()
  if err == nil {
      c.canvas.SetFillColor(r, g, b)
  } else {
      c.canvas.SetFillColor(100, 100, 100)
  }
  ```

  Con `pdf.Canvas.SetFillColor(c color.Color)` esto se reduce a
  `c.canvas.SetFillColor(bar.color)` — el fallback de error lo resuelve `Document`
  internamente (`d.addError(err)`), no cada gráfico por separado. Aplica a los tres
  archivos: `SetFillColor`, `SetDrawColor`, `SetTextColor`. `pie.go:90` tiene además un
  literal `SetDrawColor(255, 255, 255)` que pasa a `SetDrawColor(color.Color("#FFFFFF"))`.
- `c.canvas.SetFont(...)` → `c.canvas.SetDrawingFont(...)`, mismos argumentos.

### 4.3 `tests/chart_test.go` — el fix que importa de verdad

Sustituir el `fpdf.Fpdf` + `CanvasAdapter` por un `*pdf.Document` real:

```go
tf, err := pdf.LoadDeclared(font.Declare("Roboto", fontDir))
if err != nil { t.Fatalf(...) }
doc := pdf.NewDocument(tf)
doc.AddPage()

bar.New(doc).Title("Monthly Sales")./* ... */.Draw()
```

Sin adaptador, sin familia hardcodeada. Si esto no compila contra `pdf.Canvas`, ahí está
el defecto — antes de publicar, no después.

---

## 5. Verificación

1. `bar.New`, `line.New`, `pie.New` aceptan `*pdf.Document` directamente, sin adaptador.
2. `grep -rn "chart.Canvas\|fpdf.Fpdf\|CanvasAdapter" .` no encuentra nada.
3. `grep -rn "\.RGB()" bar/ line/ pie/` no encuentra nada: `color.Color` viaja entero
   hasta `pdf.Canvas`.
4. `tests/chart_test.go` construye el documento vía `font.Declare` + `pdf.LoadDeclared`
   + `pdf.NewDocument` — el mismo camino que usaría un consumidor real.
5. `gotest` en verde.

`docs/ARCHITECTURE.md` y `README.md` se actualizan en el mismo commit: la mención de
`chart.Canvas` pasa a `pdf.Canvas`.
