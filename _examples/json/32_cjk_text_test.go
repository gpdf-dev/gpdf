package json_test

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

func TestJSON_32a_CJK_Japanese(t *testing.T) {
	fontData := loadCJKFont(t, "NotoSansJP-Regular.ttf")

	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "CJK Japanese Examples", "author": "gpdf"},
		"defaultFont": {"family": "NotoSansJP", "size": 12},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Japanese (日本語)", "style": {"size": 20, "bold": true, "color": "#0D47A1"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "こんにちは世界"},
					{"type": "text", "content": "吾輩は猫である。名前はまだ無い。"},
					{"type": "text", "content": "東京都渋谷区神宮前1-2-3"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "ひらがな: あいうえお かきくけこ"},
					{"type": "text", "content": "カタカナ: アイウエオ カキクケコ"},
					{"type": "text", "content": "漢字: 春夏秋冬 東西南北"}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil,
		template.WithFont("NotoSansJP", fontData),
		template.WithDefaultFont("NotoSansJP", 12),
	)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "32a_cjk_japanese.pdf", doc)
}

func TestJSON_32b_CJK_Chinese(t *testing.T) {
	fontData := loadCJKFont(t, "NotoSansSC-Regular.ttf")

	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "CJK Chinese Examples", "author": "gpdf"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Chinese (中文)", "style": {"size": 20, "bold": true, "color": "#B71C1C"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "你好世界"},
					{"type": "text", "content": "天行健，君子以自强不息。"},
					{"type": "text", "content": "北京市朝阳区建国门外大街1号"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "简体: 学习 计算机 人工智能"},
					{"type": "text", "content": "繁體: 學習 計算機 人工智慧"},
					{"type": "text", "content": "成语: 龙飞凤舞 画龙点睛"}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil,
		template.WithFont("NotoSansSC", fontData),
		template.WithDefaultFont("NotoSansSC", 12),
	)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "32b_cjk_chinese.pdf", doc)
}

func TestJSON_32c_CJK_Korean(t *testing.T) {
	fontData := loadCJKFont(t, "NotoSansKR-Regular.ttf")

	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"metadata": {"title": "CJK Korean Examples", "author": "gpdf"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Korean (한국어)", "style": {"size": 20, "bold": true, "color": "#1B5E20"}},
					{"type": "spacer", "height": "5mm"}
				]}
			]}},
			{"row": {"cols": [
				{"span": 6, "elements": [
					{"type": "text", "content": "안녕하세요 세계"},
					{"type": "text", "content": "대한민국 서울특별시 강남구"},
					{"type": "text", "content": "가나다라마바사아자차카타파하"}
				]},
				{"span": 6, "elements": [
					{"type": "text", "content": "한글: 가갸거겨고교구규그기"},
					{"type": "text", "content": "한자혼용: 大韓民國 서울特別市"},
					{"type": "text", "content": "속담: 천리 길도 한 걸음부터"}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil,
		template.WithFont("NotoSansKR", fontData),
		template.WithDefaultFont("NotoSansKR", 12),
	)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "32c_cjk_korean.pdf", doc)
}
