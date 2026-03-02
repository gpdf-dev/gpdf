# gpdf

[![Go Reference](https://pkg.go.dev/badge/github.com/gpdf-dev/gpdf.svg)](https://pkg.go.dev/github.com/gpdf-dev/gpdf)
[![CI](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml/badge.svg)](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gpdf-dev/gpdf)](https://goreportcard.com/report/github.com/gpdf-dev/gpdf)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.22-blue)](https://go.dev/)

[English](README.md) | [日本語](README_ja.md) | [中文](README_zh.md) | [한국어](README_ko.md) | [Español](README_es.md) | **Português**

Biblioteca de geração de PDF em Go puro, sem dependências externas, com arquitetura em camadas e API declarativa de construtores.

## Características

- **Zero dependências** — apenas a biblioteca padrão do Go
- **Arquitetura em camadas** — primitivas PDF de baixo nível, modelo de documento e API de templates de alto nível
- **Sistema de grade de 12 colunas** — layout responsivo estilo Bootstrap
- **Suporte a fontes TrueType** — incorporação de fontes personalizadas com subconjuntos
- **Pronto para CJK** — suporte completo a texto chinês, japonês e coreano desde o primeiro dia
- **Tabelas** — cabeçalhos, larguras de coluna, linhas alternadas, alinhamento vertical
- **Cabeçalhos e rodapés** — com números de página, consistentes em todas as páginas
- **Listas** — listas com marcadores e numeradas
- **QR codes** — geração de QR code em Go puro (níveis de correção de erros)
- **Códigos de barras** — geração de Code 128
- **Decorações de texto** — sublinhado, tachado, espaçamento de letras, recuo
- **Números de página** — número de página automático e total de páginas
- **Integração com Go templates** — gerar PDFs a partir de templates Go
- **Componentes reutilizáveis** — templates predefinidos de Fatura, Relatório e Carta
- **Esquema JSON** — definir documentos inteiramente em JSON
- **Múltiplas unidades** — pt, mm, cm, in, em, %
- **Espaços de cor** — RGB, escala de cinza, CMYK
- **Imagens** — incorporação de JPEG e PNG com opções de ajuste
- **Metadados do documento** — título, autor, assunto, criador

## Arquitetura

```
┌─────────────────────────────────────┐
│  gpdf (ponto de entrada)            │
├─────────────────────────────────────┤
│  template  — API Builder, Grade     │  Camada 3
├─────────────────────────────────────┤
│  document  — Nós, Estilos, Layout   │  Camada 2
├─────────────────────────────────────┤
│  pdf       — Writer, Fontes, Fluxos │  Camada 1
└─────────────────────────────────────┘
```

## Requisitos

- Go 1.22 ou posterior

## Instalação

```bash
go get github.com/gpdf-dev/gpdf
```

## Início rápido

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

## Exemplos

### Estilos de texto

Tamanho da fonte, peso, estilo, cor, cor de fundo e alinhamento:

```go
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("Título grande em negrito", template.FontSize(24), template.Bold())
		c.Text("Texto em itálico", template.Italic())
		c.Text("Negrito + Itálico", template.Bold(), template.Italic())
		c.Text("Texto vermelho", template.TextColor(pdf.Red))
		c.Text("Cor personalizada", template.TextColor(pdf.RGBHex(0x336699)))
		c.Text("Com fundo", template.BgColor(pdf.Yellow))
		c.Text("Centralizado", template.AlignCenter())
		c.Text("Alinhado à direita", template.AlignRight())
	})
})
```

### Grade de 12 colunas

Construa layouts usando uma grade estilo Bootstrap de 12 colunas:

```go
// Duas colunas iguais
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(6, func(c *template.ColBuilder) {
		c.Text("Metade esquerda")
	})
	r.Col(6, func(c *template.ColBuilder) {
		c.Text("Metade direita")
	})
})

// Barra lateral + conteúdo principal
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(3, func(c *template.ColBuilder) {
		c.Text("Barra lateral")
	})
	r.Col(9, func(c *template.ColBuilder) {
		c.Text("Conteúdo principal")
	})
})
```

### Linhas de altura fixa

Use `Row()` com uma altura específica, ou `AutoRow()` para altura baseada em conteúdo:

```go
// Altura fixa: 30mm
page.Row(document.Mm(30), func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("Esta linha tem 30mm de altura")
	})
})
```

### Tabelas

Tabela básica:

```go
c.Table(
	[]string{"Nome", "Qtd.", "Preço"},
	[][]string{
		{"Widget", "10", "R$25,00"},
		{"Gadget", "3", "R$60,00"},
	},
)
```

Tabela com estilos (cores do cabeçalho, larguras de coluna, linhas alternadas):

```go
c.Table(
	[]string{"Produto", "Categoria", "Qtd.", "Preço Unit.", "Total"},
	[][]string{
		{"Laptop Pro 15", "Eletrônicos", "2", "R$6.495,00", "R$12.990,00"},
		{"Mouse Sem Fio", "Acessórios", "10", "R$149,90", "R$1.499,00"},
	},
	template.ColumnWidths(30, 20, 10, 20, 20),
	template.TableHeaderStyle(
		template.TextColor(pdf.White),
		template.BgColor(pdf.RGBHex(0x1A237E)),
	),
	template.TableStripe(pdf.RGBHex(0xF5F5F5)),
)
```

### Imagens

Incorporar imagens JPEG e PNG com opções de ajuste:

```go
c.Image(imgData)                                      // Tamanho padrão
c.Image(imgData, template.FitWidth(document.Mm(80)))   // Ajustar à largura
c.Image(imgData, template.FitHeight(document.Mm(30)))  // Ajustar à altura
```

### Linhas e espaçadores

```go
c.Line()                                           // Padrão (cinza, 1pt)
c.Line(template.LineColor(pdf.Red))                 // Com cor
c.Line(template.LineThickness(document.Pt(3)))      // Grossa
c.Spacer(document.Mm(5))                            // Espaço vertical de 5mm
```

### Cabeçalhos e rodapés

Defina cabeçalhos e rodapés que se repetem em cada página:

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
			c.Text("Gerado com gpdf", template.AlignCenter(),
				template.FontSize(8), template.TextColor(pdf.Gray(0.5)))
		})
	})
})
```

### Componentes reutilizáveis

Gere tipos de documentos comuns com uma única chamada de função:

**Fatura:**

```go
doc := template.Invoice(template.InvoiceData{
	Number:  "#INV-2026-001",
	Date:    "1 de março de 2026",
	DueDate: "31 de março de 2026",
	From:    template.InvoiceParty{Name: "ACME Corp", Address: []string{"Rua Principal 123"}},
	To:      template.InvoiceParty{Name: "Cliente Ltda.", Address: []string{"Rua Secundária 456"}},
	Items: []template.InvoiceItem{
		{Description: "Desenvolvimento Web", Quantity: "40 hrs", UnitPrice: 150, Amount: 6000},
		{Description: "Design UI/UX", Quantity: "20 hrs", UnitPrice: 120, Amount: 2400},
	},
	TaxRate: 10,
	Notes:   "Obrigado pela preferência!",
})
data, _ := doc.Generate()
```

**Relatório:**

```go
doc := template.Report(template.ReportData{
	Title:    "Relatório Trimestral",
	Subtitle: "Q1 2026",
	Author:   "ACME Corp",
	Sections: []template.ReportSection{
		{
			Title:   "Resumo Executivo",
			Content: "A receita aumentou 15% em comparação com o Q4 2025.",
			Metrics: []template.ReportMetric{
				{Label: "Receita", Value: "R$12.5M", ColorHex: 0x2E7D32},
				{Label: "Crescimento", Value: "+15%", ColorHex: 0x2E7D32},
			},
		},
		{
			Title: "Detalhamento da Receita",
			Table: &template.ReportTable{
				Header: []string{"Divisão", "Q1 2026", "Variação"},
				Rows:   [][]string{{"Nuvem", "R$5.2M", "+26.8%"}, {"Corporativo", "R$3.8M", "+8.6%"}},
			},
		},
	},
})
```

**Carta:**

```go
doc := template.Letter(template.LetterData{
	From:     template.LetterParty{Name: "ACME Corp", Address: []string{"Rua Principal 123"}},
	To:       template.LetterParty{Name: "Sr. João Silva", Address: []string{"Rua Secundária 456"}},
	Date:     "1 de março de 2026",
	Subject:  "Proposta de Parceria",
	Greeting: "Prezado Sr. Silva,",
	Body:     []string{"Estamos escrevendo para propor uma parceria estratégica..."},
	Closing:  "Atenciosamente,",
	Signature: "Maria Santos",
})
```

### Metadados do documento

```go
doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithMetadata(document.DocumentMetadata{
		Title:   "Relatório Anual 2026",
		Author:  "gpdf Library",
		Subject: "Exemplo de metadados do documento",
		Creator: "Minha Aplicação",
	}),
)
```

### Tamanhos de página e margens

```go
// Tamanhos de página disponíveis
document.A4      // 210mm x 297mm
document.A3      // 297mm x 420mm
document.Letter  // 8.5in x 11in
document.Legal   // 8.5in x 14in

// Margens uniformes
template.WithMargins(document.UniformEdges(document.Mm(20)))

// Margens assimétricas
template.WithMargins(document.Edges{
	Top:    document.Mm(10),
	Right:  document.Mm(40),
	Bottom: document.Mm(10),
	Left:   document.Mm(40),
})
```

### Opções de saída

```go
// Generate retorna []byte
data, err := doc.Generate()

// Render escreve em qualquer io.Writer
var buf bytes.Buffer
err := doc.Render(&buf)

// Escrever diretamente em um arquivo
f, _ := os.Create("output.pdf")
defer f.Close()
doc.Render(f)
```

## Referência API

### Opções do documento

| Função | Descrição |
|---|---|
| `WithPageSize(size)` | Definir tamanho da página (A4, A3, Letter, Legal) |
| `WithMargins(edges)` | Definir margens da página |
| `WithFont(family, data)` | Registrar uma fonte TrueType |
| `WithDefaultFont(family, size)` | Definir a fonte padrão |
| `WithMetadata(meta)` | Definir metadados do documento |

### Conteúdo da coluna

| Método | Descrição |
|---|---|
| `c.Text(text, opts...)` | Adicionar texto com opções de estilo |
| `c.Table(header, rows, opts...)` | Adicionar uma tabela |
| `c.Image(data, opts...)` | Adicionar uma imagem (JPEG/PNG) |
| `c.QRCode(data, opts...)` | Adicionar QR code |
| `c.Barcode(data, opts...)` | Adicionar código de barras (Code 128) |
| `c.List(items, opts...)` | Adicionar lista com marcadores |
| `c.OrderedList(items, opts...)` | Adicionar lista numerada |
| `c.PageNumber(opts...)` | Adicionar número de página atual |
| `c.TotalPages(opts...)` | Adicionar total de páginas |
| `c.Line(opts...)` | Adicionar uma linha horizontal |
| `c.Spacer(height)` | Adicionar espaço vertical |

### Opções de texto

| Opção | Descrição |
|---|---|
| `template.FontSize(size)` | Tamanho da fonte em pontos |
| `template.Bold()` | Negrito |
| `template.Italic()` | Itálico |
| `template.FontFamily(name)` | Usar fonte registrada |
| `template.TextColor(color)` | Cor do texto |
| `template.BgColor(color)` | Cor de fundo |
| `template.Underline()` | Decoração de sublinhado |
| `template.Strikethrough()` | Decoração de tachado |
| `template.LetterSpacing(pts)` | Espaçamento de letras em pontos |
| `template.TextIndent(value)` | Recuo de primeira linha |
| `template.AlignLeft()` | Alinhamento à esquerda (padrão) |
| `template.AlignCenter()` | Alinhamento centralizado |
| `template.AlignRight()` | Alinhamento à direita |

### Opções de tabela

| Opção | Descrição |
|---|---|
| `template.ColumnWidths(w...)` | Larguras de coluna em porcentagem |
| `template.TableHeaderStyle(opts...)` | Estilo da linha de cabeçalho |
| `template.TableStripe(color)` | Cor de linhas alternadas |
| `template.TableCellVAlign(align)` | Alinhamento vertical da célula (Top/Middle/Bottom) |

### Opções de imagem

| Opção | Descrição |
|---|---|
| `template.FitWidth(value)` | Escalar à largura (mantém proporção) |
| `template.FitHeight(value)` | Escalar à altura (mantém proporção) |

### Opções de QR code

| Opção | Descrição |
|---|---|
| `template.QRSize(value)` | Tamanho do QR code |
| `template.QRErrorCorrection(level)` | Nível de correção de erros (L/M/Q/H) |
| `template.QRScale(n)` | Fator de escala do módulo |

### Opções de código de barras

| Opção | Descrição |
|---|---|
| `template.BarcodeWidth(value)` | Largura do código de barras |
| `template.BarcodeHeight(value)` | Altura do código de barras |
| `template.BarcodeFormat(fmt)` | Formato do código de barras (Code 128) |

### Geração de templates

| Função | Descrição |
|---|---|
| `template.FromJSON(schema, data)` | Gerar documento a partir de esquema JSON |
| `template.FromTemplate(tmpl, data)` | Gerar documento a partir de template Go |
| `template.TemplateFuncMap()` | Obter funções auxiliares de template (inclui `toJSON`) |

### Opções de linha

| Opção | Descrição |
|---|---|
| `template.LineColor(color)` | Cor da linha |
| `template.LineThickness(value)` | Espessura da linha |

### Unidades

```go
document.Pt(72)    // Pontos (1/72 polegada)
document.Mm(10)    // Milímetros
document.Cm(2.5)   // Centímetros
document.In(1)     // Polegadas
document.Em(1.5)   // Relativo ao tamanho da fonte
document.Pct(50)   // Porcentagem
```

### Cores

```go
pdf.RGB(0.2, 0.4, 0.8)   // RGB (0.0–1.0)
pdf.RGBHex(0xFF5733)      // RGB hexadecimal
pdf.Gray(0.5)             // Escala de cinza
pdf.CMYK(0, 0.5, 1, 0)   // CMYK

// Cores predefinidas
pdf.Black, pdf.White, pdf.Red, pdf.Green, pdf.Blue
pdf.Yellow, pdf.Cyan, pdf.Magenta
```

## Benchmark

Comparação com [go-pdf/fpdf](https://github.com/go-pdf/fpdf), [signintech/gopdf](https://github.com/signintech/gopdf) e [maroto v2](https://github.com/johnfercher/maroto).
Mediana de 5 execuções, 100 iterações cada. Apple M1, Go 1.25.

**Tempo de execução** (menor é melhor):

| Benchmark | gpdf | fpdf | gopdf | maroto v2 |
|---|--:|--:|--:|--:|
| Página única | **13 µs** | 132 µs | 423 µs | 237 µs |
| Tabela (4x10) | **108 µs** | 241 µs | 835 µs | 8.6 ms |
| 100 páginas | **683 µs** | 11.7 ms | 8.6 ms | 19.8 ms |
| Documento complexo | **133 µs** | 254 µs | 997 µs | 10.4 ms |

**Uso de memória** (menor é melhor):

| Benchmark | gpdf | fpdf | gopdf | maroto v2 |
|---|--:|--:|--:|--:|
| Página única | **16 KB** | 1.2 MB | 1.8 MB | 61 KB |
| Tabela (4x10) | **209 KB** | 1.3 MB | 1.9 MB | 1.6 MB |
| 100 páginas | **909 KB** | 121 MB | 83 MB | 4.0 MB |
| Documento complexo | **246 KB** | 1.3 MB | 2.0 MB | 2.0 MB |

### Por que o gpdf é rápido?

- **Página única** — Pipeline de passagem única: construir→compor→renderizar, sem estruturas de dados intermediárias. Usa tipos struct concretos (sem boxing de `interface{}`), construindo a árvore do documento com alocações de heap mínimas.
- **Tabela** — O conteúdo das células é escrito diretamente como comandos de fluxo de conteúdo PDF através de um buffer `strings.Builder` reutilizável. Sem encapsulamento de objetos por célula ou buscas de fontes repetidas; a fonte é resolvida uma vez por documento.
- **100 páginas** — O layout escala linearmente O(n). A paginação por overflow passa os nós restantes por referência de slice (sem cópias profundas). A fonte é parseada uma vez e compartilhada entre todas as páginas.
- **Documento complexo** — O layout de passagem única sem re-medição combina todas as vantagens acima. O subsetting de fontes incorpora apenas os glifos utilizados, e a compressão Flate é aplicada por padrão, mantendo pequenos tanto a memória quanto o tamanho de saída.

Executar benchmarks:

```bash
cd _benchmark && go test -bench=. -benchmem -count=5
```

## Licença

MIT
