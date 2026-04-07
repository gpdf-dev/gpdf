# gpdf

[![Go Reference](https://pkg.go.dev/badge/github.com/gpdf-dev/gpdf.svg)](https://pkg.go.dev/github.com/gpdf-dev/gpdf)
[![CI](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml/badge.svg)](https://github.com/gpdf-dev/gpdf/actions/workflows/check-code.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gpdf-dev/gpdf)](https://goreportcard.com/report/github.com/gpdf-dev/gpdf)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.22-blue)](https://go.dev/)
[![Website](https://img.shields.io/badge/Website-gpdf.dev-blue)](https://gpdf.dev/)

[English](README.md) | [日本語](README_ja.md) | [中文](README_zh.md) | **한국어** | [Español](README_es.md) | [Português](README_pt.md)

순수 Go로 구현된 외부 의존성 없는 PDF 생성 라이브러리. 레이어드 아키텍처와 선언적 빌더 API를 제공합니다.

## 특징

- **외부 의존성 제로** — Go 표준 라이브러리만 사용
- **레이어드 아키텍처** — 저수준 PDF 프리미티브, 문서 모델, 고수준 템플릿 API
- **12컬럼 그리드 시스템** — Bootstrap 스타일의 반응형 레이아웃
- **TrueType 폰트 지원** — 커스텀 폰트 임베딩 및 서브셋팅
- **CJK 지원** — 첫날부터 한중일 텍스트 완벽 지원
- **테이블** — 헤더, 컬럼 너비, 줄무늬 행, 수직 정렬
- **머리글 및 바닥글** — 페이지 번호 포함, 모든 페이지에서 일관된 표시
- **리스트** — 글머리 기호 목록 및 번호 목록
- **QR 코드** — 순수 Go QR 코드 생성 (오류 정정 레벨 지원)
- **바코드** — Code 128 바코드 생성
- **텍스트 장식** — 밑줄, 취소선, 자간, 들여쓰기
- **페이지 번호** — 자동 페이지 번호 및 전체 페이지 수
- **Go 템플릿 통합** — Go 템플릿에서 PDF 생성
- **재사용 가능 컴포넌트** — 송장, 보고서, 레터 프리셋 템플릿 내장
- **JSON 스키마** — JSON으로만 문서 정의
- **다양한 단위** — pt, mm, cm, in, em, %
- **색상 공간** — RGB, 그레이스케일, CMYK
- **이미지** — JPEG 및 PNG 임베딩 (맞춤 옵션 지원)
- **절대 위치 지정** — 페이지의 정확한 XY 좌표에 요소 배치
- **기존 PDF 오버레이** — 기존 PDF를 열어 텍스트, 이미지, 스탬프를 위에 추가
- **폼 플래트닝** — AcroForm 필드를 정적 페이지 콘텐츠로 변환, 비위젯 주석은 보존
- **PDF 병합** — 여러 PDF를 페이지 범위 선택으로 하나로 결합
- **문서 메타데이터** — 제목, 저자, 주제, 작성자
- **암호화** — AES-256 암호화 (ISO 32000-2, Rev 6), 소유자/사용자 비밀번호 및 권한 제어
- **PDF/A** — PDF/A-1b 및 PDF/A-2b 준수, ICC 프로파일 및 XMP 메타데이터 포함
- **디지털 서명** — CMS/PKCS#7 서명, RSA/ECDSA 키 및 RFC 3161 타임스탬프 지원

## 벤치마크

[go-pdf/fpdf](https://github.com/go-pdf/fpdf), [signintech/gopdf](https://github.com/signintech/gopdf), [maroto v2](https://github.com/johnfercher/maroto)와 비교.
5회 실행 중앙값, 각 100회 반복. Apple M1, Go 1.25.

**실행 시간** (낮을수록 좋음):

| 벤치마크 | gpdf | fpdf | gopdf | maroto v2 |
|---|--:|--:|--:|--:|
| 단일 페이지 | **13 µs** | 132 µs | 423 µs | 237 µs |
| 테이블 (4x10) | **108 µs** | 241 µs | 835 µs | 8.6 ms |
| 100페이지 | **683 µs** | 11.7 ms | 8.6 ms | 19.8 ms |
| 복합 문서 | **133 µs** | 254 µs | 997 µs | 10.4 ms |

**메모리 사용량** (낮을수록 좋음):

| 벤치마크 | gpdf | fpdf | gopdf | maroto v2 |
|---|--:|--:|--:|--:|
| 단일 페이지 | **16 KB** | 1.2 MB | 1.8 MB | 61 KB |
| 테이블 (4x10) | **209 KB** | 1.3 MB | 1.9 MB | 1.6 MB |
| 100페이지 | **909 KB** | 121 MB | 83 MB | 4.0 MB |
| 복합 문서 | **246 KB** | 1.3 MB | 2.0 MB | 2.0 MB |

### gpdf가 빠른 이유

- **단일 페이지** — 빌드→레이아웃→렌더링 단일 패스 파이프라인으로 중간 데이터 구조가 없음. 전체적으로 구체적인 구조체 타입 사용(`interface{}` 박싱 없음)으로 최소한의 힙 할당으로 문서 트리를 구축.
- **테이블** — 셀 내용을 재사용 가능한 `strings.Builder` 버퍼를 통해 PDF 콘텐츠 스트림 명령으로 직접 기록. 셀별 객체 래핑이나 반복적인 폰트 조회가 없으며, 폰트는 문서당 한 번만 해석.
- **100페이지** — 레이아웃이 O(n)으로 선형 확장. 오버플로우 페이지네이션이 슬라이스 참조로 나머지 노드를 전달(딥 카피 없음). 폰트는 한 번만 파싱되어 모든 페이지에서 공유.
- **복합 문서** — 재측정 없는 단일 패스 레이아웃이 위의 모든 장점을 통합. 폰트 서브셋팅은 실제 사용된 글리프만 임베딩하고, Flate 압축이 기본 적용되어 메모리와 출력 크기 모두 작게 유지.

벤치마크 실행:

```bash
cd _benchmark && go test -bench=. -benchmem -count=5
```

## 아키텍처

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

## 요구사항

- Go 1.22 이상

## 설치

```bash
go get github.com/gpdf-dev/gpdf
```

## 빠른 시작

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

## 사용 예시

### 텍스트 스타일링

글꼴 크기, 두께, 스타일, 색상, 배경색, 정렬:

```go
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("큰 굵은 제목", template.FontSize(24), template.Bold())
		c.Text("이탤릭 텍스트", template.Italic())
		c.Text("굵게 + 이탤릭", template.Bold(), template.Italic())
		c.Text("빨간 텍스트", template.TextColor(pdf.Red))
		c.Text("커스텀 색상", template.TextColor(pdf.RGBHex(0x336699)))
		c.Text("배경색 포함", template.BgColor(pdf.Yellow))
		c.Text("가운데 정렬", template.AlignCenter())
		c.Text("오른쪽 정렬", template.AlignRight())
	})
})
```

### CJK 폰트 (한국어 / 일본어 / 중국어)

CJK 텍스트 렌더링에는 TrueType 폰트 임베딩이 필요합니다. 각 언어에 맞는 Noto Sans 폰트를 사용합니다:

```go
fontData, _ := os.ReadFile("NotoSansKR-Regular.ttf")

doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithFont("NotoSansKR", fontData),
	gpdf.WithDefaultFont("NotoSansKR", 12),
)

page := doc.AddPage()
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("안녕하세요 세계", template.FontSize(18))
	})
})
```

다국어 문서의 경우, 여러 폰트를 등록하고 `FontFamily()`로 전환합니다:

```go
jpFont, _ := os.ReadFile("NotoSansJP-Regular.ttf")
scFont, _ := os.ReadFile("NotoSansSC-Regular.ttf")
krFont, _ := os.ReadFile("NotoSansKR-Regular.ttf")

doc := gpdf.NewDocument(
	gpdf.WithFont("NotoSansJP", jpFont),
	gpdf.WithFont("NotoSansSC", scFont),
	gpdf.WithFont("NotoSansKR", krFont),
	gpdf.WithDefaultFont("NotoSansKR", 12),
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

추천 폰트 (모두 무료, OFL 라이선스):

| 폰트 | 언어 |
|---|---|
| [Noto Sans JP](https://fonts.google.com/noto/specimen/Noto+Sans+JP) | 일본어 |
| [Noto Sans SC](https://fonts.google.com/noto/specimen/Noto+Sans+SC) | 간체 중국어 |
| [Noto Sans KR](https://fonts.google.com/noto/specimen/Noto+Sans+KR) | 한국어 |

### 12컬럼 그리드 레이아웃

Bootstrap 스타일의 12컬럼 그리드로 레이아웃 구성:

```go
// 2등분 컬럼
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(6, func(c *template.ColBuilder) {
		c.Text("왼쪽 절반")
	})
	r.Col(6, func(c *template.ColBuilder) {
		c.Text("오른쪽 절반")
	})
})

// 사이드바 + 메인 콘텐츠
page.AutoRow(func(r *template.RowBuilder) {
	r.Col(3, func(c *template.ColBuilder) {
		c.Text("사이드바")
	})
	r.Col(9, func(c *template.ColBuilder) {
		c.Text("메인 콘텐츠")
	})
})
```

### 고정 높이 행

`Row()`로 높이를 지정하거나, `AutoRow()`로 콘텐츠에 맞춰 자동 조절:

```go
// 고정 높이: 30mm
page.Row(document.Mm(30), func(r *template.RowBuilder) {
	r.Col(12, func(c *template.ColBuilder) {
		c.Text("이 행의 높이는 30mm입니다")
	})
})
```

### 테이블

기본 테이블:

```go
c.Table(
	[]string{"상품명", "수량", "가격"},
	[][]string{
		{"위젯", "10", "₩5,000"},
		{"가젯", "3", "₩12,000"},
	},
)
```

스타일 적용 테이블 (헤더 색상, 컬럼 너비, 줄무늬 행):

```go
c.Table(
	[]string{"제품", "카테고리", "수량", "단가", "합계"},
	[][]string{
		{"노트북 Pro 15", "전자기기", "2", "₩1,299,000", "₩2,598,000"},
		{"무선 마우스", "주변기기", "10", "₩29,900", "₩299,000"},
	},
	template.ColumnWidths(30, 20, 10, 20, 20),
	template.TableHeaderStyle(
		template.TextColor(pdf.White),
		template.BgColor(pdf.RGBHex(0x1A237E)),
	),
	template.TableStripe(pdf.RGBHex(0xF5F5F5)),
)
```

### 이미지

JPEG 및 PNG 이미지 임베딩 (맞춤 옵션 지원):

```go
c.Image(imgData)                                      // 기본 크기
c.Image(imgData, template.FitWidth(document.Mm(80)))   // 너비에 맞춤
c.Image(imgData, template.FitHeight(document.Mm(30)))  // 높이에 맞춤
```

### 선 및 간격

```go
c.Line()                                           // 기본 (회색, 1pt)
c.Line(template.LineColor(pdf.Red))                 // 색상 지정
c.Line(template.LineThickness(document.Pt(3)))      // 굵은 선
c.Spacer(document.Mm(5))                            // 5mm 세로 간격
```

### 리스트

글머리 기호 목록 및 번호 목록:

```go
// 글머리 기호 목록
c.List([]string{"첫 번째 항목", "두 번째 항목", "세 번째 항목"})

// 번호 목록
c.OrderedList([]string{"단계 1", "단계 2", "단계 3"})
```

### QR 코드

크기와 오류 정정 레벨을 설정할 수 있는 QR 코드 생성:

```go
// 기본 QR 코드
c.QRCode("https://gpdf.dev")

// 크기와 오류 정정 레벨 지정
c.QRCode("https://gpdf.dev",
	template.QRSize(document.Mm(30)),
	template.QRErrorCorrection(qrcode.LevelH))
```

### 바코드

Code 128 바코드 생성:

```go
// 기본 바코드
c.Barcode("INV-2026-0001")

// 너비 지정
c.Barcode("INV-2026-0001", template.BarcodeWidth(document.Mm(80)))
```

### 페이지 번호

자동 페이지 번호 및 전체 페이지 수:

```go
doc.Footer(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("gpdf로 생성", template.FontSize(8))
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

### 텍스트 장식

밑줄, 취소선, 자간, 들여쓰기:

```go
c.Text("밑줄 텍스트", template.Underline())
c.Text("취소선 텍스트", template.Strikethrough())
c.Text("넓은 자간", template.LetterSpacing(3))
c.Text("들여쓰기 단락...", template.TextIndent(document.Pt(24)))
```

### 머리글 및 바닥글

모든 페이지에 반복 표시되는 머리글과 바닥글:

```go
doc.Header(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("ACME 주식회사", template.Bold(), template.FontSize(10))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("대외비", template.AlignRight(), template.FontSize(10),
				template.TextColor(pdf.Gray(0.5)))
		})
	})
})

doc.Footer(func(p *template.PageBuilder) {
	p.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("gpdf로 생성", template.AlignCenter(),
				template.FontSize(8), template.TextColor(pdf.Gray(0.5)))
		})
	})
})
```

### 다중 페이지 문서

```go
for i := 1; i <= 5; i++ {
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("페이지 콘텐츠")
		})
	})
}
```

### 재사용 가능 컴포넌트

함수 하나로 일반적인 문서 유형을 생성할 수 있습니다:

**송장:**

```go
doc := template.Invoice(template.InvoiceData{
	Number:  "#INV-2026-001",
	Date:    "2026년 3월 1일",
	DueDate: "2026년 3월 31일",
	From:    template.InvoiceParty{Name: "ACME 주식회사", Address: []string{"서울시 강남구 123"}},
	To:      template.InvoiceParty{Name: "클라이언트 주식회사", Address: []string{"부산시 해운대구 456"}},
	Items: []template.InvoiceItem{
		{Description: "웹 개발", Quantity: "40시간", UnitPrice: 150, Amount: 6000},
		{Description: "UI/UX 디자인", Quantity: "20시간", UnitPrice: 120, Amount: 2400},
	},
	TaxRate: 10,
	Notes:   "이용해 주셔서 감사합니다!",
})
data, _ := doc.Generate()
```

**보고서:**

```go
doc := template.Report(template.ReportData{
	Title:    "분기 보고서",
	Subtitle: "2026년 Q1",
	Author:   "ACME 주식회사",
	Sections: []template.ReportSection{
		{
			Title:   "경영진 요약",
			Content: "2025년 Q4 대비 매출이 15% 증가했습니다.",
			Metrics: []template.ReportMetric{
				{Label: "매출", Value: "₩12.5M", ColorHex: 0x2E7D32},
				{Label: "성장률", Value: "+15%", ColorHex: 0x2E7D32},
			},
		},
		{
			Title: "매출 내역",
			Table: &template.ReportTable{
				Header: []string{"사업부", "2026 Q1", "변화"},
				Rows:   [][]string{{"클라우드", "₩5.2M", "+26.8%"}, {"엔터프라이즈", "₩3.8M", "+8.6%"}},
			},
		},
	},
})
```

**레터:**

```go
doc := template.Letter(template.LetterData{
	From:     template.LetterParty{Name: "ACME 주식회사", Address: []string{"서울시 강남구 123"}},
	To:       template.LetterParty{Name: "김철수 님", Address: []string{"부산시 해운대구 456"}},
	Date:     "2026년 3월 1일",
	Subject:  "파트너십 제안",
	Greeting: "김철수 님께,",
	Body:     []string{"전략적 파트너십을 제안드리고자 합니다..."},
	Closing:  "감사합니다,",
	Signature: "이영희",
})
```

### 암호화

AES-256 암호화, 소유자/사용자 비밀번호 및 권한 제어:

```go
// 소유자 비밀번호만 (비밀번호 없이 PDF를 열 수 있지만 편집은 제한)
doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithEncryption(
		encrypt.WithOwnerPassword("owner-secret"),
	),
)

// 양쪽 비밀번호와 권한 제어
doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithEncryption(
		encrypt.WithOwnerPassword("owner-pass"),
		encrypt.WithUserPassword("user-pass"),
		encrypt.WithPermissions(encrypt.PermPrint|encrypt.PermCopy),
	),
)
```

### PDF/A 준수

PDF/A-1b 또는 PDF/A-2b 준수 문서 생성:

```go
doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithPDFA(
		pdfa.WithLevel(pdfa.LevelA2b),
		pdfa.WithMetadata(pdfa.MetadataInfo{
			Title:  "아카이브 보고서",
			Author: "gpdf",
		}),
	),
)
```

### 디지털 서명

RSA 또는 ECDSA 키로 CMS/PKCS#7 서명:

```go
data, _ := doc.Generate()

signed, err := gpdf.SignDocument(data, signature.Signer{
	Certificate: cert,
	PrivateKey:  key,
	Chain:       intermediates,
},
	signature.WithReason("승인됨"),
	signature.WithLocation("서울"),
	signature.WithTimestamp("http://tsa.example.com"),
)
```

### 기존 PDF 오버레이

기존 PDF를 열어 동일한 빌더 API로 콘텐츠를 오버레이:

```go
// 기존 PDF 열기
doc, err := gpdf.Open(existingPDFBytes)

// 1페이지에 "DRAFT" 워터마크 추가
doc.Overlay(0, func(p *template.PageBuilder) {
	p.Absolute(document.Mm(50), document.Mm(140), func(c *template.ColBuilder) {
		c.Text("DRAFT", template.FontSize(72),
			template.TextColor(pdf.Gray(0.85)))
	})
})

// 모든 페이지에 페이지 번호 추가
count, _ := doc.PageCount()
doc.EachPage(func(i int, p *template.PageBuilder) {
	p.Absolute(document.Mm(170), document.Mm(285), func(c *template.ColBuilder) {
		c.Text(fmt.Sprintf("%d / %d", i+1, count), template.FontSize(10))
	}, template.AbsoluteWidth(document.Mm(20)))
})

result, _ := doc.Save()
```

### JSON 스키마

JSON으로만 문서를 정의:

```go
schema := []byte(`{
	"page": {"size": "A4", "margins": "20mm"},
	"metadata": {"title": "보고서", "author": "gpdf"},
	"body": [
		{"row": {"cols": [
			{"span": 12, "text": "JSON에서 안녕하세요", "style": {"size": 24, "bold": true}}
		]}},
		{"row": {"cols": [
			{"span": 12, "table": {
				"header": ["이름", "값"],
				"rows": [["Alpha", "100"], ["Beta", "200"]],
				"headerStyle": {"bold": true, "color": "white", "background": "#1A237E"}
			}}
		]}}
	]
}`)

doc, err := template.FromJSON(schema, nil)
data, _ := doc.Generate()
```

### Go 템플릿 통합

Go 템플릿과 JSON 스키마로 동적 콘텐츠 생성:

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

data := map[string]any{"Title": "동적 보고서"}
doc, err := template.FromJSON(schema, data)
```

사전 파싱된 Go 템플릿으로 더 세밀한 제어:

```go
tmpl, _ := gotemplate.New("doc").Funcs(template.TemplateFuncMap()).Parse(schemaStr)
doc, err := template.FromTemplate(tmpl, data)
```

### 폼 플래트닝

인터랙티브 AcroForm 필드를 정적 페이지 콘텐츠로 플래트닝. 링크, 댓글 등 비위젯 주석은 보존됩니다:

```go
// 폼 필드가 있는 PDF 열기
doc, err := gpdf.Open(filledFormPDF)

// 모든 폼 필드를 정적 콘텐츠로 플래트닝
if err := doc.FlattenForms(); err != nil {
	log.Fatal(err)
}

result, _ := doc.Save()
```

### PDF 병합

여러 PDF를 페이지 범위 선택으로 하나의 문서로 결합:

```go
// 여러 PDF 합치기
merged, _ := gpdf.Merge(
	[]gpdf.Source{
		{Data: coverPage},
		{Data: report},
		{Data: appendix, Pages: gpdf.PageRange{From: 1, To: 3}}, // 처음 3페이지만
	},
	gpdf.WithMergeMetadata("My Document", "Author", ""),
)
```

### 문서 메타데이터

```go
doc := gpdf.NewDocument(
	gpdf.WithPageSize(gpdf.A4),
	gpdf.WithMetadata(document.DocumentMetadata{
		Title:   "연간 보고서 2026",
		Author:  "gpdf Library",
		Subject: "문서 메타데이터 예시",
		Creator: "My Application",
	}),
)
```

### 페이지 크기 및 여백

```go
// 사용 가능한 페이지 크기
document.A4      // 210mm x 297mm
document.A3      // 297mm x 420mm
document.Letter  // 8.5in x 11in
document.Legal   // 8.5in x 14in

// 균일 여백
template.WithMargins(document.UniformEdges(document.Mm(20)))

// 비대칭 여백
template.WithMargins(document.Edges{
	Top:    document.Mm(10),
	Right:  document.Mm(40),
	Bottom: document.Mm(10),
	Left:   document.Mm(40),
})
```

### 출력 옵션

```go
// Generate는 []byte를 반환
data, err := doc.Generate()

// Render는 임의의 io.Writer에 기록
var buf bytes.Buffer
err := doc.Render(&buf)

// 파일에 직접 기록
f, _ := os.Create("output.pdf")
defer f.Close()
doc.Render(f)
```

## API 레퍼런스

### 문서 옵션

| 함수 | 설명 |
|---|---|
| `WithPageSize(size)` | 페이지 크기 설정 (A4, A3, Letter, Legal) |
| `WithMargins(edges)` | 페이지 여백 설정 |
| `WithFont(family, data)` | TrueType 폰트 등록 |
| `WithDefaultFont(family, size)` | 기본 폰트 설정 |
| `WithMetadata(meta)` | 문서 메타데이터 설정 |
| `WithEncryption(opts...)` | AES-256 암호화 활성화 |
| `WithPDFA(opts...)` | PDF/A 준수 활성화 |

### 컬럼 콘텐츠

| 메서드 | 설명 |
|---|---|
| `c.Text(text, opts...)` | 스타일 옵션을 포함한 텍스트 추가 |
| `c.Table(header, rows, opts...)` | 테이블 추가 |
| `c.Image(data, opts...)` | 이미지 추가 (JPEG/PNG) |
| `c.QRCode(data, opts...)` | QR 코드 추가 |
| `c.Barcode(data, opts...)` | 바코드 추가 (Code 128) |
| `c.List(items, opts...)` | 글머리 기호 목록 추가 |
| `c.OrderedList(items, opts...)` | 번호 목록 추가 |
| `c.PageNumber(opts...)` | 현재 페이지 번호 추가 |
| `c.TotalPages(opts...)` | 전체 페이지 수 추가 |
| `c.Line(opts...)` | 수평선 추가 |
| `c.Spacer(height)` | 수직 공간 추가 |

### 페이지 레벨 콘텐츠

| 메서드 | 설명 |
|---|---|
| `page.AutoRow(fn)` | 자동 높이 행 추가 |
| `page.Row(height, fn)` | 고정 높이 행 추가 |
| `page.Absolute(x, y, fn, opts...)` | 정확한 XY 좌표에 콘텐츠 배치 |

#### 절대 위치 지정 옵션

| 옵션 | 설명 |
|---|---|
| `gpdf.AbsoluteWidth(value)` | 명시적 너비 설정 (기본값: 남은 공간) |
| `gpdf.AbsoluteHeight(value)` | 명시적 높이 설정 (기본값: 남은 공간) |
| `gpdf.AbsoluteOriginPage()` | 콘텐츠 영역 대신 페이지 모서리를 원점으로 사용 |

### 기존 PDF 작업

| 함수 / 메서드 | 설명 |
|---|---|
| `gpdf.Open(data, opts...)` | 기존 PDF를 오버레이용으로 열기 |
| `doc.PageCount()` | 페이지 수 가져오기 |
| `doc.Overlay(page, fn)` | 특정 페이지에 콘텐츠 오버레이 |
| `doc.EachPage(fn)` | 모든 페이지에 오버레이 적용 |
| `doc.FlattenForms()` | AcroForm 필드를 정적 페이지 콘텐츠로 플래트닝 |
| `doc.Save()` | 수정된 PDF 저장 |
| `gpdf.Merge(sources, opts...)` | 여러 PDF를 하나로 병합 |
| `WithMergeMetadata(title, author, producer)` | 병합된 출력의 메타데이터 설정 |

### 텍스트 옵션

| 옵션 | 설명 |
|---|---|
| `template.FontSize(size)` | 글꼴 크기를 포인트 단위로 설정 |
| `template.Bold()` | 굵게 |
| `template.Italic()` | 이탤릭 |
| `template.FontFamily(name)` | 등록된 폰트 사용 |
| `template.TextColor(color)` | 텍스트 색상 설정 |
| `template.BgColor(color)` | 배경 색상 설정 |
| `template.Underline()` | 밑줄 장식 |
| `template.Strikethrough()` | 취소선 장식 |
| `template.LetterSpacing(pts)` | 자간 설정 (포인트) |
| `template.TextIndent(value)` | 첫줄 들여쓰기 설정 |
| `template.AlignLeft()` | 왼쪽 정렬 (기본값) |
| `template.AlignCenter()` | 가운데 정렬 |
| `template.AlignRight()` | 오른쪽 정렬 |

### 테이블 옵션

| 옵션 | 설명 |
|---|---|
| `template.ColumnWidths(w...)` | 컬럼 너비를 백분율로 설정 |
| `template.TableHeaderStyle(opts...)` | 헤더 행 스타일 설정 |
| `template.TableStripe(color)` | 교차 행 색상 설정 |
| `template.TableCellVAlign(align)` | 셀 수직 정렬 (Top/Middle/Bottom) |

### 이미지 옵션

| 옵션 | 설명 |
|---|---|
| `template.FitWidth(value)` | 너비에 맞춰 스케일 (비율 유지) |
| `template.FitHeight(value)` | 높이에 맞춰 스케일 (비율 유지) |

### QR 코드 옵션

| 옵션 | 설명 |
|---|---|
| `template.QRSize(value)` | QR 코드 크기 설정 |
| `template.QRErrorCorrection(level)` | 오류 정정 레벨 설정 (L/M/Q/H) |
| `template.QRScale(n)` | 모듈 스케일 팩터 설정 |

### 바코드 옵션

| 옵션 | 설명 |
|---|---|
| `template.BarcodeWidth(value)` | 바코드 너비 설정 |
| `template.BarcodeHeight(value)` | 바코드 높이 설정 |
| `template.BarcodeFormat(fmt)` | 바코드 형식 설정 (Code 128) |

### 암호화 옵션

| 옵션 | 설명 |
|---|---|
| `encrypt.WithOwnerPassword(pw)` | 소유자 비밀번호 설정 |
| `encrypt.WithUserPassword(pw)` | 사용자 비밀번호 설정 |
| `encrypt.WithPermissions(perm)` | 문서 권한 설정 (PermPrint, PermCopy, PermModify 등) |

### PDF/A 옵션

| 옵션 | 설명 |
|---|---|
| `pdfa.WithLevel(level)` | 준수 레벨 설정 (LevelA1b, LevelA2b) |
| `pdfa.WithMetadata(info)` | XMP 메타데이터 설정 (Title, Author, Subject 등) |

### 디지털 서명

| 함수 / 옵션 | 설명 |
|---|---|
| `gpdf.SignDocument(data, signer, opts...)` | 디지털 서명으로 PDF 서명 |
| `signature.WithReason(reason)` | 서명 사유 설정 |
| `signature.WithLocation(location)` | 서명 위치 설정 |
| `signature.WithTimestamp(tsaURL)` | RFC 3161 타임스탬프 활성화 |
| `signature.WithSignTime(t)` | 서명 시각 설정 |

### 템플릿 생성

| 함수 | 설명 |
|---|---|
| `template.FromJSON(schema, data)` | JSON 스키마에서 문서 생성 |
| `template.FromTemplate(tmpl, data)` | Go 템플릿에서 문서 생성 |
| `template.TemplateFuncMap()` | 템플릿 헬퍼 함수 가져오기 (`toJSON` 포함) |

### 재사용 가능 컴포넌트

| 함수 | 설명 |
|---|---|
| `template.Invoice(data)` | 전문적인 송장 PDF 생성 |
| `template.Report(data)` | 구조화된 보고서 PDF 생성 |
| `template.Letter(data)` | 비즈니스 레터 PDF 생성 |

### 선 옵션

| 옵션 | 설명 |
|---|---|
| `template.LineColor(color)` | 선 색상 설정 |
| `template.LineThickness(value)` | 선 두께 설정 |

### 단위

```go
document.Pt(72)    // 포인트 (1/72 인치)
document.Mm(10)    // 밀리미터
document.Cm(2.5)   // 센티미터
document.In(1)     // 인치
document.Em(1.5)   // 폰트 크기 기준 상대값
document.Pct(50)   // 퍼센트
```

### 색상

```go
pdf.RGB(0.2, 0.4, 0.8)   // RGB (0.0–1.0)
pdf.RGBHex(0xFF5733)      // 16진수 RGB
pdf.Gray(0.5)             // 그레이스케일
pdf.CMYK(0, 0.5, 1, 0)   // CMYK

// 사전 정의 색상
pdf.Black, pdf.White, pdf.Red, pdf.Green, pdf.Blue
pdf.Yellow, pdf.Cyan, pdf.Magenta
```

## 라이선스

MIT
