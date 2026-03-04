# Go Templates

gpdf integrates with Go's `text/template` package for data-driven PDF generation. This lets you use loops, conditionals, and all standard template features to produce dynamic documents.

## Two Approaches

### 1. `FromJSON` -- Inline Templates in JSON

The simplest approach: embed Go template expressions directly in a JSON schema string.

```go
schema := []byte(`{
    "page": {"size": "A4", "margins": "20mm"},
    "body": [
        {"row": {"cols": [
            {"span": 12, "text": "Hello, {{.Name}}!", "style": {"size": 24}}
        ]}}
    ]
}`)

doc, err := template.FromJSON(schema, map[string]string{"Name": "World"})
```

When the second argument (`data`) is non-nil, gpdf processes the JSON as a Go template first, then parses the resulting JSON.

**Built-in template functions:**

| Function | Description | Usage |
|---|---|---|
| `toJSON` | Marshal a Go value to JSON | `{{toJSON .Items}}` |

### 2. `FromTemplate` -- Pre-parsed Go Templates

For more control, parse a Go template first and pass it to `FromTemplate`. This is useful when you need custom template functions or want to reuse parsed templates.

```go
import gotemplate "text/template"

tmplStr := `{
    "page": {"size": "A4", "margins": "20mm"},
    "body": [
        {"row": {"cols": [
            {"span": 12, "text": "{{.Title}}", "style": {"size": 24, "bold": true}}
        ]}},
        {{- range $i, $item := .Items}}
        {{- if $i}},{{end}}
        {"row": {"cols": [
            {"span": 12, "text": "{{$item}}"}
        ]}}
        {{- end}}
    ]
}`

tmpl, err := gotemplate.New("doc").Funcs(template.TemplateFuncMap()).Parse(tmplStr)
if err != nil {
    // handle error
}

data := map[string]any{
    "Title": "Shopping List",
    "Items": []string{"Apples", "Bananas", "Coffee"},
}

doc, err := template.FromTemplate(tmpl, data)
```

## Template Function Map

Use `template.TemplateFuncMap()` to get gpdf's built-in helper functions when parsing templates:

```go
tmpl := gotemplate.Must(
    gotemplate.New("").Funcs(template.TemplateFuncMap()).Parse(tmplStr),
)
```

### `toJSON`

Marshals a Go value to its JSON representation. Essential for embedding structured data like table rows:

```go
tmplStr := `{
    "page": {"size": "A4"},
    "body": [
        {"row": {"cols": [
            {"span": 12, "table": {
                "header": ["Name", "Age"],
                "rows": {{toJSON .Rows}}
            }}
        ]}}
    ]
}`

data := map[string]any{
    "Rows": [][]string{
        {"Alice", "30"},
        {"Bob", "25"},
    },
}
```

## Dynamic Sections with Range

Use `{{range}}` to generate rows dynamically from data:

```go
tmplStr := `{
    "page": {"size": "A4", "margins": "20mm"},
    "metadata": {"title": "{{.Title}}"},
    "body": [
        {"row": {"cols": [
            {"span": 12, "text": "{{.Title}}", "style": {"size": 24, "bold": true}}
        ]}},
        {"row": {"cols": [
            {"span": 12, "spacer": "5mm"}
        ]}}
        {{- range $i, $section := .Sections}},
        {"row": {"cols": [
            {"span": 12, "elements": [
                {"type": "text", "content": "{{$section.Heading}}", "style": {"size": 16, "bold": true}},
                {"type": "spacer", "height": "3mm"},
                {"type": "text", "content": "{{$section.Body}}"},
                {"type": "spacer", "height": "8mm"}
            ]}
        ]}}
        {{- end}}
    ]
}`
```

> **Tip:** Use `{{- if $i}},{{end}}` before each loop iteration to insert commas between JSON array elements (the leading `{{- if $i}}` skips the comma before the first item).

## Conditionals

Use `{{if}}` to conditionally include content:

```go
tmplStr := `{
    "page": {"size": "A4"},
    "body": [
        {"row": {"cols": [
            {"span": 12, "text": "{{.Title}}", "style": {"size": 24, "bold": true}}
        ]}}
        {{- if .ShowSubtitle}},
        {"row": {"cols": [
            {"span": 12, "text": "{{.Subtitle}}", "style": {"color": "gray(0.5)"}}
        ]}}
        {{- end}}
    ]
}`
```

## Complete Example: Dynamic Report

```go
type Section struct {
    Heading string
    Body    string
}

tmplStr := `{
    "page": {"size": "A4", "margins": "20mm"},
    "metadata": {"title": "{{.Title}}"},
    "header": [
        {"row": {"cols": [
            {"span": 6, "text": "{{.Company}}", "style": {"bold": true, "size": 10}},
            {"span": 6, "text": "{{.Title}}", "style": {"align": "right", "size": 10, "color": "gray(0.5)"}}
        ]}}
    ],
    "footer": [
        {"row": {"cols": [
            {"span": 12, "elements": [
                {"type": "pageNumber", "style": {"size": 8, "align": "center", "color": "gray(0.5)"}}
            ]}
        ]}}
    ],
    "body": [
        {"row": {"cols": [
            {"span": 12, "text": "{{.Title}}", "style": {"size": 28, "bold": true, "color": "#1A237E"}}
        ]}},
        {"row": {"cols": [
            {"span": 12, "spacer": "10mm"}
        ]}}
        {{- range .Sections}},
        {"row": {"cols": [
            {"span": 12, "elements": [
                {"type": "text", "content": "{{.Heading}}", "style": {"size": 16, "bold": true}},
                {"type": "spacer", "height": "3mm"},
                {"type": "text", "content": "{{.Body}}"},
                {"type": "spacer", "height": "8mm"}
            ]}
        ]}}
        {{- end}}
    ]
}`

tmpl, _ := gotemplate.New("report").Funcs(template.TemplateFuncMap()).Parse(tmplStr)

doc, err := template.FromTemplate(tmpl, map[string]any{
    "Title":   "Quarterly Report - Q1 2026",
    "Company": "ACME Corp",
    "Sections": []Section{
        {Heading: "Executive Summary", Body: "Revenue increased 25% year-over-year."},
        {Heading: "Product Development", Body: "The gpdf library reached v0.8."},
        {Heading: "Next Steps", Body: "Focus on v1.0 release preparation."},
    },
})

data, _ := doc.Generate()
```

## Options Override

Both `FromJSON` and `FromTemplate` accept optional `Option` parameters that override schema settings:

```go
doc, err := template.FromJSON(schema, data,
    template.WithFont("NotoSansJP", fontData),
    template.WithDefaultFont("NotoSansJP", 12),
    template.WithPageSize(document.Letter),  // override schema's page size
)
```

## Tips

1. **JSON comma handling**: Use `{{- if $i}},{{end}}` pattern in `range` loops to generate valid JSON comma separators.
2. **Escape special characters**: Go templates auto-escape HTML, but since we're generating JSON, be mindful of quotes in data. Use `toJSON` for structured values.
3. **Reuse parsed templates**: Parse templates once and reuse them with different data for better performance.
4. **Template errors**: `FromJSON` returns parse or execution errors with clear context (e.g., `"gpdf: parsing template: ..."` or `"gpdf: executing template: ..."`).

## See Also

- [JSON Schema](03-json-schema.md) -- Full schema reference
- [Builder API](02-builder-api.md) -- Pure Go approach without templates
