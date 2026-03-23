# gpdf

[![Go Reference](https://pkg.go.dev/badge/github.com/gpdf-dev/gpdf.svg)](https://pkg.go.dev/github.com/gpdf-dev/gpdf)
[![CI](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml/badge.svg)](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gpdf-dev/gpdf)](https://goreportcard.com/report/github.com/gpdf-dev/gpdf)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.22-blue)](https://go.dev/)
[![Website](https://img.shields.io/badge/Website-gpdf.dev-blue)](https://gpdf.dev/)

[English](README.md) | **日本語** | [中文](README_zh.md) | [한국어](README_ko.md) | [Español](README_es.md) | [Português](README_pt.md)

純粋なGoで実装された、外部依存ゼロのPDF生成ライブラリ。レイヤードアーキテクチャと宣言的なビルダーAPIを提供します。

## 特徴

- **外部依存ゼロ** — Go標準ライブラリのみ使用
- **レイヤードアーキテクチャ** — 低レベルPDFプリミティブ、ドキュメントモデル、高レベルテンプレートAPI
- **12カラムグリッドシステム** — Bootstrap風のレスポンシブレイアウト
- **TrueTypeフォント対応** — カスタムフォントの埋め込みとサブセット化
- **CJK対応** — 日中韓テキストを初日からフルサポート
- **テーブル** — ヘッダー、カラム幅指定、ストライプ行、垂直揃え
- **ヘッダー＆フッター** — ページ番号付きで全ページに一貫表示
- **リスト** — 箇条書きリストと番号付きリスト
- **QRコード** — 純GoのQRコード生成（誤り訂正レベル対応）
- **バーコード** — Code 128バーコード生成
- **テキスト装飾** — 下線、取り消し線、字間、字下げ
- **ページ番号** — 自動ページ番号と総ページ数
- **Goテンプレート統合** — GoテンプレートからPDF生成
- **再利用可能コンポーネント** — 請求書・レポート・レターのプリセットテンプレート
- **JSONスキーマ** — JSONのみでドキュメントを定義
- **複数の単位** — pt, mm, cm, in, em, %
- **カラースペース** — RGB、グレースケール、CMYK
- **画像** — JPEGとPNGの埋め込み（フィットオプション対応）
- **絶対位置指定** — ページ上の任意のXY座標に要素を配置
- **既存PDFオーバーレイ** — 既存PDFを開いてテキスト、画像、スタンプを上に追加
- **PDFマージ** — 複数のPDFをページ範囲指定付きで1つに結合
- **ドキュメントメタデータ** — タイトル、著者、件名、作成者
- **暗号化** — AES-256暗号化（ISO 32000-2, Rev 6）、オーナー/ユーザーパスワードと権限制御
- **PDF/A** — PDF/A-1bおよびPDF/A-2b準拠、ICCプロファイルとXMPメタデータ対応
- **デジタル署名** — CMS/PKCS#7署名、RSA/ECDSA鍵対応、RFC 3161タイムスタンプ対応

## ベンチマーク

[go-pdf/fpdf](https://github.com/go-pdf/fpdf)、[signintech/gopdf](https://github.com/signintech/gopdf)、[maroto v2](https://github.com/johnfercher/maroto) との比較。
5回実行の中央値、各100イテレーション。Apple M1、Go 1.25。

**実行時間**（低いほど良い）:

| ベンチマーク | gpdf | fpdf | gopdf | maroto v2 |
|---|--:|--:|--:|--:|
| 1ページ | **13 µs** | 132 µs | 423 µs | 237 µs |
| テーブル (4x10) | **108 µs** | 241 µs | 835 µs | 8.6 ms |
| 100ページ | **683 µs** | 11.7 ms | 8.6 ms | 19.8 ms |
| 複合ドキュメント | **133 µs** | 254 µs | 997 µs | 10.4 ms |

**メモリ使用量**（低いほど良い）:

| ベンチマーク | gpdf | fpdf | gopdf | maroto v2 |
|---|--:|--:|--:|--:|
| 1ページ | **16 KB** | 1.2 MB | 1.8 MB | 61 KB |
| テーブル (4x10) | **209 KB** | 1.3 MB | 1.9 MB | 1.6 MB |
| 100ページ | **909 KB** | 121 MB | 83 MB | 4.0 MB |
| 複合ドキュメント | **246 KB** | 1.3 MB | 2.0 MB | 2.0 MB |

### なぜ gpdf は速いのか？

- **1ページ** — ビルド→レイアウト→レンダリングのシングルパスパイプラインで中間データ構造を持たない。全体を通して具体的な構造体型を使用（`interface{}` ボクシングなし）し、ドキュメントツリーを最小限のヒープ割り当てで構築。
- **テーブル** — セル内容を再利用可能な `strings.Builder` バッファを通じて PDF コンテンストリームコマンドとして直接書き出す。セルごとのオブジェクトラッピングやフォントの繰り返し検索がなく、フォントはドキュメントごとに1回だけ解決。
- **100ページ** — レイアウトは O(n) で線形にスケール。オーバーフローページネーションはスライス参照で残りのノードを渡す（ディープコピーなし）。フォントは1回だけパースされ全ページで共有。
- **複合ドキュメント** — 再計測なしのシングルパスレイアウトが上記すべてを統合。フォントサブセッティングは実際に使用されたグリフのみを埋め込み、Flate 圧縮がデフォルトで適用されるため、メモリと出力サイズの両方を小さく保つ。

ベンチマーク実行:

```bash
cd _benchmark && go test -bench=. -benchmem -count=5
```

## アーキテクチャ

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

## 要件

- Go 1.22 以降

## インストール

```bash
go get github.com/gpdf-dev/gpdf
```

## クイックスタート

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

## 使用例

### テキストスタイリング

フォントサイズ、太さ、スタイル、色、背景色、配置:

```go
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("大きな太字タイトル", template.FontSize(24), template.Bold())
		c.Text("イタリックテキスト", template.Italic())
		c.Text("太字 + イタリック", template.Bold(), template.Italic())
		c.Text("赤いテキスト", template.TextColor(pdf.Red))
		c.Text("カスタムカラー", template.TextColor(pdf.RGBHex(0x336699)))
		c.Text("背景色付き", template.BgColor(pdf.Yellow))
		c.Text("中央揃え", template.AlignCenter())
		c.Text("右揃え", template.AlignRight())
	})
})
```

### CJKフォント（日本語・中国語・韓国語）

CJKテキストのレンダリングにはTrueTypeフォントの埋め込みが必要です。各言語にはそれぞれのNoto Sansフォントを使用します:

```go
fontData, _ := os.ReadFile("NotoSansJP-Regular.ttf")

doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithFont("NotoSansJP", fontData),
	gpdf.WithDefaultFont("NotoSansJP", 12),
)

page := doc.AddPage()
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("こんにちは世界", template.FontSize(18))
	})
})
```

多言語ドキュメントでは、複数のフォントを登録して`FontFamily()`で切り替えます:

```go
jpFont, _ := os.ReadFile("NotoSansJP-Regular.ttf")
scFont, _ := os.ReadFile("NotoSansSC-Regular.ttf")
krFont, _ := os.ReadFile("NotoSansKR-Regular.ttf")

doc := gpdf.NewDocument(
	gpdf.WithFont("NotoSansJP", jpFont),
	gpdf.WithFont("NotoSansSC", scFont),
	gpdf.WithFont("NotoSansKR", krFont),
	gpdf.WithDefaultFont("NotoSansJP", 12),
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

推奨フォント（すべて無料、OFLライセンス）:

| フォント | 言語 |
|---|---|
| [Noto Sans JP](https://fonts.google.com/noto/specimen/Noto+Sans+JP) | 日本語 |
| [Noto Sans SC](https://fonts.google.com/noto/specimen/Noto+Sans+SC) | 簡体字中国語 |
| [Noto Sans KR](https://fonts.google.com/noto/specimen/Noto+Sans+KR) | 韓国語 |

### 12カラムグリッドレイアウト

Bootstrap風の12カラムグリッドでレイアウトを構築:

```go
// 2等分カラム
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(6, func(c *template.ColBuilder) {
		c.Text("左半分")
	})
	r.Col(6, func(c *template.ColBuilder) {
		c.Text("右半分")
	})
})

// サイドバー + メインコンテンツ
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(3, func(c *template.ColBuilder) {
		c.Text("サイドバー")
	})
	r.Col(9, func(c *template.ColBuilder) {
		c.Text("メインコンテンツ")
	})
})
```

### 固定高さの行

`Row()` で高さを指定、`AutoRow()` でコンテンツに合わせた自動高さ:

```go
// 固定高さ: 30mm
page.Row(document.Mm(30), func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("この行の高さは30mmです")
	})
})
```

### テーブル

基本的なテーブル:

```go
c.Table(
	[]string{"商品名", "数量", "価格"},
	[][]string{
		{"ウィジェット", "10", "¥500"},
		{"ガジェット", "3", "¥1,200"},
	},
)
```

スタイル付きテーブル（ヘッダー色、カラム幅、ストライプ行）:

```go
c.Table(
	[]string{"商品", "カテゴリ", "数量", "単価", "合計"},
	[][]string{
		{"ノートPC Pro 15", "電子機器", "2", "¥129,900", "¥259,800"},
		{"ワイヤレスマウス", "周辺機器", "10", "¥2,999", "¥29,990"},
	},
	template.ColumnWidths(30, 20, 10, 20, 20),
	template.TableHeaderStyle(
		template.TextColor(pdf.White),
		template.BgColor(pdf.RGBHex(0x1A237E)),
	),
	template.TableStripe(pdf.RGBHex(0xF5F5F5)),
)
```

### 画像

JPEGとPNG画像の埋め込み（フィットオプション対応）:

```go
c.Image(imgData)                                      // デフォルトサイズ
c.Image(imgData, template.FitWidth(document.Mm(80)))   // 幅に合わせる
c.Image(imgData, template.FitHeight(document.Mm(30)))  // 高さに合わせる
```

### 罫線とスペーサー

```go
c.Line()                                           // デフォルト（グレー、1pt）
c.Line(template.LineColor(pdf.Red))                 // 色付き
c.Line(template.LineThickness(document.Pt(3)))      // 太線
c.Spacer(document.Mm(5))                            // 5mmの垂直間隔
```

### リスト

箇条書きリストと番号付きリスト:

```go
// 箇条書きリスト
c.List([]string{"項目1", "項目2", "項目3"})

// 番号付きリスト
c.OrderedList([]string{"ステップ1", "ステップ2", "ステップ3"})
```

### QRコード

サイズと誤り訂正レベルを指定可能なQRコード生成:

```go
// 基本的なQRコード
c.QRCode("https://gpdf.dev")

// サイズと誤り訂正レベルを指定
c.QRCode("https://gpdf.dev",
	template.QRSize(document.Mm(30)),
	template.QRErrorCorrection(qrcode.LevelH))
```

### バーコード

Code 128バーコード生成:

```go
// 基本的なバーコード
c.Barcode("INV-2026-0001")

// 幅を指定
c.Barcode("INV-2026-0001", template.BarcodeWidth(document.Mm(80)))
```

### ページ番号

自動ページ番号と総ページ数:

```go
doc.Footer(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("gpdfで生成", template.FontSize(8))
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

### テキスト装飾

下線、取り消し線、字間、字下げ:

```go
c.Text("下線テキスト", template.Underline())
c.Text("取り消し線テキスト", template.Strikethrough())
c.Text("広い字間", template.LetterSpacing(3))
c.Text("字下げ段落...", template.TextIndent(document.Pt(24)))
```

### ヘッダー＆フッター

全ページに繰り返し表示されるヘッダーとフッター:

```go
doc.Header(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("ACME株式会社", template.Bold(), template.FontSize(10))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("社外秘", template.AlignRight(), template.FontSize(10),
				template.TextColor(pdf.Gray(0.5)))
		})
	})
})

doc.Footer(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("gpdfで生成", template.AlignCenter(),
				template.FontSize(8), template.TextColor(pdf.Gray(0.5)))
		})
	})
})
```

### 既存PDFオーバーレイ

既存のPDFを開いて、同じビルダーAPIでコンテンツを重ねて配置:

```go
// 既存PDFを開く
doc, err := gpdf.Open(existingPDFBytes)

// 1ページ目に「DRAFT」透かしを追加
doc.Overlay(0, func(p *template.PageBuilder) {
	p.Absolute(document.Mm(50), document.Mm(140), func(c *template.ColBuilder) {
		c.Text("DRAFT", template.FontSize(72),
			template.TextColor(pdf.Gray(0.85)))
	})
})

// 全ページにページ番号を追加
count, _ := doc.PageCount()
doc.EachPage(func(i int, p *template.PageBuilder) {
	p.Absolute(document.Mm(170), document.Mm(285), func(c *template.ColBuilder) {
		c.Text(fmt.Sprintf("%d / %d", i+1, count), template.FontSize(10))
	}, template.AbsoluteWidth(document.Mm(20)))
})

result, _ := doc.Save()
```

### PDFマージ

複数のPDFをページ範囲指定付きで1つのドキュメントに結合:

```go
// 複数のPDFを結合
merged, _ := gpdf.Merge(
	[]gpdf.Source{
		{Data: coverPage},
		{Data: report},
		{Data: appendix, Pages: gpdf.PageRange{From: 1, To: 3}}, // 最初の3ページのみ
	},
	gpdf.WithMergeMetadata("My Document", "Author", ""),
)
```

### JSONスキーマ

JSONのみでドキュメントを定義:

```go
schema := []byte(`{
	"page": {"size": "A4", "margins": "20mm"},
	"metadata": {"title": "レポート", "author": "gpdf"},
	"body": [
		{"row": {"cols": [
			{"span": 12, "text": "JSONからこんにちは", "style": {"size": 24, "bold": true}}
		]}}
	]
}`)

doc, err := template.FromJSON(schema, nil)
data, _ := doc.Generate()
```

### Goテンプレート統合

GoテンプレートとJSONスキーマで動的コンテンツを生成:

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

data := map[string]any{"Title": "動的レポート"}
doc, err := template.FromJSON(schema, data)
```

事前パース済みGoテンプレートでより柔軟に:

```go
tmpl, _ := gotemplate.New("doc").Funcs(template.TemplateFuncMap()).Parse(schemaStr)
doc, err := template.FromTemplate(tmpl, data)
```

### 再利用可能コンポーネント

関数一つで一般的なドキュメントを生成:

**請求書:**

```go
doc := template.Invoice(template.InvoiceData{
	Number:  "#INV-2026-001",
	Date:    "2026年3月1日",
	DueDate: "2026年3月31日",
	From:    template.InvoiceParty{Name: "ACME株式会社", Address: []string{"東京都渋谷区1-2-3"}},
	To:      template.InvoiceParty{Name: "クライアント株式会社", Address: []string{"大阪府大阪市4-5-6"}},
	Items: []template.InvoiceItem{
		{Description: "Web開発", Quantity: "40時間", UnitPrice: 150, Amount: 6000},
		{Description: "UI/UXデザイン", Quantity: "20時間", UnitPrice: 120, Amount: 2400},
	},
	TaxRate: 10,
	Notes:   "ご利用ありがとうございます！",
})
data, _ := doc.Generate()
```

**レポート:**

```go
doc := template.Report(template.ReportData{
	Title:    "四半期レポート",
	Subtitle: "2026年 Q1",
	Author:   "ACME株式会社",
	Sections: []template.ReportSection{
		{
			Title:   "エグゼクティブサマリー",
			Content: "売上は2025年Q4と比較して15%増加しました。",
			Metrics: []template.ReportMetric{
				{Label: "売上", Value: "¥12.5M", ColorHex: 0x2E7D32},
				{Label: "成長率", Value: "+15%", ColorHex: 0x2E7D32},
			},
		},
		{
			Title: "売上内訳",
			Table: &template.ReportTable{
				Header: []string{"事業部", "2026 Q1", "変化"},
				Rows:   [][]string{{"クラウド", "¥5.2M", "+26.8%"}, {"エンタープライズ", "¥3.8M", "+8.6%"}},
			},
		},
	},
})
```

**レター:**

```go
doc := template.Letter(template.LetterData{
	From:     template.LetterParty{Name: "ACME株式会社", Address: []string{"東京都渋谷区1-2-3"}},
	To:       template.LetterParty{Name: "田中太郎 様", Address: []string{"大阪府大阪市4-5-6"}},
	Date:     "2026年3月1日",
	Subject:  "業務提携のご提案",
	Greeting: "田中様",
	Body:     []string{"戦略的パートナーシップをご提案させていただきたく..."},
	Closing:  "敬具",
	Signature: "山田花子",
})
```

### ドキュメントメタデータ

```go
doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithMetadata(document.DocumentMetadata{
		Title:   "年次報告書 2026",
		Author:  "gpdf Library",
		Subject: "ドキュメントメタデータの例",
		Creator: "My Application",
	}),
)
```

### ページサイズとマージン

```go
// 利用可能なページサイズ
document.A4      // 210mm x 297mm
document.A3      // 297mm x 420mm
document.Letter  // 8.5in x 11in
document.Legal   // 8.5in x 14in

// 均一マージン
template.WithMargins(document.UniformEdges(document.Mm(20)))

// 非対称マージン
template.WithMargins(document.Edges{
	Top:    document.Mm(10),
	Right:  document.Mm(40),
	Bottom: document.Mm(10),
	Left:   document.Mm(40),
})
```

### 出力オプション

```go
// Generateは[]byteを返す
data, err := doc.Generate()

// Renderは任意のio.Writerに書き込む
var buf bytes.Buffer
err := doc.Render(&buf)

// ファイルに直接書き込む
f, _ := os.Create("output.pdf")
defer f.Close()
doc.Render(f)
```

## APIリファレンス

### ドキュメントオプション

| 関数 | 説明 |
|---|---|
| `WithPageSize(size)` | ページサイズを設定 (A4, A3, Letter, Legal) |
| `WithMargins(edges)` | ページマージンを設定 |
| `WithFont(family, data)` | TrueTypeフォントを登録 |
| `WithDefaultFont(family, size)` | デフォルトフォントを設定 |
| `WithMetadata(meta)` | ドキュメントメタデータを設定 |

### カラムコンテンツ

| メソッド | 説明 |
|---|---|
| `c.Text(text, opts...)` | スタイルオプション付きテキストを追加 |
| `c.Table(header, rows, opts...)` | テーブルを追加 |
| `c.Image(data, opts...)` | 画像を追加 (JPEG/PNG) |
| `c.QRCode(data, opts...)` | QRコードを追加 |
| `c.Barcode(data, opts...)` | バーコードを追加 (Code 128) |
| `c.List(items, opts...)` | 箇条書きリストを追加 |
| `c.OrderedList(items, opts...)` | 番号付きリストを追加 |
| `c.PageNumber(opts...)` | 現在のページ番号を追加 |
| `c.TotalPages(opts...)` | 総ページ数を追加 |
| `c.Line(opts...)` | 水平線を追加 |
| `c.Spacer(height)` | 垂直スペースを追加 |

### ページレベルコンテンツ

| メソッド | 説明 |
|---|---|
| `page.AutoRow(fn)` | 自動高さの行を追加 |
| `page.Row(height, fn)` | 固定高さの行を追加 |
| `page.Absolute(x, y, fn, opts...)` | 指定したXY座標にコンテンツを配置 |

#### 絶対位置指定オプション

| オプション | 説明 |
|---|---|
| `gpdf.AbsoluteWidth(value)` | 明示的な幅を設定（デフォルト: 残りスペース） |
| `gpdf.AbsoluteHeight(value)` | 明示的な高さを設定（デフォルト: 残りスペース） |
| `gpdf.AbsoluteOriginPage()` | コンテンツ領域ではなくページ角を原点にする |

### 既存PDF操作

| 関数 / メソッド | 説明 |
|---|---|
| `gpdf.Open(data, opts...)` | 既存PDFをオーバーレイ用に開く |
| `doc.PageCount()` | ページ数を取得 |
| `doc.Overlay(page, fn)` | 特定ページにコンテンツを重ねて配置 |
| `doc.EachPage(fn)` | 全ページにオーバーレイを適用 |
| `doc.Save()` | 変更したPDFを保存 |
| `gpdf.Merge(sources, opts...)` | 複数のPDFを1つに結合 |
| `WithMergeMetadata(title, author, producer)` | 結合後のメタデータを設定 |

### テキストオプション

| オプション | 説明 |
|---|---|
| `template.FontSize(size)` | フォントサイズをポイント単位で設定 |
| `template.Bold()` | 太字 |
| `template.Italic()` | イタリック |
| `template.FontFamily(name)` | 登録済みフォントを使用 |
| `template.TextColor(color)` | テキスト色を設定 |
| `template.BgColor(color)` | 背景色を設定 |
| `template.Underline()` | 下線装飾 |
| `template.Strikethrough()` | 取り消し線装飾 |
| `template.LetterSpacing(pts)` | 字間をポイント単位で設定 |
| `template.TextIndent(value)` | 字下げを設定 |
| `template.AlignLeft()` | 左揃え（デフォルト） |
| `template.AlignCenter()` | 中央揃え |
| `template.AlignRight()` | 右揃え |

### テーブルオプション

| オプション | 説明 |
|---|---|
| `template.ColumnWidths(w...)` | カラム幅をパーセンテージで設定 |
| `template.TableHeaderStyle(opts...)` | ヘッダー行のスタイル設定 |
| `template.TableStripe(color)` | 交互行の色を設定 |
| `template.TableCellVAlign(align)` | セルの垂直揃え (Top/Middle/Bottom) |

### 画像オプション

| オプション | 説明 |
|---|---|
| `template.FitWidth(value)` | 幅に合わせてスケール（アスペクト比維持） |
| `template.FitHeight(value)` | 高さに合わせてスケール（アスペクト比維持） |

### QRコードオプション

| オプション | 説明 |
|---|---|
| `template.QRSize(value)` | QRコードのサイズを設定 |
| `template.QRErrorCorrection(level)` | 誤り訂正レベルを設定 (L/M/Q/H) |
| `template.QRScale(n)` | モジュールスケール係数を設定 |

### バーコードオプション

| オプション | 説明 |
|---|---|
| `template.BarcodeWidth(value)` | バーコードの幅を設定 |
| `template.BarcodeHeight(value)` | バーコードの高さを設定 |
| `template.BarcodeFormat(fmt)` | バーコードフォーマットを設定 (Code 128) |

### テンプレート生成

| 関数 | 説明 |
|---|---|
| `template.FromJSON(schema, data)` | JSONスキーマからドキュメントを生成 |
| `template.FromTemplate(tmpl, data)` | Goテンプレートからドキュメントを生成 |
| `template.TemplateFuncMap()` | テンプレートヘルパー関数を取得（`toJSON`を含む） |

### 再利用可能コンポーネント

| 関数 | 説明 |
|---|---|
| `template.Invoice(data)` | プロフェッショナルな請求書PDFを生成 |
| `template.Report(data)` | 構造化されたレポートPDFを生成 |
| `template.Letter(data)` | ビジネスレターPDFを生成 |

### 罫線オプション

| オプション | 説明 |
|---|---|
| `template.LineColor(color)` | 罫線の色を設定 |
| `template.LineThickness(value)` | 罫線の太さを設定 |

### 単位

```go
document.Pt(72)    // ポイント (1/72インチ)
document.Mm(10)    // ミリメートル
document.Cm(2.5)   // センチメートル
document.In(1)     // インチ
document.Em(1.5)   // フォントサイズに対する相対値
document.Pct(50)   // パーセント
```

### カラー

```go
pdf.RGB(0.2, 0.4, 0.8)   // RGB (0.0–1.0)
pdf.RGBHex(0xFF5733)      // 16進数からRGB
pdf.Gray(0.5)             // グレースケール
pdf.CMYK(0, 0.5, 1, 0)   // CMYK

// 定義済みカラー
pdf.Black, pdf.White, pdf.Red, pdf.Green, pdf.Blue
pdf.Yellow, pdf.Cyan, pdf.Magenta
```

## ライセンス

MIT
