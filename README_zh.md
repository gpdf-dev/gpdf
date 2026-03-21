# gpdf

[![Go Reference](https://pkg.go.dev/badge/github.com/gpdf-dev/gpdf.svg)](https://pkg.go.dev/github.com/gpdf-dev/gpdf)
[![CI](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml/badge.svg)](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gpdf-dev/gpdf)](https://goreportcard.com/report/github.com/gpdf-dev/gpdf)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.22-blue)](https://go.dev/)
[![Website](https://img.shields.io/badge/Website-gpdf.dev-blue)](https://gpdf.dev/)

[English](README.md) | [日本語](README_ja.md) | **中文** | [한국어](README_ko.md) | [Español](README_es.md) | [Português](README_pt.md)

纯 Go 实现的零依赖 PDF 生成库，采用分层架构和声明式构建器 API。

## 特性

- **零依赖** — 仅使用 Go 标准库
- **分层架构** — 底层 PDF 原语、文档模型和高层模板 API
- **12 列网格系统** — Bootstrap 风格的响应式布局
- **TrueType 字体支持** — 自定义字体嵌入与子集化
- **CJK 就绪** — 从第一天起完整支持中日韩文本
- **表格** — 表头、列宽、条纹行、垂直对齐
- **页眉和页脚** — 带页码，所有页面统一显示
- **列表** — 无序列表和有序列表
- **二维码** — 纯 Go 二维码生成（支持纠错等级）
- **条形码** — Code 128 条形码生成
- **文本装饰** — 下划线、删除线、字间距、首行缩进
- **页码** — 自动页码和总页数
- **Go 模板集成** — 从 Go 模板生成 PDF
- **可复用组件** — 内置发票、报告和信函预设模板
- **JSON 模式** — 完全用 JSON 定义文档
- **多种单位** — pt、mm、cm、in、em、%
- **色彩空间** — RGB、灰度、CMYK
- **图片** — JPEG 和 PNG 嵌入（支持缩放选项）
- **绝对定位** — 在页面上以精确 XY 坐标放置元素
- **现有 PDF 叠加** — 打开现有 PDF 并在上面添加文字、图片、印章
- **文档元数据** — 标题、作者、主题、创建者
- **加密** — AES-256 加密（ISO 32000-2, Rev 6），支持所有者/用户密码和权限控制
- **PDF/A** — PDF/A-1b 和 PDF/A-2b 合规，包含 ICC 配置文件和 XMP 元数据
- **数字签名** — CMS/PKCS#7 签名，支持 RSA/ECDSA 密钥和 RFC 3161 时间戳

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

## 架构

```
┌─────────────────────────────────────┐
│  gpdf (entry point)                 │
├─────────────────────────────────────┤
│  template  — Builder API, Grid      │  Layer 3
├─────────────────────────────────────┤
│  document  — Nodes, Style, Layout   │  Layer 2
├─────────────────────────────────────┤
│  pdf       — Writer, Fonts, Streams │  Layer 1
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

### CJK 字体（中文 / 日文 / 韩文）

渲染 CJK 文本需要嵌入 TrueType 字体。每种语言使用对应的 Noto Sans 字体：

```go
fontData, _ := os.ReadFile("NotoSansSC-Regular.ttf")

doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithFont("NotoSansSC", fontData),
	gpdf.WithDefaultFont("NotoSansSC", 12),
)

page := doc.AddPage()
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("你好世界", template.FontSize(18))
	})
})
```

对于多语言文档，注册多个字体并使用 `FontFamily()` 切换：

```go
jpFont, _ := os.ReadFile("NotoSansJP-Regular.ttf")
scFont, _ := os.ReadFile("NotoSansSC-Regular.ttf")
krFont, _ := os.ReadFile("NotoSansKR-Regular.ttf")

doc := gpdf.NewDocument(
	gpdf.WithFont("NotoSansJP", jpFont),
	gpdf.WithFont("NotoSansSC", scFont),
	gpdf.WithFont("NotoSansKR", krFont),
	gpdf.WithDefaultFont("NotoSansSC", 12),
)

page := doc.AddPage()
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(4, func(c *template.ColBuilder) {
		c.Text("日本語", template.FontFamily("NotoSansJP"))
	})
	r.Col(4, func(c *template.ColBuilder) {
		c.Text("中文", template.FontFamily("NotoSansSC"))
	})
	r.Col(4, func(c *template.ColBuilder) {
		c.Text("한국어", template.FontFamily("NotoSansKR"))
	})
})
```

推荐字体（均为免费，OFL 许可）：

| 字体 | 语言 |
|---|---|
| [Noto Sans JP](https://fonts.google.com/noto/specimen/Noto+Sans+JP) | 日文 |
| [Noto Sans SC](https://fonts.google.com/noto/specimen/Noto+Sans+SC) | 简体中文 |
| [Noto Sans KR](https://fonts.google.com/noto/specimen/Noto+Sans+KR) | 韩文 |

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
	TaxRate: 10,
	Notes:   "感谢您的惠顾！",
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

### 现有 PDF 叠加

打开现有 PDF，使用同一构建器 API 叠加内容：

```go
// 打开现有 PDF
doc, err := gpdf.Open(existingPDFBytes)

// 在第 1 页添加 "DRAFT" 水印
doc.Overlay(0, func(p *template.PageBuilder) {
	p.Absolute(document.Mm(50), document.Mm(140), func(c *template.ColBuilder) {
		c.Text("DRAFT", template.FontSize(72),
			template.TextColor(pdf.Gray(0.85)))
	})
})

// 为每页添加页码
count, _ := doc.PageCount()
doc.EachPage(func(i int, p *template.PageBuilder) {
	p.Absolute(document.Mm(170), document.Mm(285), func(c *template.ColBuilder) {
		c.Text(fmt.Sprintf("%d / %d", i+1, count), template.FontSize(10))
	}, template.AbsoluteWidth(document.Mm(20)))
})

result, _ := doc.Save()
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
| `c.List(items, opts...)` | 添加无序列表 |
| `c.OrderedList(items, opts...)` | 添加有序列表 |
| `c.PageNumber(opts...)` | 添加当前页码 |
| `c.TotalPages(opts...)` | 添加总页数 |
| `c.Line(opts...)` | 添加水平线 |
| `c.Spacer(height)` | 添加垂直间距 |

### 页面级内容

| 方法 | 说明 |
|---|---|
| `page.AutoRow(fn)` | 添加自动高度行 |
| `page.Row(height, fn)` | 添加固定高度行 |
| `page.Absolute(x, y, fn, opts...)` | 在精确 XY 坐标放置内容 |

#### 绝对定位选项

| 选项 | 说明 |
|---|---|
| `gpdf.AbsoluteWidth(value)` | 设置显式宽度（默认：剩余空间） |
| `gpdf.AbsoluteHeight(value)` | 设置显式高度（默认：剩余空间） |
| `gpdf.AbsoluteOriginPage()` | 以页面角为原点，而非内容区域 |

### 现有 PDF 操作

| 函数 / 方法 | 说明 |
|---|---|
| `gpdf.Open(data, opts...)` | 打开现有 PDF 用于叠加 |
| `doc.PageCount()` | 获取页数 |
| `doc.Overlay(page, fn)` | 在指定页上叠加内容 |
| `doc.EachPage(fn)` | 对每页应用叠加 |
| `doc.Save()` | 保存修改后的 PDF |

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

## 许可证

MIT
