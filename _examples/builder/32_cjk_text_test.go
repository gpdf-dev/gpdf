package builder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

// cjkFontRoot returns the project root directory where CJK fonts are stored.
func cjkFontRoot() string {
	// gpdf/_examples/builder/ → gpdf-dev/
	return filepath.Join("..", "..", "..")
}

func loadFont(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join(cjkFontRoot(), filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("CJK font not found: %s (run from project root)", path)
	}
	return data
}

func TestExample_32a_CJK_Japanese(t *testing.T) {
	fontData := loadFont(t, "NotoSansJP-Regular.ttf")

	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithFont("NotoSansJP", fontData),
		template.WithDefaultFont("NotoSansJP", 12),
		template.WithMetadata(document.DocumentMetadata{
			Title:  "CJK Japanese Examples",
			Author: "gpdf",
		}),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Japanese (日本語)", template.FontSize(20), template.Bold(),
				template.TextColor(pdf.RGBHex(0x0D47A1)))
			c.Spacer(document.Mm(5))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("こんにちは世界")
			c.Text("吾輩は猫である。名前はまだ無い。")
			c.Text("東京都渋谷区神宮前1-2-3")
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("ひらがな: あいうえお かきくけこ")
			c.Text("カタカナ: アイウエオ カキクケコ")
			c.Text("漢字: 春夏秋冬 東西南北")
		})
	})

	testutil.GeneratePDFSharedGolden(t, "32a_cjk_japanese.pdf", doc)
}

func TestExample_32b_CJK_Chinese(t *testing.T) {
	fontData := loadFont(t, "NotoSansSC-Regular.ttf")

	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithFont("NotoSansSC", fontData),
		template.WithDefaultFont("NotoSansSC", 12),
		template.WithMetadata(document.DocumentMetadata{
			Title:  "CJK Chinese Examples",
			Author: "gpdf",
		}),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Chinese (中文)", template.FontSize(20), template.Bold(),
				template.TextColor(pdf.RGBHex(0xB71C1C)))
			c.Spacer(document.Mm(5))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("你好世界")
			c.Text("天行健，君子以自强不息。")
			c.Text("北京市朝阳区建国门外大街1号")
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("简体: 学习 计算机 人工智能")
			c.Text("繁體: 學習 計算機 人工智慧")
			c.Text("成语: 龙飞凤舞 画龙点睛")
		})
	})

	testutil.GeneratePDFSharedGolden(t, "32b_cjk_chinese.pdf", doc)
}

func TestExample_32c_CJK_Korean(t *testing.T) {
	fontData := loadFont(t, "NotoSansKR-Regular.ttf")

	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithFont("NotoSansKR", fontData),
		template.WithDefaultFont("NotoSansKR", 12),
		template.WithMetadata(document.DocumentMetadata{
			Title:  "CJK Korean Examples",
			Author: "gpdf",
		}),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Korean (한국어)", template.FontSize(20), template.Bold(),
				template.TextColor(pdf.RGBHex(0x1B5E20)))
			c.Spacer(document.Mm(5))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("안녕하세요 세계")
			c.Text("대한민국 서울특별시 강남구")
			c.Text("가나다라마바사아자차카타파하")
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("한글: 가갸거겨고교구규그기")
			c.Text("한자혼용: 大韓民國 서울特別市")
			c.Text("속담: 천리 길도 한 걸음부터")
		})
	})

	testutil.GeneratePDFSharedGolden(t, "32c_cjk_korean.pdf", doc)
}

func TestExample_32d_CJK_Mixed(t *testing.T) {
	jpData := loadFont(t, "NotoSansJP-Regular.ttf")
	scData := loadFont(t, "NotoSansSC-Regular.ttf")
	krData := loadFont(t, "NotoSansKR-Regular.ttf")

	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithFont("NotoSansJP", jpData),
		template.WithFont("NotoSansSC", scData),
		template.WithFont("NotoSansKR", krData),
		template.WithDefaultFont("NotoSansJP", 12),
		template.WithMetadata(document.DocumentMetadata{
			Title:  "CJK Mixed Table",
			Author: "gpdf",
		}),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Mixed CJK Table", template.FontSize(20), template.Bold(),
				template.TextColor(pdf.RGBHex(0x4A148C)))
			c.Spacer(document.Mm(5))
		})
	})

	// Japanese row
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("日本語", template.FontFamily("NotoSansJP"))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("こんにちは", template.FontFamily("NotoSansJP"))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("ありがとう", template.FontFamily("NotoSansJP"))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("さようなら", template.FontFamily("NotoSansJP"))
		})
	})

	// Chinese row
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("中文", template.FontFamily("NotoSansSC"))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("你好", template.FontFamily("NotoSansSC"))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("谢谢", template.FontFamily("NotoSansSC"))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("再见", template.FontFamily("NotoSansSC"))
		})
	})

	// Korean row
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("한국어", template.FontFamily("NotoSansKR"))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("안녕하세요", template.FontFamily("NotoSansKR"))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("감사합니다", template.FontFamily("NotoSansKR"))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("안녕히 가세요", template.FontFamily("NotoSansKR"))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(5))
			c.Text("CJK characters are fully supported through TrueType font embedding.",
				template.AlignCenter(), template.Italic(), template.TextColor(pdf.Gray(0.5)),
				template.FontFamily("NotoSansJP"))
		})
	})

	testutil.GeneratePDFSharedGolden(t, "32d_cjk_mixed.pdf", doc)
}
