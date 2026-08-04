# PLAN — Extraer los gráficos de `tinywasm/pdf` a este módulo

## Antes de escribir código: lee [CONSTRUCTION_HARNESS.md](CONSTRUCTION_HARNESS.md)

**Es vinculante, no orientativo.** Los principios que gobiernan este trabajo:

| # | Principio | Cómo se aplica aquí |
|---|---|---|
| 5 | Minimal surface | La raíz expone **sólo el contrato**. Un proyecto que quiere barras no enlaza el código de tarta ni de líneas. |
| 9 | Lego pieces, never forks | Dibujar gráficos no es responsabilidad de un generador de PDF. Una responsabilidad, una pieza. |
| 1 | Typed over `any` | El contrato de dibujo **no puede** llevar un nombre de fuente en `string`: `pdf` v0.1.0 acaba de eliminar ese agujero. |

---

## 1. Por qué existe este módulo

`tinywasm/pdf` incluye hoy 398 líneas de gráficos (`chart.go`, `chart_bar.go`,
`chart_line.go`, `chart_pie.go`). Está bien que la funcionalidad exista —hay documentos
que necesitan gráficos— y está mal dónde vive:

1. **No es su responsabilidad.** `pdf` genera documentos; dibujar una tarta con
   porcentajes es otro concern.
2. **Todo proyecto paga por los tres tipos.** La fábrica actual cuelga del documento:

   ```go
   func (d *Document) Chart() *ChartFactory
   func (f *ChartFactory) Bar()  *BarChart
   func (f *ChartFactory) Line() *LineChart
   func (f *ChartFactory) Pie()  *PieChart
   ```

   Alcanzar `Bar()` obliga a enlazar `ChartFactory`, y con ella `Line` y `Pie`. Un
   informe que sólo dibuja barras carga los tres. En un binario WASM eso se paga en cada
   carga de página.

**La estructura de este módulo es la corrección:** la raíz lleva el contrato; cada tipo
de gráfico es un subpaquete. Importar `chart/bar` no enlaza `chart/pie`.

---

## 2. El problema real: el contrato no existe

El código ya está copiado en `bar/`, `line/` y `pie/`, con su `package` renombrado.
**No compila**, y la razón es la parte sustancial de este plan.

Los tres dibujan llamando a métodos **no exportados** del documento:

```go
c.doc.internal.Rect(...)      // internal es un campo privado de pdf.Document
c.doc.getActiveFontName()     // método privado
```

Un módulo aparte no puede alcanzarlos. Son 20 primitivas:

| Grupo | Métodos |
|---|---|
| Geometría | `GetMargins` `GetPageSize` `GetX` `GetY` `SetY` |
| Trazado | `MoveTo` `LineTo` `ArcTo` `DrawPath` `Line` |
| Formas | `Rect` `Circle` |
| Estilo | `SetDrawColor` `SetFillColor` `SetTextColor` `SetLineWidth` `SetFont` |
| Texto | `Text` `CellFormat` `GetStringWidth` |

Esto es exactamente lo que el harness describe:

> **A missing contract at a boundary is a defect in the library, not in the consumer.**
> If two libraries meet and there is no type to name the thing that crosses between
> them, the type is missing upstream. **Do not declare a local intersection to paper
> over it.**

Así que `tinywasm/pdf` debe **exportar una superficie de dibujo tipada**, y este módulo
consumirla. La alternativa —duplicar aquí la lógica de dibujo, o exportar `internal`—
es un fork con otro nombre.

### La restricción que no se puede violar

`SetFont(familyStr, styleStr, size)` **no puede entrar en el contrato tal cual**.
`pdf` v0.1.0 acaba de eliminar los nombres de fuente en `string` porque cualquier cadena
compilaba y sólo algunas funcionaban. Reintroducirlos aquí reabriría el mismo agujero
por la puerta de atrás.

El contrato debe pedir *estilo de texto*, no *nombre de fuente*: la familia ya la fijó
`NewDocument(Typeface)`, y un gráfico sólo necesita decir «negrita, 12pt».

---

## 3. Diseño propuesto

### La raíz: sólo el contrato

```go
package chart

// Canvas es la superficie de dibujo que un gráfico necesita.
// *pdf.Document la satisface.
type Canvas interface {
    // geometría, trazado, formas, estilo y texto — las 20 primitivas
}
```

Sin implementación, sin tipos de gráfico, sin datos de ejemplo. Un proyecto que importe
`chart` a secas no enlaza ningún dibujo.

### Los subpaquetes: un tipo por carpeta

```go
package bar

func New(c chart.Canvas) *Chart
func (b *Chart) Data(...) *Chart
func (b *Chart) Draw()
```

Cada uno depende sólo de `chart` (el contrato) y de nada más de este módulo.
`chart/bar` no importa `chart/pie`.

### Decisiones que el agente **no** debe tomar solo

1. **`Color` viene de `tinywasm/color`.** Ya no se discute: existe la pieza que lo
   posee (`color/docs/PLAN.md`), es WASM-safe y no tiene build tag. Este módulo la
   importa directamente, sin pasar por `pdf`. Declarar un `chart.Color` propio sería
   el tercer parser de hex del ecosistema.
2. **¿La interfaz `Canvas` va aquí o en `pdf`?** Recomendado: **aquí**. Go tiene
   interfaces estructurales, así que `*pdf.Document` la satisface sin declararlo. Y la
   regla es que la interfaz pertenece a quien la consume.
3. **La forma exacta del estilo de texto** (§2, la restricción). Debe quedar cerrada
   antes de escribir los subpaquetes, porque los tres la usan.

---

## 4. Trabajo en `tinywasm/pdf`

Este plan no se puede completar sin el lado de allá, y va en un commit del otro repo:

1. **Exportar la superficie de dibujo.** Las 20 primitivas, hoy alcanzables sólo vía
   `d.internal`, con el estilo de texto resuelto según §2.
2. **Borrar `chart.go`, `chart_bar.go`, `chart_line.go`, `chart_pie.go`** y el método
   `(*Document).Chart()`. Sin deprecación: un símbolo exportado que nadie llama sigue
   enlazando su código.
3. **`tests/chart_test.go`** se traslada aquí.

---

## 5. Verificación

1. `chart/bar`, `chart/line` y `chart/pie` compilan y dibujan contra un `Canvas` real.
2. **Un binario que importa `chart/bar` no contiene símbolos de `pie` ni de `line`.**
   Es el criterio que justifica la estructura, y se comprueba con `go tool nm` sobre un
   binario de prueba — no por inspección visual de los imports.
3. `tinywasm/pdf` ya no expone `Chart()` ni tipo de gráfico alguno, y **su `client.wasm`
   baja** respecto a los 3.785.258 B actuales.
4. El contrato **no** contiene ningún parámetro de nombre de fuente en `string`.
5. Los gráficos generados son visualmente equivalentes a los de `pdf` v0.1.0 — el
   `tests/test_charts.pdf` de aquel repo sirve de referencia.
6. `gotest` en verde en ambos módulos.

`docs/ARCHITECTURE.md` y `docs/SPECS.md` se escriben en el mismo commit; `README.md`
los indexa.

---

## 6. Estado actual del repositorio

Ya hecho, para que el agente no lo repita:

- `bar/bar.go`, `line/line.go`, `pie/pie.go` — copiados de `pdf` y con el `package`
  renombrado. **No compilan**: siguen llamando a `c.doc.internal.*`.
- `chart.go` de la raíz borrado: la fábrica desaparece, no se traduce.

Falta todo lo demás.
