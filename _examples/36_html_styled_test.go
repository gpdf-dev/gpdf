package examples_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/htmlpdf"
)

func TestExample_36_HTML_Styled(t *testing.T) {
	html := `
	<html>
	<head>
		<style>
		body { font-size: 12pt; }
		h1 {
			color: #1A237E;
			border-bottom: 2px solid #1A237E;
			padding-bottom: 5pt;
		}
		.section {
			background-color: #F5F5F5;
			padding: 10pt;
			margin-bottom: 10pt;
		}
		.highlight { color: #D32F2F; font-weight: bold; }
		.muted { color: #757575; font-style: italic; }
		</style>
	</head>
	<body>
		<h1>Styled Document</h1>
		<div class="section">
			<p>This section has a <strong>gray background</strong> and padding.</p>
			<p class="highlight">This paragraph is highlighted in red and bold.</p>
		</div>
		<div class="section">
			<h2>Inline Styling</h2>
			<p>Mix <em>italic</em>, <strong>bold</strong>, and <u>underlined</u> text freely.</p>
			<p style="text-align: center; font-size: 14pt;">This paragraph is centered with a larger font.</p>
		</div>
		<p class="muted">A muted footer paragraph with italic gray text.</p>
	</body>
	</html>`

	result, err := htmlpdf.FromHTML(html)
	if err != nil {
		t.Fatalf("FromHTML failed: %v", err)
	}

	data, err := result.Bytes()
	if err != nil {
		t.Fatalf("Bytes failed: %v", err)
	}

	assertValidPDF(t, data)
	writePDF(t, "36_html_styled.pdf", data)
	assertMatchesGolden(t, "36_html_styled.pdf", data)
}
