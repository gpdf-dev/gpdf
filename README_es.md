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
- **Soporte de fuentes TrueType** — incrustación de fuentes personalizadas con subconjuntos automáticos
- **Listo para CJK** — soporte completo de Unicode incluyendo texto chino, japonés y coreano
- **Texto enriquecido** — mezclar múltiples estilos en línea (negrita, cursiva, colores) en un solo párrafo
- **Tablas** — encabezados, anchos de columna, filas alternadas, alineación vertical
- **Encabezados y pies de página** — con números de página, consistentes en todas las páginas
- **Listas** — listas con viñetas y numeradas con sangría configurable
- **Códigos QR** — generación de QR en Go puro (versiones 1-40, corrección de errores L/M/Q/H)
- **Códigos de barras** — generación de Code 128
- **Decoraciones de texto** — subrayado, tachado, espaciado de letras, sangría
- **Números de página** — número de página automático y total de páginas
- **Integración con Go templates** — generar PDFs desde plantillas Go
- **Componentes reutilizables** — plantillas predefinidas de Factura, Informe y Carta
- **Esquema JSON** — definir documentos completamente en JSON
- **Múltiples unidades** — pt, mm, cm, in, em, %
- **Espacios de color** — RGB, escala de grises, CMYK
- **Imágenes** — incrustación de JPEG y PNG con opciones de ajuste
- **Compresión Flate** — compresión automática de flujos PDF para archivos más pequeños
- **Subconjuntos de fuentes** — solo incrusta los glifos utilizados, reduciendo el tamaño de salida
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

### Fuentes personalizadas

Incrustar fuentes TrueType para tipografía personalizada y texto CJK:

```go
fontData, _ := os.ReadFile("NotoSans-Regular.ttf")
boldData, _ := os.ReadFile("NotoSans-Bold.ttf")

doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithFont("NotoSans", fontData),
	gpdf.WithFont("NotoSans-Bold", boldData),
	gpdf.WithDefaultFont("NotoSans", 12),
)

page := doc.AddPage()
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("日本語テキスト — gpdf は CJK をフルサポート")
		c.Text("中文文本 — 支持中日韩文字")
		c.Text("한국어 텍스트 — CJK 완벽 지원")
		c.Text("Título en negrita", template.FontFamily("NotoSans-Bold"), template.FontSize(18))
	})
})
```

Solo se incrustan los glifos realmente utilizados (subconjuntos automáticos de fuentes), manteniendo los archivos de salida pequeños.

### Texto enriquecido

Mezcle múltiples estilos dentro de un solo párrafo usando `RichText`:

```go
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.RichText(func(rt *template.RichTextBuilder) {
			rt.Span("Esto es ")
			rt.Span("negrita", template.Bold())
			rt.Span(" y esto es ")
			rt.Span("cursiva roja", template.Italic(), template.TextColor(pdf.Red))
			rt.Span(" en un solo párrafo.")
		})
	})
})
```

Las opciones a nivel de bloque (alineación, sangría) se pueden pasar como argumentos adicionales a `RichText`:

```go
c.RichText(func(rt *template.RichTextBuilder) {
	rt.Span("Texto mixto centrado: ")
	rt.Span("$1,234.56", template.Bold(), template.TextColor(pdf.RGBHex(0x2E7D32)))
}, template.AlignCenter())
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

### Listas

Listas con viñetas y numeradas:

```go
// Lista con viñetas
c.List([]string{"Primer elemento", "Segundo elemento", "Tercer elemento"})

// Lista numerada
c.OrderedList([]string{"Paso uno", "Paso dos", "Paso tres"})

// Sangría personalizada
c.List([]string{"Con sangría", "Elementos"}, template.ListIndent(document.Mm(10)))
```

### Códigos QR

Generación de códigos QR con tamaño y corrección de errores configurables:

```go
// Código QR básico
c.QRCode("https://gpdf.dev")

// Tamaño y nivel de corrección personalizados
c.QRCode("https://gpdf.dev",
	template.QRSize(document.Mm(30)),
	template.QRErrorCorrection(qrcode.LevelH))
```

### Códigos de barras

Generación de códigos de barras Code 128:

```go
// Código de barras básico
c.Barcode("INV-2026-0001")

// Ancho personalizado
c.Barcode("INV-2026-0001", template.BarcodeWidth(document.Mm(80)))
```

### Números de página

Números de página automáticos y total de páginas:

```go
doc.Footer(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Generado con gpdf", template.FontSize(8))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.PageNumber(template.AlignRight(), template.FontSize(8))
		})
	})
})

doc.Header(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.TotalPages(template.AlignRight(), template.FontSize(9))
		})
	})
})
```

### Decoraciones de texto

Subrayado, tachado, espaciado de letras y sangría:

```go
c.Text("Texto subrayado", template.Underline())
c.Text("Texto tachado", template.Strikethrough())
c.Text("Espaciado amplio", template.LetterSpacing(3))
c.Text("Párrafo con sangría...", template.TextIndent(document.Pt(24)))
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

### Documentos de múltiples páginas

```go
for i := 1; i <= 5; i++ {
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Contenido de la página")
		})
	})
}
```

### Esquema JSON

Defina documentos completamente en JSON:

```go
schema := []byte(`{
	"page": {"size": "A4", "margins": "20mm"},
	"metadata": {"title": "Informe", "author": "gpdf"},
	"body": [
		{"row": {"cols": [
			{"span": 12, "text": "Hola desde JSON", "style": {"size": 24, "bold": true}}
		]}}
	]
}`)

doc, err := template.FromJSON(schema, nil)
data, _ := doc.Generate()
```

### Integración con Go templates

Use plantillas Go con esquema JSON para contenido dinámico:

```go
schema := []byte(`{
	"page": {"size": "A4", "margins": "20mm"},
	"metadata": {"title": "{{.Title}}"},
	"body": [
		{"row": {"cols": [
			{"span": 12, "text": "{{.Title}}", "style": {"size": 24, "bold": true}}
		]}}
	]
}`)

data := map[string]any{"Title": "Informe Dinámico"}
doc, err := template.FromJSON(schema, data)
```

Para más control, use una plantilla Go pre-parseada:

```go
tmpl, _ := gotemplate.New("doc").Funcs(template.TemplateFuncMap()).Parse(schemaStr)
doc, err := template.FromTemplate(tmpl, data)
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
	TaxRate:  10,
	Currency: "€",  // por defecto: "$"
	Notes:    "¡Gracias por su preferencia!",
	Payment: &template.InvoicePayment{
		BankName: "Banco Santander",
		Account:  "ES12 3456 7890 1234",
		Routing:  "BSCHESMM",
	},
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
| `c.RichText(fn, opts...)` | Agregar múltiples estilos en línea en un párrafo |
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

### Opciones de lista

| Opción | Descripción |
|---|---|
| `template.ListIndent(value)` | Ancho de sangría de viñeta/número |

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

### Constructor de texto enriquecido

| Método | Descripción |
|---|---|
| `rt.Span(text, opts...)` | Agregar fragmento de texto en línea con estilo |

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

## Avanzado: APIs de bajo nivel

El paquete `template` (Capa 3) cubre la mayoría de los casos de uso. Para control total, puede usar los paquetes de bajo nivel directamente:

| Paquete | Capa | Descripción |
|---|---|---|
| `template` | 3 | API declarativa de constructores, sistema de cuadrícula, componentes |
| `document` | 2 | Árbol de nodos, modelo de caja, estilos, motor de layout |
| `pdf` | 1 | PDF Writer, análisis de fuentes TrueType, flujos, imágenes |
| `qrcode` | — | Codificador independiente de QR (versiones 1-40) |
| `barcode` | — | Codificador independiente de Code 128 |

**Capa 2 (document)** proporciona acceso a:
- Modelo de caja con márgenes, relleno, bordes (sólido/punteado/líneas)
- Control de saltos de página (`BreakPolicy` — avoid, always, page)
- Celdas de tabla con `ColSpan` / `RowSpan`
- Modos de ajuste de imagen (`FitContain`, `FitCover`, `FitStretch`, `FitOriginal`)
- Dirección de layout vertical/horizontal

**Capa 1 (pdf)** proporciona acceso a:
- Escritura de objetos PDF sin procesar (`Writer`)
- Análisis y subconjuntos de fuentes TrueType
- Registro de imágenes JPEG/PNG
- Compresión Flate de flujos
- Operadores de flujo de contenido

Consulte [GoDoc](https://pkg.go.dev/github.com/gpdf-dev/gpdf) para detalles completos de la API.

## Licencia

MIT
