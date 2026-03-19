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
