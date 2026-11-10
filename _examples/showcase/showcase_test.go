// Package showcase generates the sample documents shown at the top of the
// README. The tests are skipped unless GPDF_SHOWCASE_OUT points to an output
// directory, so regular `go test ./...` runs never write anything.
//
// Regenerate the README images with:
//
//	./scripts/readme-images.sh
package showcase_test

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

// outDir returns the directory to write showcase PDFs into, skipping the
// test when GPDF_SHOWCASE_OUT is unset.
func outDir(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("GPDF_SHOWCASE_OUT")
	if dir == "" {
		t.Skip("set GPDF_SHOWCASE_OUT to regenerate README showcase assets")
	}
	return dir
}

func loadFont(t *testing.T, filename string) []byte {
	t.Helper()
	// gpdf/_examples/showcase/ → workspace root (same location as example 32).
	path := filepath.Join("..", "..", "..", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("CJK font not found: %s", path)
	}
	return data
}

func TestShowcase_InvoiceJA(t *testing.T) {
	dir := outDir(t)
	font := loadFont(t, "NotoSansJP-Regular.ttf")

	navy := pdf.RGBHex(0x1A237E)
	muted := pdf.Gray(0.45)
	tint := pdf.RGBHex(0xE8EAF6)

	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(18))),
		template.WithFont("NotoSansJP", font),
		template.WithDefaultFont("NotoSansJP", 10),
		template.WithMetadata(document.DocumentMetadata{
			Title:  "請求書 No.2026-0142",
			Author: "株式会社サンプル",
		}),
	)

	page := doc.AddPage()

	// Title + QR
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(9, func(c *template.ColBuilder) {
			c.Text("請求書", template.FontSize(28), template.TextColor(navy), template.LetterSpacing(6))
			c.Text("INVOICE  No.2026-0142", template.FontSize(9), template.TextColor(muted))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.QRCode("https://gpdf.dev", template.QRSize(document.Mm(22)))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(3))
			c.Line(template.LineColor(navy), template.LineThickness(document.Pt(2)))
			c.Spacer(document.Mm(6))
		})
	})

	// Customer / issuer
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(7, func(c *template.ColBuilder) {
			c.Text("株式会社テックソリューションズ 御中", template.FontSize(14))
			c.Spacer(document.Mm(2))
			c.Text("〒150-0001 東京都渋谷区神宮前1-2-3")
			c.Spacer(document.Mm(5))
			c.Text("下記の通りご請求申し上げます。", template.TextColor(muted))
			c.Spacer(document.Mm(3))
			c.Box(func(c *template.ColBuilder) {
				c.Text("ご請求金額（税込）", template.FontSize(9), template.TextColor(muted))
				c.Text("¥1,628,000", template.FontSize(22), template.TextColor(navy))
			},
				template.WithBoxBackground(tint),
				template.WithBoxPadding(document.UniformEdges(document.Mm(4))),
			)
		})
		r.Col(5, func(c *template.ColBuilder) {
			c.Text("発行日  2026年3月1日", template.AlignRight())
			c.Text("お支払期限  2026年3月31日", template.AlignRight())
			c.Spacer(document.Mm(6))
			c.Text("株式会社サンプル", template.FontSize(12), template.AlignRight())
			c.Text("〒100-0005", template.AlignRight(), template.TextColor(muted))
			c.Text("東京都千代田区丸の内1-1-1", template.AlignRight(), template.TextColor(muted))
			c.Text("登録番号 T1234567890123", template.AlignRight(), template.TextColor(muted))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) { c.Spacer(document.Mm(8)) })
	})

	// Line items
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(
				[]string{"品目", "数量", "単価", "金額"},
				[][]string{
					{"Webアプリケーション開発（フロントエンド）", "40 時間", "¥12,000", "¥480,000"},
					{"Webアプリケーション開発（バックエンド）", "50 時間", "¥12,000", "¥600,000"},
					{"UI/UXデザイン", "20 時間", "¥10,000", "¥200,000"},
					{"インフラ構築・運用設計", "1 式", "¥200,000", "¥200,000"},
				},
				template.ColumnWidths(52, 14, 16, 18),
				template.TableHeaderStyle(
					template.TextColor(pdf.White),
					template.BgColor(navy),
				),
				template.TableStripe(pdf.RGBHex(0xF5F5F5)),
				template.ColumnAlign(document.AlignLeft, document.AlignRight, document.AlignRight, document.AlignRight),
			)
		})
	})

	// Totals
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(7, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(6))
			c.Text("お振込先", template.FontSize(9), template.TextColor(muted))
			c.Text("サンプル銀行 本店営業部 普通 1234567")
			c.Text("カ）サンプル")
		})
		r.Col(5, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(4))
			c.Text("小計    ¥1,480,000", template.AlignRight())
			c.Text("消費税（10%）    ¥148,000", template.AlignRight())
			c.Spacer(document.Mm(2))
			c.Line(template.LineColor(navy))
			c.Spacer(document.Mm(2))
			c.Text("合計    ¥1,628,000", template.AlignRight(), template.FontSize(14), template.TextColor(navy))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(14))
			c.Line(template.LineColor(pdf.Gray(0.85)))
			c.Spacer(document.Mm(3))
			c.Text("Generated with gpdf — pure Go, zero dependencies.", template.AlignCenter(),
				template.FontSize(8), template.TextColor(muted))
		})
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "invoice_ja.pdf"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestShowcase_Frame draws a 1px border around every PNG listed in
// GPDF_FRAME_PNGS (colon-separated) so that white pages stay visible on
// GitHub's light theme.
func TestShowcase_Frame(t *testing.T) {
	list := os.Getenv("GPDF_FRAME_PNGS")
	if list == "" {
		t.Skip("set GPDF_FRAME_PNGS to frame README thumbnails")
	}
	for _, path := range filepath.SplitList(list) {
		frame(t, path)
	}
}

func frame(t *testing.T, path string) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	src, err := png.Decode(f)
	_ = f.Close()
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)
	edge := color.RGBA{0xD0, 0xD7, 0xDE, 0xFF} // GitHub border gray
	for x := b.Min.X; x < b.Max.X; x++ {
		dst.Set(x, b.Min.Y, edge)
		dst.Set(x, b.Max.Y-1, edge)
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		dst.Set(b.Min.X, y, edge)
		dst.Set(b.Max.X-1, y, edge)
	}
	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(out, dst); err != nil {
		_ = out.Close()
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
}
