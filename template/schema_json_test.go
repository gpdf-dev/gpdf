package template

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
)

// TestSchemaJSON_ValidJSON verifies that the JSON Schema file is valid JSON.
func TestSchemaJSON_ValidJSON(t *testing.T) {
	data, err := os.ReadFile("../schema/gpdf.schema.json")
	if err != nil {
		t.Fatalf("failed to read schema file: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}

	// Verify top-level required fields
	for _, key := range []string{"$schema", "$id", "title", "type", "properties", "$defs"} {
		if _, ok := schema[key]; !ok {
			t.Errorf("missing top-level key: %s", key)
		}
	}
}

// TestSchemaJSON_DefsPresent verifies that all expected $defs are present.
func TestSchemaJSON_DefsPresent(t *testing.T) {
	data, err := os.ReadFile("../schema/gpdf.schema.json")
	if err != nil {
		t.Fatalf("failed to read schema file: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}

	defs, ok := schema["$defs"].(map[string]any)
	if !ok {
		t.Fatal("$defs is not an object")
	}

	expected := []string{
		"dimension", "color",
		"page", "metadata",
		"row", "rowDef", "col", "element",
		"style", "image", "table", "list", "line",
		"qrcode", "barcode",
		"absolute", "pageBody",
	}
	sort.Strings(expected)

	var actual []string
	for k := range defs {
		actual = append(actual, k)
	}
	sort.Strings(actual)

	if len(actual) != len(expected) {
		t.Errorf("$defs count mismatch: got %d, want %d\ngot:  %v\nwant: %v", len(actual), len(expected), actual, expected)
	}

	for _, key := range expected {
		if _, ok := defs[key]; !ok {
			t.Errorf("missing $defs key: %s", key)
		}
	}
}

// TestSchemaJSON_RequiredFields verifies that required fields in key definitions
// match the Go struct expectations.
func TestSchemaJSON_RequiredFields(t *testing.T) {
	data, err := os.ReadFile("../schema/gpdf.schema.json")
	if err != nil {
		t.Fatalf("failed to read schema file: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}

	defs := schema["$defs"].(map[string]any)

	tests := []struct {
		def      string
		required []string
	}{
		{"page", []string{"size"}},
		{"row", []string{"row"}},
		{"rowDef", []string{"cols"}},
		{"col", []string{"span"}},
		{"element", []string{"type"}},
		{"image", []string{"src"}},
		{"table", []string{"header", "rows"}},
		{"list", []string{"items"}},
		{"qrcode", []string{"data"}},
		{"barcode", []string{"data"}},
		{"absolute", []string{"x", "y", "elements"}},
		{"pageBody", []string{"body"}},
	}

	for _, tt := range tests {
		t.Run(tt.def, func(t *testing.T) {
			def, ok := defs[tt.def].(map[string]any)
			if !ok {
				t.Fatalf("$defs/%s is not an object", tt.def)
			}

			reqRaw, ok := def["required"].([]any)
			if !ok {
				t.Fatalf("$defs/%s has no required array", tt.def)
			}

			var req []string
			for _, v := range reqRaw {
				req = append(req, v.(string))
			}
			sort.Strings(req)

			want := make([]string, len(tt.required))
			copy(want, tt.required)
			sort.Strings(want)

			if len(req) != len(want) {
				t.Errorf("required mismatch: got %v, want %v", req, want)
				return
			}
			for i := range req {
				if req[i] != want[i] {
					t.Errorf("required mismatch at %d: got %s, want %s", i, req[i], want[i])
				}
			}
		})
	}
}

// TestSchemaJSON_TopLevelRequired verifies that "page" is the only required
// top-level property (matching Schema struct: Page is non-pointer).
func TestSchemaJSON_TopLevelRequired(t *testing.T) {
	data, err := os.ReadFile("../schema/gpdf.schema.json")
	if err != nil {
		t.Fatalf("failed to read schema file: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}

	reqRaw, ok := schema["required"].([]any)
	if !ok {
		t.Fatal("top-level required is not an array")
	}

	if len(reqRaw) != 1 || reqRaw[0].(string) != "page" {
		t.Errorf("top-level required should be [\"page\"], got %v", reqRaw)
	}
}

// TestSchemaJSON_TopLevelProperties verifies top-level property keys match the
// Schema struct fields.
func TestSchemaJSON_TopLevelProperties(t *testing.T) {
	data, err := os.ReadFile("../schema/gpdf.schema.json")
	if err != nil {
		t.Fatalf("failed to read schema file: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties is not an object")
	}

	expected := []string{"page", "metadata", "header", "footer", "body", "pages", "absolute"}
	sort.Strings(expected)

	var actual []string
	for k := range props {
		actual = append(actual, k)
	}
	sort.Strings(actual)

	if len(actual) != len(expected) {
		t.Errorf("properties count mismatch: got %v, want %v", actual, expected)
	}
	for _, key := range expected {
		if _, ok := props[key]; !ok {
			t.Errorf("missing property: %s", key)
		}
	}
}

// TestSchemaJSON_ValidExamples verifies that known-good JSON documents can be
// unmarshaled into the Schema struct (structural conformance with Go types).
func TestSchemaJSON_ValidExamples(t *testing.T) {
	examples := []struct {
		name string
		json string
	}{
		{
			"minimal",
			`{"page": {"size": "A4"}}`,
		},
		{
			"hello_world",
			`{
				"page": {"size": "A4", "margins": "20mm"},
				"body": [
					{"row": {"cols": [
						{"span": 12, "text": "Hello, World!", "style": {"size": 24, "bold": true}}
					]}}
				]
			}`,
		},
		{
			"with_metadata",
			`{
				"page": {"size": "Letter"},
				"metadata": {"title": "Test", "author": "gpdf"},
				"body": [{"row": {"cols": [{"span": 6, "text": "Left"}, {"span": 6, "text": "Right"}]}}]
			}`,
		},
		{
			"elements_array",
			`{
				"page": {"size": "A4"},
				"body": [{"row": {"cols": [{
					"span": 12,
					"elements": [
						{"type": "text", "content": "Title", "style": {"size": 18, "bold": true}},
						{"type": "line", "line": {"color": "gray(0.5)"}},
						{"type": "spacer", "height": "5mm"},
						{"type": "text", "content": "Body text"}
					]
				}]}}]
			}`,
		},
		{
			"table",
			`{
				"page": {"size": "A4"},
				"body": [{"row": {"cols": [{
					"span": 12,
					"table": {
						"header": ["Name", "Price"],
						"rows": [["Item A", "$10"], ["Item B", "$20"]],
						"headerStyle": {"bold": true, "background": "#1A237E", "color": "white"},
						"stripeColor": "#F5F5F5"
					}
				}]}}]
			}`,
		},
		{
			"absolute",
			`{
				"page": {"size": "A4"},
				"body": [{"row": {"cols": [{"span": 12, "text": "Content"}]}}],
				"absolute": [{
					"x": "150mm", "y": "20mm",
					"elements": [{"type": "qrcode", "qrcode": {"data": "https://gpdf.dev", "size": "20mm"}}]
				}]
			}`,
		},
		{
			"multi_page",
			`{
				"page": {"size": "A4"},
				"header": [{"row": {"cols": [{"span": 12, "text": "Header"}]}}],
				"footer": [{"row": {"cols": [{"span": 12, "elements": [{"type": "pageNumber", "style": {"align": "center"}}]}]}}],
				"pages": [
					{"body": [{"row": {"cols": [{"span": 12, "text": "Page 1"}]}}]},
					{"body": [{"row": {"cols": [{"span": 12, "text": "Page 2"}]}}]}
				]
			}`,
		},
	}

	for _, tt := range examples {
		t.Run(tt.name, func(t *testing.T) {
			var s Schema
			if err := json.Unmarshal([]byte(tt.json), &s); err != nil {
				t.Errorf("failed to unmarshal valid example: %v", err)
			}
			if s.Page.Size == "" {
				t.Error("page.size should not be empty")
			}
		})
	}
}
