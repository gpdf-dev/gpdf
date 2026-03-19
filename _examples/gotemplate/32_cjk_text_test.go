package gotemplate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func loadCJKFont(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "..", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("CJK font not found: %s", path)
	}
	return data
}

func TestTmpl_32_CJK_Text(t *testing.T) {
	jpData := loadCJKFont(t, "NotoSansJP-Regular.ttf")
	scData := loadCJKFont(t, "NotoSansSC-Regular.ttf")
	krData := loadCJKFont(t, "NotoSansKR-Regular.ttf")

	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "CJK Text Examples", "author": "gpdf"},
		"defaultFont": {"family": "NotoSansJP", "size": 12},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "CJK Text Examples", "style": {"size": 24, "bold": true}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Japanese ({{.JA_Name}})", "style": {"size": 20, "bold": true, "color": "#0D47A1", "fontFamily": "NotoSansJP"}},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.JA_Hello}}", "style": {"fontFamily": "NotoSansJP"}},
					{"type": "text", "content": "{{.JA_Novel}}", "style": {"fontFamily": "NotoSansJP"}},
					{"type": "text", "content": "{{.JA_Address}}", "style": {"fontFamily": "NotoSansJP"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.JA_Hiragana}}", "style": {"fontFamily": "NotoSansJP"}},
					{"type": "text", "content": "{{.JA_Katakana}}", "style": {"fontFamily": "NotoSansJP"}},
					{"type": "text", "content": "{{.JA_Kanji}}", "style": {"fontFamily": "NotoSansJP"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "3mm"},
					{"type": "line", "thickness": "0.5pt", "color": "#B3B3B3"},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Chinese ({{.ZH_Name}})", "style": {"size": 20, "bold": true, "color": "#B71C1C", "fontFamily": "NotoSansSC"}},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.ZH_Hello}}", "style": {"fontFamily": "NotoSansSC"}},
					{"type": "text", "content": "{{.ZH_Quote}}", "style": {"fontFamily": "NotoSansSC"}},
					{"type": "text", "content": "{{.ZH_Address}}", "style": {"fontFamily": "NotoSansSC"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.ZH_Simplified}}", "style": {"fontFamily": "NotoSansSC"}},
					{"type": "text", "content": "{{.ZH_Traditional}}", "style": {"fontFamily": "NotoSansSC"}},
					{"type": "text", "content": "{{.ZH_Idiom}}", "style": {"fontFamily": "NotoSansSC"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "3mm"},
					{"type": "line", "thickness": "0.5pt", "color": "#B3B3B3"},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Korean ({{.KO_Name}})", "style": {"size": 20, "bold": true, "color": "#1B5E20", "fontFamily": "NotoSansKR"}},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.KO_Hello}}", "style": {"fontFamily": "NotoSansKR"}},
					{"type": "text", "content": "{{.KO_Address}}", "style": {"fontFamily": "NotoSansKR"}},
					{"type": "text", "content": "{{.KO_Alphabet}}", "style": {"fontFamily": "NotoSansKR"}}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.KO_Hangul}}", "style": {"fontFamily": "NotoSansKR"}},
					{"type": "text", "content": "{{.KO_Hanja}}", "style": {"fontFamily": "NotoSansKR"}},
					{"type": "text", "content": "{{.KO_Proverb}}", "style": {"fontFamily": "NotoSansKR"}}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "3mm"},
					{"type": "line", "thickness": "0.5pt", "color": "#B3B3B3"},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Mixed CJK Table", "style": {"size": 20, "bold": true, "color": "#4A148C"}},
					{"type": "spacer", "height": "3mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 3, "elements": [{"type": "text", "content": "Language", "style": {"bold": true, "color": "#FFFFFF", "background": "#37474F"}}]},
				{"span": 3, "elements": [{"type": "text", "content": "Hello", "style": {"bold": true, "color": "#FFFFFF", "background": "#37474F"}}]},
				{"span": 3, "elements": [{"type": "text", "content": "Thank you", "style": {"bold": true, "color": "#FFFFFF", "background": "#37474F"}}]},
				{"span": 3, "elements": [{"type": "text", "content": "Goodbye", "style": {"bold": true, "color": "#FFFFFF", "background": "#37474F"}}]}
			]}},
			{"row": {"cols": [
				{"span": 3, "elements": [{"type": "text", "content": "{{.JA_Lang}}", "style": {"fontFamily": "NotoSansJP"}}]},
				{"span": 3, "elements": [{"type": "text", "content": "{{.JA_HelloShort}}", "style": {"fontFamily": "NotoSansJP"}}]},
				{"span": 3, "elements": [{"type": "text", "content": "{{.JA_Thanks}}", "style": {"fontFamily": "NotoSansJP"}}]},
				{"span": 3, "elements": [{"type": "text", "content": "{{.JA_Bye}}", "style": {"fontFamily": "NotoSansJP"}}]}
			]}},
			{"row": {"cols": [
				{"span": 3, "elements": [{"type": "text", "content": "{{.ZH_Lang}}", "style": {"fontFamily": "NotoSansSC"}}]},
				{"span": 3, "elements": [{"type": "text", "content": "{{.ZH_HelloShort}}", "style": {"fontFamily": "NotoSansSC"}}]},
				{"span": 3, "elements": [{"type": "text", "content": "{{.ZH_Thanks}}", "style": {"fontFamily": "NotoSansSC"}}]},
				{"span": 3, "elements": [{"type": "text", "content": "{{.ZH_Bye}}", "style": {"fontFamily": "NotoSansSC"}}]}
			]}},
			{"row": {"cols": [
				{"span": 3, "elements": [{"type": "text", "content": "{{.KO_Lang}}", "style": {"fontFamily": "NotoSansKR"}}]},
				{"span": 3, "elements": [{"type": "text", "content": "{{.KO_HelloShort}}", "style": {"fontFamily": "NotoSansKR"}}]},
				{"span": 3, "elements": [{"type": "text", "content": "{{.KO_Thanks}}", "style": {"fontFamily": "NotoSansKR"}}]},
				{"span": 3, "elements": [{"type": "text", "content": "{{.KO_Bye}}", "style": {"fontFamily": "NotoSansKR"}}]}
			]}},
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "spacer", "height": "5mm"},
					{"type": "text", "content": "CJK characters are fully supported through TrueType font embedding.", "style": {"align": "center", "italic": true, "color": "#808080", "fontFamily": "NotoSansJP"}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"JA_Name": "日本語", "JA_Hello": "こんにちは世界",
		"JA_Novel": "吾輩は猫である。名前はまだ無い。", "JA_Address": "東京都渋谷区神宮前1-2-3",
		"JA_Hiragana": "ひらがな: あいうえお かきくけこ", "JA_Katakana": "カタカナ: アイウエオ カキクケコ",
		"JA_Kanji": "漢字: 春夏秋冬 東西南北",
		"ZH_Name":  "中文", "ZH_Hello": "你好世界",
		"ZH_Quote": "天行健，君子以自强不息。", "ZH_Address": "北京市朝阳区建国门外大街1号",
		"ZH_Simplified": "简体: 学习 计算机 人工智能", "ZH_Traditional": "繁體: 學習 計算機 人工智慧",
		"ZH_Idiom": "成语: 龙飞凤舞 画龙点睛",
		"KO_Name":  "한국어", "KO_Hello": "안녕하세요 세계",
		"KO_Address": "대한민국 서울특별시 강남구", "KO_Alphabet": "가나다라마바사아자차카타파하",
		"KO_Hangul": "한글: 가갸거겨고교구규그기", "KO_Hanja": "한자혼용: 大韓民國 서울特別市",
		"KO_Proverb": "속담: 천리 길도 한 걸음부터",
		"JA_Lang":    "日本語", "JA_HelloShort": "こんにちは", "JA_Thanks": "ありがとう", "JA_Bye": "さようなら",
		"ZH_Lang": "中文", "ZH_HelloShort": "你好", "ZH_Thanks": "谢谢", "ZH_Bye": "再见",
		"KO_Lang": "한국어", "KO_HelloShort": "안녕하세요", "KO_Thanks": "감사합니다", "KO_Bye": "안녕히 가세요",
	}

	doc, err := template.FromJSON(schema, data,
		template.WithFont("NotoSansJP", jpData),
		template.WithFont("NotoSansSC", scData),
		template.WithFont("NotoSansKR", krData),
		template.WithDefaultFont("NotoSansJP", 12),
	)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "32_cjk_text.pdf", doc)
}

func TestTmpl_32a_CJK_Japanese(t *testing.T) {
	fontData := loadCJKFont(t, "NotoSansJP-Regular.ttf")

	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "CJK Japanese Examples", "author": "gpdf"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Japanese ({{.JA_Name}})", "style": {"size": 20, "bold": true, "color": "#0D47A1"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.JA_Hello}}"},
					{"type": "text", "content": "{{.JA_Novel}}"},
					{"type": "text", "content": "{{.JA_Address}}"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.JA_Hiragana}}"},
					{"type": "text", "content": "{{.JA_Katakana}}"},
					{"type": "text", "content": "{{.JA_Kanji}}"}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"JA_Name":     "日本語",
		"JA_Hello":    "こんにちは世界",
		"JA_Novel":    "吾輩は猫である。名前はまだ無い。",
		"JA_Address":  "東京都渋谷区神宮前1-2-3",
		"JA_Hiragana": "ひらがな: あいうえお かきくけこ",
		"JA_Katakana": "カタカナ: アイウエオ カキクケコ",
		"JA_Kanji":    "漢字: 春夏秋冬 東西南北",
	}

	doc, err := template.FromJSON(schema, data,
		template.WithFont("NotoSansJP", fontData),
		template.WithDefaultFont("NotoSansJP", 12),
	)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "32a_cjk_japanese.pdf", doc)
}

func TestTmpl_32b_CJK_Chinese(t *testing.T) {
	fontData := loadCJKFont(t, "NotoSansSC-Regular.ttf")

	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "CJK Chinese Examples", "author": "gpdf"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Chinese ({{.ZH_Name}})", "style": {"size": 20, "bold": true, "color": "#B71C1C"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.ZH_Hello}}"},
					{"type": "text", "content": "{{.ZH_Quote}}"},
					{"type": "text", "content": "{{.ZH_Address}}"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.ZH_Simplified}}"},
					{"type": "text", "content": "{{.ZH_Traditional}}"},
					{"type": "text", "content": "{{.ZH_Idiom}}"}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"ZH_Name":        "中文",
		"ZH_Hello":       "你好世界",
		"ZH_Quote":       "天行健，君子以自强不息。",
		"ZH_Address":     "北京市朝阳区建国门外大街1号",
		"ZH_Simplified":  "简体: 学习 计算机 人工智能",
		"ZH_Traditional": "繁體: 學習 計算機 人工智慧",
		"ZH_Idiom":       "成语: 龙飞凤舞 画龙点睛",
	}

	doc, err := template.FromJSON(schema, data,
		template.WithFont("NotoSansSC", fontData),
		template.WithDefaultFont("NotoSansSC", 12),
	)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "32b_cjk_chinese.pdf", doc)
}

func TestTmpl_32c_CJK_Korean(t *testing.T) {
	fontData := loadCJKFont(t, "NotoSansKR-Regular.ttf")

	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "CJK Korean Examples", "author": "gpdf"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Korean ({{.KO_Name}})", "style": {"size": 20, "bold": true, "color": "#1B5E20"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.KO_Hello}}"},
					{"type": "text", "content": "{{.KO_Address}}"},
					{"type": "text", "content": "{{.KO_Alphabet}}"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "{{.KO_Hangul}}"},
					{"type": "text", "content": "{{.KO_Hanja}}"},
					{"type": "text", "content": "{{.KO_Proverb}}"}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"KO_Name":     "한국어",
		"KO_Hello":    "안녕하세요 세계",
		"KO_Address":  "대한민국 서울특별시 강남구",
		"KO_Alphabet": "가나다라마바사아자차카타파하",
		"KO_Hangul":   "한글: 가갸거겨고교구규그기",
		"KO_Hanja":    "한자혼용: 大韓民國 서울特別市",
		"KO_Proverb":  "속담: 천리 길도 한 걸음부터",
	}

	doc, err := template.FromJSON(schema, data,
		template.WithFont("NotoSansKR", fontData),
		template.WithDefaultFont("NotoSansKR", 12),
	)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "32c_cjk_korean.pdf", doc)
}
