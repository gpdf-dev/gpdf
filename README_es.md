# gpdf

[![Go Reference](https://pkg.go.dev/badge/github.com/gpdf-dev/gpdf.svg)](https://pkg.go.dev/github.com/gpdf-dev/gpdf)
[![CI](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml/badge.svg)](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gpdf-dev/gpdf)](https://goreportcard.com/report/github.com/gpdf-dev/gpdf)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.22-blue)](https://go.dev/)

[English](README.md) | [日本語](README_ja.md) | [中文](README_zh.md) | [한국어](README_ko.md) | **Español** | [Português](README_pt.md)

Biblioteca de generación de PDF en Go puro, sin dependencias externas, con arquitectura por capas y API declarativa de constructores.

## Características

- **Cero dependencias** — solo la biblioteca estándar de Go
- **Arquitectura por capas** — primitivas PDF de bajo nivel, modelo de documento y API de plantillas de alto nivel
- **Sistema de cuadrícula de 12 columnas** — diseño responsivo estilo Bootstrap
- **Soporte de fuentes TrueType** — incrustación de fuentes personalizadas con subconjuntos
- **Listo para CJK** — soporte completo de texto chino, japonés y coreano desde el primer día
- **Tablas** — encabezados, anchos de columna, filas alternadas, alineación vertical
- **Encabezados y pies de página** — con números de página, consistentes en todas las páginas
- **Listas** — listas con viñetas y numeradas
- **Códigos QR** — generación de QR en Go puro (niveles de corrección de errores)
- **Códigos de barras** — generación de Code 128
- **Decoraciones de texto** — subrayado, tachado, espaciado de letras, sangría
- **Números de página** — número de página automático y total de páginas
- **Integración con Go templates** — generar PDFs desde plantillas Go
- **Componentes reutilizables** — plantillas predefinidas de Factura, Informe y Carta
- **Esquema JSON** — definir documentos completamente en JSON
- **Múltiples unidades** — pt, mm, cm, in, em, %
- **Espacios de color** — RGB, escala de grises, CMYK
- **Imágenes** — incrustación de JPEG y PNG con opciones de ajuste
- **Metadatos del documento** — título, autor, asunto, creador

## Arquitectura

```
┌─────────────────────────────────────┐
│  gpdf (punto de entrada)            │
├─────────────────────────────────────┤
│  template  — API Builder, Cuadrícula│  Capa 3
├─────────────────────────────────────┤
│  document  — Nodos, Estilos, Layout │  Capa 2
├─────────────────────────────────────┤
│  pdf       — Writer, Fuentes, Flujos│  Capa 1
└─────────────────────────────────────┘
```

## Requisitos

- Go 1.22 o posterior

## Instalación

```bash
go get github.com/gpdf-dev/gpdf
```

## Inicio rápido

```go
package main

import (
	"os"

	"github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func main() {
	doc := gpdf.NewDocument(
		gpdf.WithPageSize(gpdf.A4),
		gpdf.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Hello, World!", template.FontSize(24), template.Bold())
		})
	})

	data, _ := doc.Generate()
	os.WriteFile("hello.pdf", data, 0644)
}
```

## Ejemplos

### Estilos de texto

Tamaño de fuente, peso, estilo, color, color de fondo y alineación:

```go
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("Título grande en negrita", template.FontSize(24), template.Bold())
		c.Text("Texto en cursiva", template.Italic())
		c.Text("Negrita + Cursiva", template.Bold(), template.Italic())
		c.Text("Texto rojo", template.TextColor(pdf.Red))
		c.Text("Color personalizado", template.TextColor(pdf.RGBHex(0x336699)))
		c.Text("Con fondo", template.BgColor(pdf.Yellow))
		c.Text("Centrado", template.AlignCenter())
		c.Text("Alineado a la derecha", template.AlignRight())
	})
})
```

### Cuadrícula de 12 columnas

Construya diseños usando una cuadrícula estilo Bootstrap de 12 columnas:

```go
// Dos columnas iguales
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(6, func(c *template.ColBuilder) {
		c.Text("Mitad izquierda")
	})
	r.Col(6, func(c *template.ColBuilder) {
		c.Text("Mitad derecha")
	})
})

// Barra lateral + contenido principal
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(3, func(c *template.ColBuilder) {
		c.Text("Barra lateral")
	})
	r.Col(9, func(c *template.ColBuilder) {
		c.Text("Contenido principal")
	})
})
```

### Filas de altura fija

Use `Row()` con una altura específica, o `AutoRow()` para altura basada en contenido:

```go
// Altura fija: 30mm
page.Row(document.Mm(30), func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("Esta fila tiene 30mm de alto")
	})
})
```

### Tablas

Tabla básica:

```go
c.Table(
	[]string{"Nombre", "Cant.", "Precio"},
	[][]string{
		{"Widget", "10", "$5.00"},
		{"Gadget", "3", "$12.00"},
	},
)
```

Tabla con estilos (colores de encabezado, anchos de columna, filas alternadas):

```go
c.Table(
	[]string{"Producto", "Categoría", "Cant.", "Precio Unit.", "Total"},
	[][]string{
		{"Laptop Pro 15", "Electrónica", "2", "$1,299.00", "$2,598.00"},
		{"Mouse Inalámbrico", "Accesorios", "10", "$29.99", "$299.90"},
	},
	template.ColumnWidths(30, 20, 10, 20, 20),
	template.TableHeaderStyle(
		template.TextColor(pdf.White),
		template.BgColor(pdf.RGBHex(0x1A237E)),
	),
	template.TableStripe(pdf.RGBHex(0xF5F5F5)),
)
```

### Imágenes

Incrustar imágenes JPEG y PNG con opciones de ajuste:

```go
c.Image(imgData)                                      // Tamaño por defecto
c.Image(imgData, template.FitWidth(document.Mm(80)))   // Ajustar al ancho
c.Image(imgData, template.FitHeight(document.Mm(30)))  // Ajustar a la altura
```

### Líneas y espaciadores

```go
c.Line()                                           // Por defecto (gris, 1pt)
c.Line(template.LineColor(pdf.Red))                 // Con color
c.Line(template.LineThickness(document.Pt(3)))      // Gruesa
c.Spacer(document.Mm(5))                            // Espacio vertical de 5mm
```

### Encabezados y pies de página

Defina encabezados y pies de página que se repiten en cada página:

```go
doc.Header(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("ACME Corporation", template.Bold(), template.FontSize(10))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Confidencial", template.AlignRight(), template.FontSize(10),
				template.TextColor(pdf.Gray(0.5)))
		})
	})
})

doc.Footer(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Generado con gpdf", template.AlignCenter(),
				template.FontSize(8), template.TextColor(pdf.Gray(0.5)))
		})
	})
})
```

### Componentes reutilizables

Genere tipos de documentos comunes con una sola llamada de función:

**Factura:**

```go
doc := template.Invoice(template.InvoiceData{
	Number:  "#INV-2026-001",
	Date:    "1 de marzo de 2026",
	DueDate: "31 de marzo de 2026",
	From:    template.InvoiceParty{Name: "ACME Corp", Address: []string{"Calle Principal 123"}},
	To:      template.InvoiceParty{Name: "Cliente S.A.", Address: []string{"Calle Secundaria 456"}},
	Items: []template.InvoiceItem{
		{Description: "Desarrollo Web", Quantity: "40 hrs", UnitPrice: 150, Amount: 6000},
		{Description: "Diseño UI/UX", Quantity: "20 hrs", UnitPrice: 120, Amount: 2400},
	},
	TaxRate: 10,
	Notes:   "¡Gracias por su preferencia!",
})
data, _ := doc.Generate()
```

**Informe:**

```go
doc := template.Report(template.ReportData{
	Title:    "Informe Trimestral",
	Subtitle: "Q1 2026",
	Author:   "ACME Corp",
	Sections: []template.ReportSection{
		{
			Title:   "Resumen Ejecutivo",
			Content: "Los ingresos aumentaron un 15% en comparación con Q4 2025.",
			Metrics: []template.ReportMetric{
				{Label: "Ingresos", Value: "$12.5M", ColorHex: 0x2E7D32},
				{Label: "Crecimiento", Value: "+15%", ColorHex: 0x2E7D32},
			},
		},
		{
			Title: "Desglose de Ingresos",
			Table: &template.ReportTable{
				Header: []string{"División", "Q1 2026", "Cambio"},
				Rows:   [][]string{{"Nube", "$5.2M", "+26.8%"}, {"Empresa", "$3.8M", "+8.6%"}},
			},
		},
	},
})
```

**Carta:**

```go
doc := template.Letter(template.LetterData{
	From:     template.LetterParty{Name: "ACME Corp", Address: []string{"Calle Principal 123"}},
	To:       template.LetterParty{Name: "Sr. Juan García", Address: []string{"Calle Secundaria 456"}},
	Date:     "1 de marzo de 2026",
	Subject:  "Propuesta de Alianza",
	Greeting: "Estimado Sr. García,",
	Body:     []string{"Nos dirigimos a usted para proponer una alianza estratégica..."},
	Closing:  "Atentamente,",
	Signature: "María López",
})
```

### Metadatos del documento

```go
doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithMetadata(document.DocumentMetadata{
		Title:   "Informe Anual 2026",
		Author:  "gpdf Library",
		Subject: "Ejemplo de metadatos del documento",
		Creator: "Mi Aplicación",
	}),
)
```

### Tamaños de página y márgenes

```go
// Tamaños de página disponibles
document.A4      // 210mm x 297mm
document.A3      // 297mm x 420mm
document.Letter  // 8.5in x 11in
document.Legal   // 8.5in x 14in

// Márgenes uniformes
template.WithMargins(document.UniformEdges(document.Mm(20)))

// Márgenes asimétricos
template.WithMargins(document.Edges{
	Top:    document.Mm(10),
	Right:  document.Mm(40),
	Bottom: document.Mm(10),
	Left:   document.Mm(40),
})
```

### Opciones de salida

```go
// Generate devuelve []byte
data, err := doc.Generate()

// Render escribe en cualquier io.Writer
var buf bytes.Buffer
err := doc.Render(&buf)

// Escribir directamente a un archivo
f, _ := os.Create("output.pdf")
defer f.Close()
doc.Render(f)
```

## Referencia API

### Opciones del documento

| Función | Descripción |
|---|---|
| `WithPageSize(size)` | Establecer tamaño de página (A4, A3, Letter, Legal) |
| `WithMargins(edges)` | Establecer márgenes de página |
| `WithFont(family, data)` | Registrar una fuente TrueType |
| `WithDefaultFont(family, size)` | Establecer la fuente predeterminada |
| `WithMetadata(meta)` | Establecer metadatos del documento |

### Contenido de columna

| Método | Descripción |
|---|---|
| `c.Text(text, opts...)` | Agregar texto con opciones de estilo |
| `c.Table(header, rows, opts...)` | Agregar una tabla |
| `c.Image(data, opts...)` | Agregar una imagen (JPEG/PNG) |
| `c.QRCode(data, opts...)` | Agregar un código QR |
| `c.Barcode(data, opts...)` | Agregar un código de barras (Code 128) |
| `c.List(items, opts...)` | Agregar lista con viñetas |
| `c.OrderedList(items, opts...)` | Agregar lista numerada |
| `c.PageNumber(opts...)` | Agregar número de página actual |
| `c.TotalPages(opts...)` | Agregar total de páginas |
| `c.Line(opts...)` | Agregar una línea horizontal |
| `c.Spacer(height)` | Agregar espacio vertical |

### Opciones de texto

| Opción | Descripción |
|---|---|
| `template.FontSize(size)` | Tamaño de fuente en puntos |
| `template.Bold()` | Negrita |
| `template.Italic()` | Cursiva |
| `template.FontFamily(name)` | Usar fuente registrada |
| `template.TextColor(color)` | Color del texto |
| `template.BgColor(color)` | Color de fondo |
| `template.Underline()` | Decoración de subrayado |
| `template.Strikethrough()` | Decoración de tachado |
| `template.LetterSpacing(pts)` | Espaciado de letras en puntos |
| `template.TextIndent(value)` | Sangría de primera línea |
| `template.AlignLeft()` | Alineación izquierda (por defecto) |
| `template.AlignCenter()` | Alineación centrada |
| `template.AlignRight()` | Alineación derecha |

### Opciones de tabla

| Opción | Descripción |
|---|---|
| `template.ColumnWidths(w...)` | Anchos de columna en porcentaje |
| `template.TableHeaderStyle(opts...)` | Estilo de la fila de encabezado |
| `template.TableStripe(color)` | Color de filas alternadas |
| `template.TableCellVAlign(align)` | Alineación vertical de celda (Top/Middle/Bottom) |

### Opciones de imagen

| Opción | Descripción |
|---|---|
| `template.FitWidth(value)` | Escalar al ancho (mantiene proporción) |
| `template.FitHeight(value)` | Escalar a la altura (mantiene proporción) |

### Opciones de código QR

| Opción | Descripción |
|---|---|
| `template.QRSize(value)` | Tamaño del código QR |
| `template.QRErrorCorrection(level)` | Nivel de corrección de errores (L/M/Q/H) |
| `template.QRScale(n)` | Factor de escala del módulo |

### Opciones de código de barras

| Opción | Descripción |
|---|---|
| `template.BarcodeWidth(value)` | Ancho del código de barras |
| `template.BarcodeHeight(value)` | Altura del código de barras |
| `template.BarcodeFormat(fmt)` | Formato del código de barras (Code 128) |

### Generación de plantillas

| Función | Descripción |
|---|---|
| `template.FromJSON(schema, data)` | Generar documento desde esquema JSON |
| `template.FromTemplate(tmpl, data)` | Generar documento desde plantilla Go |
| `template.TemplateFuncMap()` | Obtener funciones auxiliares de plantilla (incluye `toJSON`) |

### Opciones de línea

| Opción | Descripción |
|---|---|
| `template.LineColor(color)` | Color de la línea |
| `template.LineThickness(value)` | Grosor de la línea |

### Unidades

```go
document.Pt(72)    // Puntos (1/72 pulgada)
document.Mm(10)    // Milímetros
document.Cm(2.5)   // Centímetros
document.In(1)     // Pulgadas
document.Em(1.5)   // Relativo al tamaño de fuente
document.Pct(50)   // Porcentaje
```

### Colores

```go
pdf.RGB(0.2, 0.4, 0.8)   // RGB (0.0–1.0)
pdf.RGBHex(0xFF5733)      // RGB hexadecimal
pdf.Gray(0.5)             // Escala de grises
pdf.CMYK(0, 0.5, 1, 0)   // CMYK

// Colores predefinidos
pdf.Black, pdf.White, pdf.Red, pdf.Green, pdf.Blue
pdf.Yellow, pdf.Cyan, pdf.Magenta
```

## Benchmark

Comparación con [go-pdf/fpdf](https://github.com/go-pdf/fpdf), [signintech/gopdf](https://github.com/signintech/gopdf) y [maroto v2](https://github.com/johnfercher/maroto).
Mediana de 5 ejecuciones, 100 iteraciones cada una. Apple M1, Go 1.25.

**Tiempo de ejecución** (menor es mejor):

| Benchmark | gpdf | fpdf | gopdf | maroto v2 |
|---|--:|--:|--:|--:|
| Página única | **13 µs** | 132 µs | 423 µs | 237 µs |
| Tabla (4x10) | **108 µs** | 241 µs | 835 µs | 8.6 ms |
| 100 páginas | **683 µs** | 11.7 ms | 8.6 ms | 19.8 ms |
| Documento complejo | **133 µs** | 254 µs | 997 µs | 10.4 ms |

**Uso de memoria** (menor es mejor):

| Benchmark | gpdf | fpdf | gopdf | maroto v2 |
|---|--:|--:|--:|--:|
| Página única | **16 KB** | 1.2 MB | 1.8 MB | 61 KB |
| Tabla (4x10) | **209 KB** | 1.3 MB | 1.9 MB | 1.6 MB |
| 100 páginas | **909 KB** | 121 MB | 83 MB | 4.0 MB |
| Documento complejo | **246 KB** | 1.3 MB | 2.0 MB | 2.0 MB |

### ¿Por qué gpdf es rápido?

- **Página única** — Pipeline de un solo paso: construir→componer→renderizar, sin estructuras de datos intermedias. Usa tipos struct concretos (sin boxing de `interface{}`), construyendo el árbol del documento con asignaciones de heap mínimas.
- **Tabla** — El contenido de las celdas se escribe directamente como comandos de flujo de contenido PDF a través de un buffer `strings.Builder` reutilizable. Sin envoltura de objetos por celda ni búsquedas de fuentes repetidas; la fuente se resuelve una vez por documento.
- **100 páginas** — El layout escala linealmente O(n). La paginación por desbordamiento pasa los nodos restantes por referencia de slice (sin copias profundas). La fuente se parsea una vez y se comparte entre todas las páginas.
- **Documento complejo** — El layout de un solo paso sin re-medición combina todas las ventajas anteriores. El subsetting de fuentes incrusta solo los glifos utilizados, y la compresión Flate se aplica por defecto, manteniendo pequeños tanto la memoria como el tamaño de salida.

Ejecutar benchmarks:

```bash
cd _benchmark && go test -bench=. -benchmem -count=5
```

## Licencia

MIT
