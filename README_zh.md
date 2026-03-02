# gpdf

[![Go Reference](https://pkg.go.dev/badge/github.com/gpdf-dev/gpdf.svg)](https://pkg.go.dev/github.com/gpdf-dev/gpdf)
[![CI](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml/badge.svg)](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gpdf-dev/gpdf)](https://goreportcard.com/report/github.com/gpdf-dev/gpdf)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.22-blue)](https://go.dev/)

[English](README.md) | [日本語](README_ja.md) | **中文** | [한국어](README_ko.md) | [Español](README_es.md) | [Português](README_pt.md)

纯 Go 实现的零依赖 PDF 生成库，采用分层架构和声明式构建器 API。

## 特性

- **零依赖** — 仅使用 Go 标准库
- **分层架构** — 底层 PDF 原语、文档模型和高层模板 API
- **12 列网格系统** — Bootstrap 风格的响应式布局
- **TrueType 字体支持** — 自定义字体嵌入与自动子集化
- **CJK 就绪** — 完整的 Unicode 支持，包括中文、日文和韩文
- **富文本** — 在单个段落中混合多种内联样式（粗体、斜体、颜色）
- **表格** — 表头、列宽、条纹行、垂直对齐
- **页眉和页脚** — 带页码，所有页面统一显示
- **列表** — 可配置缩进的无序列表和有序列表
- **二维码** — 纯 Go 二维码生成（版本 1-40，纠错等级 L/M/Q/H）
- **条形码** — Code 128 条形码生成
- **文本装饰** — 下划线、删除线、字间距、首行缩进
- **页码** — 自动页码和总页数
- **Go 模板集成** — 从 Go 模板生成 PDF
- **可复用组件** — 内置发票、报告和信函预设模板
- **JSON 模式** — 完全用 JSON 定义文档
- **多种单位** — pt、mm、cm、in、em、%
- **色彩空间** — RGB、灰度、CMYK
- **图片** — JPEG 和 PNG 嵌入（支持缩放选项）
- **Flate 压缩** — 自动 PDF 流压缩，减小文件大小
- **字体子集化** — 仅嵌入实际使用的字形，减小输出大小
- **文档元数据** — 标题、作者、主题、创建者

## 架构

```
┌─────────────────────────────────────┐
│  gpdf（入口点）                       │
├─────────────────────────────────────┤
│  template  — 构建器 API、网格         │  第 3 层
├─────────────────────────────────────┤
│  document  — 节点、样式、布局         │  第 2 层
├─────────────────────────────────────┤
│  pdf       — Writer、字体、流         │  第 1 层
└─────────────────────────────────────┘
```

## 环境要求

- Go 1.22 或更高版本

## 安装

```bash
go get github.com/gpdf-dev/gpdf
```

## 快速入门

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

## 使用示例

### 文本样式

字号、字重、样式、颜色、背景色和对齐方式：

```go
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("大号粗体标题", template.FontSize(24), template.Bold())
		c.Text("斜体文本", template.Italic())
		c.Text("粗体 + 斜体", template.Bold(), template.Italic())
		c.Text("红色文本", template.TextColor(pdf.Red))
		c.Text("自定义颜色", template.TextColor(pdf.RGBHex(0x336699)))
		c.Text("带背景色", template.BgColor(pdf.Yellow))
		c.Text("居中对齐", template.AlignCenter())
		c.Text("右对齐", template.AlignRight())
	})
})
```

### 自定义字体

嵌入 TrueType 字体以支持自定义排版和 CJK 文本：

```go
fontData, _ := os.ReadFile("NotoSansSC-Regular.ttf")
boldData, _ := os.ReadFile("NotoSansSC-Bold.ttf")

doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithFont("NotoSansSC", fontData),
	gpdf.WithFont("NotoSansSC-Bold", boldData),
	gpdf.WithDefaultFont("NotoSansSC", 12),
)

page := doc.AddPage()
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("中文文本 — gpdf 完整支持 CJK")
		c.Text("日本語テキスト — CJK をフルサポート")
		c.Text("한국어 텍스트 — CJK 완벽 지원")
		c.Text("粗体标题", template.FontFamily("NotoSansSC-Bold"), template.FontSize(18))
	})
})
```

仅嵌入实际使用的字形（自动字体子集化），保持输出文件较小。

### 富文本

使用 `RichText` 在单个段落中混合多种样式：

```go
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.RichText(func(rt *template.RichTextBuilder) {
			rt.Span("这是 ")
			rt.Span("粗体", template.Bold())
			rt.Span("，这是 ")
			rt.Span("红色斜体", template.Italic(), template.TextColor(pdf.Red))
			rt.Span("，在同一段落中。")
		})
	})
})
```

块级选项（对齐、缩进）可以作为 `RichText` 的附加参数传递：

```go
c.RichText(func(rt *template.RichTextBuilder) {
	rt.Span("居中混合文本：")
	rt.Span("¥1,234", template.Bold(), template.TextColor(pdf.RGBHex(0x2E7D32)))
}, template.AlignCenter())
```

### 12 列网格布局

使用 Bootstrap 风格的 12 列网格构建布局：

```go
// 两等分列
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(6, func(c *template.ColBuilder) {
		c.Text("左半部分")
	})
	r.Col(6, func(c *template.ColBuilder) {
		c.Text("右半部分")
	})
})

// 侧边栏 + 主内容
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(3, func(c *template.ColBuilder) {
		c.Text("侧边栏")
	})
	r.Col(9, func(c *template.ColBuilder) {
		c.Text("主内容")
	})
})
```

### 固定高度行

使用 `Row()` 指定高度，或使用 `AutoRow()` 自适应内容高度：

```go
// 固定高度：30mm
page.Row(document.Mm(30), func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("此行高度为 30mm")
	})
})
```

### 表格

基本表格：

```go
c.Table(
	[]string{"名称", "数量", "价格"},
	[][]string{
		{"组件", "10", "¥50.00"},
		{"配件", "3", "¥120.00"},
	},
)
```

带样式的表格（表头颜色、列宽、条纹行）：

```go
c.Table(
	[]string{"产品", "类别", "数量", "单价", "合计"},
	[][]string{
		{"笔记本 Pro 15", "电子产品", "2", "¥12,990", "¥25,980"},
		{"无线鼠标", "配件", "10", "¥299", "¥2,990"},
	},
	template.ColumnWidths(30, 20, 10, 20, 20),
	template.TableHeaderStyle(
		template.TextColor(pdf.White),
		template.BgColor(pdf.RGBHex(0x1A237E)),
	),
	template.TableStripe(pdf.RGBHex(0xF5F5F5)),
)
```

### 图片

嵌入 JPEG 和 PNG 图片（支持缩放选项）：

```go
c.Image(imgData)                                      // 默认尺寸
c.Image(imgData, template.FitWidth(document.Mm(80)))   // 按宽度缩放
c.Image(imgData, template.FitHeight(document.Mm(30)))  // 按高度缩放
```

### 线条和间距

```go
c.Line()                                           // 默认（灰色，1pt）
c.Line(template.LineColor(pdf.Red))                 // 带颜色
c.Line(template.LineThickness(document.Pt(3)))      // 粗线
c.Spacer(document.Mm(5))                            // 5mm 垂直间距
```

### 列表

无序列表和有序列表：

```go
// 无序列表
c.List([]string{"第一项", "第二项", "第三项"})

// 有序列表
c.OrderedList([]string{"步骤一", "步骤二", "步骤三"})

// 自定义缩进
c.List([]string{"缩进", "项目"}, template.ListIndent(document.Mm(10)))
```

### 二维码

可配置大小和纠错等级的二维码生成：

```go
// 基本二维码
c.QRCode("https://gpdf.dev")

// 自定义大小和纠错等级
c.QRCode("https://gpdf.dev",
	template.QRSize(document.Mm(30)),
	template.QRErrorCorrection(qrcode.LevelH))
```

### 条形码

Code 128 条形码生成：

```go
// 基本条形码
c.Barcode("INV-2026-0001")

// 自定义宽度
c.Barcode("INV-2026-0001", template.BarcodeWidth(document.Mm(80)))
```

### 页码

自动页码和总页数：

```go
doc.Footer(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("由 gpdf 生成", template.FontSize(8))
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

### 文本装饰

下划线、删除线、字间距和首行缩进：

```go
c.Text("下划线文本", template.Underline())
c.Text("删除线文本", template.Strikethrough())
c.Text("宽字间距", template.LetterSpacing(3))
c.Text("首行缩进段落...", template.TextIndent(document.Pt(24)))
```

### 页眉和页脚

定义在每一页重复显示的页眉和页脚：

```go
doc.Header(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("ACME 公司", template.Bold(), template.FontSize(10))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("机密文件", template.AlignRight(), template.FontSize(10),
				template.TextColor(pdf.Gray(0.5)))
		})
	})
})

doc.Footer(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("由 gpdf 生成", template.AlignCenter(),
				template.FontSize(8), template.TextColor(pdf.Gray(0.5)))
		})
	})
})
```

### 多页文档

```go
for i := 1; i <= 5; i++ {
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("页面内容")
		})
	})
}
```

### JSON 模式

完全用 JSON 定义文档：

```go
schema := []byte(`{
	"page": {"size": "A4", "margins": "20mm"},
	"metadata": {"title": "报告", "author": "gpdf"},
	"body": [
		{"row": {"cols": [
			{"span": 12, "text": "来自 JSON 的问候", "style": {"size": 24, "bold": true}}
		]}}
	]
}`)

doc, err := template.FromJSON(schema, nil)
data, _ := doc.Generate()
```

### Go 模板集成

使用 Go 模板与 JSON 模式生成动态内容：

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

data := map[string]any{"Title": "动态报告"}
doc, err := template.FromJSON(schema, data)
```

使用预解析的 Go 模板获得更多控制：

```go
tmpl, _ := gotemplate.New("doc").Funcs(template.TemplateFuncMap()).Parse(schemaStr)
doc, err := template.FromTemplate(tmpl, data)
```

### 可复用组件

一个函数调用即可生成常见文档类型：

**发票：**

```go
doc := template.Invoice(template.InvoiceData{
	Number:  "#INV-2026-001",
	Date:    "2026年3月1日",
	DueDate: "2026年3月31日",
	From:    template.InvoiceParty{Name: "ACME公司", Address: []string{"北京市朝阳区123号"}},
	To:      template.InvoiceParty{Name: "客户有限公司", Address: []string{"上海市浦东新区456号"}},
	Items: []template.InvoiceItem{
		{Description: "Web开发", Quantity: "40小时", UnitPrice: 150, Amount: 6000},
		{Description: "UI/UX设计", Quantity: "20小时", UnitPrice: 120, Amount: 2400},
	},
	TaxRate:  10,
	Currency: "¥",  // 默认: "$"
	Notes:    "感谢您的惠顾！",
	Payment: &template.InvoicePayment{
		BankName: "中国工商银行",
		Account:  "1234567890",
		Routing:  "102100099996",
	},
})
data, _ := doc.Generate()
```

**报告：**

```go
doc := template.Report(template.ReportData{
	Title:    "季度报告",
	Subtitle: "2026年 Q1",
	Author:   "ACME公司",
	Sections: []template.ReportSection{
		{
			Title:   "执行摘要",
			Content: "与2025年Q4相比，收入增长了15%。",
			Metrics: []template.ReportMetric{
				{Label: "收入", Value: "¥12.5M", ColorHex: 0x2E7D32},
				{Label: "增长", Value: "+15%", ColorHex: 0x2E7D32},
			},
		},
		{
			Title: "收入明细",
			Table: &template.ReportTable{
				Header: []string{"部门", "2026 Q1", "变化"},
				Rows:   [][]string{{"云服务", "¥5.2M", "+26.8%"}, {"企业服务", "¥3.8M", "+8.6%"}},
			},
		},
	},
})
```

**信函：**

```go
doc := template.Letter(template.LetterData{
	From:     template.LetterParty{Name: "ACME公司", Address: []string{"北京市朝阳区123号"}},
	To:       template.LetterParty{Name: "张先生", Address: []string{"上海市浦东新区456号"}},
	Date:     "2026年3月1日",
	Subject:  "合作提案",
	Greeting: "尊敬的张先生：",
	Body:     []string{"我们希望向您提出战略合作伙伴关系的建议..."},
	Closing:  "此致敬礼",
	Signature: "李明",
})
```

### 文档元数据

```go
doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithMetadata(document.DocumentMetadata{
		Title:   "年度报告 2026",
		Author:  "gpdf Library",
		Subject: "文档元数据示例",
		Creator: "My Application",
	}),
)
```

### 页面尺寸和边距

```go
// 可用页面尺寸
document.A4      // 210mm x 297mm
document.A3      // 297mm x 420mm
document.Letter  // 8.5in x 11in
document.Legal   // 8.5in x 14in

// 统一边距
template.WithMargins(document.UniformEdges(document.Mm(20)))

// 非对称边距
template.WithMargins(document.Edges{
	Top:    document.Mm(10),
	Right:  document.Mm(40),
	Bottom: document.Mm(10),
	Left:   document.Mm(40),
})
```

### 输出选项

```go
// Generate 返回 []byte
data, err := doc.Generate()

// Render 写入任意 io.Writer
var buf bytes.Buffer
err := doc.Render(&buf)

// 直接写入文件
f, _ := os.Create("output.pdf")
defer f.Close()
doc.Render(f)
```

## API 参考

### 文档选项

| 函数 | 说明 |
|---|---|
| `WithPageSize(size)` | 设置页面大小（A4、A3、Letter、Legal） |
| `WithMargins(edges)` | 设置页面边距 |
| `WithFont(family, data)` | 注册 TrueType 字体 |
| `WithDefaultFont(family, size)` | 设置默认字体 |
| `WithMetadata(meta)` | 设置文档元数据 |

### 列内容

| 方法 | 说明 |
|---|---|
| `c.Text(text, opts...)` | 添加带样式选项的文本 |
| `c.Table(header, rows, opts...)` | 添加表格 |
| `c.Image(data, opts...)` | 添加图片（JPEG/PNG） |
| `c.QRCode(data, opts...)` | 添加二维码 |
| `c.Barcode(data, opts...)` | 添加条形码（Code 128） |
| `c.RichText(fn, opts...)` | 在一个段落中添加多种内联样式 |
| `c.List(items, opts...)` | 添加无序列表 |
| `c.OrderedList(items, opts...)` | 添加有序列表 |
| `c.PageNumber(opts...)` | 添加当前页码 |
| `c.TotalPages(opts...)` | 添加总页数 |
| `c.Line(opts...)` | 添加水平线 |
| `c.Spacer(height)` | 添加垂直间距 |

### 文本选项

| 选项 | 说明 |
|---|---|
| `template.FontSize(size)` | 设置字号（单位：磅） |
| `template.Bold()` | 粗体 |
| `template.Italic()` | 斜体 |
| `template.FontFamily(name)` | 使用已注册的字体 |
| `template.TextColor(color)` | 设置文本颜色 |
| `template.BgColor(color)` | 设置背景颜色 |
| `template.Underline()` | 下划线装饰 |
| `template.Strikethrough()` | 删除线装饰 |
| `template.LetterSpacing(pts)` | 设置字间距（磅） |
| `template.TextIndent(value)` | 设置首行缩进 |
| `template.AlignLeft()` | 左对齐（默认） |
| `template.AlignCenter()` | 居中对齐 |
| `template.AlignRight()` | 右对齐 |

### 表格选项

| 选项 | 说明 |
|---|---|
| `template.ColumnWidths(w...)` | 设置列宽百分比 |
| `template.TableHeaderStyle(opts...)` | 设置表头行样式 |
| `template.TableStripe(color)` | 设置交替行颜色 |
| `template.TableCellVAlign(align)` | 设置单元格垂直对齐（Top/Middle/Bottom） |

### 图片选项

| 选项 | 说明 |
|---|---|
| `template.FitWidth(value)` | 按宽度缩放（保持宽高比） |
| `template.FitHeight(value)` | 按高度缩放（保持宽高比） |

### 列表选项

| 选项 | 说明 |
|---|---|
| `template.ListIndent(value)` | 设置项目符号/编号缩进宽度 |

### 二维码选项

| 选项 | 说明 |
|---|---|
| `template.QRSize(value)` | 设置二维码大小 |
| `template.QRErrorCorrection(level)` | 设置纠错等级（L/M/Q/H） |
| `template.QRScale(n)` | 设置模块缩放因子 |

### 条形码选项

| 选项 | 说明 |
|---|---|
| `template.BarcodeWidth(value)` | 设置条形码宽度 |
| `template.BarcodeHeight(value)` | 设置条形码高度 |
| `template.BarcodeFormat(fmt)` | 设置条形码格式（Code 128） |

### 模板生成

| 函数 | 说明 |
|---|---|
| `template.FromJSON(schema, data)` | 从 JSON 模式生成文档 |
| `template.FromTemplate(tmpl, data)` | 从 Go 模板生成文档 |
| `template.TemplateFuncMap()` | 获取模板辅助函数（包含 `toJSON`） |

### 可复用组件

| 函数 | 说明 |
|---|---|
| `template.Invoice(data)` | 生成专业发票 PDF |
| `template.Report(data)` | 生成结构化报告 PDF |
| `template.Letter(data)` | 生成商务信函 PDF |

### 富文本构建器

| 方法 | 说明 |
|---|---|
| `rt.Span(text, opts...)` | 添加带样式的内联文本片段 |

### 线条选项

| 选项 | 说明 |
|---|---|
| `template.LineColor(color)` | 设置线条颜色 |
| `template.LineThickness(value)` | 设置线条粗细 |

### 单位

```go
document.Pt(72)    // 点（1/72 英寸）
document.Mm(10)    // 毫米
document.Cm(2.5)   // 厘米
document.In(1)     // 英寸
document.Em(1.5)   // 相对于字体大小
document.Pct(50)   // 百分比
```

### 颜色

```go
pdf.RGB(0.2, 0.4, 0.8)   // RGB（0.0–1.0）
pdf.RGBHex(0xFF5733)      // 十六进制 RGB
pdf.Gray(0.5)             // 灰度
pdf.CMYK(0, 0.5, 1, 0)   // CMYK

// 预定义颜色
pdf.Black, pdf.White, pdf.Red, pdf.Green, pdf.Blue
pdf.Yellow, pdf.Cyan, pdf.Magenta
```

## 基准测试

与 [go-pdf/fpdf](https://github.com/go-pdf/fpdf)、[signintech/gopdf](https://github.com/signintech/gopdf)、[maroto v2](https://github.com/johnfercher/maroto) 对比。
5次运行取中位数，每次100次迭代。Apple M1，Go 1.25。

**执行时间**（越低越好）:

| 基准测试 | gpdf | fpdf | gopdf | maroto v2 |
|---|--:|--:|--:|--:|
| 单页 | **13 µs** | 132 µs | 423 µs | 237 µs |
| 表格 (4x10) | **108 µs** | 241 µs | 835 µs | 8.6 ms |
| 100页 | **683 µs** | 11.7 ms | 8.6 ms | 19.8 ms |
| 复杂文档 | **133 µs** | 254 µs | 997 µs | 10.4 ms |

**内存使用**（越低越好）:

| 基准测试 | gpdf | fpdf | gopdf | maroto v2 |
|---|--:|--:|--:|--:|
| 单页 | **16 KB** | 1.2 MB | 1.8 MB | 61 KB |
| 表格 (4x10) | **209 KB** | 1.3 MB | 1.9 MB | 1.6 MB |
| 100页 | **909 KB** | 121 MB | 83 MB | 4.0 MB |
| 复杂文档 | **246 KB** | 1.3 MB | 2.0 MB | 2.0 MB |

### 为什么 gpdf 更快？

- **单页** — 构建→布局→渲染的单次管道，无中间数据结构。全程使用具体结构体类型（无 `interface{}` 装箱），以最少的堆分配构建文档树。
- **表格** — 单元格内容通过可复用的 `strings.Builder` 缓冲区直接写入 PDF 内容流命令。无逐单元格的对象包装或重复字体查找，字体在每个文档中仅解析一次。
- **100页** — 布局以 O(n) 线性扩展。溢出分页通过切片引用传递剩余节点（无深拷贝）。字体仅解析一次并在所有页面间共享。
- **复杂文档** — 无需重新测量的单次布局整合了以上所有优势。字体子集化仅嵌入实际使用的字形，且默认启用 Flate 压缩，使内存和输出大小保持较小。

运行基准测试:

```bash
cd _benchmark && go test -bench=. -benchmem -count=5
```

## 高级：低级 API

`template` 包（第 3 层）涵盖大多数使用场景。如需完全控制，可以直接使用低级包：

| 包 | 层级 | 说明 |
|---|---|---|
| `template` | 3 | 声明式构建器 API、网格系统、组件 |
| `document` | 2 | 节点树、盒模型、样式、布局引擎 |
| `pdf` | 1 | PDF Writer、TrueType 字体解析、流、图片 |
| `qrcode` | — | 独立二维码编码器（版本 1-40） |
| `barcode` | — | 独立 Code 128 条形码编码器 |

**第 2 层 (document)** 提供：
- 带边距、内边距、边框（实线/虚线/点线）的盒模型
- 分页控制（`BreakPolicy` — avoid、always、page）
- 支持 `ColSpan` / `RowSpan` 的表格单元格
- 图片适应模式（`FitContain`、`FitCover`、`FitStretch`、`FitOriginal`）
- 垂直/水平布局方向

**第 1 层 (pdf)** 提供：
- 原始 PDF 对象写入（`Writer`）
- TrueType 字体解析和子集化
- JPEG/PNG 图片注册
- Flate 流压缩
- 内容流运算符

详细 API 请参阅 [GoDoc](https://pkg.go.dev/github.com/gpdf-dev/gpdf)。

## 许可证

MIT
