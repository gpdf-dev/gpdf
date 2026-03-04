package template

import (
	"encoding/json"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
)

func TestAbsolute_BuilderAPI(t *testing.T) {
	doc := New(
		WithPageSize(document.A4),
		WithMargins(document.UniformEdges(document.Mm(15))),
	)

	page := doc.AddPage()

	// Normal flow content.
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Text("Flow content")
		})
	})

	// Absolute positioned content.
	page.Absolute(document.Mm(100), document.Mm(200), func(c *ColBuilder) {
		c.Text("Absolute content", FontSize(16))
	}, AbsoluteWidth(document.Mm(50)))

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PDF output")
	}

	// Verify PDF header.
	if string(data[:5]) != "%PDF-" {
		t.Error("output does not start with PDF header")
	}
}

func TestAbsolute_OriginPage(t *testing.T) {
	doc := New(
		WithPageSize(document.A4),
		WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()
	page.Absolute(document.Mm(10), document.Mm(10), func(c *ColBuilder) {
		c.Text("Page origin")
	}, AbsoluteOriginPage())

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PDF output")
	}
}

func TestAbsolute_MultipleElements(t *testing.T) {
	doc := New(WithPageSize(document.A4))

	page := doc.AddPage()

	// Multiple absolute elements on the same page.
	page.Absolute(document.Mm(10), document.Mm(10), func(c *ColBuilder) {
		c.Text("Element 1")
	})
	page.Absolute(document.Mm(100), document.Mm(100), func(c *ColBuilder) {
		c.Text("Element 2")
	})
	page.Absolute(document.Mm(50), document.Mm(250), func(c *ColBuilder) {
		c.Text("Element 3")
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PDF output")
	}
}

func TestAbsolute_JSONSchema(t *testing.T) {
	schema := `{
		"page": {"size": "A4", "margins": "15mm"},
		"body": [
			{"row": {"cols": [{"span": 12, "text": "Flow text"}]}}
		],
		"absolute": [
			{
				"x": "100mm",
				"y": "200mm",
				"width": "50mm",
				"elements": [
					{"type": "text", "content": "Absolute text", "style": {"size": 16}}
				]
			}
		]
	}`

	doc, err := FromJSON([]byte(schema), nil)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PDF output")
	}
}

func TestAbsolute_JSONSchema_OriginPage(t *testing.T) {
	schema := `{
		"page": {"size": "A4"},
		"body": [],
		"absolute": [
			{
				"x": "10mm",
				"y": "10mm",
				"origin": "page",
				"elements": [
					{"type": "text", "content": "Page corner"}
				]
			}
		]
	}`

	doc, err := FromJSON([]byte(schema), nil)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PDF output")
	}
}

func TestAbsolute_JSONSchema_MultiplePages(t *testing.T) {
	schema := `{
		"page": {"size": "A4"},
		"pages": [
			{
				"body": [{"row": {"cols": [{"span": 12, "text": "Page 1"}]}}],
				"absolute": [
					{"x": "50mm", "y": "50mm", "elements": [{"type": "text", "content": "Abs on P1"}]}
				]
			},
			{
				"body": [{"row": {"cols": [{"span": 12, "text": "Page 2"}]}}],
				"absolute": [
					{"x": "80mm", "y": "80mm", "elements": [{"type": "text", "content": "Abs on P2"}]}
				]
			}
		]
	}`

	doc, err := FromJSON([]byte(schema), nil)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PDF output")
	}
}

func TestSchemaAbsolute_JSONParsing(t *testing.T) {
	input := `{
		"page": {"size": "A4"},
		"body": [],
		"absolute": [
			{
				"x": "100mm",
				"y": "200mm",
				"width": "50mm",
				"height": "30mm",
				"origin": "page",
				"elements": [
					{"type": "text", "content": "test"}
				]
			}
		]
	}`

	var s Schema
	if err := json.Unmarshal([]byte(input), &s); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if len(s.Absolute) != 1 {
		t.Fatalf("expected 1 absolute entry, got %d", len(s.Absolute))
	}

	abs := s.Absolute[0]
	if abs.X != "100mm" {
		t.Errorf("X: got %q, want %q", abs.X, "100mm")
	}
	if abs.Y != "200mm" {
		t.Errorf("Y: got %q, want %q", abs.Y, "200mm")
	}
	if abs.Width != "50mm" {
		t.Errorf("Width: got %q, want %q", abs.Width, "50mm")
	}
	if abs.Height != "30mm" {
		t.Errorf("Height: got %q, want %q", abs.Height, "30mm")
	}
	if abs.Origin != "page" {
		t.Errorf("Origin: got %q, want %q", abs.Origin, "page")
	}
	if len(abs.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(abs.Elements))
	}
}
